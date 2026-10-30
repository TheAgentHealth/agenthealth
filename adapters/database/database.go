// Package database implements fixed, read-only native database health probes.
package database

import (
	"bufio"
	"context"
	"crypto/tls"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgproto3"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "database", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"database"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}

type endpoint struct {
	kind, address, host, database string
	options                       core.DatabaseOptions
}

func parse(t core.Target) (endpoint, error) {
	invalid := &core.Failure{Status: core.Misconfigured}
	u, err := url.Parse(t.Endpoint)
	if err != nil || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.Opaque != "" || strings.Contains(u.Host, ",") {
		return endpoint{}, invalid
	}
	ports := map[string]string{"postgres": "5432", "mysql": "3306", "redis": "6379", "mongodb": "27017"}
	port, ok := ports[u.Scheme]
	if !ok {
		return endpoint{}, invalid
	}
	if strings.Contains(u.Host, ":") && strings.HasSuffix(u.Host, ":") {
		return endpoint{}, invalid
	}
	if u.Port() != "" {
		port = u.Port()
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return endpoint{}, invalid
	}
	o := core.DatabaseOptions{}
	if t.Database != nil {
		o = *t.Database
	}
	if t.Auth != nil || (o.TLS != "" && o.TLS != "required" && o.TLS != "disabled") {
		return endpoint{}, invalid
	}
	db := strings.TrimPrefix(u.Path, "/")
	if strings.ContainsAny(db, "/\x00") || strings.ContainsRune(o.Username, 0) {
		return endpoint{}, invalid
	}
	if (u.Scheme == "postgres" || u.Scheme == "mysql") && (o.Username == "" || db == "") {
		return endpoint{}, invalid
	}
	if u.Scheme == "redis" && db != "" {
		i, e := strconv.Atoi(db)
		if e != nil || i < 0 || i > 15 {
			return endpoint{}, invalid
		}
	}
	return endpoint{u.Scheme, net.JoinHostPort(u.Hostname(), port), u.Hostname(), db, o}, nil
}
func (e endpoint) tlsConfig() *tls.Config {
	if e.options.TLS == "disabled" {
		return nil
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, ServerName: e.host}
}
func observation(status core.Status) core.Observation {
	return core.Observation{Check: core.CheckResult{Status: status}, ResponseReceived: true}
}

func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	e, err := parse(r.Target)
	if err != nil {
		return core.Observation{}, err
	}
	if dimension == "configuration" {
		return observation(core.Healthy), nil
	}
	if dimension == "reachability" {
		conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", e.address)
		if err != nil {
			return core.Observation{}, err
		}
		conn.Close()
		if r.RecordStep != nil {
			r.RecordStep("tcp", core.Healthy)
		}
		return observation(core.Healthy), nil
	}
	if dimension != "authentication" && dimension != "protocol" && dimension != "capability" && dimension != "functional" {
		return observation(core.Misconfigured), nil
	}
	// A probe may have reached the server even if a driver hides the first byte.
	// Conservatively prevent retries once native protocol work begins.
	if r.MarkResponse != nil {
		r.MarkResponse()
	}
	if cached, ok := r.RunState.Load("database-probe"); ok && dimension != "functional" {
		return cached.(core.Observation), nil
	}
	err = probe(ctx, e, r.Environment[e.options.PasswordEnv], dimension == "functional")
	result := observation(core.Healthy)
	if err != nil {
		if ctx.Err() != nil {
			return core.Observation{ResponseReceived: true}, ctx.Err()
		}
		result = observation(core.Unhealthy)
		var pg *pgconn.PgError
		var my *mysql.MySQLError
		var command mongo.CommandError
		if (errors.As(err, &pg) && strings.HasPrefix(pg.Code, "28")) || (errors.As(err, &my) && (my.Number == 1045 || my.Number == 1698)) || (errors.As(err, &command) && (command.Code == 18 || command.Code == 13)) || errors.Is(err, errRedisAuth) {
			result = observation(core.Misconfigured)
		}
	}
	if dimension != "functional" {
		r.RunState.Store("database-probe", result)
	}
	return result, nil
}

// Each native operation owns and closes its connection. Queries are constants;
// callers cannot provide SQL, commands, driver options or insecure TLS modes.
func probe(ctx context.Context, e endpoint, password string, functional bool) error {
	switch e.kind {
	case "postgres":
		// Override environment-derived settings; never permit TLS fallback.
		u := url.URL{Scheme: "postgres", Host: e.address, Path: "/" + e.database, User: url.UserPassword(e.options.Username, "unused-placeholder"), RawQuery: "sslmode=disable&target_session_attrs=any&connect_timeout=0"}
		cfg, err := pgx.ParseConfig(u.String())
		if err != nil {
			return err
		}
		cfg.User = e.options.Username
		cfg.Password = password
		cfg.Database = e.database
		cfg.TLSConfig = e.tlsConfig()
		cfg.Fallbacks = nil
		cfg.RuntimeParams = map[string]string{}
		cfg.ConnectTimeout = 0
		cfg.ValidateConnect = nil
		cfg.BuildFrontend = func(r io.Reader, w io.Writer) *pgproto3.Frontend {
			f := pgproto3.NewFrontend(r, w)
			f.SetMaxBodyLen(4 << 20)
			return f
		}
		conn, err := pgx.ConnectConfig(ctx, cfg)
		if err != nil {
			return err
		}
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = conn.Close(cleanup)
		}()
		if !functional {
			return nil
		}
		var value int
		err = conn.QueryRow(ctx, "SELECT 1").Scan(&value)
		if err == nil && value != 1 {
			return errors.New("invalid constant result")
		}
		return err
	case "mysql":
		cfg := mysql.NewConfig()
		cfg.Net = "tcp"
		cfg.Addr = e.address
		cfg.User = e.options.Username
		cfg.Passwd = password
		cfg.DBName = e.database
		cfg.TLS = e.tlsConfig()
		cfg.AllowNativePasswords = true
		cfg.Logger = discardLog{}
		connector, err := mysql.NewConnector(cfg)
		if err != nil {
			return err
		}
		conn, err := connector.Connect(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()
		if !functional {
			return nil
		}
		rows, err := conn.(driver.QueryerContext).QueryContext(ctx, "SELECT 1", nil)
		if err != nil {
			return err
		}
		defer rows.Close()
		values := make([]driver.Value, 1)
		if err = rows.Next(values); err != nil {
			return err
		}
		if stringValue := values[0]; stringValue != int64(1) && stringValue != "1" {
			if b, ok := stringValue.([]byte); !ok || string(b) != "1" {
				return errors.New("invalid constant result")
			}
		}
		return nil
	case "redis":
		return redisProbe(ctx, e, password)

	case "mongodb":
		opts := options.Client().SetLoggerOptions(options.Logger().SetSink(discardLog{})).SetHosts([]string{e.address}).SetDirect(true).SetTLSConfig(e.tlsConfig()).SetRetryReads(false).SetRetryWrites(false).SetMaxPoolSize(1)
		if e.options.Username != "" {
			source := e.database
			if source == "" {
				source = "admin"
			}
			opts.SetAuth(options.Credential{Username: e.options.Username, Password: password, AuthSource: source})
		}
		client, err := mongo.Connect(opts)
		if err != nil {
			return err
		}
		defer func() {
			cleanup, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = client.Disconnect(cleanup)
		}()
		db := e.database
		if db == "" {
			db = "admin"
		}
		return client.Database(db).RunCommand(ctx, bson.D{{Key: "ping", Value: 1}}).Err()
	}
	return errors.New("unsupported database")
}

// The health-only RESP2 client sends just AUTH, SELECT and PING. It has no
// background pool, retry machinery, topology discovery or driver logging.
var errRedisAuth = errors.New("redis authentication denied")

func redisProbe(ctx context.Context, e endpoint, password string) error {
	dialer := &net.Dialer{}
	var conn net.Conn
	var err error
	if cfg := e.tlsConfig(); cfg != nil {
		conn, err = (&tls.Dialer{NetDialer: dialer, Config: cfg}).DialContext(ctx, "tcp", e.address)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", e.address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		if err = conn.SetDeadline(deadline); err != nil {
			return err
		}
	}
	reader := bufio.NewReaderSize(conn, 1024)
	command := func(expected string, args ...string) error {
		var request strings.Builder
		fmt.Fprintf(&request, "*%d\r\n", len(args))
		for _, arg := range args {
			fmt.Fprintf(&request, "$%d\r\n%s\r\n", len(arg), arg)
		}
		if _, err := io.WriteString(conn, request.String()); err != nil {
			return err
		}
		reply, err := reader.ReadSlice('\n')
		if err != nil {
			return err
		}
		text := string(reply)
		if strings.HasPrefix(text, "-WRONGPASS ") || strings.HasPrefix(text, "-NOAUTH ") || strings.HasPrefix(text, "-NOPERM ") {
			return errRedisAuth
		}
		if text != expected {
			return errors.New("unexpected RESP reply")
		}
		return nil
	}
	if password != "" || e.options.Username != "" {
		args := []string{"AUTH"}
		if e.options.Username != "" {
			args = append(args, e.options.Username)
		}
		args = append(args, password)
		if err = command("+OK\r\n", args...); err != nil {
			return err
		}
	}
	if e.database != "" && e.database != "0" {
		if err = command("+OK\r\n", "SELECT", e.database); err != nil {
			return err
		}
	}
	return command("+PONG\r\n", "PING")
}

// Driver diagnostics may include server-controlled text. Preserve engine-owned
// diagnostics even when a process environment enables native debug logging.
type discardLog struct{}

func (discardLog) Print(...any)                {}
func (discardLog) Info(int, string, ...any)    {}
func (discardLog) Error(error, string, ...any) {}

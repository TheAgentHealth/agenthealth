package database

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestEndpointSafety(t *testing.T) {
	for _, endpoint := range []string{"http://localhost", "postgres://user:secret@localhost/db", "mysql://localhost/db?tls=skip-verify", "redis://localhost/16", "mongodb://a,b/db", "redis://localhost/#x", "postgres://localhost/", "redis://localhost:0", "redis://localhost:99999", "redis://localhost:/", "redis://localhost/?", "redis://localhost/a%2fb"} {
		if _, err := parse(core.Target{Endpoint: endpoint}); err == nil {
			t.Errorf("accepted %q", endpoint)
		}
	}
	for _, kind := range []string{"postgres", "mysql", "redis", "mongodb"} {
		e, err := parse(core.Target{Endpoint: kind + "://localhost/1", Database: &core.DatabaseOptions{Username: "health"}})
		if err != nil || e.tlsConfig() == nil || e.tlsConfig().InsecureSkipVerify {
			t.Fatalf("unsafe default %s: %v", kind, err)
		}
	}
}
func TestProbeCacheAndUnknownDimension(t *testing.T) {
	r := core.Request{Target: core.Target{Endpoint: "redis://localhost"}, RunState: &sync.Map{}}
	r.RunState.Store("database-probe", observation(core.Degraded))
	got, err := (Adapter{}).Check(context.Background(), r, "capability")
	if err != nil || got.Check.Status != core.Degraded {
		t.Fatal(got, err)
	}
	got, err = (Adapter{}).Check(context.Background(), r, "invalid")
	if err != nil || got.Check.Status != core.Misconfigured {
		t.Fatal(got, err)
	}
}
func TestReachabilityAndCancellation(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	r := core.Request{Target: core.Target{Endpoint: "redis://" + listener.Addr().String()}, RunState: &sync.Map{}}
	got, err := (Adapter{}).Check(context.Background(), r, "reachability")
	if err != nil || got.Check.Status != core.Healthy {
		t.Fatal(got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = (Adapter{}).Check(ctx, r, "reachability"); err == nil {
		t.Fatal("cancellation ignored")
	}
}
func TestMissingPasswordNeverConnects(t *testing.T) {
	t.Setenv("AGENTHEALTH_DB_MISSING", "")
	registry := core.NewRegistry()
	if err := registry.Register(Adapter{}); err != nil {
		t.Fatal(err)
	}
	results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "db", Type: "database", Endpoint: "redis://127.0.0.1:1", Database: &core.DatabaseOptions{PasswordEnv: "AGENTHEALTH_DB_MISSING"}}}})
	if err != nil || results[0].Status != core.Misconfigured {
		t.Fatal(results, err)
	}
}
func TestRealDatabases(t *testing.T) {
	if os.Getenv("AGENTHEALTH_DATABASE_TESTS") != "1" {
		t.Skip("set AGENTHEALTH_DATABASE_TESTS=1 with the documented disposable services")
	}
	for _, tt := range []struct{ kind, port, user, db string }{{"postgres", "15432", "health", "health"}, {"mysql", "13306", "health", "health"}, {"redis", "16379", "", "0"}, {"mongodb", "17017", "health", "admin"}} {
		t.Run(tt.kind, func(t *testing.T) {
			target := core.Target{Name: "db", Type: "database", Endpoint: fmt.Sprintf("%s://127.0.0.1:%s/%s", tt.kind, tt.port, tt.db), Database: &core.DatabaseOptions{Username: tt.user, PasswordEnv: "AGENTHEALTH_DB_PASSWORD", TLS: "disabled"}, Checks: []string{"reachability", "authentication", "protocol", "capability", "functional", "latency"}}
			run := func(password string) core.Result {
				t.Setenv("AGENTHEALTH_DB_PASSWORD", password)
				reg := core.NewRegistry()
				if err := reg.Register(Adapter{}); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				results, err := core.NewEngine(reg).Run(ctx, core.Config{Version: "v1", Targets: []core.Target{target}})
				if err != nil {
					t.Fatal(err)
				}
				var out bytes.Buffer
				if err := core.WriteJSON(&out, results); err != nil {
					t.Fatal(err)
				}
				if strings.Contains(out.String(), password) {
					t.Fatal("password leaked")
				}
				return results[0]
			}
			if result := run("health-test-password"); result.Status != core.Healthy {
				t.Fatalf("healthy probe: %+v", result)
			}
			if result := run("wrong-secret-password"); result.Status != core.Misconfigured {
				t.Fatalf("authentication failure: %+v", result)
			}
		})
	}
}

func TestRedisBoundedResponsesAndCancellation(t *testing.T) {
	for _, reply := range []string{"+PONG\r\n", "+WRONG\r\n", "-NOAUTH secret-server-message\r\n", "-ERR invalid password\r\n", strings.Repeat("x", 2048) + "\r\n", ""} {
		t.Run(fmt.Sprintf("response-%d", len(reply)), func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			done := make(chan struct{})
			go func() {
				defer close(done)
				conn, err := listener.Accept()
				if err != nil {
					return
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(time.Second))
				b := make([]byte, 64)
				_, _ = conn.Read(b)
				if reply != "" {
					_, _ = conn.Write([]byte(reply))
				} else {
					_, _ = conn.Read(b)
				}
			}()
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()
			err = redisProbe(ctx, endpoint{address: listener.Addr().String(), options: core.DatabaseOptions{TLS: "disabled"}}, "")
			if (reply == "+PONG\r\n") != (err == nil) {
				t.Fatalf("unexpected success/error: %v", err)
			}
			<-done
		})
	}
}

func TestUntrustedTLSIsRejected(t *testing.T) {
	server := httptest.NewTLSServer(nil)
	defer server.Close()
	e := endpoint{kind: "redis", address: server.Listener.Addr().String(), host: "localhost"}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := redisProbe(ctx, e, ""); err == nil {
		t.Fatal("untrusted server accepted")
	}
}

func TestNativeConnectionFailureKeepsEngineClassification(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	listener.Close()
	marked := false
	r := core.Request{Target: core.Target{Endpoint: "redis://" + address, Database: &core.DatabaseOptions{TLS: "disabled"}}, RunState: &sync.Map{}, MarkResponse: func() { marked = true }}
	got, err := (Adapter{}).Check(context.Background(), r, "authentication")
	if err == nil || marked || got.ResponseReceived || core.NormalizeError(err, got.ResponseReceived).Status != core.Unreachable {
		t.Fatalf("connection error lost classification: %+v %v marked=%v", got, err, marked)
	}
	if _, ok := r.RunState.Load("database-probe"); ok {
		t.Fatal("no-response failure cached")
	}
}

func TestObservedBytesControlRetryEvidence(t *testing.T) {
	left, right := net.Pipe()
	defer left.Close()
	defer right.Close()
	marked := false
	conn := observingConn{Conn: left, mark: func() { marked = true }}
	go func() { _, _ = right.Write([]byte("x")) }()
	b := make([]byte, 1)
	if _, err := conn.Read(b); err != nil || !marked {
		t.Fatal("first response byte not marked", err)
	}
}

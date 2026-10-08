package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func runServe(ctx context.Context, args []string, out, diagnostics io.Writer) int {
	fail := func(message string) int { fmt.Fprintln(diagnostics, message); return 6 }
	if len(args) == 0 {
		return fail("usage: agenthealth serve <configuration.yaml> [--listen host:port] [--token-env NAME]")
	}
	if args[0] == "--help" || args[0] == "-h" {
		_, err := io.WriteString(out, usage)
		if err != nil {
			return fail("cannot write help")
		}
		return 0
	}
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	listen := flags.String("listen", "127.0.0.1:8080", "listen address")
	tokenEnv := flags.String("token-env", "", "bearer token environment reference")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
		return fail("invalid serve options; use --help")
	}
	config, err := core.LoadConfigFile(args[0])
	if err != nil {
		return fail("cannot load serving configuration")
	}
	token := ""
	if *tokenEnv != "" {
		token = os.Getenv(*tokenEnv)
		if token == "" {
			return fail("serving token environment reference is empty")
		}
	}
	registry, err := newRegistry()
	if err != nil {
		return fail("cannot register adapters")
	}
	handler, err := core.NewAHPServer(core.NewEngine(registry), config, core.AHPOptions{Token: token})
	if err != nil {
		return fail("invalid serving configuration")
	}
	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		return fail("cannot bind AHP listener")
	}
	defer listener.Close()
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	go func() { defer close(done); handler.Run(runCtx) }()
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		<-runCtx.Done()
		shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		_ = server.Shutdown(shutdownCtx)
	}()
	err = server.Serve(listener)
	cancel()
	<-stopped
	<-done
	if err != nil && err != http.ErrServerClosed {
		return fail("AHP server failed")
	}
	return 0
}

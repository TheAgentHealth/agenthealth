package main

import (
	"bytes"
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestServeInvocation(t *testing.T) {
	for _, args := range [][]string{{"serve"}, {"serve", "missing.yaml"}, {"serve", "missing.yaml", "--unknown"}, {"serve", "missing.yaml", "--format", "json"}} {
		var out, diagnostics bytes.Buffer
		if code := run(context.Background(), args, &out, &diagnostics); code != 6 || diagnostics.Len() == 0 {
			t.Fatalf("%v: %d %s", args, code, diagnostics.String())
		}
	}
	var out, diagnostics bytes.Buffer
	if code := run(context.Background(), []string{"serve", "--help"}, &out, &diagnostics); code != 0 || out.Len() == 0 {
		t.Fatal(code)
	}
}

func TestServeSuccessfulLifecycle(t *testing.T) {
	configuration := filepath.Join(t.TempDir(), "health.yaml")
	if err := os.WriteFile(configuration, []byte("version: v1\ntargets:\n  - name: service\n    type: http\n    endpoint: http://127.0.0.1:9000\n    checks: []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AHP_TEST_TOKEN", "1234567890123456")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	var out, diagnostics bytes.Buffer
	done := make(chan int, 1)
	go func() {
		done <- run(ctx, []string{"serve", configuration, "--listen", "127.0.0.1:0", "--token-env", "AHP_TEST_TOKEN"}, &out, &diagnostics)
	}()
	select {
	case code := <-done:
		if code != 0 || diagnostics.Len() != 0 {
			t.Fatalf("serve exited %d: %s", code, diagnostics.String())
		}
		if ctx.Err() == nil {
			t.Fatal("serve returned before cancellation")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("serve failed to shut down")
	}
}

func TestServeConfigurationAndBindingFailures(t *testing.T) {
	configuration := filepath.Join(t.TempDir(), "health.yaml")
	if err := os.WriteFile(configuration, []byte("version: v1\ntargets:\n  - name: service\n    type: http\n    endpoint: http://127.0.0.1:9000\n    checks: []\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AHP_EMPTY_TEST_TOKEN", "")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	for _, options := range [][]string{{"--token-env", "AHP_EMPTY_TEST_TOKEN"}, {"--listen", "invalid-address"}, {"--listen", listener.Addr().String()}} {
		var out, diagnostics bytes.Buffer
		if code := run(context.Background(), append([]string{"serve", configuration}, options...), &out, &diagnostics); code != 6 {
			t.Fatalf("expected failure, got %d", code)
		}
	}
}

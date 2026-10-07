package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func execute(t *testing.T, target core.Target) core.Result {
	t.Helper()
	registry := core.NewRegistry()
	if err := registry.Register(Adapter{}); err != nil {
		t.Fatal(err)
	}
	results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	return results[0]
}
func TestHTTPDefaultsAuthenticationAndFunctional(t *testing.T) {
	var gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			gets.Add(1)
		}
		if r.Header.Get("Authorization") != "Bearer super-secret" {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	t.Setenv("AGENTHEALTH_HTTP_TOKEN", "super-secret")
	target := core.Target{Name: "http", Type: "http", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "AGENTHEALTH_HTTP_TOKEN"}}
	result := execute(t, target)
	if result.Status != core.Healthy || result.LatencyMS == nil || gets.Load() != 0 {
		t.Fatal(result, gets.Load())
	}
	target.Checks = []string{"functional"}
	result = execute(t, target)
	if result.Status != core.Healthy || len(result.Checks) != 1 || gets.Load() != 1 {
		t.Fatal(result, gets.Load())
	}
	target.Auth = nil
	result = execute(t, target)
	if result.Status != core.Misconfigured || gets.Load() != 1 {
		t.Fatal("auth failure did not block functional", result)
	}
}
func TestTLSVerificationAndRedirectSafety(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }))
	defer server.Close()
	result := execute(t, core.Target{Name: "tls", Type: "http", Endpoint: server.URL})
	if result.Status != core.Unreachable {
		t.Fatal("untrusted certificate accepted", result)
	}
	var forwarded atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { forwarded.Add(1) }))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, destination.URL, 302) }))
	defer redirect.Close()
	result = execute(t, core.Target{Name: "redirect", Type: "http", Endpoint: redirect.URL, Checks: []string{"reachability"}})
	if forwarded.Load() != 0 || result.Status != core.Healthy {
		t.Fatal("followed redirect", result)
	}
}
func TestHTTPTimeoutErrorsAndCredentialURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	timeout := 20
	result := execute(t, core.Target{Name: "slow", Type: "http", Endpoint: server.URL, TimeoutMS: &timeout})
	if result.Status != core.Unreachable || result.LatencyMS != nil {
		t.Fatal(result)
	}
	server.Close()
	result = execute(t, core.Target{Name: "offline", Type: "http", Endpoint: server.URL})
	if result.Status != core.Unreachable {
		t.Fatal(result)
	}
	result = execute(t, core.Target{Name: "url", Type: "http", Endpoint: "https://user:password@example.com/?token=secret"})
	if result.Status != core.Misconfigured {
		t.Fatal(result)
	}
	var b bytes.Buffer
	if err := core.WriteJSON(&b, []core.Result{result}); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"password", "token=secret"} {
		if strings.Contains(b.String(), secret) {
			t.Fatal("URL credential leaked")
		}
	}
}
func TestFunctionalHTTPFailureAndLatency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { time.Sleep(time.Millisecond); w.WriteHeader(500) }))
	defer server.Close()
	result := execute(t, core.Target{Name: "broken", Type: "http", Endpoint: server.URL, Checks: []string{"functional"}})
	if result.Status != core.Unhealthy {
		t.Fatal(result)
	}
}

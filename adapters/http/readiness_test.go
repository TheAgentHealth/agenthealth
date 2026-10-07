package http

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestPassiveReadinessStatus(t *testing.T) {
	for _, code := range []int{200, 204, 302, 404, 500, 401, 403} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "HEAD" {
					t.Errorf("passive check issued %s", r.Method)
				}
				w.WriteHeader(code)
			}))
			defer server.Close()
			target := core.Target{Name: "status", Type: "http", Endpoint: server.URL}
			result := execute(t, target)
			want := core.Unhealthy
			if code < 300 {
				want = core.Healthy
			}
			if code == 401 || code == 403 {
				want = core.Misconfigured
			}
			if result.Status != want || result.Checks["reachability"].Status != core.Healthy {
				t.Fatal(result)
			}
			target.Checks = []string{"reachability"}
			if result := execute(t, target); result.Status != core.Healthy {
				t.Fatal("explicit connectivity should remain independent of status", result)
			}
			target.Checks = nil
			target.HTTP = &core.HTTPOptions{ExpectedStatus: []int{code}}
			result = execute(t, target)
			if code != 401 && code != 403 && result.Status != core.Healthy {
				t.Fatal("custom status rejected", result)
			}
		})
	}
}

func TestHeadersAndBody(t *testing.T) {
	var gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			gets.Add(1)
		}
		w.Header().Set("X-Ready", "yes")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ready sensitive-response-text"))
	}))
	defer server.Close()
	target := core.Target{Name: "response", Type: "http", Endpoint: server.URL, HTTP: &core.HTTPOptions{Headers: map[string]string{"x-ready": "yes"}}}
	if result := execute(t, target); result.Status != core.Healthy || gets.Load() != 0 {
		t.Fatal(result)
	}
	target.HTTP.Headers["x-ready"] = "no"
	if result := execute(t, target); result.Status != core.Unhealthy || result.Checks["protocol"].Message != "required HTTP response header did not match" {
		t.Fatal(result)
	}
	target.HTTP.Headers = nil
	target.Checks = []string{"functional"}
	target.HTTP.BodyContains = "ready"
	if result := execute(t, target); result.Status != core.Healthy || gets.Load() != 1 {
		t.Fatal(result)
	}
	target.HTTP.BodyContains = "missing-secret"
	result := execute(t, target)
	if result.Status != core.Unhealthy {
		t.Fatal(result)
	}
	limit := 4
	target.HTTP.MaxBodyBytes = &limit
	result = execute(t, target)
	if result.Status != core.Unknown || result.Checks["functional"].Message != "HTTP response body exceeded the configured size limit" {
		t.Fatal(result)
	}
	var out bytes.Buffer
	if err := core.WriteJSON(&out, []core.Result{result}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "sensitive-response-text") || strings.Contains(out.String(), "missing-secret") {
		t.Fatal("response/expectation leaked")
	}
}

func TestTransportSteps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	target := core.Target{Name: "transport", Type: "http", Endpoint: strings.Replace(server.URL, "127.0.0.1", "localhost", 1)}
	result := execute(t, target)
	if result.Checks["reachability"].Steps["dns"] != core.Healthy || result.Checks["reachability"].Steps["tcp"] != core.Healthy || result.Checks["reachability"].Steps["http"] != core.Healthy {
		t.Fatal(result)
	}
	if _, ok := result.Checks["reachability"].Steps["tls"]; ok {
		t.Fatal("plain HTTP must not claim TLS")
	}
	server.Close()
	result = execute(t, target)
	if result.Status != core.Unreachable || result.Checks["reachability"].Steps["tcp"] != core.Unreachable {
		t.Fatal(result)
	}
	tlsServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	defer tlsServer.Close()
	target.Endpoint = tlsServer.URL
	result = execute(t, target)
	if result.Status != core.Unreachable || result.Checks["reachability"].Steps["tls"] != core.Unreachable {
		t.Fatal(result)
	}
	// Verify the adapter accepts a trusted TLS connection, without disabling verification.
	obs, err := (Adapter{}).Check(context.Background(), core.Request{Target: target, Client: tlsServer.Client()}, "reachability")
	if err != nil || obs.Check.Status != core.Healthy {
		t.Fatal(obs, err)
	}
}

func TestBodyTimeoutDoesNotRetry(t *testing.T) {
	var gets atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			gets.Add(1)
			w.Write([]byte("partial"))
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}
	}))
	defer server.Close()
	timeout, retries := 40, 3
	target := core.Target{Name: "slow-body", Type: "http", Endpoint: server.URL, Checks: []string{"functional"}, TimeoutMS: &timeout, Retries: &retries, HTTP: &core.HTTPOptions{BodyContains: "ready"}}
	if result := execute(t, target); result.Status != core.Unknown || gets.Load() != 1 {
		t.Fatal(result, gets.Load())
	}
}

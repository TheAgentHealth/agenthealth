package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

func TestRouterAndBackendEvidence(t *testing.T) {
	for _, tc := range []struct{ router, backend int }{{200, 200}, {503, 200}, {200, 503}, {401, 200}} {
		t.Run(http.StatusText(tc.router)+http.StatusText(tc.backend), func(t *testing.T) {
			router := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("health method = %s", r.Method)
				}
				w.WriteHeader(tc.router)
			}))
			defer router.Close()
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.backend) }))
			defer backend.Close()
			registry := core.NewRegistry()
			if err := registry.Register(Adapter{}); err != nil {
				t.Fatal(err)
			}
			if err := registry.Register(httpadapter.Adapter{}); err != nil {
				t.Fatal(err)
			}
			config := core.Config{Version: "v1", Targets: []core.Target{{Name: "router", Type: "router", Endpoint: router.URL, HTTP: &core.HTTPOptions{ExpectedStatus: []int{200}}, Dependencies: []core.Dependency{{Target: core.Target{Name: "backend", Type: "http", Endpoint: backend.URL}}}}}}
			if err := config.Validate(); err != nil {
				t.Fatal(err)
			}
			results, err := core.NewEngine(registry).Run(context.Background(), config)
			if err != nil {
				t.Fatal(err)
			}
			want := core.Healthy
			if tc.router == 503 {
				want = core.Unhealthy
			}
			if tc.router == 401 {
				want = core.Misconfigured
			}
			if results[0].Checks["protocol"].Status != want {
				t.Fatalf("router checks: %+v", results[0].Checks)
			}
			backendWant := core.Healthy
			if tc.backend == 503 {
				backendWant = core.Unhealthy
			}
			if results[0].Dependencies[0].Status != backendWant {
				t.Fatalf("backend: %+v", results[0].Dependencies)
			}
		})
	}
}

func TestIndependentRouteEvidence(t *testing.T) {
	for _, tc := range []struct {
		signal, backend, route int
		optional               bool
		want                   core.Status
	}{
		{200, 200, 503, false, core.Unhealthy},
		{503, 200, 200, false, core.Unhealthy},
		{200, 503, 200, false, core.Unhealthy},
		{200, 200, 503, true, core.Degraded},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/signal":
				w.WriteHeader(tc.signal)
			case "/backend":
				w.WriteHeader(tc.backend)
			case "/route":
				w.WriteHeader(tc.route)
			}
		}))
		registry := core.NewRegistry()
		if err := registry.Register(Adapter{}); err != nil {
			t.Fatal(err)
		}
		target := func(name string) core.Target {
			return core.Target{Name: name, Type: "router", Endpoint: server.URL + "/" + name}
		}
		root := target("signal")
		critical := !tc.optional
		root.Dependencies = []core.Dependency{{Target: target("backend")}, {Target: target("route"), Critical: &critical}}
		results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{root}})
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		if results[0].Status != tc.want || len(results[0].Dependencies) != 2 {
			t.Fatalf("unexpected evidence: %+v", results)
		}
		for i, code := range []int{tc.backend, tc.route} {
			want := core.Healthy
			if code == 503 {
				want = core.Unhealthy
			}
			if results[0].Dependencies[i].Checks["protocol"].Status != want {
				t.Fatalf("lost independent evidence: %+v", results)
			}
		}
	}
}

func TestFunctionalRouteBoundsAndNoRetry(t *testing.T) {
	for _, body := range []string{"ready", "absent", "ready-too-long"} {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.Method != "GET" {
				t.Errorf("method = %s", r.Method)
			}
			_, _ = w.Write([]byte(body))
		}))
		registry := core.NewRegistry()
		if err := registry.Register(Adapter{}); err != nil {
			t.Fatal(err)
		}
		retries, limit := 3, 8
		config := core.Config{Version: "v1", Targets: []core.Target{{Name: "read-only-route", Type: "router", Endpoint: server.URL, Checks: []string{"functional"}, Retries: &retries, HTTP: &core.HTTPOptions{BodyContains: "ready", MaxBodyBytes: &limit}}}}
		if err := config.Validate(); err != nil {
			t.Fatal(err)
		}
		results, err := core.NewEngine(registry).Run(context.Background(), config)
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		want := core.Healthy
		if body == "absent" {
			want = core.Unhealthy
		}
		if len(body) > limit {
			want = core.Unknown
		}
		// Reachability and authentication precede the single functional request.
		if calls != 3 || results[0].Checks["functional"].Status != want {
			t.Fatalf("body=%s calls=%d results=%+v", body, calls, results)
		}
	}
}

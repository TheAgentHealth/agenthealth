package gateway

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

func TestGatewayAndBackendEvidence(t *testing.T) {
	for _, tc := range []struct{ gateway, backend int }{{200, 200}, {503, 200}, {200, 503}, {401, 200}} {
		t.Run(http.StatusText(tc.gateway)+http.StatusText(tc.backend), func(t *testing.T) {
			gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("health method = %s", r.Method)
				}
				w.WriteHeader(tc.gateway)
			}))
			defer gateway.Close()
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.backend) }))
			defer backend.Close()
			registry := core.NewRegistry()
			if err := registry.Register(Adapter{}); err != nil {
				t.Fatal(err)
			}
			if err := registry.Register(httpadapter.Adapter{}); err != nil {
				t.Fatal(err)
			}
			config := core.Config{Version: "v1", Targets: []core.Target{{Name: "gateway", Type: "gateway", Endpoint: gateway.URL, HTTP: &core.HTTPOptions{ExpectedStatus: []int{200}}, Dependencies: []core.Dependency{{Target: core.Target{Name: "backend", Type: "http", Endpoint: backend.URL}}}}}}
			if err := config.Validate(); err != nil {
				t.Fatal(err)
			}
			results, err := core.NewEngine(registry).Run(context.Background(), config)
			if err != nil {
				t.Fatal(err)
			}
			want := core.Healthy
			if tc.gateway == 503 {
				want = core.Unhealthy
			}
			if tc.gateway == 401 {
				want = core.Misconfigured
			}
			if results[0].Checks["protocol"].Status != want {
				t.Fatalf("gateway checks: %+v", results[0].Checks)
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

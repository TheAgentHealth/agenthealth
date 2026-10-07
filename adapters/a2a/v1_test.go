package a2a

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestV1DefaultAndExplicitLegacy(t *testing.T) {
	for _, mode := range []string{"", "1.0", "0.3.0"} {
		t.Run(mode, func(t *testing.T) {
			var reads, lookups, sends atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					reads.Add(1)
					c := agentCard(server.URL + "/rpc")
					if mode != "0.3.0" {
						delete(c, "protocolVersion")
						delete(c, "url")
						c["supportedInterfaces"] = []any{map[string]string{"url": server.URL + "/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "1.0", "tenant": "health"}}
					}
					json.NewEncoder(w).Encode(c)
					return
				}
				var q struct {
					ID     string
					Method string
					Params map[string]json.RawMessage
				}
				json.NewDecoder(r.Body).Decode(&q)
				if mode != "0.3.0" && (r.Header.Get("A2A-Version") != "1.0" || string(q.Params["tenant"]) != `"health"`) {
					t.Error("missing version/tenant")
				}
				if q.Method == "GetTask" || q.Method == "tasks/get" {
					lookups.Add(1)
					json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "error": map[string]any{"code": -32001, "message": "not found"}})
					return
				}
				sends.Add(1)
				var result any = map[string]any{"kind": "message", "messageId": "reply", "role": "agent", "parts": []any{map[string]string{"kind": "text", "text": "OK"}}}
				if mode != "0.3.0" {
					if q.Method != "SendMessage" {
						t.Error("wrong method")
					}
					var message map[string]json.RawMessage
					json.Unmarshal(q.Params["message"], &message)
					if string(message["role"]) != `"ROLE_USER"` || string(message["parts"]) != `[{"text":"health"}]` {
						t.Errorf("bad v1 input: %s", q.Params["message"])
					}
					result = map[string]any{"message": map[string]any{"messageId": "reply", "contextId": "context", "role": "ROLE_AGENT", "parts": []any{map[string]string{"text": "OK"}}}}
				}
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": result})
			}))
			defer server.Close()
			reg := core.NewRegistry()
			reg.Register(Adapter{})
			target := core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, Checks: []string{"authentication", "protocol", "capability", "functional"}, A2A: &core.A2AOptions{ProtocolVersion: mode, RequiredSkills: []string{"health"}, Functional: &core.A2AInteraction{Safe: true, Text: "health"}}}
			for i := 0; i < 2; i++ {
				results, err := core.NewEngine(reg).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target}})
				if err != nil || results[0].Status != core.Healthy {
					t.Fatalf("%+v %v", results, err)
				}
			}
			if reads.Load() != 2 || lookups.Load() != 2 || sends.Load() != 2 {
				t.Fatalf("requests: %d/%d/%d", reads.Load(), lookups.Load(), sends.Load())
			}
		})
	}
}
func TestV1ResultStatesAndOneof(t *testing.T) {
	cases := []struct {
		raw  string
		want core.Status
	}{
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_COMPLETED"}}}`, core.Healthy},
		{`{"task":{"id":"t","status":{"state":3}}}`, core.Healthy},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_FAILED"}}}`, core.Unhealthy},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_CANCELED"}}}`, core.Unhealthy},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_REJECTED"}}}`, core.Unhealthy},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_AUTH_REQUIRED"}}}`, core.Misconfigured},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_INPUT_REQUIRED"}}}`, core.Unknown},
		{`{"task":{"id":"t","status":{"state":"TASK_STATE_WORKING"}}}`, core.Unknown},
		{`{"task":{"id":"t","status":{"state":99}}}`, core.Unhealthy},
		{`{"task":{},"message":{}}`, core.Unhealthy},
		{`{"message":{"messageId":"m","contextId":"c","role":"ROLE_AGENT","parts":[{"text":"OK","url":"https://example.com"}]}}`, core.Unhealthy},
		{`{"message":{"messageId":"m","contextId":"c","role":2,"parts":[{"text":"OK"}]}}`, core.Healthy},
	}
	for _, tc := range cases {
		if got := validateV1Result(json.RawMessage(tc.raw), "functional"); got.Check.Status != tc.want {
			t.Errorf("%s: %+v", tc.raw, got)
		}
	}
}

func TestV1InterfaceAndSecurityValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
		want   core.Status
	}{
		{"no compatible version", func(c map[string]any) {
			c["supportedInterfaces"] = []any{map[string]string{"url": "https://example.com/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "9.0"}}
		}, core.Unhealthy},
		{"no JSONRPC", func(c map[string]any) {
			c["supportedInterfaces"] = []any{map[string]string{"url": "https://example.com/rpc", "protocolBinding": "GRPC", "protocolVersion": "1.0"}}
		}, core.Unhealthy},
		{"missing binding", func(c map[string]any) {
			c["supportedInterfaces"] = []any{map[string]string{"url": "https://example.com/rpc", "protocolVersion": "1.0"}}
		}, core.Unhealthy},
		{"invalid security oneof", func(c map[string]any) {
			c["securitySchemes"] = map[string]any{"bearer": map[string]any{"httpAuthSecurityScheme": map[string]string{"scheme": "bearer"}, "apiKeySecurityScheme": map[string]string{"name": "key"}}}
		}, core.Unhealthy},
		{"valid bearer", func(c map[string]any) {
			c["securitySchemes"] = map[string]any{"bearer": map[string]any{"httpAuthSecurityScheme": map[string]string{"scheme": "bearer"}}}
			c["securityRequirements"] = []any{map[string]any{"schemes": map[string]any{"bearer": map[string]any{"list": []string{}}}}}
		}, core.Healthy},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := agentCard("https://example.com/rpc")
			c["supportedInterfaces"] = []any{map[string]string{"url": "https://example.com/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}}
			tc.mutate(c)
			raw, _ := json.Marshal(c)
			card, obs := v1Card(raw)
			if obs.Check.Status != tc.want {
				t.Fatalf("%+v", obs)
			}
			if tc.name == "valid bearer" && (card.SecuritySchemes["bearer"].Scheme != "bearer" || len(card.Security) != 1) {
				t.Fatalf("%+v", card)
			}
		})
	}
}

func TestRequiredCapabilityAbsentOrFalse(t *testing.T) {
	for _, version := range []string{"1.0", "0.3.0"} {
		for _, present := range []bool{false, true} {
			t.Run(version+map[bool]string{false: "/absent", true: "/false"}[present], func(t *testing.T) {
				var server *httptest.Server
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == "GET" {
						c := agentCard(server.URL + "/rpc")
						c["capabilities"] = map[string]any{}
						if present {
							c["capabilities"] = map[string]any{"streaming": false}
						}
						if version == "1.0" {
							c["supportedInterfaces"] = []any{map[string]string{"url": server.URL + "/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"}}
						}
						json.NewEncoder(w).Encode(c)
						return
					}
					var q struct{ ID string }
					json.NewDecoder(r.Body).Decode(&q)
					json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "error": map[string]any{"code": -32001, "message": "not found"}})
				}))
				defer server.Close()
				reg := core.NewRegistry()
				reg.Register(Adapter{})
				results, err := core.NewEngine(reg).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "peer", Type: "a2a", Endpoint: server.URL, A2A: &core.A2AOptions{ProtocolVersion: version, RequiredCapabilities: []string{"streaming"}}}}})
				if err != nil {
					t.Fatal(err)
				}
				check := results[0].Checks["capability"]
				if results[0].Status != core.Unhealthy || check.Status != core.Unhealthy || check.Code != "a2a_required" {
					t.Fatalf("%+v", results[0])
				}
			})
		}
	}
}

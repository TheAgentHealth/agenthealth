package a2a

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func agentCard(endpoint string) map[string]any {
	return map[string]any{"protocolVersion": "0.3.0", "name": "Peer", "description": "Health peer", "version": "1", "url": endpoint, "capabilities": map[string]any{"streaming": true}, "defaultInputModes": []string{"text/plain"}, "defaultOutputModes": []string{"text/plain"}, "skills": []any{map[string]any{"id": "health", "name": "Health", "description": "Read-only health response", "tags": []string{"health"}}}}
}
func execute(t *testing.T, target core.Target) core.Result {
	t.Helper()
	return executeAdapter(t, target, Adapter{})
}
func executeAdapter(t *testing.T, target core.Target, adapter Adapter) core.Result {
	t.Helper()
	reg := core.NewRegistry()
	if err := reg.Register(adapter); err != nil {
		t.Fatal(err)
	}
	results, err := core.NewEngine(reg).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	return results[0]
}
func TestPassiveAndFunctional(t *testing.T) {
	var gets, posts, sends atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer super-secret" {
			t.Error("missing bearer")
		}
		if r.Method == "GET" {
			gets.Add(1)
			if r.URL.Path != "/.well-known/agent-card.json" {
				t.Error("discovery path")
			}
			json.NewEncoder(w).Encode(agentCard(server.URL + "/rpc"))
			return
		}
		posts.Add(1)
		if r.URL.Path != "/rpc" {
			t.Error("RPC path")
		}
		var req struct {
			ID     string
			Method string
			Params map[string]any
		}
		json.NewDecoder(r.Body).Decode(&req)
		reply := map[string]any{"jsonrpc": "2.0", "id": req.ID}
		if req.Method == "tasks/get" {
			reply["error"] = map[string]any{"code": -32001, "message": "missing super-secret"}
		} else if req.Method == "message/send" {
			sends.Add(1)
			m := req.Params["message"].(map[string]any)
			if m["kind"] != "message" || m["role"] != "user" {
				t.Error("invalid message")
			}
			reply["result"] = map[string]any{"kind": "message", "role": "agent", "messageId": "reply", "parts": []any{map[string]any{"kind": "text", "text": "super-secret"}}}
		} else {
			t.Error("unexpected method")
		}
		json.NewEncoder(w).Encode(reply)
	}))
	defer server.Close()
	t.Setenv("A2A_TOKEN", "super-secret")
	target := core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL + "/ignored", Auth: &core.AuthReference{BearerEnv: "A2A_TOKEN"}, A2A: &core.A2AOptions{RequiredSkills: []string{"health"}, RequiredCapabilities: []string{"streaming"}}}
	r := execute(t, target)
	if r.Status != core.Healthy || gets.Load() != 1 || posts.Load() != 2 || sends.Load() != 0 {
		t.Fatalf("passive: %+v, gets %d posts %d", r, gets.Load(), posts.Load())
	}
	var output strings.Builder
	if err := core.WriteJSON(&output, []core.Result{r}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "super-secret") {
		t.Fatal("secret exposed")
	}
	target.Checks = []string{"functional"}
	target.A2A.RequiredSkills = nil
	target.A2A.RequiredCapabilities = nil
	target.A2A.Functional = &core.A2AInteraction{Safe: true, Text: "Return OK without external actions"}
	r = execute(t, target)
	if r.Status != core.Healthy || sends.Load() != 1 {
		t.Fatalf("functional: %+v", r)
	}
}
func TestCardAndRPCFailures(t *testing.T) {
	cases := []struct {
		name       string
		mutate     func(map[string]any)
		cardStatus int
		rpc        string
		want       core.Status
	}{
		{"required extension", func(c map[string]any) {
			c["capabilities"] = map[string]any{"extensions": []any{map[string]any{"uri": "https://example.com/required", "required": true}}}
		}, 0, "", core.Unhealthy},
		{"malformed extension", func(c map[string]any) {
			c["capabilities"] = map[string]any{"extensions": []any{map[string]any{"uri": "", "required": false}}}
		}, 0, "", core.Unhealthy},
		{"missing name", func(c map[string]any) { delete(c, "name") }, 0, "", core.Unhealthy},
		{"unsupported version", func(c map[string]any) { c["protocolVersion"] = "1.0" }, 0, "", core.Unhealthy},
		{"transport", func(c map[string]any) { c["preferredTransport"] = "GRPC" }, 0, "", core.Unhealthy},
		{"foreign origin", func(c map[string]any) { c["url"] = "https://foreign.example/rpc" }, 0, "", core.Misconfigured},
		{"user info", func(c map[string]any) { c["url"] = "http://secret@example.com" }, 0, "", core.Unhealthy},
		{"bad capability", func(c map[string]any) { c["capabilities"] = map[string]any{"streaming": "true"} }, 0, "", core.Unhealthy},
		{"duplicate skill", func(c map[string]any) { c["skills"] = append(c["skills"].([]any), c["skills"].([]any)[0]) }, 0, "", core.Unhealthy},
		{"declared auth", func(c map[string]any) {
			c["security"] = []any{map[string]any{"token": []string{}}}
			c["securitySchemes"] = map[string]any{"token": map[string]any{"type": "http", "scheme": "bearer"}}
		}, 0, "", core.Misconfigured},
		{"missing skill", func(c map[string]any) { c["skills"] = []any{} }, 0, "", core.Unhealthy},
		{"card auth", nil, 401, "", core.Misconfigured},
		{"card status", nil, 500, "", core.Unhealthy},
		{"redirect", nil, 302, "", core.Unhealthy},
		{"rpc malformed", nil, 0, `{"jsonrpc":"2.0","id":"wrong","result":null}`, core.Unhealthy},
		{"rpc both", nil, 0, `{"jsonrpc":"2.0","id":"ID","result":null,"error":{"code":-32001,"message":"missing"}}`, core.Unhealthy},
		{"rpc error null", nil, 0, `{"jsonrpc":"2.0","id":"ID","error":null}`, core.Unhealthy},
		{"rpc rejection", nil, 0, `{"jsonrpc":"2.0","id":"ID","error":{"code":-32601,"message":"unknown"}}`, core.Unhealthy},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					if tc.cardStatus != 0 {
						w.WriteHeader(tc.cardStatus)
						return
					}
					c := agentCard(server.URL + "/rpc")
					if tc.mutate != nil {
						tc.mutate(c)
					}
					json.NewEncoder(w).Encode(c)
					return
				}
				var req struct{ ID string }
				json.NewDecoder(r.Body).Decode(&req)
				if tc.rpc != "" {
					w.Write([]byte(strings.ReplaceAll(tc.rpc, "ID", req.ID)))
					return
				}
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "error": map[string]any{"code": -32001, "message": "missing"}})
			}))
			defer server.Close()
			r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, A2A: &core.A2AOptions{RequiredSkills: []string{"health"}}})
			if r.Status != tc.want {
				t.Fatalf("got %+v want %s", r, tc.want)
			}
		})
	}
}
func TestFunctionalStates(t *testing.T) {
	for _, tc := range []struct {
		state string
		want  core.Status
	}{{"completed", core.Healthy}, {"working", core.Unknown}, {"input-required", core.Unknown}, {"failed", core.Unhealthy}, {"rejected", core.Unhealthy}, {"auth-required", core.Misconfigured}, {"nonsense", core.Unhealthy}} {
		t.Run(tc.state, func(t *testing.T) {
			var sends atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					json.NewEncoder(w).Encode(agentCard(server.URL))
					return
				}
				var req struct{ ID, Method string }
				json.NewDecoder(r.Body).Decode(&req)
				res := map[string]any{"jsonrpc": "2.0", "id": req.ID}
				if req.Method == "tasks/get" {
					res["error"] = map[string]any{"code": -32001, "message": "missing"}
				} else {
					sends.Add(1)
					res["result"] = map[string]any{"kind": "task", "id": "task", "contextId": "context", "status": map[string]any{"state": tc.state}}
				}
				json.NewEncoder(w).Encode(res)
			}))
			defer server.Close()
			retries := 3
			r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, Retries: &retries, Checks: []string{"functional"}, A2A: &core.A2AOptions{Functional: &core.A2AInteraction{Safe: true, Text: "Read-only health"}}})
			if r.Status != tc.want || sends.Load() != 1 {
				t.Fatalf("result %+v sends %d", r, sends.Load())
			}
		})
	}
}
func TestLimitsTimeoutAndTLS(t *testing.T) {
	for _, tc := range []struct {
		name    string
		handler http.HandlerFunc
		want    core.Status
	}{
		{"oversize", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(strings.Repeat("x", bodyLimit+1))) }, core.Unknown},
		{"partial timeout", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{"))
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}, core.Unknown},
		{"no response timeout", func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }, core.Unreachable},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()
			timeout := 100
			r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, TimeoutMS: &timeout})
			if r.Status != tc.want {
				t.Fatalf("%+v", r)
			}
		})
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("untrusted TLS contacted") }))
	defer server.Close()
	r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL})
	if r.Status != core.Unreachable {
		t.Fatalf("TLS %+v", r)
	}
}
func TestUnsafeDiscoveryAndMissingCredentials(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	for _, target := range []core.Target{
		{Name: "peer", Type: "a2a", Endpoint: server.URL, A2A: &core.A2AOptions{CardURL: "https://foreign.example/card"}},
		{Name: "peer", Type: "a2a", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "A2A_MISSING"}},
		{Name: "peer", Type: "a2a", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "A2A_BAD"}},
	} {
		t.Setenv("A2A_MISSING", "")
		t.Setenv("A2A_BAD", "bad\r\nvalue")
		r := execute(t, target)
		if r.Status != core.Misconfigured {
			t.Fatalf("%+v", r)
		}
	}
	if calls.Load() != 0 {
		t.Fatal("network used before configuration gate")
	}
}
func TestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	reg := core.NewRegistry()
	reg.Register(Adapter{})
	start := time.Now()
	_, err := core.NewEngine(reg).Run(ctx, core.Config{Version: "v1", Targets: []core.Target{{Name: "peer", Type: "a2a", Endpoint: "http://localhost"}}})
	if err != nil || time.Since(start) > time.Second {
		t.Fatalf("cancellation %v", err)
	}
}

func TestRedirectDoesNotForwardCredential(t *testing.T) {
	var leaked atomic.Int32
	foreign := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked.Add(1) }))
	defer foreign.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", foreign.URL)
		w.WriteHeader(302)
	}))
	defer server.Close()
	t.Setenv("A2A_TOKEN", "secret")
	r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "A2A_TOKEN"}})
	if r.Status != core.Unhealthy || leaked.Load() != 0 {
		t.Fatalf("redirect %+v leaked %d", r, leaked.Load())
	}
}
func TestRPCAuthenticationBlocksInteraction(t *testing.T) {
	for _, status := range []int{401, 403} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var calls atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					json.NewEncoder(w).Encode(agentCard(server.URL))
					return
				}
				calls.Add(1)
				w.WriteHeader(status)
			}))
			defer server.Close()
			r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, Checks: []string{"functional"}, A2A: &core.A2AOptions{Functional: &core.A2AInteraction{Safe: true, Text: "Read-only"}}})
			if r.Status != core.Misconfigured || calls.Load() != 1 {
				t.Fatalf("%+v calls %d", r, calls.Load())
			}
			if _, exists := r.Checks["functional"]; exists {
				t.Fatal("interaction ran after rejected authentication")
			}
		})
	}
}
func TestCustomCardAndOptionalSecurity(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			if r.URL.Path != "/custom/card" {
				t.Error("wrong custom card path")
			}
			c := agentCard(server.URL + "/rpc")
			c["security"] = []any{map[string]any{}, map[string]any{"unsupported": []string{}}}
			c["capabilities"] = map[string]any{"streaming": true, "extensions": []any{map[string]any{"uri": "https://example.com/optional", "required": false}}}
			json.NewEncoder(w).Encode(c)
			return
		}
		var req struct {
			ID     string
			Params struct{ ID string }
		}
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": map[string]any{"kind": "task", "id": req.Params.ID, "contextId": "context", "status": map[string]any{"state": "working"}}})
	}))
	defer server.Close()
	r := execute(t, core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, A2A: &core.A2AOptions{CardURL: server.URL + "/custom/card"}})
	if r.Status != core.Healthy {
		t.Fatalf("%+v", r)
	}
}

func TestMIMEModes(t *testing.T) {
	for _, tc := range []struct {
		name, input, output string
		functional          bool
		want                core.Status
	}{
		{"invalid input", "not-a-media-type", "text/plain", false, core.Unhealthy},
		{"invalid output", "text/plain", "text/plain; charset", false, core.Unhealthy},
		{"missing subtype", "text/", "text/plain", false, core.Unhealthy},
		{"missing type", "/plain", "text/plain", false, core.Unhealthy},
		{"invalid syntax", "text/pl ain", "text/plain", false, core.Unhealthy},
		{"case and parameters", "Text/Plain; Charset=UTF-8", "TEXT/PLAIN; charset=utf-8", true, core.Healthy},
		{"valid nontext", "application/json", "application/json", false, core.Healthy},
		{"unsupported interaction modes", "application/json", "application/json", true, core.Misconfigured},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var sends atomic.Int32
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					c := agentCard(server.URL)
					c["defaultInputModes"] = []string{tc.input}
					c["defaultOutputModes"] = []string{tc.output}
					json.NewEncoder(w).Encode(c)
					return
				}
				var req struct{ ID, Method string }
				json.NewDecoder(r.Body).Decode(&req)
				reply := map[string]any{"jsonrpc": "2.0", "id": req.ID}
				if req.Method == "tasks/get" {
					reply["error"] = map[string]any{"code": -32001, "message": "not found"}
				} else {
					sends.Add(1)
					reply["result"] = map[string]any{"kind": "message", "role": "agent", "messageId": "reply", "parts": []any{map[string]any{"kind": "text", "text": "OK"}}}
				}
				json.NewEncoder(w).Encode(reply)
			}))
			defer server.Close()
			target := core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL}
			if tc.functional {
				target.Checks = []string{"functional"}
				target.A2A = &core.A2AOptions{Functional: &core.A2AInteraction{Safe: true, Text: "Read-only health"}}
			}
			r := execute(t, target)
			if r.Status != tc.want {
				t.Fatalf("%+v", r)
			}
			wantSends := int32(0)
			if tc.functional && tc.want == core.Healthy {
				wantSends = 1
			}
			if sends.Load() != wantSends {
				t.Fatalf("sends %d want %d", sends.Load(), wantSends)
			}
		})
	}
}

type failingRandom struct{}

func (failingRandom) Read([]byte) (int, error) {
	return 0, errors.New("random failure with secret details")
}

func TestRandomFailuresBeforeRPC(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
	defer server.Close()
	target := core.Target{Name: "peer", Type: "a2a", Endpoint: server.URL, Checks: []string{"functional"}, A2A: &core.A2AOptions{Functional: &core.A2AInteraction{Safe: true, Text: "Read-only health"}}}
	for _, dimension := range []string{"authentication", "protocol", "functional"} {
		for _, prefix := range []int{0, 16} {
			t.Run(dimension+"/"+string(rune('0'+prefix/16)), func(t *testing.T) {
				source := io.MultiReader(bytes.NewReader(make([]byte, prefix)), failingRandom{})
				state := &sync.Map{}
				raw, _ := json.Marshal(agentCard(server.URL))
				state.Store("a2a-card", fetched{status: 200, body: raw})
				adapter := Adapter{random: source}
				_, err := adapter.Check(context.Background(), core.Request{Target: target, RunState: state, Client: server.Client()}, dimension)
				if err == nil {
					t.Fatal("missing random-source error")
				}
			})
		}
	}
	if calls.Load() != 0 {
		t.Fatal("RPC sent after random-source failure")
	}
	// The engine must classify the returned error without exposing reader details.
	var cardServer *httptest.Server
	cardServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { json.NewEncoder(w).Encode(agentCard(cardServer.URL)) }))
	defer cardServer.Close()
	target.Endpoint = cardServer.URL
	r := executeAdapter(t, target, Adapter{random: failingRandom{}})
	if r.Status != core.Unknown {
		t.Fatalf("random failure: %+v", r)
	}
	var output strings.Builder
	if err := core.WriteJSON(&output, []core.Result{r}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "secret details") {
		t.Fatal("raw random-source error exposed")
	}
	if calls.Load() != 0 {
		t.Fatal("RPC sent despite engine random-source failure")
	}
}

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

type fixture struct {
	mu                                 sync.Mutex
	methods                            []string
	sse, pagination, unsafe, callError bool
	version, auth                      string
	override                           func(string) any
}

func (f *fixture) server(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.auth != "" && r.Header.Get("Authorization") != "Bearer "+f.auth {
			w.WriteHeader(401)
			return
		}
		if r.Method == "HEAD" {
			w.WriteHeader(405)
			return
		}
		if r.Method == "DELETE" {
			f.mu.Lock()
			f.methods = append(f.methods, "DELETE")
			f.mu.Unlock()
			w.WriteHeader(204)
			return
		}
		var msg struct {
			ID     int                        `json:"id"`
			Method string                     `json:"method"`
			Params map[string]json.RawMessage `json:"params"`
		}
		if json.NewDecoder(r.Body).Decode(&msg) != nil {
			t.Error("invalid request")
			w.WriteHeader(400)
			return
		}
		if r.Header.Get("Accept") != "application/json, text/event-stream" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("missing transport headers")
		}
		f.mu.Lock()
		f.methods = append(f.methods, msg.Method)
		f.mu.Unlock()
		if msg.Method == "server/discover" {
			w.WriteHeader(400)
			return
		}
		if msg.Method != "initialize" {
			version := f.version
			if version == "" {
				version = "2025-11-25"
			}
			if r.Header.Get("Mcp-Session-Id") != "test-session" || r.Header.Get("MCP-Protocol-Version") != version {
				t.Error("missing negotiated headers")
			}
		}
		if msg.Method == "server/discover" {
			w.WriteHeader(400)
			return
		}
		if msg.Method == "notifications/initialized" {
			w.WriteHeader(202)
			return
		}
		var result any
		switch msg.Method {
		case "initialize":
			w.Header().Set("Mcp-Session-Id", "test-session")
			version := f.version
			if version == "" {
				version = "2025-11-25"
			}
			result = map[string]any{"protocolVersion": version, "serverInfo": map[string]string{"name": "fixture", "version": "1"}, "capabilities": map[string]any{"tools": map[string]any{}, "resources": map[string]any{}, "prompts": map[string]any{}}}
		case "tools/list":
			tools := []any{map[string]any{"name": "search", "inputSchema": map[string]string{"type": "object"}, "annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": f.unsafe}}}
			result = map[string]any{"tools": tools}
			if f.pagination && msg.Params["cursor"] == nil {
				result = map[string]any{"tools": []any{}, "nextCursor": "page-2"}
			}
		case "resources/list":
			result = map[string]any{"resources": []any{map[string]string{"name": "ready", "uri": "health://ready"}}}
		case "prompts/list":
			result = map[string]any{"prompts": []any{map[string]string{"name": "summary"}}}
		case "tools/call":
			if string(msg.Params["name"]) != `"search"` || string(msg.Params["arguments"]) != `{"query":"health"}` {
				t.Errorf("unexpected invocation: %s", msg.Params)
			}
			result = map[string]any{"content": []any{map[string]string{"type": "text", "text": "secret-response"}}, "isError": f.callError}
		default:
			t.Errorf("unexpected method %s", msg.Method)
		}
		if f.override != nil {
			if replacement := f.override(msg.Method); replacement != nil {
				result = replacement
			}
		}
		envelope := map[string]any{"jsonrpc": "2.0", "id": msg.ID, "result": result}
		if f.sse {
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: message\ndata: {\"jsonrpc\":\"2.0\",\"method\":\"notifications/message\",\"params\":{}}\n\n")
			b, _ := json.Marshal(envelope)
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", b)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(envelope)
	}))
}
func runTarget(t *testing.T, target core.Target) core.Result {
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
func TestDiscoveryAndFunctional(t *testing.T) {
	for _, sse := range []bool{false, true} {
		for _, active := range []bool{false, true} {
			t.Run(fmt.Sprintf("sse=%t/active=%t", sse, active), func(t *testing.T) {
				f := &fixture{sse: sse, pagination: true, auth: "credential-value", version: "2025-06-18"}
				server := f.server(t)
				defer server.Close()
				t.Setenv("MCP_TEST_TOKEN", f.auth)
				target := core.Target{Name: "test", Type: "mcp", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "MCP_TEST_TOKEN"}, MCP: &core.MCPOptions{RequiredTools: []string{"search"}, RequiredResources: []string{"health://ready"}, RequiredPrompts: []string{"summary"}}}
				if active {
					retries := 3
					target.Retries = &retries
					target.Checks = []string{"protocol", "capability", "functional", "latency"}
					target.MCP.Functional = &core.MCPInvocation{Tool: "search", Safe: true, ArgumentsJSON: `{"query":"health"}`}
				}
				result := runTarget(t, target)
				if result.Status != core.Healthy || result.LatencyMS == nil {
					t.Fatalf("result: %+v", result)
				}
				f.mu.Lock()
				methods := append([]string{}, f.methods...)
				f.mu.Unlock()
				calls, deletes := 0, 0
				for _, m := range methods {
					if m == "tools/call" {
						calls++
					}
					if m == "DELETE" {
						deletes++
					}
				}
				expected := 0
				if active {
					expected = 1
				}
				if calls != expected || deletes < 3 {
					t.Fatalf("calls=%d deletes=%d methods=%v", calls, deletes, methods)
				}
				var output bytes.Buffer
				core.WriteJSON(&output, []core.Result{result})
				if strings.Contains(output.String(), f.auth) || strings.Contains(output.String(), "secret-response") {
					t.Fatal("sensitive response escaped")
				}
			})
		}
	}
}
func TestClassification(t *testing.T) {
	for _, tc := range []struct {
		name    string
		fixture *fixture
		options *core.MCPOptions
		active  bool
		status  core.Status
	}{
		{name: "missing-tool", options: &core.MCPOptions{RequiredTools: []string{"missing"}}, status: core.Degraded},
		{name: "missing-resource", options: &core.MCPOptions{RequiredResources: []string{"missing"}}, status: core.Degraded},
		{name: "missing-prompt", options: &core.MCPOptions{RequiredPrompts: []string{"missing"}}, status: core.Degraded},
		{name: "unsupported-version", fixture: &fixture{version: "2099-01-01"}, status: core.Misconfigured},
		{name: "pinned-version", fixture: &fixture{version: "2025-06-18"}, options: &core.MCPOptions{ProtocolVersion: "2025-11-25"}, status: core.Misconfigured},
		{name: "bad-init", fixture: &fixture{override: func(m string) any {
			if m == "initialize" {
				return map[string]any{}
			}
			return nil
		}}, status: core.Unhealthy},
		{name: "malformed-list", fixture: &fixture{override: func(m string) any {
			if m == "tools/list" {
				return map[string]any{"tools": nil}
			}
			return nil
		}}, status: core.Unhealthy},
		{name: "pagination-loop", fixture: &fixture{override: func(m string) any {
			if m == "tools/list" {
				return map[string]any{"tools": []any{}, "nextCursor": "repeat"}
			}
			return nil
		}}, status: core.Unknown},
		{name: "unsafe-probe", fixture: &fixture{unsafe: true}, active: true, status: core.Misconfigured},
		{name: "failed-probe", fixture: &fixture{callError: true}, active: true, status: core.Unhealthy},
		{name: "authentication", fixture: &fixture{auth: "required-token"}, status: core.Misconfigured},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.fixture == nil {
				tc.fixture = &fixture{}
			}
			server := tc.fixture.server(t)
			defer server.Close()
			target := core.Target{Name: "test", Type: "mcp", Endpoint: server.URL, MCP: tc.options}
			if tc.active {
				target.Checks = []string{"functional"}
				target.MCP = &core.MCPOptions{Functional: &core.MCPInvocation{Tool: "search", Safe: true, ArgumentsJSON: `{"query":"health"}`}}
			}
			result := runTarget(t, target)
			if result.Status != tc.status {
				t.Fatalf("got %+v expected %s", result, tc.status)
			}
			if tc.name == "unsafe-probe" {
				tc.fixture.mu.Lock()
				defer tc.fixture.mu.Unlock()
				for _, m := range tc.fixture.methods {
					if m == "tools/call" {
						t.Fatal("unsafe invocation")
					}
				}
			}
		})
	}
}
func TestTransportSafety(t *testing.T) {
	t.Run("TLS", func(t *testing.T) {
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer server.Close()
		result := runTarget(t, core.Target{Name: "tls", Type: "mcp", Endpoint: server.URL})
		if result.Status != core.Unreachable {
			t.Fatalf("%+v", result)
		}
	})
	t.Run("redirect", func(t *testing.T) {
		var reached bool
		dst := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
		defer dst.Close()
		src := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, dst.URL, 307) }))
		defer src.Close()
		result := runTarget(t, core.Target{Name: "redirect", Type: "mcp", Endpoint: src.URL})
		if reached || result.Status != core.Unhealthy {
			t.Fatalf("followed=%v result=%+v", reached, result)
		}
	})
	for _, mode := range []string{"size", "timeout", "bad-id", "rpc-error"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "HEAD" {
					w.WriteHeader(200)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				switch mode {
				case "size":
					fmt.Fprint(w, strings.Repeat("x", maxBody+1))
				case "timeout":
					w.WriteHeader(200)
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				case "bad-id":
					fmt.Fprint(w, `{"jsonrpc":"2.0","id":999,"result":{}}`)
				case "rpc-error":
					fmt.Fprint(w, `{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"secret-value"}}`)
				}
			}))
			defer server.Close()
			timeout := 80
			result := runTarget(t, core.Target{Name: "test", Type: "mcp", Endpoint: server.URL, Checks: []string{"protocol"}, TimeoutMS: &timeout})
			expected := core.Unhealthy
			if mode == "size" || mode == "timeout" {
				expected = core.Unknown
			}
			if result.Status != expected {
				t.Fatalf("%+v", result)
			}
		})
	}
}
func TestSSECompletesBeforeStreamCloses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			return
		}
		var msg struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		json.NewDecoder(r.Body).Decode(&msg)
		if msg.Method == "server/discover" {
			w.WriteHeader(400)
			return
		}
		if msg.Method == "notifications/initialized" {
			w.WriteHeader(202)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: {\"jsonrpc\":\"2.0\",\"id\":%d,\"result\":{\"protocolVersion\":\"2025-11-25\",\"capabilities\":{},\"serverInfo\":{\"name\":\"test\",\"version\":\"1\"}}}\n\n", msg.ID)
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	timeout := 500
	started := time.Now()
	result := runTarget(t, core.Target{Name: "sse", Type: "mcp", Endpoint: server.URL, Checks: []string{"protocol"}, TimeoutMS: &timeout})
	if result.Status != core.Healthy || time.Since(started) > 400*time.Millisecond {
		t.Fatalf("%+v", result)
	}
}

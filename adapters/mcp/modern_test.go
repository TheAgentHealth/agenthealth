package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestModernHTTP(t *testing.T) {
	for _, sse := range []bool{false, true} {
		t.Run(fmt.Sprint(sse), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "HEAD" {
					w.WriteHeader(405)
					return
				}
				if r.Method != "POST" {
					t.Errorf("modern session operation %s", r.Method)
					w.WriteHeader(400)
					return
				}
				var request struct {
					ID     int                        `json:"id"`
					Method string                     `json:"method"`
					Params map[string]json.RawMessage `json:"params"`
				}
				json.NewDecoder(r.Body).Decode(&request)
				var meta map[string]json.RawMessage
				json.Unmarshal(request.Params["_meta"], &meta)
				if string(meta["io.modelcontextprotocol/protocolVersion"]) != `"2026-07-28"` || r.Header.Get("MCP-Protocol-Version") != modernVersion || r.Header.Get("Mcp-Method") != request.Method || r.Header.Get("Mcp-Session-Id") != "" {
					t.Error("incorrect modern metadata")
				}
				var result any
				switch request.Method {
				case "server/discover":
					result = map[string]any{"resultType": "complete", "supportedVersions": []string{modernVersion}, "capabilities": map[string]any{"tools": map[string]any{}, "resources": map[string]any{}, "prompts": map[string]any{}}}
				case "tools/list":
					result = map[string]any{"resultType": "complete", "tools": []any{map[string]any{"name": "search 世界", "inputSchema": map[string]any{"type": "object", "properties": map[string]any{"query": map[string]string{"type": "string", "x-mcp-header": "Query"}}}, "annotations": map[string]bool{"readOnlyHint": true, "destructiveHint": false}}}}
				case "resources/list":
					result = map[string]any{"resultType": "complete", "resources": []any{}}
				case "prompts/list":
					result = map[string]any{"resultType": "complete", "prompts": []any{}}
				case "tools/call":
					calls.Add(1)
					if r.Header.Get("Mcp-Name") != encodeHeader("search 世界") || r.Header.Get("Mcp-Param-Query") != encodeHeader("hello\n世界") {
						t.Error("incorrect mirrored headers")
					}
					result = map[string]any{"resultType": "complete", "content": []any{map[string]string{"type": "text", "text": "ok"}}}
				default:
					t.Errorf("unexpected method %s", request.Method)
				}
				envelope := map[string]any{"jsonrpc": "2.0", "id": request.ID, "result": result}
				data, _ := json.Marshal(envelope)
				if sse {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprintf(w, "data: %s\n\n", data)
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.Write(data)
				}
			}))
			defer server.Close()
			retries := 3
			target := core.Target{Name: "modern", Type: "mcp", Endpoint: server.URL, Retries: &retries, Checks: []string{"protocol", "capability", "functional"}, MCP: &core.MCPOptions{Functional: &core.MCPInvocation{Tool: "search 世界", Safe: true, ArgumentsJSON: `{"query":"hello\n世界"}`}}}
			result := runTarget(t, target)
			if result.Status != core.Healthy || calls.Load() != 1 {
				t.Fatalf("%+v calls=%d", result, calls.Load())
			}
		})
	}
}
func TestModernErrorsDoNotFallBack(t *testing.T) {
	for _, code := range []int{-32022, -32020, -32601} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "HEAD" {
					return
				}
				var q struct {
					ID     int    `json:"id"`
					Method string `json:"method"`
				}
				json.NewDecoder(r.Body).Decode(&q)
				if q.Method != "server/discover" {
					t.Error("fell back to initialization")
				}
				w.Header().Set("Content-Type", "application/json")
				status := 400
				if code == -32601 {
					status = 404
				}
				w.WriteHeader(status)
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "error": map[string]any{"code": code, "message": "private", "data": map[string]any{"supported": []string{"2099-01-01"}}}})
			}))
			defer server.Close()
			result := runTarget(t, core.Target{Name: "modern", Type: "mcp", Endpoint: server.URL, Checks: []string{"protocol"}})
			expected := core.Unhealthy

			if result.Status != expected {
				t.Fatalf("%+v", result)
			}
		})
	}
}
func TestHeaderBindings(t *testing.T) {
	valid := json.RawMessage(`{"inputSchema":{"type":"object","properties":{"nested":{"type":"object","properties":{"number":{"type":"integer","x-mcp-header":"Number"},"enabled":{"type":"boolean","x-mcp-header":"Enabled"}}},"text":{"type":"string","x-mcp-header":"Text"}}}}`)
	headers, err := invocationHeaders(valid, json.RawMessage(`{"nested":{"number":42,"enabled":true},"text":" padded "}`))
	if err != nil || headers["Mcp-Param-Number"] != "42" || headers["Mcp-Param-Enabled"] != "true" || !strings.HasPrefix(headers["Mcp-Param-Text"], "=?base64?") {
		t.Fatalf("%v %v", headers, err)
	}
	for _, schema := range []string{`{"type":"object","x-mcp-header":"Root"}`, `{"properties":{"x":{"type":"number","x-mcp-header":"Num"}}}`, `{"properties":{"x":{"type":"string","x-mcp-header":"bad name"}}}`, `{"properties":{"x":{"type":"string","x-mcp-header":"Name"},"y":{"type":"string","x-mcp-header":"name"}}}`, `{"anyOf":[{"properties":{"x":{"type":"string","x-mcp-header":"X"}}}]}`} {
		if _, err := headerBindings(json.RawMessage(`{"inputSchema":` + schema + `}`)); err == nil {
			t.Fatalf("invalid accepted %s", schema)
		}
	}
	if _, err := invocationHeaders(valid, json.RawMessage(`{"nested":{"number":9007199254740992}}`)); err == nil {
		t.Fatal("unsafe integer accepted")
	}
	if _, err := invocationHeaders(valid, json.RawMessage(`{"nested":{"number":null}}`)); err != nil {
		t.Fatal("null should omit header")
	}
}

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TheAgentHealth/agenthealth/core"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestContentRequiredFields(t *testing.T) {
	for _, raw := range []string{`{"type":"text"}`, `{"type":"unknown"}`, `{"type":"text","text":null}`, `{"type":"image","data":"a"}`, `{"type":"audio","data":3,"mimeType":"audio/wav"}`, `{"type":"resource_link","uri":"x"}`, `{"type":"resource","resource":{"uri":"x"}}`} {
		if validContent(json.RawMessage(raw)) {
			t.Errorf("accepted malformed content %s", raw)
		}
	}
	for _, raw := range []string{`{"type":"text","text":""}`, `{"type":"image","data":"","mimeType":"image/png"}`, `{"type":"audio","data":"","mimeType":"audio/wav"}`, `{"type":"resource_link","uri":"health://x","name":"x"}`, `{"type":"resource","resource":{"uri":"health://x","text":""}}`, `{"type":"resource","resource":{"uri":"health://x","blob":""}}`} {
		if !validContent(json.RawMessage(raw)) {
			t.Errorf("rejected valid content %s", raw)
		}
	}
}

func TestMalformedFunctionalContent(t *testing.T) {
	f := &fixture{override: func(method string) any {
		if method == "tools/call" {
			return map[string]any{"content": []any{map[string]any{"type": "text"}}}
		}
		return nil
	}}
	server := f.server(t)
	defer server.Close()
	target := core.Target{Name: "test", Type: "mcp", Endpoint: server.URL, Checks: []string{"functional"}, MCP: &core.MCPOptions{Functional: &core.MCPInvocation{Tool: "search", Safe: true, ArgumentsJSON: `{"query":"health"}`}}}
	if result := runTarget(t, target); result.Status != core.Unhealthy {
		t.Fatalf("malformed content passed: %+v", result)
	}
}

func TestSSEResumption(t *testing.T) {
	for _, mode := range []string{"legacy", "modern", "no-id", "limit", "deadline", "budget"} {
		t.Run(mode, func(t *testing.T) {
			gets, posts := 0, 0
			start := time.Now()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				if r.Method == "POST" {
					posts++
					if mode != "no-id" {
						fmt.Fprint(w, "id: cursor\nretry: 20\ndata:\n\n")
					}
					if mode == "deadline" {
						fmt.Fprint(w, "retry: 999999\n\n")
					}
					if mode == "budget" {
						fmt.Fprint(w, ":"+strings.Repeat("x", maxBody/2)+"\n\n")
					}
					return
				}
				gets++
				if r.Method != "GET" || r.Header.Get("Last-Event-ID") != "cursor" || r.Header.Get("Mcp-Session-Id") != "session" || r.Header.Get("MCP-Protocol-Version") != "2025-11-25" {
					t.Error("invalid resumption headers")
				}
				if mode == "limit" {
					fmt.Fprint(w, "id: cursor\ndata:\n\n")
					return
				}
				if mode == "budget" {
					fmt.Fprint(w, ":"+strings.Repeat("x", maxBody/2)+"\n\n")
					return
				}
				fmt.Fprint(w, "data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{}}\n\n")
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
			defer cancel()
			s := session{r: core.Request{Target: core.Target{Endpoint: server.URL}, Client: server.Client()}, version: "2025-11-25", initialized: true, id: "session", modern: mode == "modern"}
			_, err := s.rpc(ctx, "tools/call", map[string]any{"name": "test"})
			if mode == "legacy" {
				if err != nil || gets != 1 || time.Since(start) < 20*time.Millisecond {
					t.Fatalf("resume: gets=%d err=%v", gets, err)
				}
			} else if err == nil {
				t.Fatal("expected failure")
			}
			if posts != 1 {
				t.Fatalf("repeated POST: %d", posts)
			}
			if (mode == "modern" || mode == "no-id" || mode == "deadline") && gets != 0 {
				t.Fatalf("unexpected GET: %d", gets)
			}
			if mode == "limit" && gets != 3 {
				t.Fatalf("GET bound: %d", gets)
			}
		})
	}
}

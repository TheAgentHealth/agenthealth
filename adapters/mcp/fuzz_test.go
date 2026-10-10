package mcp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

func FuzzJSONRPC(f *testing.F) {
	f.Add([]byte(`{"jsonrpc":"2.0","id":1,"result":{}}`))
	f.Add([]byte(`{"jsonrpc":"2.0","id":2,"error":{"code":-32601,"message":"missing"}}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		_, _ = decodeResponse(data, 1)
	})
}
func FuzzSSE(f *testing.F) {
	f.Add([]byte("data: {\"jsonrpc\":\"2.0\",\"id\":1,\"result\":{}}\n\n"))
	f.Add([]byte("id: 1\nretry: 999999999999999999999\ndata: {}\n\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		// Modern transport forbids reconnection: malformed framing cannot access a network.
		s := &session{modern: true}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, _ = s.readSSE(ctx, &http.Response{Body: io.NopCloser(bytes.NewReader(data))}, 1)
	})
}

type fuzzTransport func(*http.Request) (*http.Response, error)

func (fn fuzzTransport) RoundTrip(r *http.Request) (*http.Response, error) { return fn(r) }
func FuzzOAuthDiscovery(f *testing.F) {
	f.Add([]byte(`{"resource":"https://resource.example/mcp","authorization_servers":["https://issuer.example"]}`), []byte(`{"issuer":"https://issuer.example","token_endpoint":"https://issuer.example/token"}`))
	f.Add([]byte(`{}`), []byte(`{"token_endpoint":"http://169.254.169.254"}`))
	f.Fuzz(func(t *testing.T, protected, metadata []byte) {
		if len(protected)+len(metadata) > 65536 {
			t.Skip()
		}
		client := &http.Client{Transport: fuzzTransport(func(r *http.Request) (*http.Response, error) {
			if r.Header.Get("Authorization") != "" {
				t.Fatal("credentials sent to discovery")
			}
			data := metadata
			if strings.Contains(r.URL.Path, "oauth-protected-resource") {
				data = protected
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(bytes.NewReader(data))}, nil
		})}
		request := core.Request{Client: client, Target: core.Target{Endpoint: "https://resource.example/mcp", MCP: &core.MCPOptions{OAuth: &core.MCPOAuth{Issuer: "https://issuer.example"}}}}
		_, _ = discoverOAuth(context.Background(), request)
	})
}

package mcp

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

type oauthFixture struct {
	t                            *testing.T
	server                       *httptest.Server
	tokens                       atomic.Int32
	mu                           sync.Mutex
	challenge                    string
	invalidIssuer, redirectToken bool
	inventory                    bool
}

func newOAuthFixture(t *testing.T) *oauthFixture {
	f := &oauthFixture{t: t}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		root := f.server.URL
		if r.URL.Path != "/mcp" && r.Header.Get("Authorization") != "" && r.URL.Path != "/token" {
			t.Error("credential sent to metadata endpoint")
		}
		switch r.URL.Path {
		case "/.well-known/oauth-protected-resource/mcp":
			json.NewEncoder(w).Encode(map[string]any{"resource": root + "/mcp", "authorization_servers": []string{root}, "scopes_supported": []string{"read"}})
		case "/.well-known/oauth-authorization-server", "/.well-known/openid-configuration":
			issuer := root
			if f.invalidIssuer {
				issuer = "https://untrusted.example"
			}
			json.NewEncoder(w).Encode(map[string]any{"issuer": issuer, "token_endpoint": root + "/token", "authorization_endpoint": root + "/authorize", "code_challenge_methods_supported": []string{"S256"}, "token_endpoint_auth_methods_supported": []string{"client_secret_basic", "none"}, "authorization_response_iss_parameter_supported": true})
		case "/token":
			f.tokens.Add(1)
			r.ParseForm()
			if f.redirectToken {
				http.Redirect(w, r, root+"/steal", 307)
				return
			}
			if r.Form.Get("resource") != root+"/mcp" || r.Form.Get("client_id") != "registered-client" {
				t.Error("missing token resource binding")
			}
			switch r.Form.Get("grant_type") {
			case "client_credentials":
				id, secret, ok := r.BasicAuth()
				if !ok || id != "registered-client" || secret != "client-secret-value" {
					t.Error("invalid client authentication")
				}
			case "refresh_token":
				if r.Form.Get("refresh_token") != "refresh-secret-value" {
					t.Error("invalid refresh token")
				}
			case "authorization_code":
				if r.Form.Get("code") != "authorization-code" || r.Form.Get("redirect_uri") == "" {
					t.Error("invalid code exchange")
				}
				hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
				f.mu.Lock()
				expected := f.challenge
				f.mu.Unlock()
				if base64.RawURLEncoding.EncodeToString(hash[:]) != expected {
					t.Error("PKCE mismatch")
				}
			default:
				t.Error("unsupported grant")
			}
			json.NewEncoder(w).Encode(map[string]any{"access_token": "acquired-sensitive-value", "refresh_token": "rotated-sensitive-value", "token_type": "Bearer", "expires_in": 3600})
		case "/steal":
			t.Error("followed credential redirect")
		case "/mcp":
			if r.Method == "HEAD" || r.Header.Get("Authorization") != "Bearer acquired-sensitive-value" {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s/.well-known/oauth-protected-resource/mcp", scope="read"`, root))
				w.WriteHeader(401)
				return
			}
			var q struct {
				ID     int                        `json:"id"`
				Method string                     `json:"method"`
				Params map[string]json.RawMessage `json:"params"`
			}
			json.NewDecoder(r.Body).Decode(&q)
			if f.inventory && q.Method == "tools/list" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": map[string]any{"resultType": "complete", "tools": []any{map[string]any{"name": "health", "inputSchema": map[string]any{"type": "object"}}}}})
				return
			}
			if q.Method != "server/discover" {
				t.Errorf("unexpected method %s", q.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": q.ID, "result": map[string]any{"resultType": "complete", "supportedVersions": []string{modernVersion}, "capabilities": func() map[string]any {
				if f.inventory {
					return map[string]any{"tools": map[string]any{}}
				}
				return map[string]any{}
			}()}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}
func (f *oauthFixture) target(grant string) core.Target {
	return core.Target{Name: "acquired-sensitive-value", Type: "mcp", Endpoint: f.server.URL + "/mcp", MCP: &core.MCPOptions{OAuth: &core.MCPOAuth{Issuer: f.server.URL, ClientID: "registered-client", Grant: grant}}}
}
func TestOAuthAcquisitionAndRedaction(t *testing.T) {
	for _, grant := range []string{"client_credentials", "refresh_token"} {
		t.Run(grant, func(t *testing.T) {
			f := newOAuthFixture(t)
			target := f.target(grant)
			if grant == "client_credentials" {
				target.MCP.OAuth.ClientSecretEnv = "OAUTH_TEST_SECRET"
				t.Setenv("OAUTH_TEST_SECRET", "client-secret-value")
			} else {
				target.MCP.OAuth.RefreshTokenEnv = "OAUTH_TEST_REFRESH"
				t.Setenv("OAUTH_TEST_REFRESH", "refresh-secret-value")
			}
			result := runTarget(t, target)
			if result.Status != core.Healthy || result.Target.Name != "[REDACTED]" || f.tokens.Load() != 1 {
				t.Fatalf("%+v token requests=%d", result, f.tokens.Load())
			}
		})
	}
}
func TestOAuthFilesAndRefresh(t *testing.T) {
	f := newOAuthFixture(t)
	target := f.target("authorization_code")
	target.MCP.OAuth.TokenFile = filepath.Join(t.TempDir(), "credentials", "token.json")
	old := oauthToken{AccessToken: "old-token", RefreshToken: "refresh-secret-value", ExpiresAt: time.Now().Add(-time.Hour), Issuer: f.server.URL, Resource: target.Endpoint, ClientID: target.MCP.OAuth.ClientID}
	if err := writeToken(target.MCP.OAuth.TokenFile, old); err != nil {
		t.Fatal(err)
	}
	result := runTarget(t, target)
	if result.Status != core.Healthy || f.tokens.Load() != 1 {
		t.Fatalf("%+v requests=%d", result, f.tokens.Load())
	}
	saved, err := readToken(target.MCP.OAuth.TokenFile)
	if err != nil || saved.RefreshToken != "rotated-sensitive-value" || saved.AccessToken != "acquired-sensitive-value" {
		t.Fatalf("failed rotated token storage: %v", err)
	}
	result = runTarget(t, target)
	if result.Status != core.Healthy || f.tokens.Load() != 1 {
		t.Fatalf("cached token was reacquired: %+v", result)
	}
	if runtime.GOOS != "windows" {
		os.Chmod(target.MCP.OAuth.TokenFile, 0644)
		if _, err := readToken(target.MCP.OAuth.TokenFile); err == nil {
			t.Fatal("insecure token permissions accepted")
		}
		os.Chmod(target.MCP.OAuth.TokenFile, 0600)
	}
	link := filepath.Join(t.TempDir(), "symlink.json")
	if err := os.Symlink(target.MCP.OAuth.TokenFile, link); err != nil {
		t.Fatal(err)
	}
	if _, err := readToken(link); err == nil {
		t.Fatal("symlink token accepted")
	}
	if err := writeToken(link, old); err == nil {
		t.Fatal("symlink token overwritten")
	}
}
func TestOAuthFailures(t *testing.T) {
	for _, mode := range []string{"issuer", "redirect", "missing-login", "bound-token"} {
		t.Run(mode, func(t *testing.T) {
			f := newOAuthFixture(t)
			target := f.target("client_credentials")
			target.MCP.OAuth.ClientSecretEnv = "OAUTH_TEST_SECRET"
			t.Setenv("OAUTH_TEST_SECRET", "client-secret-value")
			switch mode {
			case "issuer":
				f.invalidIssuer = true
			case "redirect":
				f.redirectToken = true
			case "missing-login":
				target.MCP.OAuth.Grant = "authorization_code"
				target.MCP.OAuth.TokenFile = filepath.Join(t.TempDir(), "missing.json")
			case "bound-token":
				target.MCP.OAuth.TokenFile = filepath.Join(t.TempDir(), "credentials", "token.json")
				if err := writeToken(target.MCP.OAuth.TokenFile, oauthToken{AccessToken: "bad", Issuer: "https://other.example", ClientID: target.MCP.OAuth.ClientID, Resource: target.Endpoint, ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
					t.Fatal(err)
				}
			}
			result := runTarget(t, target)
			if result.Status != core.Misconfigured {
				t.Fatalf("%+v", result)
			}
		})
	}
}

type loginWriter struct {
	f            *oauthFixture
	mode         string
	buffer       bytes.Buffer
	callbackDone chan struct{}
}

func (w *loginWriter) Write(data []byte) (int, error) {
	w.buffer.Write(data)
	text := string(data)
	if strings.Contains(text, "Open this URL") {
		lines := strings.Split(strings.TrimSpace(text), "\n")
		authorization, err := url.Parse(lines[len(lines)-1])
		if err != nil {
			return 0, err
		}
		q := authorization.Query()
		if q.Get("code_challenge_method") != "S256" || q.Get("resource") != w.f.server.URL+"/mcp" {
			w.f.t.Error("missing authorization resource or PKCE")
		}
		w.f.mu.Lock()
		w.f.challenge = q.Get("code_challenge")
		w.f.mu.Unlock()
		callback, _ := url.Parse(q.Get("redirect_uri"))
		values := url.Values{"code": {"authorization-code"}, "state": {q.Get("state")}, "iss": {w.f.server.URL}}
		if w.mode == "issuer" {
			values.Set("iss", "https://wrong.example")
		}
		if w.mode == "missing-issuer" {
			values.Del("iss")
		}
		go func() {
			defer close(w.callbackDone)
			if w.mode == "bad-state-first" {
				bad := url.Values{"state": {"wrong"}, "code": {"authorization-code"}}
				callback.RawQuery = bad.Encode()
				response, err := http.Get(callback.String())
				if err == nil {
					if response.StatusCode != 400 {
						w.f.t.Error("wrong state accepted")
					}
					response.Body.Close()
				}
			}
			callback.RawQuery = values.Encode()
			response, err := http.Get(callback.String())
			if err != nil {
				w.f.t.Error(err)
				return
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				w.f.t.Error(err)
				return
			}
			wantStatus := http.StatusOK
			if w.mode == "issuer" || w.mode == "missing-issuer" {
				wantStatus = http.StatusBadRequest
			}
			if response.StatusCode != wantStatus || len(body) == 0 {
				w.f.t.Errorf("callback status=%d body length=%d", response.StatusCode, len(body))
			}
		}()
	}
	return len(data), nil
}
func TestOAuthPKCELogin(t *testing.T) {
	for _, mode := range []string{"success", "bad-state-first", "issuer", "missing-issuer"} {
		t.Run(mode, func(t *testing.T) {
			f := newOAuthFixture(t)
			target := f.target("authorization_code")
			target.MCP.OAuth.TokenFile = filepath.Join(t.TempDir(), "credentials", "token.json")
			writer := &loginWriter{f: f, mode: mode, callbackDone: make(chan struct{})}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := Login(ctx, target, writer)
			select {
			case <-writer.callbackDone:
			case <-ctx.Done():
				t.Fatal("browser callback did not finish")
			}
			if mode == "issuer" || mode == "missing-issuer" {
				if err == nil || f.tokens.Load() != 0 {
					t.Fatal("invalid issuer accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(writer.buffer.String(), "acquired-sensitive-value") || strings.Contains(writer.buffer.String(), "rotated-sensitive-value") {
				t.Fatal("login printed tokens")
			}
			file, err := os.Open(target.MCP.OAuth.TokenFile)
			if err != nil {
				t.Fatal(err)
			}
			private := privateTokenFile(file)
			file.Close()
			if !private {
				t.Fatal("token storage permissions")
			}
			result := runTarget(t, target)
			if result.Status != core.Healthy || f.tokens.Load() != 1 {
				t.Fatalf("%+v", result)
			}
		})
	}
}

func TestOAuthCacheIsolation(t *testing.T) {
	f := newOAuthFixture(t)
	target := f.target("client_credentials")
	target.MCP.OAuth.ClientSecretEnv = "OAUTH_TEST_SECRET"
	t.Setenv("OAUTH_TEST_SECRET", "client-secret-value")
	registry := core.NewRegistry()
	registry.Register(Adapter{})
	engine := core.NewEngine(registry)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results, err := engine.Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target, target}})
			if err != nil || len(results) != 2 || results[0].Status != core.Healthy || results[1].Status != core.Healthy {
				t.Errorf("isolated run failed: %v", err)
			}
		}()
	}
	wg.Wait()
	if f.tokens.Load() != 4 {
		t.Fatalf("token cache shared across targets/runs: %d", f.tokens.Load())
	}
}

func TestOAuthReachabilityDoesNotLogin(t *testing.T) {
	f := newOAuthFixture(t)
	target := f.target("authorization_code")
	target.MCP.OAuth.TokenFile = filepath.Join(t.TempDir(), "missing.json")
	target.Checks = []string{"reachability"}
	result := runTarget(t, target)
	if result.Status != core.Healthy || f.tokens.Load() != 0 {
		t.Fatalf("connectivity-only check attempted OAuth: %+v", result)
	}
}

func TestSharedSessionRetainsOAuthCredential(t *testing.T) {
	f := newOAuthFixture(t)
	f.inventory = true
	target := f.target("client_credentials")
	target.MCP.RequiredTools = []string{"health"}
	target.MCP.OAuth.ClientSecretEnv = "OAUTH_TEST_SECRET"
	t.Setenv("OAUTH_TEST_SECRET", "client-secret-value")
	result := runTarget(t, target)
	if result.Status != core.Healthy || f.tokens.Load() != 1 {
		t.Fatalf("%+v token requests=%d", result, f.tokens.Load())
	}
}

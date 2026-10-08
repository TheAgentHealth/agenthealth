package model

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
)

func run(t *testing.T, target core.Target) core.Result {
	t.Helper()
	reg := core.NewRegistry()
	if err := reg.Register(Adapter{}); err != nil {
		t.Fatal(err)
	}
	results, err := core.NewEngine(reg).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	return results[0]
}

func models(ids ...string) map[string]any {
	data := []any{}
	for _, id := range ids {
		data = append(data, map[string]any{"id": id, "object": "model"})
	}
	return map[string]any{"object": "list", "data": data}
}

func TestOpenAIPassiveChecksShareOneListing(t *testing.T) {
	t.Setenv("MODEL_TOKEN", "super-secret")
	var lists atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models" || r.URL.Query().Get("api-version") != "1" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL)
		}
		if r.Header.Get("Authorization") != "Bearer super-secret" || r.Header.Get("x-api-key") != "" {
			t.Error("OpenAI-compatible APIs use bearer authentication")
		}
		lists.Add(1)
		json.NewEncoder(w).Encode(models("gpt-test", "embed-test"))
	}))
	defer server.Close()
	result := run(t, core.Target{Name: "llm", Type: "llm", Endpoint: server.URL + "/v1/?api-version=1", Auth: &core.AuthReference{BearerEnv: "MODEL_TOKEN"},
		Model: &core.ModelOptions{RequiredModels: []string{"gpt-test"}}})
	if result.Status != core.Healthy || lists.Load() != 1 {
		t.Fatalf("status %s after %d listings: %+v", result.Status, lists.Load(), result.Checks)
	}
	for _, dimension := range []string{"configuration", "reachability", "authentication", "protocol", "capability", "latency"} {
		if result.Checks[dimension].Status != core.Healthy {
			t.Fatalf("%s: %+v", dimension, result.Checks[dimension])
		}
	}
	if result.Checks["reachability"].Steps["http"] != core.Healthy {
		t.Fatalf("missing transport steps: %+v", result.Checks["reachability"].Steps)
	}
}

func TestListingClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   core.Status
		code   string
	}{
		{"missing model", 200, `{"data":[{"id":"other"}]}`, core.Unhealthy, "model_required"},
		{"rejected credential", 401, "", core.Misconfigured, "model_auth"},
		{"forbidden", 403, "", core.Misconfigured, "model_auth"},
		{"rate limited", 429, "", core.Degraded, "model_rate_limit"},
		{"overloaded", 529, "", core.Degraded, "model_overloaded"},
		{"server error", 503, "", core.Unhealthy, "model_http"},
		{"wrong base path", 404, "", core.Unhealthy, "model_http"},
		{"not JSON", 200, "<html>", core.Unhealthy, "model_protocol"},
		{"no data", 200, `{"object":"list"}`, core.Unhealthy, "model_protocol"},
		{"blank id", 200, `{"data":[{"id":" "}]}`, core.Unhealthy, "model_protocol"},
		{"oversized", 200, `{"data":[` + strings.Repeat(" ", listLimit) + `]}`, core.Unknown, "model_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			result := run(t, core.Target{Name: "m", Type: "model", Endpoint: server.URL, Model: &core.ModelOptions{RequiredModels: []string{"needed"}}})
			if result.Status != tc.want {
				t.Fatalf("status %s, want %s: %+v", result.Status, tc.want, result.Checks)
			}
			found := false
			for _, check := range result.Checks {
				found = found || check.Code == tc.code
			}
			if !found {
				t.Fatalf("missing code %s: %+v", tc.code, result.Checks)
			}
		})
	}
}

func TestOpenAIMinimalInference(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(models("gpt-test"))
			return
		}
		posts.Add(1)
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/v1/chat/completions" || body["model"] != "gpt-test" || body["max_tokens"] != float64(4) || body["max_completion_tokens"] != nil || body["stream"] != false {
			t.Errorf("unexpected inference request %s %v", r.URL.Path, body)
		}
		if messages := body["messages"].([]any); len(messages) != 1 || messages[0].(map[string]any)["content"] != "Reply with OK." {
			t.Errorf("unexpected messages %v", messages)
		}
		json.NewEncoder(w).Encode(map[string]any{"object": "chat.completion", "choices": []any{map[string]any{"index": 0, "message": map[string]any{"role": "assistant", "content": "OK"}, "finish_reason": "length"}}})
	}))
	defer server.Close()
	tokens := 4
	result := run(t, core.Target{Name: "m", Type: "model", Endpoint: server.URL + "/v1", Checks: []string{"reachability", "functional"},
		Model: &core.ModelOptions{Functional: &core.ModelInference{Safe: true, Model: "gpt-test", Prompt: "Reply with OK.", MaxOutputTokens: &tokens}}})
	if result.Status != core.Healthy || result.Checks["functional"].Status != core.Healthy || posts.Load() != 1 {
		t.Fatalf("status %s after %d inferences: %+v", result.Status, posts.Load(), result.Checks)
	}
}

func TestInferenceFailuresAreClassifiedAndNeverRetried(t *testing.T) {
	retries := 3
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   core.Status
		code   string
	}{
		{"unknown model", 404, "", core.Unhealthy, "model_unavailable"},
		{"server error", 500, "", core.Unhealthy, "model_functional"},
		{"rate limited", 429, "", core.Degraded, "model_rate_limit"},
		{"rejected credential", 401, "", core.Misconfigured, "model_auth"},
		{"no choices", 200, `{"choices":[]}`, core.Unhealthy, "model_protocol"},
		{"embedded error", 200, `{"error":{"message":"quota"}}`, core.Unhealthy, "model_protocol"},
		{"limit ignored", 200, `{"choices":[{"message":{}}],"usage":{"completion_tokens":415}}`, core.Degraded, "model_token_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var posts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					json.NewEncoder(w).Encode(models("gpt-test"))
					return
				}
				posts.Add(1)
				w.WriteHeader(tc.status)
				w.Write([]byte(tc.body))
			}))
			defer server.Close()
			result := run(t, core.Target{Name: "m", Type: "model", Endpoint: server.URL, Retries: &retries, Checks: []string{"functional"},
				Model: &core.ModelOptions{Functional: &core.ModelInference{Safe: true, Model: "gpt-test", Prompt: "Reply with OK."}}})
			check := result.Checks["functional"]
			if check.Status != tc.want || check.Code != tc.code || posts.Load() != 1 {
				t.Fatalf("got %+v after %d inferences", check, posts.Load())
			}
		})
	}
}

func TestAnthropicHeadersPathsAndPagination(t *testing.T) {
	t.Setenv("ANTHROPIC_KEY", "super-secret")
	hasMore := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "super-secret" || r.Header.Get("anthropic-version") != anthropicVersion || r.Header.Get("Authorization") != "" {
			t.Error("Anthropic requests use x-api-key and anthropic-version")
		}
		if r.Method == http.MethodGet {
			if r.URL.Path != "/v1/models" || r.URL.Query().Get("limit") != "1000" {
				t.Errorf("unexpected listing %s", r.URL)
			}
			json.NewEncoder(w).Encode(map[string]any{"data": []any{map[string]any{"id": "claude-test", "type": "model"}}, "has_more": hasMore})
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path != "/v1/messages" || body["max_tokens"] != float64(defaultOutputTokens) || body["model"] != "claude-test" {
			t.Errorf("unexpected inference request %s %v", r.URL.Path, body)
		}
		json.NewEncoder(w).Encode(map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "text", "text": "OK"}}})
	}))
	defer server.Close()
	target := core.Target{Name: "claude", Type: "llm", Endpoint: server.URL + "/v1", Auth: &core.AuthReference{BearerEnv: "ANTHROPIC_KEY"},
		Checks: []string{"authentication", "capability", "functional"},
		Model: &core.ModelOptions{API: "anthropic", RequiredModels: []string{"claude-test"},
			Functional: &core.ModelInference{Safe: true, Model: "claude-test", Prompt: "Reply with OK."}}}
	if result := run(t, target); result.Status != core.Healthy {
		t.Fatalf("status %s: %+v", result.Status, result.Checks)
	}
	// A required model beyond an incomplete listing is inconclusive, not missing.
	hasMore = true
	target.Model.RequiredModels = []string{"claude-other"}
	target.Checks = []string{"capability"}
	target.Model.Functional = nil
	if result := run(t, target); result.Checks["capability"].Code != "model_limit" || result.Status != core.Unknown {
		t.Fatalf("got %+v", result.Checks)
	}
}

func TestCredentialsAndResponseBodiesNeverReachOutput(t *testing.T) {
	t.Setenv("MODEL_TOKEN", "super-secret")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"bad key super-secret for account acct-12345"}`))
	}))
	defer server.Close()
	result := run(t, core.Target{Name: "m", Type: "model", Endpoint: server.URL, Auth: &core.AuthReference{BearerEnv: "MODEL_TOKEN"}})
	var out bytes.Buffer
	if err := core.WriteJSON(&out, []core.Result{result}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out.String(), "super-secret") || strings.Contains(out.String(), "acct-12345") {
		t.Fatalf("leaked response or credential: %s", out.String())
	}
}

func TestConfigurationRejectsInvalidEndpointsAndCredentials(t *testing.T) {
	t.Setenv("MODEL_TOKEN", "bad\nvalue")
	for _, target := range []core.Target{
		{Name: "m", Type: "model", Endpoint: "ftp://models.example.com"},
		{Name: "m", Type: "model", Endpoint: "https://user:pass@models.example.com/v1"},
		{Name: "m", Type: "model", Endpoint: "https://models.example.com/v1", Auth: &core.AuthReference{BearerEnv: "MODEL_TOKEN"}},
	} {
		if result := run(t, target); result.Status != core.Misconfigured {
			t.Fatalf("%s: got %s", target.Endpoint, result.Status)
		}
	}
}

func TestFunctionalRequiresExplicitOptIn(t *testing.T) {
	for _, target := range []core.Target{
		{Name: "m", Type: "model", Endpoint: "https://m.example.com", Checks: []string{"functional"}},
		{Name: "m", Type: "model", Endpoint: "https://m.example.com", Model: &core.ModelOptions{Functional: &core.ModelInference{Safe: true, Model: "x", Prompt: "OK"}}},
		{Name: "m", Type: "model", Endpoint: "https://m.example.com", Checks: []string{"functional"}, Model: &core.ModelOptions{Functional: &core.ModelInference{Model: "x", Prompt: "OK"}}},
		{Name: "m", Type: "http", Endpoint: "https://m.example.com", Model: &core.ModelOptions{}},
		{Name: "m", Type: "model", Endpoint: "https://m.example.com", Model: &core.ModelOptions{API: "gemini"}},
	} {
		if err := (core.Config{Version: "v1", Targets: []core.Target{target}}).Validate(); err == nil {
			t.Fatalf("accepted %+v", target)
		}
	}
}

func TestTokenParameterSelection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(models("o-test"))
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["max_completion_tokens"] != float64(defaultOutputTokens) || body["max_tokens"] != nil {
			t.Errorf("unexpected token parameter %v", body)
		}
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{}}, "usage": map[string]any{"completion_tokens": defaultOutputTokens}})
	}))
	defer server.Close()
	result := run(t, core.Target{Name: "m", Type: "model", Endpoint: server.URL, Checks: []string{"functional"},
		Model: &core.ModelOptions{Functional: &core.ModelInference{Safe: true, Model: "o-test", Prompt: "OK", TokenParameter: "max_completion_tokens"}}})
	if result.Status != core.Healthy {
		t.Fatalf("got %+v", result.Checks)
	}
	anthropic := core.Target{Name: "m", Type: "model", Endpoint: server.URL, Checks: []string{"functional"},
		Model: &core.ModelOptions{API: "anthropic", Functional: &core.ModelInference{Safe: true, Model: "c", Prompt: "OK", TokenParameter: "max_completion_tokens"}}}
	if err := (core.Config{Version: "v1", Targets: []core.Target{anthropic}}).Validate(); err == nil {
		t.Fatal("accepted max_completion_tokens for anthropic")
	}
}

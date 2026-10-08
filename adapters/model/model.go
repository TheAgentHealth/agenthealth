// Package model implements passive model API health and an opt-in, token-bounded
// minimal inference for OpenAI-compatible and Anthropic inference endpoints.
package model

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

const (
	listLimit           = 4 << 20
	inferenceLimit      = 1 << 20
	defaultOutputTokens = 16
	anthropicVersion    = "2023-06-01"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "model", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"model", "llm"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}

func api(t core.Target) string {
	if t.Model != nil && t.Model.API != "" {
		return t.Model.API
	}
	return "openai"
}

// base accepts an absolute HTTP(S) API base such as https://api.openai.com/v1.
func base(t core.Target) (*url.URL, bool) {
	u, err := url.Parse(t.Endpoint)
	return u, err == nil && u.Hostname() != "" && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.Fragment == "" && u.Opaque == ""
}

// resolve appends an API path to the base, preserving any configured query.
func resolve(t core.Target, path string, extra url.Values) string {
	u, _ := base(t)
	u.Path = strings.TrimSuffix(u.Path, "/") + path
	u.RawPath = ""
	query := u.Query()
	for key, values := range extra {
		query[key] = values
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func observed(status core.Status, code string) core.Observation {
	return core.Observation{Check: core.CheckResult{Status: status}, Code: code, ResponseReceived: true}
}

type fetched struct {
	status int
	body   []byte
}

func send(ctx context.Context, r core.Request, method, endpoint string, body []byte, limit int, trace bool) (fetched, core.Observation, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return fetched{}, core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if api(r.Target) == "anthropic" {
		req.Header.Set("anthropic-version", anthropicVersion)
		if r.Credential != "" {
			req.Header.Set("x-api-key", r.Credential)
		}
	} else if r.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+r.Credential)
	}
	hooks := &httptrace.ClientTrace{GotFirstResponseByte: func() {
		if r.MarkResponse != nil {
			r.MarkResponse()
		}
	}}
	if trace && r.RecordStep != nil {
		step := func(name string, err error) {
			status := core.Healthy
			if err != nil {
				status = core.Unreachable
			}
			r.RecordStep(name, status)
		}
		hooks.DNSStart = func(httptrace.DNSStartInfo) { r.RecordStep("dns", core.Unknown) }
		hooks.DNSDone = func(info httptrace.DNSDoneInfo) { step("dns", info.Err) }
		hooks.ConnectStart = func(string, string) { r.RecordStep("tcp", core.Unknown) }
		hooks.ConnectDone = func(_, _ string, err error) { step("tcp", err) }
		hooks.TLSHandshakeStart = func() { r.RecordStep("tls", core.Unknown) }
		hooks.TLSHandshakeDone = func(_ tls.ConnectionState, err error) { step("tls", err) }
		hooks.GotConn = func(httptrace.GotConnInfo) { r.RecordStep("http", core.Unknown) }
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, hooks))
	resp, err := r.Client.Do(req)
	if err != nil {
		return fetched{}, core.Observation{}, err
	}
	defer resp.Body.Close()
	if r.MarkResponse != nil {
		r.MarkResponse()
	}
	if r.RecordStep != nil {
		r.RecordStep("http", core.Healthy)
	}
	f := fetched{status: resp.StatusCode}
	// Error bodies may echo request content or account details; never read them.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return f, observed(core.Healthy, ""), nil
	}
	f.body, err = io.ReadAll(io.LimitReader(resp.Body, int64(limit)+1))
	if err != nil {
		return fetched{}, core.Observation{ResponseReceived: true}, err
	}
	if len(f.body) > limit {
		return fetched{}, observed(core.Unknown, "model_limit"), nil
	}
	return f, observed(core.Healthy, ""), nil
}

// failure classifies non-success statuses shared by listing and inference.
func failure(status int) (core.Observation, bool) {
	switch {
	case status >= 200 && status < 300:
		return core.Observation{}, false
	case status == 401 || status == 403:
		return observed(core.Misconfigured, "model_auth"), true
	case status == 429:
		return observed(core.Degraded, "model_rate_limit"), true
	case status == 529:
		return observed(core.Degraded, "model_overloaded"), true
	default:
		return observed(core.Unhealthy, "model_http"), true
	}
}

type listing struct {
	Data *[]struct {
		ID string `json:"id"`
	} `json:"data"`
	HasMore bool `json:"has_more"`
}

func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	if dimension == "configuration" {
		if _, ok := base(r.Target); !ok || !validHeaderValue(r.Credential) {
			return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
		}
		return core.Observation{Check: core.CheckResult{Status: core.Healthy}}, nil
	}
	if _, ok := base(r.Target); !ok {
		return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
	}
	if dimension == "functional" {
		return infer(ctx, r)
	}
	// Passive dimensions share one model listing per target run.
	var f fetched
	if cached, exists := r.RunState.Load("model-list"); exists {
		f = cached.(fetched)
	} else {
		var page url.Values
		if api(r.Target) == "anthropic" {
			page = url.Values{"limit": {"1000"}}
		}
		var obs core.Observation
		var err error
		f, obs, err = send(ctx, r, http.MethodGet, resolve(r.Target, "/models", page), nil, listLimit, dimension == "reachability")
		if err != nil || obs.Check.Status != core.Healthy {
			return obs, err
		}
		r.RunState.Store("model-list", f)
	}
	switch dimension {
	case "reachability":
		return observed(core.Healthy, ""), nil
	case "authentication":
		if f.status == 401 || f.status == 403 {
			return observed(core.Misconfigured, "model_auth"), nil
		}
		return observed(core.Healthy, ""), nil
	}
	if obs, failed := failure(f.status); failed {
		return obs, nil
	}
	var models listing
	if json.Unmarshal(f.body, &models) != nil || models.Data == nil {
		return observed(core.Unhealthy, "model_protocol"), nil
	}
	available := map[string]bool{}
	for _, m := range *models.Data {
		if strings.TrimSpace(m.ID) == "" {
			return observed(core.Unhealthy, "model_protocol"), nil
		}
		available[m.ID] = true
	}
	if dimension == "capability" && r.Target.Model != nil {
		for _, name := range r.Target.Model.RequiredModels {
			if !available[name] {
				if models.HasMore {
					return observed(core.Unknown, "model_limit"), nil
				}
				return observed(core.Unhealthy, "model_required"), nil
			}
		}
	}
	return observed(core.Healthy, ""), nil
}

// infer sends one minimal, non-streaming, token-bounded request. Active checks
// are never retried by the engine, so it runs at most once per target run.
func infer(ctx context.Context, r core.Request) (core.Observation, error) {
	if r.Target.Model == nil || r.Target.Model.Functional == nil {
		return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
	}
	f := r.Target.Model.Functional
	tokens := defaultOutputTokens
	if f.MaxOutputTokens != nil {
		tokens = *f.MaxOutputTokens
	}
	messages := []map[string]string{{"role": "user", "content": f.Prompt}}
	// max_tokens is the default: servers that reject it fail visibly, while some
	// compatible servers silently ignore max_completion_tokens.
	parameter := "max_tokens"
	if f.TokenParameter != "" {
		parameter = f.TokenParameter
	}
	path, request := "/chat/completions", map[string]any{"model": f.Model, "messages": messages, parameter: tokens, "stream": false}
	if api(r.Target) == "anthropic" {
		path, request = "/messages", map[string]any{"model": f.Model, "messages": messages, "max_tokens": tokens}
	}
	body, err := json.Marshal(request)
	if err != nil {
		return core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	response, obs, err := send(ctx, r, http.MethodPost, resolve(r.Target, path, nil), body, inferenceLimit, false)
	if err != nil || obs.Check.Status != core.Healthy {
		return obs, err
	}
	if response.status == 404 {
		return observed(core.Unhealthy, "model_unavailable"), nil
	}
	if obs, failed := failure(response.status); failed {
		if obs.Check.Status == core.Unhealthy {
			obs.Code = "model_functional"
		}
		return obs, nil
	}
	var completion struct {
		Type    string            `json:"type"`
		Choices []json.RawMessage `json:"choices"`
		Content []json.RawMessage `json:"content"`
		Usage   struct {
			CompletionTokens *int `json:"completion_tokens"`
			OutputTokens     *int `json:"output_tokens"`
		} `json:"usage"`
	}
	if json.Unmarshal(response.body, &completion) != nil {
		return observed(core.Unhealthy, "model_protocol"), nil
	}
	if api(r.Target) == "anthropic" {
		if completion.Type != "message" || completion.Content == nil {
			return observed(core.Unhealthy, "model_protocol"), nil
		}
	} else if len(completion.Choices) == 0 {
		return observed(core.Unhealthy, "model_protocol"), nil
	}
	// Reported usage above the bound means the limit is not being enforced.
	for _, used := range []*int{completion.Usage.CompletionTokens, completion.Usage.OutputTokens} {
		if used != nil && *used > tokens {
			return observed(core.Degraded, "model_token_limit"), nil
		}
	}
	return observed(core.Healthy, ""), nil
}

// Match net/http header validation so invalid credentials are MISCONFIGURED.
func validHeaderValue(value string) bool {
	for i := 0; i < len(value); i++ {
		if b := value[i]; (b < 0x20 && b != '\t') || b == 0x7f {
			return false
		}
	}
	return true
}

// Package a2a implements bounded A2A 0.3.0 JSON-RPC health validation.
package a2a

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

const bodyLimit = 1048576

type Adapter struct {
	// random is normally nil, selecting the OS-backed source. Tests can inject
	// a failing reader without mutating the process-wide cryptographic source.
	random io.Reader
}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "a2a", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"a2a"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}

type skill struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description *string  `json:"description"`
	Tags        []string `json:"tags"`
}
type card struct {
	ProtocolVersion    string                     `json:"protocolVersion"`
	Name               string                     `json:"name"`
	Description        *string                    `json:"description"`
	Version            string                     `json:"version"`
	URL                string                     `json:"url"`
	PreferredTransport string                     `json:"preferredTransport"`
	Capabilities       map[string]json.RawMessage `json:"capabilities"`
	Skills             []skill                    `json:"skills"`
	InputModes         []string                   `json:"defaultInputModes"`
	OutputModes        []string                   `json:"defaultOutputModes"`
	Security           []map[string][]string      `json:"security"`
	SecuritySchemes    map[string]struct {
		Type   string `json:"type"`
		Scheme string `json:"scheme"`
	} `json:"securitySchemes"`
}

func observed(status core.Status, code string) core.Observation {
	return core.Observation{Check: core.CheckResult{Status: status}, Code: code, ResponseReceived: true}
}
func validURL(raw string) (*url.URL, bool) {
	u, err := url.Parse(raw)
	return u, err == nil && u.Hostname() != "" && (u.Scheme == "http" || u.Scheme == "https") && u.User == nil && u.Fragment == "" && u.Opaque == ""
}
func sameOrigin(a, b *url.URL) bool {
	port := func(u *url.URL) string {
		if u.Port() != "" {
			return u.Port()
		}
		if u.Scheme == "https" {
			return "443"
		}
		return "80"
	}
	return a.Scheme == b.Scheme && strings.EqualFold(a.Hostname(), b.Hostname()) && port(a) == port(b)
}
func discovery(t core.Target) (string, bool) {
	base, ok := validURL(t.Endpoint)
	if !ok {
		return "", false
	}
	if t.A2A != nil && t.A2A.CardURL != "" {
		u, ok := validURL(t.A2A.CardURL)
		if !ok || !sameOrigin(base, u) {
			return "", false
		}
		return u.String(), true
	}
	base.Path = "/.well-known/agent-card.json"
	base.RawPath = ""
	base.RawQuery = ""
	return base.String(), true
}

type fetched struct {
	status int
	body   []byte
}

func fetch(ctx context.Context, r core.Request, method, endpoint string, body []byte) (fetched, core.Observation, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		return fetched{}, core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+r.Credential)
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{GotFirstResponseByte: func() {
		if r.MarkResponse != nil {
			r.MarkResponse()
		}
	}}))
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
	// No need to read authentication errors or unexpected HTTP responses.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return f, observed(core.Healthy, ""), nil
	}
	f.body, err = io.ReadAll(io.LimitReader(resp.Body, bodyLimit+1))
	if err != nil {
		return fetched{}, core.Observation{ResponseReceived: true}, err
	}
	if len(f.body) > bodyLimit {
		return fetched{}, observed(core.Unknown, "a2a_limit"), nil
	}
	return f, observed(core.Healthy, ""), nil
}
func (a Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	endpoint, ok := discovery(r.Target)
	if dimension == "configuration" {
		if !ok || strings.ContainsAny(r.Credential, "\r\n") {
			return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
		}
		for i := 0; i < len(r.Credential); i++ {
			if r.Credential[i] < 32 && r.Credential[i] != '\t' || r.Credential[i] == 127 {
				return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
			}
		}
		return core.Observation{Check: core.CheckResult{Status: core.Healthy}}, nil
	}
	if !ok {
		return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
	}
	var f fetched
	cached, exists := r.RunState.Load("a2a-card")
	if exists {
		f = cached.(fetched)
	} else {
		var obs core.Observation
		var err error
		f, obs, err = fetch(ctx, r, http.MethodGet, endpoint, nil)
		if err != nil || obs.Check.Status != core.Healthy {
			return obs, err
		}
		r.RunState.Store("a2a-card", f)
	}
	if dimension == "reachability" {
		return observed(core.Healthy, ""), nil
	}
	if f.status == 401 || f.status == 403 {
		return observed(core.Misconfigured, "a2a_auth"), nil
	}
	if f.status < 200 || f.status >= 300 {
		return observed(core.Unhealthy, "a2a_http"), nil
	}
	var c card
	if json.Unmarshal(f.body, &c) != nil || strings.TrimSpace(c.Name) == "" || strings.TrimSpace(c.Version) == "" || c.Description == nil || c.Capabilities == nil || c.Skills == nil || len(c.InputModes) == 0 || len(c.OutputModes) == 0 {
		return observed(core.Unhealthy, "a2a_card"), nil
	}
	if c.ProtocolVersion != "0.3.0" || c.PreferredTransport != "" && c.PreferredTransport != "JSONRPC" {
		return observed(core.Misconfigured, "a2a_version"), nil
	}
	rpc, valid := validURL(c.URL)
	base, _ := validURL(r.Target.Endpoint)
	if !valid {
		return observed(core.Unhealthy, "a2a_card"), nil
	}
	if !sameOrigin(base, rpc) {
		return observed(core.Misconfigured, "a2a_origin"), nil
	}
	seen := map[string]bool{}
	for _, s := range c.Skills {
		if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.Name) == "" || s.Description == nil || s.Tags == nil || seen[s.ID] {
			return observed(core.Unhealthy, "a2a_card"), nil
		}
		seen[s.ID] = true
	}
	for _, name := range []string{"streaming", "pushNotifications", "stateTransitionHistory"} {
		if raw, exists := c.Capabilities[name]; exists {
			var b bool
			if string(raw) == "null" || json.Unmarshal(raw, &b) != nil {
				return observed(core.Unhealthy, "a2a_card"), nil
			}
		}
	}
	for _, modes := range [][]string{c.InputModes, c.OutputModes} {
		for i, mode := range modes {
			mediaType, _, err := mime.ParseMediaType(mode)
			// ParseMediaType also accepts disposition tokens such as "inline";
			// agent modes must contain both a MIME type and subtype.
			if err != nil || !strings.Contains(mediaType, "/") {
				return observed(core.Unhealthy, "a2a_card"), nil
			}
			modes[i] = mediaType
		}
	}
	if raw, exists := c.Capabilities["extensions"]; exists {
		var extensions []struct {
			URI      string `json:"uri"`
			Required bool   `json:"required"`
		}
		if string(raw) == "null" || json.Unmarshal(raw, &extensions) != nil {
			return observed(core.Unhealthy, "a2a_card"), nil
		}
		for _, extension := range extensions {
			if strings.TrimSpace(extension.URI) == "" {
				return observed(core.Unhealthy, "a2a_card"), nil
			}
			if extension.Required {
				return observed(core.Misconfigured, "a2a_version"), nil
			}
		}
	}
	// Security alternatives are OR; schemes within each alternative are AND.
	supported := len(c.Security) == 0
	for _, alternative := range c.Security {
		accepted := true
		for name := range alternative {
			s, exists := c.SecuritySchemes[name]
			if !exists || s.Type != "http" || !strings.EqualFold(s.Scheme, "bearer") || r.Credential == "" {
				accepted = false
			}
		}
		supported = supported || accepted
	}
	if !supported {
		return observed(core.Misconfigured, "a2a_auth"), nil
	}
	if dimension == "capability" {
		if o := r.Target.A2A; o != nil {
			for _, name := range o.RequiredSkills {
				if !seen[name] {
					return observed(core.Degraded, "a2a_required"), nil
				}
			}
			for _, name := range o.RequiredCapabilities {
				var enabled bool
				if json.Unmarshal(c.Capabilities[name], &enabled) != nil || !enabled {
					return observed(core.Degraded, "a2a_required"), nil
				}
			}
		}
		return observed(core.Healthy, ""), nil
	}
	interactionID, err := a.identifier()
	if err != nil {
		return core.Observation{ResponseReceived: true}, err
	}
	method := "tasks/get"
	params := map[string]any{"id": interactionID}
	if dimension == "functional" {
		if r.Target.A2A == nil || r.Target.A2A.Functional == nil || !r.Target.A2A.Functional.Safe {
			return observed(core.Misconfigured, "a2a_functional"), nil
		}
		if !has(c.InputModes, "text/plain") || !has(c.OutputModes, "text/plain") {
			return observed(core.Misconfigured, "a2a_functional"), nil
		}
		method = "message/send"
		params = map[string]any{"message": map[string]any{"kind": "message", "role": "user", "messageId": interactionID, "parts": []any{map[string]string{"kind": "text", "text": r.Target.A2A.Functional.Text}}}, "configuration": map[string]any{"blocking": true, "acceptedOutputModes": []string{"text/plain"}, "historyLength": 0}}
	}
	id, err := a.identifier()
	if err != nil {
		return core.Observation{ResponseReceived: true}, err
	}
	body, _ := json.Marshal(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": params})
	f, obs, err := fetch(ctx, r, http.MethodPost, c.URL, body)
	if err != nil || obs.Check.Status != core.Healthy {
		return obs, err
	}
	if f.status == 401 || f.status == 403 {
		return observed(core.Misconfigured, "a2a_auth"), nil
	}
	if f.status < 200 || f.status >= 300 {
		return observed(core.Unhealthy, "a2a_http"), nil
	}
	var response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      string          `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   json.RawMessage `json:"error"`
	}
	if json.Unmarshal(f.body, &response) != nil || response.JSONRPC != "2.0" || response.ID != id || (response.Result != nil) == (response.Error != nil) {
		return observed(core.Unhealthy, "a2a_protocol"), nil
	}
	if response.Error != nil {
		var e struct {
			Code    *int    `json:"code"`
			Message *string `json:"message"`
		}
		if json.Unmarshal(response.Error, &e) != nil || e.Code == nil || e.Message == nil {
			return observed(core.Unhealthy, "a2a_protocol"), nil
		}
		if method == "tasks/get" && *e.Code == -32001 {
			return observed(core.Healthy, ""), nil
		}
		return observed(core.Unhealthy, "a2a_rpc"), nil
	}
	if method == "tasks/get" {
		var task struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(response.Result, &task) != nil || task.ID != params["id"] {
			return observed(core.Unhealthy, "a2a_protocol"), nil
		}
	}
	return validateResult(response.Result, dimension), nil
}
func (a Adapter) identifier() (string, error) {
	source := a.random
	if source == nil {
		source = rand.Reader
	}
	var b [16]byte
	if _, err := io.ReadFull(source, b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
func has(values []string, s string) bool {
	for _, v := range values {
		if v == s {
			return true
		}
	}
	return false
}
func validateResult(raw json.RawMessage, dimension string) core.Observation {
	var result struct {
		Kind      string            `json:"kind"`
		ID        string            `json:"id"`
		ContextID string            `json:"contextId"`
		MessageID string            `json:"messageId"`
		Role      string            `json:"role"`
		Parts     []json.RawMessage `json:"parts"`
		Status    struct {
			State string `json:"state"`
		} `json:"status"`
	}
	if json.Unmarshal(raw, &result) != nil {
		return observed(core.Unhealthy, "a2a_protocol")
	}
	if result.Kind == "message" && dimension == "functional" && result.Role == "agent" && result.MessageID != "" && len(result.Parts) > 0 {
		for _, raw := range result.Parts {
			var p struct {
				Kind string  `json:"kind"`
				Text *string `json:"text"`
			}
			if json.Unmarshal(raw, &p) != nil || p.Kind != "text" || p.Text == nil {
				return observed(core.Unhealthy, "a2a_protocol")
			}
		}
		return observed(core.Healthy, "")
	}
	if result.Kind != "task" || result.ID == "" || result.ContextID == "" || !has([]string{"submitted", "working", "input-required", "completed", "canceled", "failed", "rejected", "auth-required", "unknown"}, result.Status.State) {
		return observed(core.Unhealthy, "a2a_protocol")
	}
	if dimension != "functional" || result.Status.State == "completed" {
		return observed(core.Healthy, "")
	}
	if result.Status.State == "failed" || result.Status.State == "rejected" || result.Status.State == "canceled" {
		return observed(core.Unhealthy, "a2a_functional")
	}
	if result.Status.State == "auth-required" {
		return observed(core.Misconfigured, "a2a_auth")
	}
	return observed(core.Unknown, "a2a_pending")
}

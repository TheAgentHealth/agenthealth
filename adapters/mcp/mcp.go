// Package mcp implements bounded, session-isolated Streamable HTTP health checks.
package mcp

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

const maxBody = 1048576
const maxPages = 100

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "mcp", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"mcp"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}

type probeFailure struct {
	status core.Status
	code   string
}

func (e *probeFailure) Error() string            { return "MCP health check failed" }
func fail(status core.Status, code string) error { return &probeFailure{status, code} }
func protocolError() error                       { return fail(core.Unhealthy, "mcp_protocol") }

type session struct {
	r            core.Request
	version, id  string
	sequence     int
	capabilities map[string]json.RawMessage
	received     atomic.Bool
}

func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	s := &session{r: r, version: "2025-11-25"}
	observation := func(err error) (core.Observation, error) {
		o := core.Observation{Check: core.CheckResult{Status: core.Healthy}, ResponseReceived: s.received.Load()}
		var f *probeFailure
		if errors.As(err, &f) {
			o.Check.Status = f.status
			o.Code = f.code
			return o, nil
		}
		return o, err
	}
	if dimension == "configuration" {
		u, err := url.Parse(r.Target.Endpoint)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" || !headerValue(r.Credential) {
			return observation(fail(core.Misconfigured, ""))
		}
		return observation(nil)
	}
	if dimension == "reachability" {
		req, err := s.request(ctx, http.MethodHead, nil)
		if err != nil {
			return observation(err)
		}
		start := func(stage string) {
			if r.RecordStep != nil {
				r.RecordStep(stage, core.Unknown)
			}
		}
		done := func(stage string, err error) {
			if r.RecordStep != nil {
				state := core.Healthy
				if err != nil {
					state = core.Unreachable
				}
				r.RecordStep(stage, state)
			}
		}
		trace := &httptrace.ClientTrace{
			DNSStart: func(httptrace.DNSStartInfo) { start("dns") }, DNSDone: func(i httptrace.DNSDoneInfo) { done("dns", i.Err) },
			ConnectStart: func(string, string) { start("tcp") }, ConnectDone: func(_, _ string, e error) { done("tcp", e) },
			TLSHandshakeStart: func() { start("tls") }, TLSHandshakeDone: func(_ tls.ConnectionState, e error) { done("tls", e) },
			GotConn: func(httptrace.GotConnInfo) { start("http") },
		}
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
		// Reinstall response tracing after adding transport hooks.
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{GotFirstResponseByte: s.mark}))
		resp, err := r.Client.Do(req)
		if err != nil {
			return observation(err)
		}
		defer resp.Body.Close()
		s.mark()
		done("http", nil)
		return observation(nil)
	}
	defer s.close(ctx)
	err := s.initialize(ctx)
	if err != nil {
		return observation(err)
	}
	switch dimension {
	case "authentication", "protocol":
		return observation(nil)
	case "capability", "functional":
		inventories := map[string]map[string]json.RawMessage{}
		for _, kind := range []string{"tools", "resources", "prompts"} {
			if _, exists := s.capabilities[kind]; !exists {
				continue
			}
			items, e := s.list(ctx, kind)
			if e != nil {
				return observation(e)
			}
			inventories[kind] = items
		}
		if options := r.Target.MCP; options != nil {
			for kind, required := range map[string][]string{"tools": options.RequiredTools, "resources": options.RequiredResources, "prompts": options.RequiredPrompts} {
				for _, name := range required {
					if _, ok := inventories[kind][name]; !ok {
						return observation(fail(core.Degraded, "mcp_required"))
					}
				}
			}
		}
		if dimension == "functional" {
			if r.Target.MCP == nil || r.Target.MCP.Functional == nil || !r.Target.MCP.Functional.Safe {
				return observation(fail(core.Misconfigured, ""))
			}
			f := r.Target.MCP.Functional
			tool, ok := inventories["tools"][f.Tool]
			if !ok {
				return observation(fail(core.Degraded, "mcp_required"))
			}
			var definition struct {
				Annotations struct {
					ReadOnly    bool  `json:"readOnlyHint"`
					Destructive *bool `json:"destructiveHint"`
				} `json:"annotations"`
			}
			if json.Unmarshal(tool, &definition) != nil || !definition.Annotations.ReadOnly || definition.Annotations.Destructive == nil || *definition.Annotations.Destructive {
				return observation(fail(core.Misconfigured, "mcp_unsafe"))
			}
			arguments := json.RawMessage(`{}`)
			if f.ArgumentsJSON != "" {
				arguments = json.RawMessage(f.ArgumentsJSON)
			}
			result, e := s.rpc(ctx, "tools/call", map[string]any{"name": f.Tool, "arguments": arguments})
			if e != nil {
				return observation(e)
			}
			var reply struct {
				Content []json.RawMessage `json:"content"`
				IsError bool              `json:"isError"`
			}
			if json.Unmarshal(result, &reply) != nil || reply.Content == nil {
				return observation(protocolError())
			}
			for _, item := range reply.Content {
				var c map[string]json.RawMessage
				if json.Unmarshal(item, &c) != nil || c == nil || !nonemptyString(c["type"]) {
					return observation(protocolError())
				}
			}
			if reply.IsError {
				return observation(fail(core.Unhealthy, "mcp_functional"))
			}
		}
		return observation(nil)
	default:
		return observation(fail(core.Unknown, ""))
	}
}
func headerValue(v string) bool {
	for i := 0; i < len(v); i++ {
		if v[i] < 32 || v[i] == 127 {
			return false
		}
	}
	return true
}
func sessionValue(v string) bool {
	if len(v) == 0 || len(v) > 4096 {
		return false
	}
	for i := 0; i < len(v); i++ {
		if v[i] < 33 || v[i] > 126 {
			return false
		}
	}
	return true
}
func (s *session) mark() {
	s.received.Store(true)
	if s.r.MarkResponse != nil {
		s.r.MarkResponse()
	}
}
func (s *session) request(ctx context.Context, method string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, s.r.Target.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, &core.Failure{Status: core.Misconfigured}
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s.r.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+s.r.Credential)
	}
	if s.id != "" {
		req.Header.Set("Mcp-Session-Id", s.id)
	}
	if s.sequence > 0 {
		req.Header.Set("MCP-Protocol-Version", s.version)
	}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{GotFirstResponseByte: s.mark})), nil
}
func (s *session) close(ctx context.Context) {
	if s.id == "" || ctx.Err() != nil {
		return
	}
	cleanup, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	req, err := s.request(cleanup, http.MethodDelete, nil)
	if err != nil {
		return
	}
	resp, err := s.r.Client.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}
func (s *session) initialize(ctx context.Context) error {
	if o := s.r.Target.MCP; o != nil && o.ProtocolVersion != "" {
		s.version = o.ProtocolVersion
	}
	result, err := s.rpc(ctx, "initialize", map[string]any{"protocolVersion": s.version, "capabilities": map[string]any{}, "clientInfo": map[string]string{"name": "agenthealth", "version": "0.1.0"}})
	if err != nil {
		return err
	}
	var init struct {
		Version      string                     `json:"protocolVersion"`
		Capabilities map[string]json.RawMessage `json:"capabilities"`
		ServerInfo   struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"serverInfo"`
	}
	if json.Unmarshal(result, &init) != nil || init.Capabilities == nil || init.ServerInfo.Name == "" || init.ServerInfo.Version == "" {
		return protocolError()
	}
	if !core.SupportedMCPVersion(init.Version) || (s.r.Target.MCP != nil && s.r.Target.MCP.ProtocolVersion != "" && init.Version != s.version) {
		return fail(core.Misconfigured, "mcp_version")
	}
	for _, value := range init.Capabilities {
		var obj map[string]json.RawMessage
		if json.Unmarshal(value, &obj) != nil || obj == nil {
			return protocolError()
		}
	}
	s.version = init.Version
	s.capabilities = init.Capabilities
	_, err = s.rpc(ctx, "notifications/initialized", nil)
	return err
}
func (s *session) rpc(ctx context.Context, method string, params any) (json.RawMessage, error) {
	notification := method == "notifications/initialized"
	payload := map[string]any{"jsonrpc": "2.0", "method": method}
	expected := s.sequence + 1
	if !notification {
		payload["id"] = expected
	}
	if params != nil {
		payload["params"] = params
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, protocolError()
	}
	req, err := s.request(ctx, http.MethodPost, body)
	if err != nil {
		return nil, err
	}
	s.sequence++
	resp, err := s.r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	s.mark()
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return nil, fail(core.Misconfigured, "mcp_auth")
	}
	if notification {
		if resp.StatusCode != http.StatusAccepted {
			return nil, fail(core.Unhealthy, "mcp_http")
		}
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fail(core.Unhealthy, "mcp_http")
	}
	if method == "initialize" {
		if ids := resp.Header.Values("Mcp-Session-Id"); len(ids) > 0 {
			if len(ids) != 1 || !sessionValue(ids[0]) {
				return nil, protocolError()
			}
			s.id = ids[0]
		}
	}
	media, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, protocolError()
	}
	reader := &io.LimitedReader{R: resp.Body, N: maxBody + 1}
	if media == "application/json" {
		data, e := io.ReadAll(reader)
		if e != nil {
			return nil, e
		}
		if len(data) > maxBody {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		return decodeResponse(data, expected)
	}
	if media != "text/event-stream" {
		return nil, protocolError()
	}
	// Consume bounded SSE events until the matching response; do not wait for EOF.
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxBody+1)
	var data []byte
	total := 0
	for scanner.Scan() {
		line := scanner.Text()
		total += len(line) + 1
		if total > maxBody {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		if line == "" {
			if len(data) > 0 {
				var envelope map[string]json.RawMessage
				if json.Unmarshal(data, &envelope) != nil {
					return nil, protocolError()
				}
				if _, hasID := envelope["id"]; hasID {
					return decodeResponse(data, expected)
				}
				if string(envelope["jsonrpc"]) != `"2.0"` || !nonemptyString(envelope["method"]) {
					return nil, protocolError()
				}
			}
			data = nil
			continue
		}
		if strings.HasPrefix(line, "data:") {
			if len(data) > 0 {
				data = append(data, '\n')
			}
			data = append(data, strings.TrimPrefix(strings.TrimPrefix(line, "data:"), " ")...)
		}
	}
	if err := scanner.Err(); err != nil {
		if errors.Is(err, bufio.ErrTooLong) {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		return nil, err
	}
	if reader.N == 0 {
		return nil, fail(core.Unknown, "mcp_limit")
	}
	return nil, protocolError()
}
func decodeResponse(data []byte, expected int) (json.RawMessage, error) {
	var e map[string]json.RawMessage
	if json.Unmarshal(data, &e) != nil || e == nil || string(e["jsonrpc"]) != `"2.0"` {
		return nil, protocolError()
	}
	var id int
	if json.Unmarshal(e["id"], &id) != nil || id != expected {
		return nil, protocolError()
	}
	result, hasResult := e["result"]
	rpcError, hasError := e["error"]
	if hasResult == hasError {
		return nil, protocolError()
	}
	if _, hasMethod := e["method"]; hasMethod {
		return nil, protocolError()
	}
	if hasError {
		var failure struct {
			Code    *int    `json:"code"`
			Message *string `json:"message"`
		}
		if json.Unmarshal(rpcError, &failure) != nil || failure.Code == nil || failure.Message == nil {
			return nil, protocolError()
		}
		return nil, fail(core.Unhealthy, "mcp_rpc")
	}
	var obj map[string]json.RawMessage
	if json.Unmarshal(result, &obj) != nil || obj == nil {
		return nil, protocolError()
	}
	return result, nil
}
func nonemptyString(raw json.RawMessage) bool {
	var s string
	return json.Unmarshal(raw, &s) == nil && strings.TrimSpace(s) != ""
}
func (s *session) list(ctx context.Context, kind string) (map[string]json.RawMessage, error) {
	items := map[string]json.RawMessage{}
	cursor := ""
	seen := map[string]bool{}
	for page := 0; page < maxPages; page++ {
		params := map[string]string{}
		if cursor != "" {
			params["cursor"] = cursor
		}
		result, err := s.rpc(ctx, kind+"/list", params)
		if err != nil {
			return nil, err
		}
		var obj map[string]json.RawMessage
		if json.Unmarshal(result, &obj) != nil {
			return nil, protocolError()
		}
		var entries []json.RawMessage
		if json.Unmarshal(obj[kind], &entries) != nil || entries == nil {
			return nil, protocolError()
		}
		for _, entry := range entries {
			var definition map[string]json.RawMessage
			if json.Unmarshal(entry, &definition) != nil || definition == nil {
				return nil, protocolError()
			}
			key := "name"
			if kind == "resources" {
				key = "uri"
			}
			var name string
			if json.Unmarshal(definition[key], &name) != nil || strings.TrimSpace(name) == "" {
				return nil, protocolError()
			}
			if !nonemptyString(definition["name"]) {
				return nil, protocolError()
			}
			if kind == "tools" {
				var schema map[string]json.RawMessage
				if json.Unmarshal(definition["inputSchema"], &schema) != nil || schema == nil || string(schema["type"]) != `"object"` {
					return nil, protocolError()
				}
			}
			if _, duplicate := items[name]; duplicate {
				return nil, protocolError()
			}
			items[name] = entry
			if len(items) > 10000 {
				return nil, fail(core.Unknown, "mcp_limit")
			}
		}
		next, exists := obj["nextCursor"]
		if !exists {
			return items, nil
		}
		if json.Unmarshal(next, &cursor) != nil || cursor == "" {
			return nil, protocolError()
		}
		if seen[cursor] {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		seen[cursor] = true
	}
	return nil, fail(core.Unknown, "mcp_limit")
}

// Package mcp implements bounded, session-isolated Streamable HTTP health checks.
package mcp

import (
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
	"os"
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
	modern       bool
	initialized  bool
	stdio        *stdioTransport
	toolHeaders  map[string]string
	responseHook atomic.Value
	ready        bool
	initErr      error
	inventories  map[string]map[string]json.RawMessage
}

func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	s := &session{r: r, version: "2025-11-25"}
	shared := r.TargetContext != nil && r.RunState != nil && dimension != "configuration"
	if shared {
		h := sessionState(r.RunState)
		if !h.acquire(ctx) {
			return core.Observation{}, ctx.Err()
		}
		defer h.release()
		if h.session == nil {
			h.session = s
		} else {
			s = h.session
			token := s.r.Credential
			s.r = r
			if r.Target.MCP != nil && r.Target.MCP.OAuth != nil {
				s.r.Credential = token
			}
		}
		defer func() {
			if ctx.Err() != nil || r.TargetContext.Err() != nil || (!s.ready && !s.received.Load()) {
				cleanupSession(s)
				h.session = nil
			}
		}()
	}
	hook := r.MarkResponse
	if hook == nil {
		hook = func() {}
	}
	s.responseHook.Store(hook)
	s.received.Store(false)
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
		if o := r.Target.MCP; o != nil && o.Transport == "stdio" {
			u, err := url.Parse(r.Target.Endpoint)
			if err != nil || u.Scheme != "stdio" || u.Host == "" || u.User != nil || u.Fragment != "" || u.RawQuery != "" || o.Stdio == nil {
				return observation(fail(core.Misconfigured, ""))
			}
			if _, err := stdioCommand(o.Stdio); err != nil {
				return observation(fail(core.Misconfigured, "mcp_process"))
			}
			if o.Stdio.Directory != "" {
				info, err := os.Stat(o.Stdio.Directory)
				if err != nil || !info.IsDir() {
					return observation(fail(core.Misconfigured, "mcp_process"))
				}
			}
			return observation(nil)
		}
		if o := r.Target.MCP; o != nil && o.OAuth != nil {
			if !secureURL(o.OAuth.Issuer) {
				return observation(fail(core.Misconfigured, "mcp_oauth"))
			}
		}
		u, err := url.Parse(r.Target.Endpoint)
		if err != nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" || !headerValue(r.Credential) {
			return observation(fail(core.Misconfigured, ""))
		}
		return observation(nil)
	}
	if o := r.Target.MCP; o != nil && o.Transport == "stdio" && s.stdio == nil {
		stdioRequest := r
		stdioRequest.MarkResponse = s.mark
		processCtx := ctx
		if shared {
			processCtx = r.TargetContext
		}
		transport, err := startStdio(processCtx, stdioRequest)
		if err != nil {
			return observation(err)
		}
		s.stdio = transport
	}
	if !shared {
		defer s.close(ctx)
	}
	if dimension == "reachability" && s.stdio == nil {
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
	if s.initErr != nil {
		s.mark()
		return observation(s.initErr)
	}
	if !s.ready {
		err := s.initialize(ctx)
		if err != nil {
			if s.received.Load() && ctx.Err() == nil {
				s.initErr = err
			}
			return observation(err)
		}
		s.ready = true
	} else {
		s.mark()
	}
	switch dimension {
	case "reachability", "authentication", "protocol":
		return observation(nil)
	case "capability", "functional":
		inventories := s.inventories
		if inventories == nil {
			inventories = map[string]map[string]json.RawMessage{}
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
			s.inventories = inventories
		}
		if options := r.Target.MCP; options != nil {
			for kind, required := range map[string][]string{"tools": options.RequiredTools, "resources": options.RequiredResources, "prompts": options.RequiredPrompts} {
				for _, name := range required {
					if _, ok := inventories[kind][name]; !ok {
						return observation(fail(core.Unhealthy, "mcp_required"))
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
				return observation(fail(core.Unhealthy, "mcp_required"))
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
			if s.modern && s.stdio == nil {
				headers, e := invocationHeaders(tool, arguments)
				if e != nil {
					return observation(e)
				}
				s.toolHeaders = headers
			}
			result, e := s.rpc(ctx, "tools/call", map[string]any{"name": f.Tool, "arguments": arguments})
			if e != nil {
				return observation(e)
			}
			var reply struct {
				Content []json.RawMessage `json:"content"`
				IsError bool              `json:"isError"`
			}
			if s.modern {
				if err := completeResult(result); err != nil {
					return observation(err)
				}
			}
			if json.Unmarshal(result, &reply) != nil || reply.Content == nil {
				return observation(protocolError())
			}
			for _, item := range reply.Content {
				if !validContent(item) {
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
	if hook := s.responseHook.Load(); hook != nil {
		hook.(func())()
	} else if s.r.MarkResponse != nil {
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
	if s.modern || s.initialized {
		req.Header.Set("MCP-Protocol-Version", s.version)
	}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{GotFirstResponseByte: s.mark})), nil
}
func (s *session) close(ctx context.Context) {
	if s.stdio != nil {
		s.stdio.close()
		return
	}
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
	if s.r.Target.MCP != nil && s.r.Target.MCP.OAuth != nil {
		token, err := acquireToken(ctx, s.r)
		if err != nil {
			return err
		}
		s.r.Credential = token
	}
	pin := ""
	if s.r.Target.MCP != nil {
		pin = s.r.Target.MCP.ProtocolVersion
	}
	if pin == "" || pin == modernVersion {
		s.modern = true
		s.version = modernVersion
		probeCtx := ctx
		cancel := func() {}
		if s.stdio != nil && pin == "" {
			probeCtx, cancel = context.WithTimeout(ctx, 500*time.Millisecond)
		}
		err := s.discover(probeCtx)
		cancel()
		if err == nil {
			return nil
		}
		if pin != "" || !s.legacyFallback(err) {
			return err
		}
		s.modern = false
		s.version = "2025-11-25"
	}
	return s.initializeLegacy(ctx)
}
func (s *session) initializeLegacy(ctx context.Context) error {
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
	if init.Version == modernVersion || !core.SupportedMCPVersion(init.Version) || (s.r.Target.MCP != nil && s.r.Target.MCP.ProtocolVersion != "" && init.Version != s.version) {
		return fail(core.Unhealthy, "mcp_version")
	}
	for _, value := range init.Capabilities {
		var obj map[string]json.RawMessage
		if json.Unmarshal(value, &obj) != nil || obj == nil {
			return protocolError()
		}
	}
	s.version = init.Version
	s.initialized = true
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
	if s.modern {
		raw, err := json.Marshal(params)
		if err != nil {
			return nil, protocolError()
		}
		var p map[string]any
		if params != nil && json.Unmarshal(raw, &p) != nil {
			return nil, protocolError()
		}
		if p == nil {
			p = map[string]any{}
		}
		caps := map[string]any{}
		if o := s.r.Target.MCP; o != nil && o.OAuth != nil && o.OAuth.Grant == "client_credentials" {
			caps["extensions"] = map[string]any{"io.modelcontextprotocol/oauth-client-credentials": map[string]any{}}
		}
		p["_meta"] = map[string]any{"io.modelcontextprotocol/protocolVersion": s.version, "io.modelcontextprotocol/clientInfo": map[string]string{"name": "agenthealth", "version": "0.1.0"}, "io.modelcontextprotocol/clientCapabilities": caps}
		payload["params"] = p
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, protocolError()
	}
	if s.stdio != nil {
		s.sequence++
		return s.stdio.rpc(ctx, body, expected, notification, s.mark)
	}
	req, err := s.request(ctx, http.MethodPost, body)
	if err != nil {
		return nil, err
	}
	if s.modern {
		req.Header.Set("MCP-Protocol-Version", s.version)
		req.Header.Set("Mcp-Method", method)
		if method == "tools/call" {
			var p struct {
				Name string `json:"name"`
			}
			raw, _ := json.Marshal(params)
			json.Unmarshal(raw, &p)
			req.Header.Set("Mcp-Name", encodeHeader(p.Name))
			for name, value := range s.toolHeaders {
				req.Header.Set(name, value)
			}
		}
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
		data, e := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
		if e != nil {
			return nil, e
		}
		if len(data) > maxBody {
			return nil, fail(core.Unknown, "mcp_limit")
		}
		return nil, httpRPCError(data, resp.StatusCode)
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
	return s.readSSE(ctx, resp, expected)
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
		return nil, parseRPCError(rpcError)
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
		if s.modern {
			if err := completeResult(result); err != nil {
				return nil, err
			}
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
			if s.modern && s.stdio == nil && kind == "tools" {
				if _, err := headerBindings(entry); err != nil {
					continue
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

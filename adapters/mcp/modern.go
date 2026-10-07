package mcp

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
)

const modernVersion = "2026-07-28"

type rpcFailure struct {
	code      int
	supported []string
}

func (e *rpcFailure) Error() string { return "MCP RPC rejected" }
func (e *rpcFailure) Unwrap() error {
	if e.code == -32022 {
		return fail("MISCONFIGURED", "mcp_version")
	}
	return fail("UNHEALTHY", "mcp_rpc")
}

type httpFailure struct {
	status int
	cause  error
}

func (e *httpFailure) Error() string { return "MCP HTTP rejected" }
func (e *httpFailure) Unwrap() error { return e.cause }
func parseRPCError(raw json.RawMessage) error {
	var e struct {
		Code int `json:"code"`
		Data struct {
			Supported []string `json:"supported"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &e) != nil {
		return protocolError()
	}
	return &rpcFailure{code: e.Code, supported: e.Data.Supported}
}
func httpRPCError(data []byte, status int) error {
	cause := fail("UNHEALTHY", "mcp_http")
	var obj map[string]json.RawMessage
	if json.Unmarshal(data, &obj) == nil && string(obj["jsonrpc"]) == `"2.0"` {
		if raw, ok := obj["error"]; ok {
			cause = parseRPCError(raw)
		}
	}
	return &httpFailure{status: status, cause: cause}
}
func (s *session) legacyFallback(err error) bool {
	var rpc *rpcFailure
	if errors.As(err, &rpc) && (rpc.code == -32022 || rpc.code == -32020) {
		return false
	}
	if s.stdio != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return true
		}
		return errors.As(err, &rpc)
	}
	var h *httpFailure
	if !errors.As(err, &h) {
		return false
	}
	if h.status == http.StatusNotFound && rpc != nil && rpc.code == -32601 {
		return false
	}
	return h.status >= 400 && h.status < 500 && h.status != 401 && h.status != 403
}
func (s *session) discover(ctx context.Context) error {
	result, err := s.rpc(ctx, "server/discover", map[string]any{})
	if err != nil {
		return err
	}
	var d struct {
		ResultType   string                     `json:"resultType"`
		Versions     []string                   `json:"supportedVersions"`
		Capabilities map[string]json.RawMessage `json:"capabilities"`
	}
	if json.Unmarshal(result, &d) != nil || d.ResultType != "complete" || d.Capabilities == nil || len(d.Versions) == 0 {
		return protocolError()
	}
	found := false
	for _, v := range d.Versions {
		if v == modernVersion {
			found = true
		}
	}
	if !found {
		return fail("MISCONFIGURED", "mcp_version")
	}
	for _, raw := range d.Capabilities {
		var obj map[string]json.RawMessage
		if json.Unmarshal(raw, &obj) != nil || obj == nil {
			return protocolError()
		}
	}
	s.capabilities = d.Capabilities
	return nil
}
func encodeHeader(value string) string {
	safe := strings.TrimSpace(value) == value && !(strings.HasPrefix(value, "=?base64?") && strings.HasSuffix(value, "?="))
	for i := 0; i < len(value); i++ {
		if (value[i] < 32 && value[i] != 9) || value[i] > 126 {
			safe = false
		}
	}
	if safe {
		return value
	}
	return "=?base64?" + base64.StdEncoding.EncodeToString([]byte(value)) + "?="
}

type headerBinding struct {
	name, kind string
	path       []string
}

func headerBindings(tool json.RawMessage) ([]headerBinding, error) {
	var definition map[string]json.RawMessage
	if json.Unmarshal(tool, &definition) != nil {
		return nil, protocolError()
	}
	var root any
	if json.Unmarshal(definition["inputSchema"], &root) != nil {
		return nil, protocolError()
	}
	bindings := []headerBinding{}
	seen := map[string]bool{}
	var walk func(any, []string, bool, int) error
	walk = func(node any, path []string, reachable bool, depth int) error {
		if depth > 64 {
			return protocolError()
		}
		switch obj := node.(type) {
		case map[string]any:
			if annotation, ok := obj["x-mcp-header"]; ok {
				name, valid := annotation.(string)
				kind, _ := obj["type"].(string)
				if !valid || name == "" || !reachable || len(path) == 0 || (kind != "string" && kind != "boolean" && kind != "integer") || seen[strings.ToLower(name)] {
					return protocolError()
				}
				for _, c := range name {
					if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
						return protocolError()
					}
				}
				seen[strings.ToLower(name)] = true
				bindings = append(bindings, headerBinding{name: name, kind: kind, path: append([]string{}, path...)})
			}
			for key, value := range obj {
				if key == "properties" {
					props, ok := value.(map[string]any)
					if !ok {
						return protocolError()
					}
					for name, child := range props {
						p := append(append([]string{}, path...), name)
						if err := walk(child, p, reachable, depth+1); err != nil {
							return err
						}
					}
				} else if err := walk(value, path, false, depth+1); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range obj {
				if err := walk(child, path, false, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := walk(root, nil, true, 0); err != nil {
		return nil, err
	}
	return bindings, nil
}
func invocationHeaders(tool, arguments json.RawMessage) (map[string]string, error) {
	bindings, err := headerBindings(tool)
	if err != nil {
		return nil, err
	}
	var args any
	decoder := json.NewDecoder(strings.NewReader(string(arguments)))
	decoder.UseNumber()
	if decoder.Decode(&args) != nil {
		return nil, protocolError()
	}
	headers := map[string]string{}
	for _, binding := range bindings {
		value := args
		for _, key := range binding.path {
			obj, ok := value.(map[string]any)
			if !ok {
				value = nil
				break
			}
			value = obj[key]
		}
		if value == nil {
			continue
		}
		var text string
		switch binding.kind {
		case "string":
			v, ok := value.(string)
			if !ok {
				return nil, protocolError()
			}
			text = v
		case "boolean":
			v, ok := value.(bool)
			if !ok {
				return nil, protocolError()
			}
			text = strconv.FormatBool(v)
		case "integer":
			v, ok := value.(json.Number)
			if !ok {
				return nil, protocolError()
			}
			n, err := v.Float64()
			if err != nil || math.Trunc(n) != n || math.Abs(n) > 9007199254740991 {
				return nil, protocolError()
			}
			text = strconv.FormatFloat(n, 'f', -1, 64)
		}
		headers["Mcp-Param-"+binding.name] = encodeHeader(text)
	}
	return headers, nil
}

func completeResult(raw json.RawMessage) error {
	var obj map[string]json.RawMessage
	if json.Unmarshal(raw, &obj) != nil {
		return protocolError()
	}
	switch string(obj["resultType"]) {
	case `"complete"`:
		return nil
	case `"input_required"`:
		return fail("UNKNOWN", "mcp_input")
	default:
		return protocolError()
	}
}

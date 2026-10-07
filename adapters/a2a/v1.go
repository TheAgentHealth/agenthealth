package a2a

import (
	"encoding/json"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

// v1Card normalizes only supported JSON-RPC interfaces. Discovery never causes
// credentials to be forwarded to another origin, or silently downgrades v1.
func v1Card(raw []byte) (card, core.Observation) {
	var c card
	if json.Unmarshal(raw, &c) != nil {
		return c, observed(core.Unhealthy, "a2a_card")
	}
	var wire struct {
		Interfaces []struct {
			URL     string `json:"url"`
			Binding string `json:"protocolBinding"`
			Version string `json:"protocolVersion"`
			Tenant  string `json:"tenant"`
		} `json:"supportedInterfaces"`
		Requirements []struct {
			Schemes map[string]struct {
				List []string `json:"list"`
			} `json:"schemes"`
		} `json:"securityRequirements"`
		Schemes map[string]json.RawMessage `json:"securitySchemes"`
	}
	if json.Unmarshal(raw, &wire) != nil || len(wire.Interfaces) == 0 {
		return c, observed(core.Unhealthy, "a2a_card")
	}
	c.ProtocolVersion = "1.0"
	c.PreferredTransport = "JSONRPC"
	c.URL = ""
	for _, i := range wire.Interfaces {
		if i.URL == "" || i.Binding == "" || i.Version == "" {
			return c, observed(core.Unhealthy, "a2a_card")
		}
		if c.URL == "" && i.Binding == "JSONRPC" && i.Version == "1.0" {
			c.URL = i.URL
			c.Tenant = i.Tenant
		}
	}
	if c.URL == "" {
		return c, observed(core.Unhealthy, "a2a_version")
	}
	c.Security = nil
	c.SecuritySchemes = make(map[string]securityScheme)
	for name, raw := range wire.Schemes {
		var schemes map[string]json.RawMessage
		if json.Unmarshal(raw, &schemes) != nil || len(schemes) != 1 {
			return c, observed(core.Unhealthy, "a2a_card")
		}
		var h struct {
			Scheme string `json:"scheme"`
		}
		if value, ok := schemes["httpAuthSecurityScheme"]; ok && json.Unmarshal(value, &h) == nil {
			c.SecuritySchemes[name] = securityScheme{Type: "http", Scheme: h.Scheme}
		}
	}
	for _, requirement := range wire.Requirements {
		if requirement.Schemes == nil {
			return c, observed(core.Unhealthy, "a2a_card")
		}
		alternative := make(map[string][]string)
		for name, scopes := range requirement.Schemes {
			alternative[name] = scopes.List
		}
		c.Security = append(c.Security, alternative)
	}
	return c, observed(core.Healthy, "")
}

func validateV1Result(raw json.RawMessage, dimension string) core.Observation {
	if dimension == "functional" {
		var envelope map[string]json.RawMessage
		if json.Unmarshal(raw, &envelope) != nil {
			return observed(core.Unhealthy, "a2a_protocol")
		}
		task, hasTask := envelope["task"]
		message, hasMessage := envelope["message"]
		if hasTask == hasMessage {
			return observed(core.Unhealthy, "a2a_protocol")
		}
		if hasMessage {
			var m struct {
				ID        string                       `json:"messageId"`
				ContextID string                       `json:"contextId"`
				Role      json.RawMessage              `json:"role"`
				Parts     []map[string]json.RawMessage `json:"parts"`
			}
			if json.Unmarshal(message, &m) != nil || m.ID == "" || m.ContextID == "" || (string(m.Role) != `"ROLE_AGENT"` && string(m.Role) != "2") || len(m.Parts) == 0 {
				return observed(core.Unhealthy, "a2a_protocol")
			}
			for _, part := range m.Parts {
				count := 0
				for _, key := range []string{"text", "raw", "url", "data"} {
					if _, ok := part[key]; ok {
						count++
					}
				}
				var text string
				if count != 1 || part["text"] == nil || string(part["text"]) == "null" || json.Unmarshal(part["text"], &text) != nil {
					return observed(core.Unhealthy, "a2a_protocol")
				}
			}
			return observed(core.Healthy, "")
		}
		raw = task
	}
	var t struct {
		ID     string `json:"id"`
		Status struct {
			State json.RawMessage `json:"state"`
		} `json:"status"`
	}
	if json.Unmarshal(raw, &t) != nil || t.ID == "" {
		return observed(core.Unhealthy, "a2a_protocol")
	}
	if len(t.Status.State) == 0 || string(t.Status.State) == "null" {
		return observed(core.Unhealthy, "a2a_protocol")
	}
	var state string
	if json.Unmarshal(t.Status.State, &state) != nil {
		var n int
		states := []string{"UNSPECIFIED", "SUBMITTED", "WORKING", "COMPLETED", "FAILED", "CANCELED", "INPUT_REQUIRED", "REJECTED", "AUTH_REQUIRED"}
		if json.Unmarshal(t.Status.State, &n) != nil || n < 0 || n >= len(states) {
			return observed(core.Unhealthy, "a2a_protocol")
		}
		state = "TASK_STATE_" + states[n]
	}
	switch state {
	case "TASK_STATE_UNSPECIFIED", "TASK_STATE_SUBMITTED", "TASK_STATE_WORKING", "TASK_STATE_INPUT_REQUIRED":
		if dimension == "functional" {
			return observed(core.Unknown, "a2a_pending")
		}
	case "TASK_STATE_COMPLETED":
	case "TASK_STATE_FAILED", "TASK_STATE_CANCELED", "TASK_STATE_REJECTED":
		if dimension == "functional" {
			return observed(core.Unhealthy, "a2a_functional")
		}
	case "TASK_STATE_AUTH_REQUIRED":
		if dimension == "functional" {
			return observed(core.Misconfigured, "a2a_auth")
		}
	default:
		return observed(core.Unhealthy, "a2a_protocol")
	}
	return observed(core.Healthy, "")
}

func protocolVersion(t core.Target) string {
	if t.A2A != nil && strings.TrimSpace(t.A2A.ProtocolVersion) != "" {
		return t.A2A.ProtocolVersion
	}
	return "1.0"
}

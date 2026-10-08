// Package agent implements a framework-neutral, bounded agent health interface.
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"strings"

	httpadapter "github.com/TheAgentHealth/agenthealth/adapters/http"
	"github.com/TheAgentHealth/agenthealth/core"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "agent", Version: "0.5.0", CompatibilityVersion: "v1", TargetTypes: []string{"agent", "multi-agent"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}

type document struct {
	Version      string   `json:"version"`
	Name         string   `json:"name"`
	Live         *bool    `json:"live"`
	Ready        *bool    `json:"ready"`
	Capabilities []string `json:"capabilities"`
	Dependencies []string `json:"dependencies"`
}
type evidence struct {
	doc         document
	observation core.Observation
}

func observation(status core.Status, code string) core.Observation {
	return core.Observation{Check: core.CheckResult{Status: status}, ResponseReceived: true, Code: code}
}
func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	if dimension == "configuration" {
		for _, check := range r.Target.Checks {
			if check == "functional" && (r.Target.Agent == nil || r.Target.Agent.Functional == nil) {
				return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
			}
		}
	}
	if dimension == "configuration" || dimension == "reachability" {
		return (httpadapter.Adapter{}).Check(ctx, r, dimension)
	}
	if dimension == "functional" {
		return task(ctx, r)
	}
	var e evidence
	if r.RunState != nil {
		if v, ok := r.RunState.Load("agent.document"); ok {
			e = v.(evidence)
		} else {
			var err error
			e, err = readDocument(ctx, r)
			if err != nil {
				return e.observation, err
			}
			r.RunState.Store("agent.document", e)
		}
	} else {
		var err error
		e, err = readDocument(ctx, r)
		if err != nil {
			return e.observation, err
		}
	}
	if e.observation.Check.Status != core.Healthy {
		return e.observation, nil
	}
	if dimension == "protocol" {
		if !*e.doc.Live || !*e.doc.Ready {
			return observation(core.Unhealthy, "agent_readiness"), nil
		}
	}
	if dimension == "capability" {
		has := func(names []string, name string) bool {
			for _, v := range names {
				if v == name {
					return true
				}
			}
			return false
		}
		if r.Target.Agent != nil {
			for _, name := range r.Target.Agent.RequiredCapabilities {
				if !has(e.doc.Capabilities, name) {
					return observation(core.Unhealthy, "agent_required"), nil
				}
			}
		}
		// Remote discovery supplies names only. Explicit configuration authorizes execution.
		for _, name := range e.doc.Dependencies {
			found := false
			for _, dep := range r.Target.Dependencies {
				if dep.Name == name {
					found = true
				}
			}
			if !found {
				return observation(core.Unknown, "agent_required"), nil
			}
		}
	}
	return observation(core.Healthy, ""), nil
}
func exchange(ctx context.Context, r core.Request, method string, body []byte, out any) (core.Observation, error) {
	req, err := http.NewRequestWithContext(ctx, method, r.Target.Endpoint, bytes.NewReader(body))
	if err != nil {
		return core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.Credential != "" {
		req.Header.Set("Authorization", "Bearer "+r.Credential)
	}
	res, err := r.Client.Do(req)
	if err != nil {
		return core.Observation{}, err
	}
	defer res.Body.Close()
	if r.MarkResponse != nil {
		r.MarkResponse()
	}
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return observation(core.Misconfigured, "http_auth"), nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return observation(core.Unhealthy, "http_status"), nil
	}
	media, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return observation(core.Unhealthy, "agent_document"), nil
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 65537))
	if err != nil {
		return core.Observation{ResponseReceived: true}, err
	}
	if len(data) > 65536 {
		return observation(core.Unknown, "http_body_limit"), nil
	}
	if json.Unmarshal(data, out) != nil {
		return observation(core.Unhealthy, "agent_document"), nil
	}
	return observation(core.Healthy, ""), nil
}
func readDocument(ctx context.Context, r core.Request) (e evidence, err error) {
	e.observation, err = exchange(ctx, r, http.MethodGet, nil, &e.doc)
	if err == nil && e.observation.Check.Status == core.Healthy {
		valid := e.doc.Version == "v1" && strings.TrimSpace(e.doc.Name) != "" && e.doc.Live != nil && e.doc.Ready != nil && e.doc.Capabilities != nil && e.doc.Dependencies != nil
		for _, names := range [][]string{e.doc.Capabilities, e.doc.Dependencies} {
			seen := map[string]bool{}
			for _, name := range names {
				if strings.TrimSpace(name) == "" || seen[name] {
					valid = false
				}
				seen[name] = true
			}
		}
		if !valid {
			e.observation = observation(core.Unhealthy, "agent_document")
		}
	}
	return
}
func task(ctx context.Context, r core.Request) (core.Observation, error) {
	if r.Target.Agent == nil || r.Target.Agent.Functional == nil || !r.Target.Agent.Functional.Safe {
		return core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	f := r.Target.Agent.Functional
	data, err := json.Marshal(f)
	if err != nil {
		return core.Observation{}, err
	}
	var result struct {
		Completed  *bool  `json:"completed"`
		Downstream string `json:"downstream"`
		Success    *bool  `json:"success"`
	}
	o, err := exchange(ctx, r, http.MethodPost, data, &result)
	if err != nil || o.Check.Status != core.Healthy {
		return o, err
	}
	if result.Completed == nil || !*result.Completed {
		return observation(core.Unknown, "agent_pending"), nil
	}
	if result.Success == nil {
		return observation(core.Unknown, "agent_pending"), nil
	}
	if !*result.Success || result.Downstream != f.Downstream {
		return observation(core.Unhealthy, "agent_task"), nil
	}
	return observation(core.Healthy, ""), nil
}

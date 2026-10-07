// Package http provides a small reference adapter for exercising the core.
// Advanced HTTP diagnostics belong to roadmap Phase 4.
package http

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "http", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"http", "api"}, Dimensions: []string{"configuration", "reachability", "authentication", "functional"}, ActiveChecks: []string{"functional"}}
}
func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	if dimension == "configuration" {
		endpoint, err := url.Parse(r.Target.Endpoint)
		if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.User != nil || strings.ContainsAny(r.Credential, "\r\n") {
			return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
		}
		return core.Observation{Check: core.CheckResult{Status: core.Healthy}}, nil
	}
	method := http.MethodHead
	if dimension == "functional" {
		method = http.MethodGet
	}
	request, err := http.NewRequestWithContext(ctx, method, r.Target.Endpoint, nil)
	if err != nil {
		return core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	if r.Credential != "" {
		request.Header.Set("Authorization", "Bearer "+r.Credential)
	}
	response, err := r.Client.Do(request)
	if err != nil {
		return core.Observation{}, err
	}
	defer response.Body.Close()
	if r.MarkResponse != nil {
		r.MarkResponse()
	}
	status := core.Healthy
	if dimension != "reachability" && (response.StatusCode == 401 || response.StatusCode == 403) {
		status = core.Misconfigured
	}
	if dimension == "functional" && status == core.Healthy && (response.StatusCode < 200 || response.StatusCode >= 300) {
		status = core.Unhealthy
	}
	return core.Observation{Check: core.CheckResult{Status: status}, ResponseReceived: true}, nil
}

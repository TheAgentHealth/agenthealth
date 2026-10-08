// Package http implements passive HTTP readiness and opt-in response matching.
package http

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

// ReadinessGET selects GET for read-only health endpoints that do not support HEAD.
type Adapter struct{ ReadinessGET bool }

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "http", Version: "0.2.0", CompatibilityVersion: "v1", TargetTypes: []string{"http", "api"}, Dimensions: []string{"configuration", "reachability", "protocol", "authentication", "functional"}, ActiveChecks: []string{"functional"}}
}
func (a Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	if dimension == "configuration" {
		endpoint, err := url.Parse(r.Target.Endpoint)
		if err != nil || endpoint.Hostname() == "" || (endpoint.Scheme != "https" && endpoint.Scheme != "http") || endpoint.User != nil || endpoint.Fragment != "" || !validHeaderValue(r.Credential) {
			return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
		}
		return core.Observation{Check: core.CheckResult{Status: core.Healthy}}, nil
	}
	method := http.MethodHead
	if dimension == "functional" || a.ReadinessGET {
		method = http.MethodGet
	}
	request, err := http.NewRequestWithContext(ctx, method, r.Target.Endpoint, nil)
	if err != nil {
		return core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	if r.Credential != "" {
		request.Header.Set("Authorization", "Bearer "+r.Credential)
	}
	if dimension == "reachability" {
		step := func(name string, err error) {
			if r.RecordStep == nil {
				return
			}
			status := core.Healthy
			if err != nil {
				status = core.Unreachable
			}
			r.RecordStep(name, status)
		}
		start := func(name string) {
			if r.RecordStep != nil {
				r.RecordStep(name, core.Unknown)
			}
		}
		trace := &httptrace.ClientTrace{
			DNSStart:          func(httptrace.DNSStartInfo) { start("dns") },
			DNSDone:           func(info httptrace.DNSDoneInfo) { step("dns", info.Err) },
			ConnectStart:      func(string, string) { start("tcp") },
			ConnectDone:       func(_, _ string, err error) { step("tcp", err) },
			TLSHandshakeStart: func() { start("tls") },
			TLSHandshakeDone:  func(_ tls.ConnectionState, err error) { step("tls", err) },
			GotConn:           func(httptrace.GotConnInfo) { start("http") },
			GotFirstResponseByte: func() {
				if r.MarkResponse != nil {
					r.MarkResponse()
				}
			},
		}
		request = request.WithContext(httptrace.WithClientTrace(request.Context(), trace))
	}
	response, err := r.Client.Do(request)
	if err != nil {
		return core.Observation{}, err
	}
	defer response.Body.Close()
	if r.MarkResponse != nil {
		r.MarkResponse()
	}
	if dimension == "reachability" && r.RecordStep != nil {
		r.RecordStep("http", core.Healthy)
	}
	status := core.Healthy
	if dimension != "reachability" && (response.StatusCode == 401 || response.StatusCode == 403) {
		return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}, ResponseReceived: true, Code: "http_auth"}, nil
	}
	if dimension == "protocol" || dimension == "functional" {
		accepted := response.StatusCode >= 200 && response.StatusCode < 300
		if r.Target.HTTP != nil && r.Target.HTTP.ExpectedStatus != nil {
			accepted = false
			for _, code := range r.Target.HTTP.ExpectedStatus {
				if response.StatusCode == code {
					accepted = true
				}
			}
		}
		failure := func(code string, status core.Status) (core.Observation, error) {
			return core.Observation{Check: core.CheckResult{Status: status}, ResponseReceived: true, Code: code}, nil
		}
		if !accepted {
			return failure("http_status", core.Unhealthy)
		}
		if r.Target.HTTP != nil {
			for name, expected := range r.Target.HTTP.Headers {
				values, exists := response.Header[http.CanonicalHeaderKey(name)]
				matched := false
				for _, value := range values {
					if value == expected {
						matched = true
					}
				}
				if !exists || !matched {
					return failure("http_headers", core.Unhealthy)
				}
			}
			if dimension == "functional" && r.Target.HTTP.BodyContains != "" {
				limit := 65536
				if r.Target.HTTP.MaxBodyBytes != nil {
					limit = *r.Target.HTTP.MaxBodyBytes
				}
				body, err := io.ReadAll(io.LimitReader(response.Body, int64(limit)+1))
				if err != nil {
					return core.Observation{ResponseReceived: true}, err
				}
				if len(body) > limit {
					return failure("http_body_limit", core.Unknown)
				}
				if !strings.Contains(string(body), r.Target.HTTP.BodyContains) {
					return failure("http_body", core.Unhealthy)
				}
			}
		}
	}
	return core.Observation{Check: core.CheckResult{Status: status}, ResponseReceived: true}, nil
}

// Match net/http header validation: HTAB is permitted, other control bytes
// and DEL are invalid. Check before networking so failures are MISCONFIGURED.
func validHeaderValue(value string) bool {
	for i := 0; i < len(value); i++ {
		b := value[i]
		if (b < 0x20 && b != '\t') || b == 0x7f {
			return false
		}
	}
	return true
}

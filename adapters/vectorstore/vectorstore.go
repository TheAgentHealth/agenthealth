// Package vectorstore implements read-only Qdrant and Pinecone dense-vector health.
package vectorstore

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"

	"github.com/TheAgentHealth/agenthealth/core"
)

const bodyLimit = 1 << 20

type Adapter struct{}

func (Adapter) Metadata() core.Metadata {
	return core.Metadata{Name: "vector-store", Version: "0.1.0", CompatibilityVersion: "v1", TargetTypes: []string{"vector-store"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: []string{"functional"}}
}
func observation(status core.Status, code string) core.Observation {
	return core.Observation{Check: core.CheckResult{Status: status}, Code: code, ResponseReceived: true}
}

type response struct {
	status int
	body   []byte
}

func endpoint(t core.Target, path string) string {
	u, _ := url.Parse(t.Endpoint)
	u.Path = strings.TrimRight(u.Path, "/") + path
	u.RawPath = ""
	return u.String()
}
func validCredential(s string) bool {
	for i := 0; i < len(s); i++ {
		if (s[i] < 32 && s[i] != '\t') || s[i] == 127 {
			return false
		}
	}
	return true
}
func send(ctx context.Context, r core.Request, method, path string, body []byte) (response, core.Observation, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint(r.Target, path), bytes.NewReader(body))
	if err != nil {
		return response{}, core.Observation{}, &core.Failure{Status: core.Misconfigured}
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if r.Credential != "" {
		req.Header.Set("api-key", r.Credential)
	}
	if r.Target.VectorStore.Provider == "pinecone" {
		req.Header.Set("X-Pinecone-Api-Version", "2025-10")
	}
	mark := func() {
		if r.MarkResponse != nil {
			r.MarkResponse()
		}
	}
	step := func(name string, err error) {
		if r.RecordStep != nil {
			s := core.Healthy
			if err != nil {
				s = core.Unreachable
			}
			r.RecordStep(name, s)
		}
	}
	trace := &httptrace.ClientTrace{GotFirstResponseByte: mark,
		DNSDone:          func(i httptrace.DNSDoneInfo) { step("dns", i.Err) },
		ConnectDone:      func(_, _ string, err error) { step("tcp", err) },
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) { step("tls", err) },
	}
	req = req.WithContext(httptrace.WithClientTrace(ctx, trace))
	resp, err := r.Client.Do(req)
	if err != nil {
		return response{}, core.Observation{}, err
	}
	defer resp.Body.Close()
	mark()
	step("http", nil)
	f := response{status: resp.StatusCode}
	// Never parse or expose error bodies, which may contain credentials or data.
	if f.status < 200 || f.status >= 300 {
		return f, observation(core.Healthy, ""), nil
	}
	f.body, err = io.ReadAll(io.LimitReader(resp.Body, bodyLimit+1))
	if err != nil {
		return response{}, core.Observation{ResponseReceived: true}, err
	}
	if len(f.body) > bodyLimit {
		return response{}, observation(core.Unknown, "vector_limit"), nil
	}
	return f, observation(core.Healthy, ""), nil
}
func failure(status int) (core.Observation, bool) {
	switch {
	case status >= 200 && status < 300:
		return core.Observation{}, false
	case status == 401 || status == 403:
		return observation(core.Misconfigured, "vector_auth"), true
	case status == 404:
		return observation(core.Unhealthy, "vector_missing"), true
	case status == 429:
		return observation(core.Degraded, "vector_rate_limit"), true
	default:
		return observation(core.Unhealthy, "vector_http"), true
	}
}
func (Adapter) Check(ctx context.Context, r core.Request, dimension string) (core.Observation, error) {
	if !core.ValidVectorStore(r.Target) || !validCredential(r.Credential) {
		return core.Observation{Check: core.CheckResult{Status: core.Misconfigured}}, nil
	}
	if dimension == "configuration" {
		return core.Observation{Check: core.CheckResult{Status: core.Healthy}}, nil
	}
	o := r.Target.VectorStore
	var f response
	if cached, ok := r.RunState.Load("vector-metadata"); ok {
		f = cached.(response)
	} else {
		method, path, body := http.MethodGet, "/collections/"+o.Collection, []byte(nil)
		if o.Provider == "pinecone" {
			method, path, body = http.MethodPost, "/describe_index_stats", []byte(`{}`)
		}
		var obs core.Observation
		var err error
		f, obs, err = send(ctx, r, method, path, body)
		if err != nil || obs.Check.Status != core.Healthy {
			return obs, err
		}
		r.RunState.Store("vector-metadata", f)
	}
	if dimension == "reachability" {
		return observation(core.Healthy, ""), nil
	}
	if obs, failed := failure(f.status); failed {
		return obs, nil
	}
	if dimension == "authentication" {
		return observation(core.Healthy, ""), nil
	}
	size, status, ok := metadata(f.body, o)
	if !ok {
		return observation(core.Unhealthy, "vector_protocol"), nil
	}
	if dimension == "capability" {
		if o.Provider == "pinecone" && o.Namespace != "" {
			var m struct {
				Namespaces map[string]json.RawMessage `json:"namespaces"`
			}
			_ = json.Unmarshal(f.body, &m)
			if _, exists := m.Namespaces[o.Namespace]; !exists {
				return observation(core.Unhealthy, "vector_missing"), nil
			}
		}
		if status == "red" {
			return observation(core.Unhealthy, "vector_collection"), nil
		}
		if status == "yellow" || status == "grey" {
			return observation(core.Degraded, "vector_collection"), nil
		}
	}
	if dimension != "functional" {
		return observation(core.Healthy, ""), nil
	}
	if o.Functional == nil || !o.Functional.Safe {
		return observation(core.Misconfigured, ""), nil
	}
	if len(o.Functional.Vector) != size {
		return observation(core.Misconfigured, "vector_dimension"), nil
	}
	path, query := "/collections/"+o.Collection+"/points/query", map[string]any{"query": o.Functional.Vector, "limit": 1, "with_payload": false, "with_vector": false}
	if o.Provider == "pinecone" {
		path = "/query"
		query = map[string]any{"vector": o.Functional.Vector, "topK": 1, "namespace": o.Namespace, "includeMetadata": false, "includeValues": false}
	}
	body, err := json.Marshal(query)
	if err != nil {
		return observation(core.Misconfigured, ""), nil
	}
	f, obs, err := send(ctx, r, http.MethodPost, path, body)
	if err != nil || obs.Check.Status != core.Healthy {
		return obs, err
	}
	if obs, failed := failure(f.status); failed {
		return obs, nil
	}
	if !queryValid(f.body, o.Provider) {
		return observation(core.Unhealthy, "vector_protocol"), nil
	}
	return observation(core.Healthy, ""), nil
}
func metadata(body []byte, o *core.VectorStoreOptions) (int, string, bool) {
	if o.Provider == "pinecone" {
		var m struct {
			Dimension  int    `json:"dimension"`
			Total      *int64 `json:"totalVectorCount"`
			Namespaces map[string]struct {
				Count *int64 `json:"vectorCount"`
			} `json:"namespaces"`
		}
		if json.Unmarshal(body, &m) != nil || m.Dimension < 1 || m.Total == nil || *m.Total < 0 {
			return 0, "", false
		}
		for _, n := range m.Namespaces {
			if n.Count == nil || *n.Count < 0 {
				return 0, "", false
			}
		}
		// An absent namespace is valid for an empty index; query checks it explicitly.
		return m.Dimension, "", true
	}
	var m struct {
		Status string `json:"status"`
		Result *struct {
			Status string `json:"status"`
			Config struct {
				Params struct {
					Vectors struct {
						Size     int    `json:"size"`
						Distance string `json:"distance"`
					} `json:"vectors"`
				} `json:"params"`
			} `json:"config"`
		} `json:"result"`
	}
	if json.Unmarshal(body, &m) != nil || m.Status != "ok" || m.Result == nil {
		return 0, "", false
	}
	x := m.Result
	if x.Status != "green" && x.Status != "yellow" && x.Status != "red" && x.Status != "grey" {
		return 0, "", false
	}
	if x.Config.Params.Vectors.Size < 1 || x.Config.Params.Vectors.Distance == "" {
		return 0, "", false
	}
	return x.Config.Params.Vectors.Size, x.Status, true
}

type hit struct {
	ID    json.RawMessage `json:"id"`
	Score *float64        `json:"score"`
}

func queryValid(body []byte, provider string) bool {
	var hits *[]hit
	if provider == "pinecone" {
		var m struct {
			Matches *[]hit `json:"matches"`
		}
		if json.Unmarshal(body, &m) != nil {
			return false
		}
		hits = m.Matches
	} else {
		var m struct {
			Status string `json:"status"`
			Result *struct {
				Points *[]hit `json:"points"`
			} `json:"result"`
		}
		if json.Unmarshal(body, &m) != nil || m.Status != "ok" || m.Result == nil {
			return false
		}
		hits = m.Result.Points
	}
	if hits == nil || len(*hits) > 1 {
		return false
	}
	for _, h := range *hits {
		if h.Score == nil || math.IsNaN(*h.Score) || math.IsInf(*h.Score, 0) {
			return false
		}
		if len(h.ID) == 0 || bytes.Equal(bytes.TrimSpace(h.ID), []byte("null")) {
			return false
		}
		var id string
		var n uint64
		if json.Unmarshal(h.ID, &id) == nil && id != "" {
			continue
		}
		if provider != "qdrant" || json.Unmarshal(h.ID, &n) != nil {
			return false
		}
	}
	return true
}

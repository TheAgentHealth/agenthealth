package vectorstore

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

const qMetadata = `{"status":"ok","result":{"status":"green","config":{"params":{"vectors":{"size":3,"distance":"Cosine"}}}}}`
const pMetadata = `{"dimension":3,"totalVectorCount":0,"namespaces":{}}`

func run(t *testing.T, target core.Target) core.Result {
	t.Helper()
	reg := core.NewRegistry()
	if err := reg.Register(Adapter{}); err != nil {
		t.Fatal(err)
	}
	results, err := core.NewEngine(reg).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	return results[0]
}
func target(endpoint, provider string, active bool) core.Target {
	o := &core.VectorStoreOptions{Provider: provider}
	if provider == "qdrant" {
		o.Collection = "docs"
	}
	t := core.Target{Name: "retrieval", Type: "vector-store", Endpoint: endpoint, VectorStore: o}
	if active {
		t.Checks = []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional", "latency"}
		o.Functional = &core.VectorQuery{Safe: true, Vector: []float64{1, 0, 0}}
	}
	return t
}
func TestProviderMetadataAndBoundedQuery(t *testing.T) {
	for _, provider := range []string{"qdrant", "pinecone"} {
		t.Run(provider, func(t *testing.T) {
			t.Setenv("VECTOR_KEY", "private-key")
			var metadataCalls, queries atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("api-key") != "private-key" || r.Header.Get("Authorization") != "" {
					t.Error("incorrect credential header")
				}
				if provider == "pinecone" && r.Header.Get("X-Pinecone-Api-Version") != "2025-10" {
					t.Error("missing API version")
				}
				w.Header().Set("Content-Type", "application/json")
				if r.URL.Path == "/prefix/collections/docs" || r.URL.Path == "/prefix/describe_index_stats" {
					metadataCalls.Add(1)
					if provider == "qdrant" {
						if r.Method != "GET" {
							t.Error(r.Method)
						}
						w.Write([]byte(qMetadata))
					} else {
						if r.Method != "POST" {
							t.Error(r.Method)
						}
						w.Write([]byte(pMetadata))
					}
					return
				}
				queries.Add(1)
				if r.Method != "POST" {
					t.Error(r.Method)
				}
				var body map[string]any
				if json.NewDecoder(r.Body).Decode(&body) != nil {
					t.Error("invalid query")
				}
				if provider == "qdrant" {
					if r.URL.Path != "/prefix/collections/docs/points/query" || body["limit"] != float64(1) || body["with_payload"] != false || body["with_vector"] != false {
						t.Errorf("unsafe query %v", body)
					}
					w.Write([]byte(`{"status":"ok","result":{"points":[]}}`))
				} else {
					if r.URL.Path != "/prefix/query" || body["topK"] != float64(1) || body["includeMetadata"] != false || body["includeValues"] != false {
						t.Errorf("unsafe query %v", body)
					}
					w.Write([]byte(`{"matches":[]}`))
				}
			}))
			defer server.Close()
			v := target(server.URL+"/prefix", provider, false)
			v.Auth = &core.AuthReference{BearerEnv: "VECTOR_KEY"}
			if result := run(t, v); result.Status != core.Healthy {
				t.Fatalf("passive: %+v", result)
			}
			if queries.Load() != 0 || metadataCalls.Load() != 1 {
				t.Fatal("passive query or duplicate metadata")
			}
			v = target(server.URL+"/prefix", provider, true)
			v.Auth = &core.AuthReference{BearerEnv: "VECTOR_KEY"}
			if result := run(t, v); result.Status != core.Healthy {
				t.Fatalf("active: %+v", result)
			}
			if queries.Load() != 1 || metadataCalls.Load() != 2 {
				t.Fatal("duplicate requests")
			}
		})
	}
}
func TestStatusesAndProtocol(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   core.Status
		code   string
	}{
		{"auth", 401, "private", core.Misconfigured, "vector_auth"},
		{"forbidden", 403, "private", core.Misconfigured, "vector_auth"},
		{"missing", 404, "private", core.Unhealthy, "vector_missing"},
		{"rate", 429, "private", core.Degraded, "vector_rate_limit"},
		{"server", 503, "private", core.Unhealthy, "vector_http"},
		{"redirect", 302, "private", core.Unhealthy, "vector_http"},
		{"json", 200, `{}`, core.Unhealthy, "vector_protocol"},
		{"null", 200, `null`, core.Unhealthy, "vector_protocol"},
		{"named", 200, `{"status":"ok","result":{"status":"green","config":{"params":{"vectors":{"named":{"size":3}}}}}}`, core.Unhealthy, "vector_protocol"},
		{"red", 200, strings.Replace(qMetadata, "green", "red", 1), core.Unhealthy, "vector_collection"},
		{"yellow", 200, strings.Replace(qMetadata, "green", "yellow", 1), core.Degraded, "vector_collection"},
		{"large", 200, strings.Repeat("x", bodyLimit+1), core.Unknown, "vector_limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); w.Write([]byte(tc.body)) }))
			defer server.Close()
			result := run(t, target(server.URL, "qdrant", false))
			if result.Status != tc.want {
				t.Fatalf("%+v", result)
			}
			var out bytes.Buffer
			if err := core.WriteJSON(&out, []core.Result{result}); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), tc.code) || strings.Contains(out.String(), "private") {
				t.Fatal(out.String())
			}
		})
	}
}
func TestFunctionalFailureNeverRetries(t *testing.T) {
	for _, status := range []int{429, 503} {
		var queries atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				w.Write([]byte(qMetadata))
				return
			}
			queries.Add(1)
			w.WriteHeader(status)
		}))
		v := target(server.URL, "qdrant", true)
		retries := 3
		v.Retries = &retries
		result := run(t, v)
		server.Close()
		if result.Status == core.Healthy || queries.Load() != 1 {
			t.Fatalf("%+v calls=%d", result, queries.Load())
		}
	}
}
func TestQueryResponseValidation(t *testing.T) {
	for _, tc := range []struct {
		provider, body string
		valid          bool
	}{
		{"qdrant", `{"status":"ok","result":{"points":[{"id":1,"score":0.9}]}}`, true},
		{"pinecone", `{"matches":[{"id":"doc","score":0.9}]}`, true},
		{"qdrant", `{"status":"ok","result":{"points":null}}`, false},
		{"pinecone", `{"matches":null}`, false},
		{"pinecone", `{"matches":[{"id":null,"score":1}]}`, false},
		{"pinecone", `{"matches":[{"id":"doc"}]}`, false},
		{"qdrant", `{"status":"error","result":{"points":[]}}`, false},
		{"qdrant", `{"status":"ok","result":{"points":[{"id":{},"score":1}]}}`, false},
		{"pinecone", `{"matches":[{"id":"a","score":1},{"id":"b","score":1}]}`, false},
	} {
		if queryValid([]byte(tc.body), tc.provider) != tc.valid {
			t.Errorf("%s: %s", tc.provider, tc.body)
		}
	}
}
func TestTLSRedirectAndDeadline(t *testing.T) {
	var forwarded atomic.Int32
	sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
	defer sink.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, sink.URL, 302) }))
	defer redirect.Close()
	if result := run(t, target(redirect.URL, "qdrant", false)); result.Status != core.Unhealthy || forwarded.Load() != 0 {
		t.Fatalf("redirect: %+v", result)
	}
	tls := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(qMetadata)) }))
	defer tls.Close()
	if result := run(t, target(tls.URL, "qdrant", false)); result.Status != core.Unreachable {
		t.Fatalf("TLS: %+v", result)
	}
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer slow.Close()
	v := target(slow.URL, "qdrant", false)
	ms := 20
	v.TimeoutMS = &ms
	start := time.Now()
	if result := run(t, v); result.Status != core.Unreachable {
		t.Fatalf("timeout: %+v", result)
	}
	if time.Since(start) > time.Second {
		t.Fatal("deadline ignored")
	}
}
func TestInvalidCredentialsAndDimensionsAvoidQuery(t *testing.T) {
	for _, credential := range []string{"", "invalid\r\nvalue"} {
		t.Setenv("VECTOR_KEY", credential)
		var calls atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
		v := target(server.URL, "qdrant", false)
		v.Auth = &core.AuthReference{BearerEnv: "VECTOR_KEY"}
		result := run(t, v)
		server.Close()
		if result.Status != core.Misconfigured || calls.Load() != 0 {
			t.Fatalf("%+v", result)
		}
	}
	var queries atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			queries.Add(1)
		}
		w.Write([]byte(qMetadata))
	}))
	defer server.Close()
	v := target(server.URL, "qdrant", true)
	v.VectorStore.Functional.Vector = []float64{1}
	if result := run(t, v); result.Status != core.Misconfigured || queries.Load() != 0 {
		t.Fatalf("dimension: %+v", result)
	}
}

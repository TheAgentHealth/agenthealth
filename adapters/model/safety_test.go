package model

import (
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

func TestTransportSafety(t *testing.T) {
	var forwarded atomic.Int32
	destination := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
	defer destination.Close()
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, destination.URL, http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	tls := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { forwarded.Add(1) }))
	defer tls.Close()
	t.Setenv("MODEL_SECRET", "private-key")
	for _, endpoint := range []string{redirect.URL, tls.URL} {
		result := run(t, core.Target{Name: "model", Type: "model", Endpoint: endpoint, Auth: &core.AuthReference{BearerEnv: "MODEL_SECRET"}})
		if result.Status == core.Healthy || forwarded.Load() != 0 {
			t.Fatalf("unsafe request: %+v", result)
		}
	}
}

func TestPartialResponseIsNeverRetried(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Write([]byte(`{"data":`))
		w.(http.Flusher).Flush()
		<-r.Context().Done()
	}))
	defer server.Close()
	retries, timeout := 3, 30
	result := run(t, core.Target{Name: "model", Type: "model", Endpoint: server.URL, Retries: &retries, TimeoutMS: &timeout})
	if result.Status != core.Unknown || calls.Load() != 1 {
		t.Fatalf("partial response: %+v calls=%d", result, calls.Load())
	}
}

func TestInferenceDeadlineNeverRetries(t *testing.T) {
	var posts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(models("local"))
			return
		}
		posts.Add(1)
		// Read the body so the server can observe client cancellation.
		var request any
		json.NewDecoder(r.Body).Decode(&request)
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	defer server.Close()
	retries, timeout := 3, 30
	result := run(t, core.Target{Name: "model", Type: "model", Endpoint: server.URL, Retries: &retries, TimeoutMS: &timeout, Checks: []string{"functional"}, Model: &core.ModelOptions{Functional: &core.ModelInference{Safe: true, Model: "local", Prompt: "OK"}}})
	if result.Checks["functional"].Status == core.Healthy || posts.Load() != 1 {
		t.Fatalf("inference: %+v posts=%d", result, posts.Load())
	}
}

func TestMalformedInferenceCannotPass(t *testing.T) {
	for _, api := range []string{"openai", "anthropic"} {
		bodies := []string{`null`, `{}`, `{"choices":[null]}`, `{"choices":[{}]}`, `{"choices":[{"message":null}]}`, `{"choices":[{"message":{"role":"user"}}]}`, `{"type":"message","role":"assistant","content":[]}`, `{"type":"message","role":"assistant","content":[null]}`}
		for _, body := range bodies {
			t.Run(api+body, func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method == http.MethodGet {
						json.NewEncoder(w).Encode(models("local"))
						return
					}
					w.Write([]byte(body))
				}))
				defer server.Close()
				result := run(t, core.Target{Name: "model", Type: "model", Endpoint: server.URL, Checks: []string{"functional"}, Model: &core.ModelOptions{API: api, Functional: &core.ModelInference{Safe: true, Model: "local", Prompt: "OK"}}})
				if result.Checks["functional"].Code != "model_protocol" {
					t.Fatalf("accepted malformed response: %+v", result)
				}
			})
		}
	}
}

func TestSharedModelGraphNodeRunsOncePerRun(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1); json.NewEncoder(w).Encode(models("local")) }))
	defer server.Close()
	config, err := core.LoadConfig(strings.NewReader("version: v1\ntargets:\n- name: first\n  type: model\n  endpoint: " + server.URL + "\n  checks: []\n  dependencies:\n  - ref: shared\n- name: second\n  type: model\n  endpoint: " + server.URL + "\n  checks: []\n  dependencies:\n  - ref: shared\n- id: shared\n  name: model\n  type: model\n  endpoint: " + server.URL + "\n"))
	if err != nil {
		t.Fatal(err)
	}
	registry := core.NewRegistry()
	registry.Register(Adapter{})
	engine := core.NewEngine(registry)
	for i := int32(1); i <= 2; i++ {
		results, err := engine.Run(context.Background(), config)
		if err != nil || calls.Load() != i || results[0].Dependencies[0].Status != core.Healthy || results[1].Dependencies[0].Status != core.Healthy {
			t.Fatalf("graph: %+v err=%v calls=%d", results, err, calls.Load())
		}
	}
}

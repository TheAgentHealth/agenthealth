package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/TheAgentHealth/agenthealth/core"
)

func TestAgentHealth(t *testing.T) {
	for _, targetType := range []string{"agent", "multi-agent"} {
		for _, tc := range []struct {
			name, doc string
			want      core.Status
		}{
			{"healthy", `{"version":"v1","name":"first","live":true,"ready":true,"capabilities":["answer"],"dependencies":[]}`, core.Healthy},
			{"unready", `{"version":"v1","name":"first","live":true,"ready":false,"capabilities":["answer"],"dependencies":[]}`, core.Unhealthy},
			{"missing metadata", `{}`, core.Unhealthy},
			{"missing capability", `{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":[]}`, core.Unhealthy},
			{"unconfigured discovery", `{"version":"v1","name":"first","live":true,"ready":true,"capabilities":["answer"],"dependencies":["unknown"]}`, core.Unknown},
		} {
			t.Run(targetType+"/"+tc.name, func(t *testing.T) {
				var gets, posts atomic.Int32
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					if r.Method == "GET" {
						gets.Add(1)
					}
					if r.Method == "POST" {
						posts.Add(1)
					}
					fmt.Fprint(w, tc.doc)
				}))
				defer server.Close()
				registry := core.NewRegistry()
				if err := registry.Register(Adapter{}); err != nil {
					t.Fatal(err)
				}
				results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "first", Type: targetType, Endpoint: server.URL, Agent: &core.AgentOptions{RequiredCapabilities: []string{"answer"}}}}})
				if err != nil {
					t.Fatal(err)
				}
				if results[0].Status != tc.want {
					t.Fatalf("%+v", results[0])
				}
				if gets.Load() != 1 || posts.Load() != 0 {
					t.Fatalf("gets %d posts %d", gets.Load(), posts.Load())
				}
			})
		}
	}
}
func TestFunctionalAndPath(t *testing.T) {
	for _, tc := range []struct {
		body string
		want core.Status
	}{
		{`{"completed":true,"success":true,"downstream":"peer"}`, core.Healthy},
		{`{"completed":true,"success":true,"downstream":"wrong"}`, core.Unhealthy},
		{`{"completed":false}`, core.Unknown},
		{`{"completed":true,"success":false,"downstream":"peer"}`, core.Unhealthy},
		{`{"completed":true}`, core.Unknown},
	} {
		var posts atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if r.Method == "POST" {
				posts.Add(1)
				fmt.Fprint(w, tc.body)
			} else {
				fmt.Fprint(w, `{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":[]}`)
			}
		}))
		retries := 3
		registry := core.NewRegistry()
		registry.Register(Adapter{})
		results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "path", Type: "agent", Endpoint: server.URL, Checks: []string{"functional"}, Retries: &retries, Agent: &core.AgentOptions{Functional: &core.AgentTask{Safe: true, Text: "health probe", Downstream: "peer"}}}}})
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		if results[0].Status != tc.want || posts.Load() != 1 {
			t.Fatalf("%+v posts=%d", results, posts.Load())
		}
	}
}
func TestResponseBoundsAndAuth(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   core.Status
	}{
		{401, `secret`, core.Misconfigured}, {503, `secret`, core.Unhealthy}, {200, string(make([]byte, 65537)), core.Unknown}, {200, `secret`, core.Unhealthy},
	} {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(tc.status)
			fmt.Fprint(w, tc.body)
		}))
		client := core.NewHTTPClient()
		o, err := (Adapter{}).Check(context.Background(), core.Request{Target: core.Target{Endpoint: s.URL}, Client: client}, "protocol")
		client.CloseIdleConnections()
		s.Close()
		if err != nil || o.Check.Status != tc.want {
			t.Fatalf("%+v %v", o, err)
		}
	}
}

func TestIndependentPeerAndPathEvidence(t *testing.T) {
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" {
			w.Write([]byte(`{"completed":true,"success":false,"downstream":"peer"}`))
			return
		}
		w.Write([]byte(`{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":["peer"]}`))
	}))
	defer first.Close()
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"v1","name":"peer","live":true,"ready":true,"capabilities":[],"dependencies":[]}`))
	}))
	defer peer.Close()
	for _, critical := range []bool{true, false} {
		registry := core.NewRegistry()
		registry.Register(Adapter{})
		results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "first", Type: "agent", Endpoint: first.URL, Dependencies: []core.Dependency{
			{Target: core.Target{Name: "peer", Type: "agent", Endpoint: peer.URL}},
			{Critical: &critical, Target: core.Target{Name: "path", Type: "agent", Endpoint: first.URL, Checks: []string{"functional"}, Agent: &core.AgentOptions{Functional: &core.AgentTask{Safe: true, Text: "safe health probe", Downstream: "peer"}}}},
		}}}})
		if err != nil {
			t.Fatal(err)
		}
		want := core.Unhealthy
		if !critical {
			want = core.Degraded
		}
		r := results[0]
		if r.Status != want || r.Checks["protocol"].Status != core.Healthy || r.Dependencies[0].Status != core.Healthy || r.Dependencies[1].Status != core.Unhealthy {
			t.Fatalf("%+v", r)
		}
	}
}
func TestMissingTaskFailsBeforeNetworking(t *testing.T) {
	registry := core.NewRegistry()
	registry.Register(Adapter{})
	results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "legacy", Type: "agent", Endpoint: "http://127.0.0.1:1", Checks: []string{"functional"}}}})
	if err != nil || results[0].Status != core.Misconfigured {
		t.Fatalf("%+v %v", results, err)
	}
}

func TestFirstRuntimeContactsSelectedPeer(t *testing.T) {
	var peerCalls atomic.Int32
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.Header.Get("X-Probe-Origin") != "first-runtime" {
			t.Error("probe bypassed first runtime")
		}
		peerCalls.Add(1)
		w.Write([]byte(`{"completed":true,"success":true}`))
	}))
	defer peer.Close()
	first := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "POST" {
			w.Write([]byte(`{"version":"v1","name":"first","live":true,"ready":true,"capabilities":[],"dependencies":[]}`))
			return
		}
		var task core.AgentTask
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil || task.Downstream != "peer" || !task.Safe {
			t.Error("invalid path selector")
			w.WriteHeader(400)
			return
		}
		// Simulate the runtime's configured direct integration, not remote discovery.
		request, err := http.NewRequestWithContext(r.Context(), "POST", peer.URL, strings.NewReader(`{"safe":true,"text":"health"}`))
		if err != nil {
			t.Error(err)
			w.WriteHeader(500)
			return
		}
		request.Header.Set("X-Probe-Origin", "first-runtime")
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			w.Write([]byte(`{"completed":true,"success":false,"downstream":"peer"}`))
			return
		}
		defer response.Body.Close()
		w.Write([]byte(`{"completed":true,"success":true,"downstream":"peer"}`))
	}))
	defer first.Close()
	registry := core.NewRegistry()
	registry.Register(Adapter{})
	results, err := core.NewEngine(registry).Run(context.Background(), core.Config{Version: "v1", Targets: []core.Target{{Name: "first-to-peer", Type: "agent", Endpoint: first.URL, Checks: []string{"functional"}, Agent: &core.AgentOptions{Functional: &core.AgentTask{Safe: true, Text: "safe health probe", Downstream: "peer"}}}}})
	if err != nil || results[0].Status != core.Healthy || peerCalls.Load() != 1 {
		t.Fatalf("%+v calls=%d err=%v", results, peerCalls.Load(), err)
	}
}

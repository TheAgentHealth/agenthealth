package http

import (
	"github.com/TheAgentHealth/agenthealth/core"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestConfiguredTimeoutExceedsFiveSeconds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { time.Sleep(5200 * time.Millisecond); w.WriteHeader(204) }))
	defer server.Close()
	timeout := 7000
	result := execute(t, core.Target{Name: "slow", Type: "http", Endpoint: server.URL, Checks: []string{"reachability"}, CheckTimeouts: map[string]int{"reachability": timeout}})
	if result.Status != core.Healthy {
		t.Fatal("fixed transport cap overrode configured deadline", result)
	}
}

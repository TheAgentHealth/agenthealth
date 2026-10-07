package http

import (
	"github.com/TheAgentHealth/agenthealth/core"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestInvalidCredentialControlsFailBeforeNetworking(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(204) }))
	defer server.Close()
	for _, b := range []byte{1, 7, 8, 10, 13, 31, 127} {
		t.Setenv("AGENTHEALTH_HTTP_TOKEN", "token"+string(b))
		retries := 3
		result := execute(t, core.Target{Name: "bad-header", Type: "http", Endpoint: server.URL, Retries: &retries, Auth: &core.AuthReference{BearerEnv: "AGENTHEALTH_HTTP_TOKEN"}})
		if result.Status != core.Misconfigured || calls.Load() != 0 {
			t.Fatal("invalid header reached network", b, result)
		}
	}
	if !validHeaderValue("token\tvalue") {
		t.Fatal("valid HTTP horizontal tab rejected")
	}
}

package core

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAHPHTTPContract(t *testing.T) {
	config := Config{Version: "v1", Targets: []Target{{Name: "private-agent-name", Type: "http", Endpoint: "http://localhost", Checks: []string{}}}}
	s, err := NewAHPServer(NewEngine(NewRegistry()), config, AHPOptions{Token: "1234567890123456"})
	if err != nil {
		t.Fatal(err)
	}
	request := func(path, token string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("GET", path, nil)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		return w
	}
	if w := request("/ready", ""); w.Code != 503 {
		t.Fatal(w.Code)
	}
	if w := request("/health/dependencies", ""); w.Code != 401 || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.Run(ctx); close(done) }()
	defer func() { cancel(); <-done }()
	deadline := time.Now().Add(time.Second)
	for {
		s.mu.RLock()
		ready := !s.updated.IsZero()
		s.mu.RUnlock()
		if ready {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("no snapshot")
		}
		time.Sleep(time.Millisecond)
	}
	if w := request("/live", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := request("/health", ""); strings.Contains(w.Body.String(), "private") || strings.Contains(w.Body.String(), "evidence") {
		t.Fatal(w.Body.String())
	}
	w := request("/health/dependencies", "1234567890123456")
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body["evidence"] == nil {
		t.Fatal(w.Body.String(), err)
	}
	s.mu.Lock()
	s.updated = time.Now().Add(-time.Hour)
	s.mu.Unlock()
	if w := request("/health/dependencies", "1234567890123456"); w.Code != 503 || strings.Contains(w.Body.String(), "private") {
		t.Fatal(w.Body.String())
	}
	for _, path := range []string{"/missing", "/ready"} {
		r := httptest.NewRequest("POST", path, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		if w.Code != 404 && w.Code != 405 {
			t.Fatal(w.Code)
		}
	}
}

func TestAHPOptions(t *testing.T) {
	c := Config{Version: "v1", Targets: []Target{{Name: "test", Type: "http", Endpoint: "http://localhost"}}}
	for _, o := range []AHPOptions{{Token: "short"}, {Interval: time.Millisecond}, {Interval: time.Minute, MaxAge: time.Second}} {
		if _, err := NewAHPServer(NewEngine(NewRegistry()), c, o); err == nil {
			t.Fatal("accepted invalid options")
		}
	}
}

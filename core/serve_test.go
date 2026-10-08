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

func TestAHPStatusMappingAndGraphEvidence(t *testing.T) {
	config := Config{Version: "v1", Targets: []Target{{Name: "test", Type: "http", Endpoint: "http://localhost"}}}
	s, err := NewAHPServer(NewEngine(NewRegistry()), config, AHPOptions{Token: "1234567890123456"})
	if err != nil {
		t.Fatal(err)
	}
	optional := false
	result := Result{Target: TargetIdentity{ID: "direct", Name: "direct", Type: "agent"}, Status: Healthy, Checks: map[string]CheckResult{"capability": {Status: Healthy}, "functional": {Status: Unhealthy}}, Dependencies: []Result{{Target: TargetIdentity{ID: "peer", Name: "peer", Type: "a2a"}, Relationship: "downstream", Critical: &optional, Status: Unknown, Checks: map[string]CheckResult{}, Dependencies: []Result{}}}}
	var snapshot strings.Builder
	if err := WriteJSON(&snapshot, []Result{result}); err != nil {
		t.Fatal(err)
	}
	s.running = true
	s.updated = time.Now()
	s.snapshot = []byte(snapshot.String())
	for _, status := range statuses {
		s.status = status
		r := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, r)
		want := 503
		if status == Healthy || status == Degraded {
			want = 200
		}
		if w.Code != want {
			t.Fatalf("%s: %d", status, w.Code)
		}
	}
	r := httptest.NewRequest("GET", "/health/capabilities", nil)
	r.Header.Set("Authorization", "Bearer 1234567890123456")
	w := httptest.NewRecorder()
	s.ServeHTTP(w, r)
	var envelope struct{ Evidence struct{ Result } }
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	got := envelope.Evidence.Result
	if got.Target.ID != "direct" || got.Checks["capability"].Status != Healthy || got.Checks["functional"].Status != Unhealthy || len(got.Dependencies) != 1 || got.Dependencies[0].Target.ID != "peer" || got.Dependencies[0].Relationship != "downstream" || got.Dependencies[0].Critical == nil || *got.Dependencies[0].Critical {
		t.Fatal(w.Body.String())
	}
	r.Header.Set("Authorization", "Bearer incorrect-token")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatal(w.Code)
	}
	r = httptest.NewRequest("GET", "/health", nil)
	r.Header.Set("AHP-Version", "v2")
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	s.running = false
	r = httptest.NewRequest("GET", "/live", nil)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, r)
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}

func TestAHPBearerTokenValidation(t *testing.T) {
	config := Config{Version: "v1", Targets: []Target{{Name: "test", Type: "http", Endpoint: "http://localhost"}}}
	for _, token := range []string{"1234567890123456 ", " 1234567890123456", "1234567890123456\t", "12345678 90123456", "1234567890123456\n", "1234567890123456é", "12345678=90123456", "================"} {
		if _, err := NewAHPServer(NewEngine(NewRegistry()), config, AHPOptions{Token: token}); err == nil {
			t.Fatal("accepted invalid bearer token")
		}
	}
	for _, token := range []string{"1234567890123456", "aAzZ019-._~+/token=="} {
		s, err := NewAHPServer(NewEngine(NewRegistry()), config, AHPOptions{Token: token})
		if err != nil {
			t.Fatal(err)
		}
		if !s.authorized("Bearer "+token) || !s.authorized("bearer "+token) || !s.authorized("bEaReR  "+token) {
			t.Fatal("valid token denied")
		}
		for _, header := range []string{"", "Bearer short", "Bearer " + token + "x", "Bearer " + strings.Repeat("x", len(token)), "Bearer " + token + " ", "Basic " + token} {
			if s.authorized(header) {
				t.Fatal("invalid header authorized")
			}
		}
	}
}

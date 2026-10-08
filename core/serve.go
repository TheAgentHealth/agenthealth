package core

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

// AHPOptions controls the experimental v1 HTTP binding. Token grants access to
// topology and evidence; it must be resolved by the caller from a secret reference.
type AHPOptions struct {
	Token    string
	Interval time.Duration
	MaxAge   time.Duration
}

// AHPServer runs checks independently of requests and serves immutable snapshots.
// Construct one server per configuration; Run must remain active while serving.
type AHPServer struct {
	engine   *Engine
	config   Config
	options  AHPOptions
	mu       sync.RWMutex
	snapshot []byte
	status   Status
	updated  time.Time
	running  bool
}

func NewAHPServer(engine *Engine, config Config, options AHPOptions) (*AHPServer, error) {
	if engine == nil {
		return nil, errors.New("AHP requires an engine")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if options.Interval == 0 {
		options.Interval = 30 * time.Second
	}
	if options.MaxAge == 0 {
		options.MaxAge = 2 * time.Minute
	}
	if options.Interval < time.Second || options.Interval > time.Hour || options.MaxAge < options.Interval || options.MaxAge > time.Hour {
		return nil, errors.New("invalid AHP refresh policy")
	}
	if options.Token != "" && (len(options.Token) < 16 || !validHTTPHeaderValue(options.Token)) {
		return nil, errors.New("invalid AHP bearer token")
	}
	return &AHPServer{engine: engine, config: config, options: options, status: Unknown}, nil
}

// Run refreshes serially, without overlapping runs or request-triggered probes.
func (s *AHPServer) Run(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	defer func() { s.mu.Lock(); s.running = false; s.mu.Unlock() }()
	for {
		results, err := s.engine.Run(ctx, s.config)
		var b bytes.Buffer
		if err == nil {
			for i := range results {
				results[i] = NewRedactor(s.options.Token).Result(results[i])
			}
			err = WriteJSON(&b, results)
		}
		if err == nil && ctx.Err() == nil {
			states := make([]Status, len(results))
			for i, r := range results {
				states[i] = r.Status
			}
			s.mu.Lock()
			s.snapshot = append([]byte(nil), b.Bytes()...)
			s.status = Worst(states...)
			s.updated = time.Now()
			s.mu.Unlock()
		}
		timer := time.NewTimer(s.options.Interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (s *AHPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("AHP-Version", "v1")
	writeError := func(code int, name string) {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{"ahp_version": "v1", "error": name})
	}
	switch r.URL.Path {
	case "/health", "/ready", "/live", "/health/dependencies", "/health/capabilities":
	default:
		writeError(404, "not_found")
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(405, "method_not_allowed")
		return
	}
	if v := r.Header.Get("AHP-Version"); v != "" && v != "v1" {
		writeError(400, "unsupported_version")
		return
	}
	detailed := r.URL.Path == "/health/dependencies" || r.URL.Path == "/health/capabilities"
	if detailed && (s.options.Token == "" || subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.options.Token)) != 1) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		writeError(401, "unauthorized")
		return
	}
	s.mu.RLock()
	snapshot := s.snapshot
	status := s.status
	updated := s.updated
	running := s.running
	s.mu.RUnlock()
	code := http.StatusOK
	if r.URL.Path == "/live" {
		status = Healthy
		if !running {
			status = Unknown
			code = 503
		}
	} else if !running || updated.IsZero() || time.Since(updated) > s.options.MaxAge {
		status = Unknown
		code = 503
		snapshot = nil
	} else if status != Healthy && status != Degraded {
		code = 503
	}
	if detailed && snapshot == nil {
		writeError(503, "snapshot_unavailable")
		return
	}
	envelope := map[string]any{"ahp_version": "v1", "spec_version": "v1", "status": status}
	if r.URL.Path != "/live" && !updated.IsZero() {
		envelope["observed_at"] = updated.UTC().Format(time.RFC3339Nano)
	}
	if detailed {
		// The existing formatter validates and redacts the full recursive result,
		// retaining IDs, edge policy and separate dimension evidence.
		envelope["evidence"] = json.RawMessage(snapshot)
	}
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(envelope)
}

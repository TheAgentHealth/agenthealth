package core

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Engine executes targets in declaration order with a bounded adapter-call
// pool. Construct it with NewEngine; share it across concurrent runs if needed.
type Engine struct {
	registry *Registry
	slots    chan struct{}
}

func NewEngine(registry *Registry) *Engine {
	return &Engine{registry: registry, slots: make(chan struct{}, 16)}
}

// Run applies a 60-second whole-run safety budget in addition to per-check
// deadlines. A shorter caller deadline takes precedence. Dependencies execute
// independently even when a parent's prerequisites fail.
func (e *Engine) Run(ctx context.Context, config Config) ([]Result, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	credentials := map[string]string{}
	secrets := []string{}
	var collect func(Target)
	collect = func(t Target) {
		if t.Auth != nil {
			value := os.Getenv(t.Auth.BearerEnv)
			credentials[t.Auth.BearerEnv] = value
			secrets = append(secrets, value)
		}
		for _, d := range t.Dependencies {
			collect(d.Target)
		}
	}
	for _, t := range config.Targets {
		collect(t)
	}
	redactor := NewRedactor(secrets...)
	client := NewHTTPClient()
	defer client.CloseIdleConnections()
	results := make([]Result, 0, len(config.Targets))
	for _, t := range config.Targets {
		results = append(results, redactor.Result(e.runTarget(ctx, t, client, credentials)))
	}
	return results, nil
}
func passed(s Status) bool { return s == Healthy || s == Degraded }
func (e *Engine) runTarget(ctx context.Context, t Target, client *http.Client, credentials map[string]string) Result {
	result := Result{Target: TargetIdentity{Name: t.Name, Type: t.Type}, Checks: map[string]CheckResult{}, Dependencies: []Result{}}
	a := e.registry.lookup(t.Type)
	requested := t.Checks
	var metadata Metadata
	if a != nil {
		metadata = a.Metadata()
		if requested == nil {
			requested = []string{"configuration", "reachability"}
			for _, dimension := range []string{"protocol", "authentication", "capability"} {
				if contains(metadata.Dimensions, dimension) && !contains(metadata.ActiveChecks, dimension) {
					requested = append(requested, dimension)
				}
			}
			requested = append(requested, "latency")
			if t.Type == "agent" {
				requested = append(requested, "dependency")
			}
		}
	} else if requested == nil {
		requested = []string{"configuration"}
	}
	direct := false
	for _, dimension := range requested {
		if dimension != "dependency" {
			direct = true
		}
	}
	if direct {
		request := Request{Target: t, Client: client}
		configCheck := CheckResult{Status: Healthy}
		if a == nil {
			configCheck = CheckResult{Status: Misconfigured, Message: "no adapter registered for target type"}
		}
		if t.Auth != nil {
			request.Credential = credentials[t.Auth.BearerEnv]
			if request.Credential == "" {
				configCheck = CheckResult{Status: Misconfigured, Message: "credential environment variable is missing or empty"}
			}
		}
		for _, d := range requested {
			if d != "latency" && d != "dependency" && !contains(metadata.Dimensions, d) {
				configCheck = CheckResult{Status: Misconfigured, Message: "requested dimension is not supported by adapter"}
			}
		}
		if configCheck.Status == Healthy {
			observation, _ := e.execute(ctx, a, metadata, request, "configuration")
			configCheck = observation.Check
			if configCheck.Status == Unreachable {
				configCheck = CheckResult{Status: Unknown, Message: "configuration check did not complete"}
			}
		}
		if contains(requested, "configuration") || !passed(configCheck.Status) {
			result.Checks["configuration"] = configCheck
		}
		networkNeeded := false
		for _, d := range requested {
			if d != "configuration" && d != "dependency" {
				networkNeeded = true
			}
		}
		if passed(configCheck.Status) && networkNeeded {
			reach, elapsed := e.execute(ctx, a, metadata, request, "reachability")
			if contains(requested, "reachability") || !passed(reach.Check.Status) {
				result.Checks["reachability"] = reach.Check
			}
			if passed(reach.Check.Status) {
				ms := float64(elapsed) / float64(time.Millisecond)
				result.LatencyMS = &ms
				auth := CheckResult{Status: Healthy}
				needsAuth := contains(requested, "authentication") || contains(requested, "capability") || contains(requested, "functional")
				if needsAuth && contains(metadata.Dimensions, "authentication") {
					obs, _ := e.execute(ctx, a, metadata, request, "authentication")
					auth = obs.Check
					if contains(requested, "authentication") || !passed(auth.Status) {
						result.Checks["authentication"] = auth
					}
				}
				for _, d := range []string{"protocol", "capability", "functional"} {
					if !contains(requested, d) || ((d == "capability" || d == "functional") && !passed(auth.Status)) {
						continue
					}
					obs, _ := e.execute(ctx, a, metadata, request, d)
					result.Checks[d] = obs.Check
				}
				if contains(requested, "latency") {
					check := CheckResult{Status: Healthy}
					if t.Thresholds != nil && t.Thresholds.LatencyMS != nil && ms > *t.Thresholds.LatencyMS {
						check = CheckResult{Status: Degraded, Message: fmt.Sprintf("latency exceeds threshold %.2fms", *t.Thresholds.LatencyMS)}
					}
					result.Checks["latency"] = check
				}
			}
		}
	}
	critical := make([]bool, 0, len(t.Dependencies))
	contribution := Healthy
	for _, d := range t.Dependencies {
		dep := e.runTarget(ctx, d.Target, client, credentials)
		result.Dependencies = append(result.Dependencies, dep)
		critical = append(critical, d.IsCritical())
		contribution = Worst(contribution, DependencyContribution(dep.Status, d.IsCritical()))
	}
	if contains(requested, "dependency") {
		result.Checks["dependency"] = CheckResult{Status: contribution}
	}
	result.Status = Aggregate(result.Checks, result.Dependencies, critical)
	return result
}

func (e *Engine) execute(ctx context.Context, a Adapter, m Metadata, request Request, dimension string) (Observation, time.Duration) {
	active := contains(m.ActiveChecks, dimension)
	if active && !contains(request.Target.Checks, dimension) {
		return Observation{Check: CheckResult{Status: Misconfigured, Message: "active check requires explicit opt-in"}}, 0
	}
	var observation Observation
	var elapsed time.Duration
	for attempt := 0; attempt <= request.Target.retryCount(); attempt++ {
		observation, elapsed = e.attempt(ctx, a, request, dimension)
		// Retrying an active operation or a partial response could duplicate work.
		if active || observation.ResponseReceived || observation.Check.Status != Unreachable || attempt == request.Target.retryCount() {
			break
		}
		timer := time.NewTimer(request.Target.retryDelay())
		select {
		case <-ctx.Done():
			timer.Stop()
			return Observation{Check: NormalizeError(ctx.Err(), false)}, elapsed
		case <-timer.C:
		}
	}
	return observation, elapsed
}
func (e *Engine) attempt(ctx context.Context, a Adapter, request Request, dimension string) (Observation, time.Duration) {
	ctx, cancel := context.WithTimeout(ctx, request.Target.timeout(dimension))
	defer cancel()
	var start time.Time
	if ctx.Err() != nil {
		return Observation{Check: NormalizeError(ctx.Err(), false)}, 0
	}
	var received atomic.Bool
	var stepsMu sync.Mutex
	steps := map[string]Status{}
	request.RecordStep = func(name string, status Status) {
		if !contains([]string{"dns", "tcp", "tls", "http"}, name) || !status.Valid() {
			return
		}
		stepsMu.Lock()
		// A successful dial wins over failed parallel address attempts.
		if steps[name] != Healthy {
			steps[name] = status
		}
		stepsMu.Unlock()
	}
	snapshot := func() map[string]Status {
		stepsMu.Lock()
		defer stepsMu.Unlock()
		if len(steps) == 0 {
			return nil
		}
		copy := make(map[string]Status, len(steps))
		for key, value := range steps {
			copy[key] = value
		}
		return copy
	}
	request.MarkResponse = func() { received.Store(true) }
	client := *request.Client
	client.Timeout = request.Target.timeout(dimension)
	request.Client = &client
	select {
	case e.slots <- struct{}{}:
	case <-ctx.Done():
		return Observation{Check: NormalizeError(ctx.Err(), received.Load()), ResponseReceived: received.Load()}, 0
	}
	start = time.Now()
	type answer struct {
		observation Observation
		err         error
	}
	done := make(chan answer, 1)
	go func() {
		defer func() { <-e.slots }()
		response := answer{}
		defer func() {
			if recover() != nil {
				response = answer{err: fmt.Errorf("adapter panic")}
			}
			done <- response
		}()
		response.observation, response.err = a.Check(ctx, request, dimension)
	}()
	select {
	case response := <-done:
		obs := response.observation
		obs.ResponseReceived = obs.ResponseReceived || received.Load()
		if response.err == nil && ctx.Err() != nil {
			response.err = ctx.Err()
		}
		if response.err != nil {
			obs.Check = NormalizeError(response.err, obs.ResponseReceived)
		}
		if !obs.Check.Status.Valid() {
			obs.Check = CheckResult{Status: Unknown, Message: "invalid adapter result"}
		}
		// Use canonical diagnostics only: adapter text may contain unknown secrets.
		if obs.Check.Status != Healthy {
			obs.Check.Message = string(obs.Check.Status) + " in " + dimension
		} else {
			obs.Check.Message = ""
		}
		if response.err == nil {
			if message, ok := diagnosticMessages[obs.Code]; ok {
				obs.Check.Message = message
			}
		}
		obs.Check.Steps = snapshot()
		return obs, time.Since(start)
	case <-ctx.Done():
		check := NormalizeError(ctx.Err(), received.Load())
		check.Steps = snapshot()
		return Observation{Check: check, ResponseReceived: received.Load()}, time.Since(start)
	}
}

var diagnosticMessages = map[string]string{
	"http_status":     "HTTP status did not match expected status",
	"http_headers":    "required HTTP response header did not match",
	"http_body":       "HTTP response body did not contain required text",
	"http_body_limit": "HTTP response body exceeded the configured size limit",
	"http_auth":       "HTTP authentication rejected (401 or 403)",
}

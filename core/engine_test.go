package core

import (
	"bytes"
	"context"
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testAdapter struct {
	mu     sync.Mutex
	calls  []string
	fn     func(context.Context, Request, string) (Observation, error)
	active []string
}

func (a *testAdapter) Metadata() Metadata {
	return Metadata{Name: "test", Version: "1", CompatibilityVersion: "v1", TargetTypes: []string{"custom"}, Dimensions: []string{"configuration", "reachability", "authentication", "protocol", "capability", "functional"}, ActiveChecks: append([]string{"functional"}, a.active...)}
}
func (a *testAdapter) Check(ctx context.Context, r Request, d string) (Observation, error) {
	a.mu.Lock()
	a.calls = append(a.calls, d)
	a.mu.Unlock()
	if a.fn != nil {
		return a.fn(ctx, r, d)
	}
	return Observation{Check: CheckResult{Status: Healthy}, ResponseReceived: d != "configuration"}, nil
}
func runTest(t *testing.T, a *testAdapter, target Target) Result {
	t.Helper()
	registry := NewRegistry()
	if err := registry.Register(a); err != nil {
		t.Fatal(err)
	}
	results, err := NewEngine(registry).Run(context.Background(), Config{Version: "v1", Targets: []Target{target}})
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateResult(results[0]); err != nil {
		t.Fatal(err)
	}
	return results[0]
}
func targetForTest(checks ...string) Target {
	return Target{Name: "target", Type: "custom", Endpoint: "test", Checks: checks}
}
func TestPrerequisiteGates(t *testing.T) {
	for _, failed := range []string{"configuration", "reachability", "authentication"} {
		t.Run(failed, func(t *testing.T) {
			a := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
				status := Healthy
				if d == failed {
					status = Misconfigured
					if d == "reachability" {
						status = Unreachable
					}
				}
				return Observation{Check: CheckResult{Status: status}}, nil
			}}
			r := runTest(t, a, targetForTest("functional", "protocol", "latency"))
			if _, ok := r.Checks["functional"]; ok {
				t.Fatal("blocked functional check was emitted")
			}
			if _, ok := r.Checks[failed]; !ok {
				t.Fatal("failed implicit gate missing")
			}
			if failed == "authentication" {
				if _, ok := r.Checks["protocol"]; !ok {
					t.Fatal("authentication must not block protocol")
				}
				if _, ok := r.Checks["latency"]; !ok {
					t.Fatal("authentication must not block latency")
				}
			}
		})
	}
	a := &testAdapter{}
	r := runTest(t, a, targetForTest("functional"))
	if len(r.Checks) != 1 || r.Checks["functional"].Status != Healthy {
		t.Fatal("passing implicit gates must be hidden", r)
	}
	if strings.Join(a.calls, ",") != "configuration,reachability,authentication,functional" {
		t.Fatal(a.calls)
	}
}
func TestPassiveDefaultsAndEmptyChecks(t *testing.T) {
	a := &testAdapter{active: []string{"capability"}}
	r := runTest(t, a, targetForTest())
	if contains(a.calls, "functional") || contains(a.calls, "capability") || r.Status != Healthy {
		t.Fatal("default executed active check", a.calls)
	}
	a = &testAdapter{}
	target := targetForTest()
	target.Checks = []string{}
	r = runTest(t, a, target)
	if len(a.calls) != 0 || r.Status != Unknown {
		t.Fatal("empty checks ran", r, a.calls)
	}
}
func TestDependenciesRunAfterParentFailure(t *testing.T) {
	a := &testAdapter{fn: func(_ context.Context, r Request, d string) (Observation, error) {
		s := Healthy
		if r.Target.Name == "parent" && d == "configuration" {
			s = Misconfigured
		}
		if r.Target.Name == "child" && d == "reachability" {
			s = Unknown
		}
		return Observation{Check: CheckResult{Status: s}}, nil
	}}
	parent := targetForTest("reachability")
	parent.Name = "parent"
	child := targetForTest("reachability")
	child.Name = "child"
	parent.Dependencies = []Dependency{{Target: child}}
	r := runTest(t, a, parent)
	if len(r.Dependencies) != 1 || r.Status != Unknown || r.Checks["configuration"].Status != Misconfigured {
		t.Fatal(r)
	}
	if _, ok := r.Checks["dependency"]; ok {
		t.Fatal("unrequested dependency summary")
	}
}
func TestRetriesAndActiveNeverRetried(t *testing.T) {
	retries, delay := 3, 0
	for _, dimension := range []string{"reachability", "functional"} {
		attempts := 0
		a := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
			if d == dimension {
				attempts++
				return Observation{}, &net.DNSError{Err: "no response"}
			}
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		}}
		target := targetForTest(dimension)
		target.Retries = &retries
		target.RetryDelayMS = &delay
		runTest(t, a, target)
		expected := 4
		if dimension == "functional" {
			expected = 1
		}
		if attempts != expected {
			t.Fatal(dimension, attempts)
		}
	}
	attempts := 0
	a := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
		if d == "reachability" {
			attempts++
			return Observation{ResponseReceived: true}, context.DeadlineExceeded
		}
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}}
	target := targetForTest("reachability")
	target.Retries = &retries
	if r := runTest(t, a, target); r.Status != Unknown || attempts != 1 {
		t.Fatal("partial response retried", r, attempts)
	}
}
func TestTimeoutPartialAndIgnoringAdapter(t *testing.T) {
	for _, partial := range []bool{false, true} {
		a := &testAdapter{fn: func(ctx context.Context, r Request, d string) (Observation, error) {
			if d == "reachability" {
				if partial {
					r.MarkResponse()
				}
				<-ctx.Done()
				return Observation{}, ctx.Err()
			}
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		}}
		target := targetForTest("reachability")
		target.CheckTimeouts = map[string]int{"reachability": 10}
		r := runTest(t, a, target)
		expected := Unreachable
		if partial {
			expected = Unknown
		}
		if r.Status != expected {
			t.Fatal(r)
		}
	}
	release := make(chan struct{})
	defer close(release)
	a := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
		if d == "reachability" {
			<-release
		}
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}}
	target := targetForTest("reachability")
	timeout := 10
	target.TimeoutMS = &timeout
	started := time.Now()
	r := runTest(t, a, target)
	if r.Status != Unreachable || time.Since(started) > time.Second {
		t.Fatal("uncooperative adapter hung caller", r)
	}
}
func TestAdapterPanicInvalidResultAndSecrets(t *testing.T) {
	t.Setenv("AGENTHEALTH_TEST_TOKEN", "private-token")
	for _, mode := range []string{"panic", "invalid", "error", "message"} {
		a := &testAdapter{fn: func(_ context.Context, _ Request, d string) (Observation, error) {
			if d == "reachability" {
				switch mode {
				case "panic":
					panic("private-token")
				case "invalid":
					return Observation{Check: CheckResult{Status: "bad", Message: "private-token"}}, nil
				case "error":
					return Observation{}, errors.New("private-token")
				case "message":
					return Observation{Check: CheckResult{Status: Unhealthy, Message: "unknown-secret"}}, nil
				}
			}
			return Observation{Check: CheckResult{Status: Healthy}}, nil
		}}
		target := targetForTest("reachability")
		target.Name = "private-token"
		target.Auth = &AuthReference{BearerEnv: "AGENTHEALTH_TEST_TOKEN"}
		r := runTest(t, a, target)
		var b bytes.Buffer
		if err := WriteJSON(&b, []Result{r}); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(b.String(), "private-token") || strings.Contains(b.String(), "unknown-secret") {
			t.Fatal("secret leaked", b.String())
		}
	}
	a := &testAdapter{}
	target := targetForTest("functional")
	target.Auth = &AuthReference{BearerEnv: "AGENTHEALTH_NONEXISTENT_TOKEN"}
	t.Setenv(target.Auth.BearerEnv, "")
	r := runTest(t, a, target)
	if r.Status != Misconfigured || len(a.calls) != 0 {
		t.Fatal("missing credentials attempted check", r)
	}
}
func TestLatencyThresholdAndCancellation(t *testing.T) {
	threshold := 0.0
	target := targetForTest("latency")
	target.Thresholds = &Thresholds{LatencyMS: &threshold}
	r := runTest(t, &testAdapter{}, target)
	if r.Status != Degraded || r.LatencyMS == nil {
		t.Fatal(r)
	}
	a := &testAdapter{}
	registry := NewRegistry()
	if err := registry.Register(a); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	results, err := NewEngine(registry).Run(ctx, Config{Version: "v1", Targets: []Target{targetForTest("reachability")}})
	if err != nil || results[0].Status != Unknown {
		t.Fatal(results, err)
	}
}
func TestRegistryAndBoundedCalls(t *testing.T) {
	registry := NewRegistry()
	a := &testAdapter{}
	if err := registry.Register(a); err != nil {
		t.Fatal(err)
	}
	if registry.Register(a) == nil {
		t.Fatal("duplicate adapter accepted")
	}
	var running, max atomic.Int32
	release := make(chan struct{})
	defer close(release)
	a.fn = func(_ context.Context, _ Request, _ string) (Observation, error) {
		n := running.Add(1)
		defer running.Add(-1)
		for {
			old := max.Load()
			if n <= old || max.CompareAndSwap(old, n) {
				break
			}
		}
		<-release
		return Observation{Check: CheckResult{Status: Healthy}}, nil
	}
	engine := NewEngine(registry)
	var wg sync.WaitGroup
	for i := 0; i < 24; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			target := targetForTest("configuration")
			timeout := 10
			target.TimeoutMS = &timeout
			_, err := engine.Run(context.Background(), Config{Version: "v1", Targets: []Target{target}})
			if err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if max.Load() > 16 {
		t.Fatal("adapter call pool exceeded bound", max.Load())
	}
}

func TestLatencyExcludesAdapterQueue(t *testing.T) {
	engine := NewEngine(NewRegistry())
	for i := 0; i < cap(engine.slots); i++ {
		engine.slots <- struct{}{}
	}
	done := make(chan time.Duration, 1)
	go func() {
		_, elapsed := engine.attempt(context.Background(), &testAdapter{}, Request{Target: targetForTest("reachability"), Client: NewHTTPClient()}, "reachability")
		done <- elapsed
	}()
	// Hold every slot for much longer than the fast adapter's work.
	time.Sleep(100 * time.Millisecond)
	<-engine.slots
	elapsed := <-done
	if elapsed >= 50*time.Millisecond {
		t.Fatalf("latency includes local queue: %v", elapsed)
	}
	for i := 0; i < cap(engine.slots)-1; i++ {
		<-engine.slots
	}
}

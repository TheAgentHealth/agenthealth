package core

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"sync"
)

// Metadata declares compatibility, implemented dimensions, and active checks.
// All active checks must remain non-destructive, even when opted in.
type Metadata struct {
	Name, Version, CompatibilityVersion   string
	TargetTypes, Dimensions, ActiveChecks []string
}

// Request contains per-target state. Credentials must never be logged.
// Client verifies TLS, restricts redirects, and follows the check context.
type Request struct {
	// TargetContext bounds resources shared across dimensions; Check must still honor its attempt context.
	TargetContext context.Context
	// MarkResponse must be called as soon as any response arrives, before reading its body.
	MarkResponse func()
	// RecordStep records only allowlisted transport stages and health states.
	RecordStep func(string, Status)
	// Environment is a per-target snapshot of configured environment references.
	Environment map[string]string
	// RunState belongs to one target execution; adapters may cache acquired credentials.
	RunState *sync.Map
	// RememberSecret registers dynamically acquired credentials for output redaction.
	RememberSecret func(string)
	Target         Target
	Credential     string
	Client         *http.Client
}

// Observation records whether a response was received, distinguishing partial
// responses before a timeout from failures to establish communication.
type Observation struct {
	Check            CheckResult
	ResponseReceived bool
	// Code selects an engine-owned diagnostic; arbitrary adapter text is suppressed.
	Code string
}

// Adapter implementations must honor context cancellation and use the supplied
// HTTP client for HTTP requests. The engine trusts compiled-in adapter code;
// arbitrary code cannot be sandboxed or forcibly stopped in-process.
type Adapter interface {
	Metadata() Metadata
	Check(context.Context, Request, string) (Observation, error)
}

// TargetCloser optionally releases per-target resources after all dimensions.
// It must honor the bounded cleanup context and is not called concurrently with another CloseTarget.
type TargetCloser interface {
	CloseTarget(context.Context, *sync.Map)
}

// Failure carries a normalized health classification, never a raw diagnostic.
type Failure struct{ Status Status }

func (e *Failure) Error() string { return "health check failed" }

// NormalizeError never incorporates raw errors (which may contain credentials).
func NormalizeError(err error, responseReceived bool) CheckResult {
	if err == nil {
		return CheckResult{Status: Healthy}
	}
	var failure *Failure
	if errors.As(err, &failure) && failure.Status.Valid() {
		return CheckResult{Status: failure.Status, Message: "health check failed"}
	}
	if errors.Is(err, context.Canceled) {
		return CheckResult{Status: Unknown, Message: "check canceled"}
	}
	var network net.Error
	var cert *tls.CertificateVerificationError
	var unknownAuthority x509.UnknownAuthorityError
	if errors.Is(err, context.DeadlineExceeded) || errors.As(err, &network) || errors.As(err, &cert) || errors.As(err, &unknownAuthority) {
		if responseReceived {
			return CheckResult{Status: Unknown, Message: "incomplete response"}
		}
		return CheckResult{Status: Unreachable, Message: "communication failed"}
	}
	return CheckResult{Status: Unknown, Message: "adapter error"}
}

// Registry supports concurrent lookup and registration; duplicate types fail
// atomically rather than silently replacing an adapter.
type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
	cleanup  chan struct{}
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter), cleanup: make(chan struct{}, 1)}
}
func (r *Registry) Register(a Adapter) error {
	if a == nil {
		return errors.New("nil adapter")
	}
	m := a.Metadata()
	if m.Name == "" || m.Version == "" || m.CompatibilityVersion != "v1" || len(m.TargetTypes) == 0 {
		return errors.New("invalid adapter metadata")
	}
	types := map[string]bool{}
	for _, typ := range m.TargetTypes {
		if !contains(targetTypes, typ) || types[typ] {
			return errors.New("invalid adapter target types")
		}
		types[typ] = true
	}
	checks := map[string]bool{}
	for _, dimension := range m.Dimensions {
		if !contains(dimensions, dimension) || dimension == "dependency" || dimension == "latency" || checks[dimension] {
			return errors.New("invalid adapter dimensions")
		}
		checks[dimension] = true
	}
	if !checks["configuration"] || !checks["reachability"] {
		return errors.New("adapter requires configuration and reachability gates")
	}
	for _, active := range m.ActiveChecks {
		if !checks[active] || active == "configuration" || active == "reachability" || active == "authentication" {
			return errors.New("invalid active dimension")
		}
	}
	if checks["functional"] && !contains(m.ActiveChecks, "functional") {
		return errors.New("functional dimension must be declared active")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.adapters == nil {
		r.adapters = make(map[string]Adapter)
	}
	for typ := range types {
		if r.adapters[typ] != nil {
			return errors.New("target type already registered")
		}
	}
	for typ := range types {
		r.adapters[typ] = a
	}
	return nil
}
func (r *Registry) lookup(typ string) Adapter {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.adapters[typ]
}

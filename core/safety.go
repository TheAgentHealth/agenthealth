package core

import (
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var urlPattern = regexp.MustCompile(`(?i)\b(?:https?|postgres(?:ql)?|mysql|mongodb|redis)://[^\s"<>]+`)
var credentialPattern = regexp.MustCompile(`(?i)(authorization|bearer|api[_-]?key|token|password|secret)(\s*[:=]?\s+|\s*[:=]\s*)[^\s,;]+`)

// Redactor masks resolved credentials, URL addresses (including userinfo and
// query strings), and recognizable authentication assignments.
type Redactor struct {
	mu      sync.RWMutex
	secrets []string
}

func NewRedactor(secrets ...string) *Redactor {
	values := append([]string(nil), secrets...)
	sort.Slice(values, func(i, j int) bool { return len(values[i]) > len(values[j]) })
	return &Redactor{secrets: values}
}

// Remember safely registers credentials obtained while a run is in progress.
func (r *Redactor) Remember(secret string) {
	if r == nil || secret == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.secrets = append(r.secrets, secret)
	sort.Slice(r.secrets, func(i, j int) bool { return len(r.secrets[i]) > len(r.secrets[j]) })
}
func (r *Redactor) Redact(s string) string {
	if r != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		for _, secret := range r.secrets {
			if secret != "" {
				s = strings.ReplaceAll(s, secret, "[REDACTED]")
			}
		}
	}
	s = urlPattern.ReplaceAllString(s, "[REDACTED URL]")
	return credentialPattern.ReplaceAllString(s, "[REDACTED CREDENTIAL]")
}
func (r *Redactor) Result(value Result) Result {
	value.Target.Name = r.Redact(value.Target.Name)
	if value.Target.ID != "" && r.Redact(value.Target.ID) != value.Target.ID {
		value.Target.ID = "redacted"
	}
	checks := make(map[string]CheckResult, len(value.Checks))
	for key, check := range value.Checks {
		check.Message = r.Redact(check.Message)
		checks[key] = check
	}
	value.Checks = checks
	dependencies := make([]Result, len(value.Dependencies))
	for i, dep := range value.Dependencies {
		dependencies[i] = r.Result(dep)
	}
	value.Dependencies = dependencies
	return value
}

// NewHTTPClient creates an isolated TLS-verifying transport. Redirects are not
// followed, preventing probes or credentials from moving to another endpoint.
func NewHTTPClient() *http.Client {
	transport := &http.Transport{
		Proxy:           http.ProxyFromEnvironment,
		IdleConnTimeout: 30 * time.Second,
		MaxIdleConns:    16,
	}
	return &http.Client{Transport: transport, Timeout: DefaultTimeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
}

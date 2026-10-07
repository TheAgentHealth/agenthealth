// Package core implements the shared Agent Health Specification model.
package core

// Status is ordered by the specification's severity, not by reachability.
type Status string

const (
	Healthy       Status = "HEALTHY"
	Degraded      Status = "DEGRADED"
	Unhealthy     Status = "UNHEALTHY"
	Unreachable   Status = "UNREACHABLE"
	Misconfigured Status = "MISCONFIGURED"
	Unknown       Status = "UNKNOWN"
)

var statuses = []Status{Healthy, Degraded, Unhealthy, Unreachable, Misconfigured, Unknown}

func (s Status) Valid() bool {
	for _, v := range statuses {
		if s == v {
			return true
		}
	}
	return false
}
func severity(s Status) int {
	for i, v := range statuses {
		if s == v {
			return i
		}
	}
	return len(statuses) - 1
}

// Worst returns UNKNOWN when there is no evidence or an invalid status.
func Worst(values ...Status) Status {
	if len(values) == 0 {
		return Unknown
	}
	result := Healthy
	for _, s := range values {
		if !s.Valid() {
			s = Unknown
		}
		if severity(s) > severity(result) {
			result = s
		}
	}
	return result
}

// DependencyContribution maps dependency health onto the parent's ability to work.
func DependencyContribution(s Status, critical bool) Status {
	if !s.Valid() {
		s = Unknown
	}
	if s == Healthy || s == Degraded {
		return s
	}
	if !critical {
		return Degraded
	}
	if s == Unknown {
		return Unknown
	}
	return Unhealthy
}

type TargetIdentity struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
type CheckResult struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}
type Result struct {
	Target       TargetIdentity         `json:"target"`
	Status       Status                 `json:"status"`
	LatencyMS    *float64               `json:"latency_ms"`
	Checks       map[string]CheckResult `json:"checks"`
	Dependencies []Result               `json:"dependencies"`
}

// Aggregate excludes the dependency diagnostic from own health; dependencies
// contribute independently even when the diagnostic was not requested.
func Aggregate(checks map[string]CheckResult, dependencies []Result, critical []bool) Status {
	own := []Status{}
	for dimension, check := range checks {
		if dimension != "dependency" {
			own = append(own, check.Status)
		}
	}
	values := []Status{Worst(own...)}
	for i, dependency := range dependencies {
		required := true
		if i < len(critical) {
			required = critical[i]
		}
		values = append(values, DependencyContribution(dependency.Status, required))
	}
	return Worst(values...)
}

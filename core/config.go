package core

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string   `yaml:"version"`
	Targets []Target `yaml:"targets"`
}
type Target struct {
	Name          string         `yaml:"name"`
	Type          string         `yaml:"type"`
	Endpoint      string         `yaml:"endpoint"`
	Checks        []string       `yaml:"checks"`
	Thresholds    *Thresholds    `yaml:"thresholds"`
	Dependencies  []Dependency   `yaml:"dependencies"`
	TimeoutMS     *int           `yaml:"timeout_ms"`
	Retries       *int           `yaml:"retries"`
	RetryDelayMS  *int           `yaml:"retry_delay_ms"`
	CheckTimeouts map[string]int `yaml:"check_timeouts_ms"`
	Auth          *AuthReference `yaml:"auth"`
}

// AuthReference resolves a credential from the environment; never a literal secret.
type AuthReference struct {
	BearerEnv string `yaml:"bearer_env"`
}

const DefaultTimeout = 5 * time.Second
const DefaultRetryDelay = 100 * time.Millisecond

func (t Target) timeout(dimension string) time.Duration {
	if ms, ok := t.CheckTimeouts[dimension]; ok {
		return time.Duration(ms) * time.Millisecond
	}
	if t.TimeoutMS != nil {
		return time.Duration(*t.TimeoutMS) * time.Millisecond
	}
	return DefaultTimeout
}
func (t Target) retryCount() int {
	if t.Retries != nil {
		return *t.Retries
	}
	return 0
}
func (t Target) retryDelay() time.Duration {
	if t.RetryDelayMS != nil {
		return time.Duration(*t.RetryDelayMS) * time.Millisecond
	}
	return DefaultRetryDelay
}

type Thresholds struct {
	LatencyMS *float64 `yaml:"latency_ms"`
}
type Dependency struct {
	Target   `yaml:",inline"`
	Critical *bool `yaml:"critical"`
}

func (d Dependency) IsCritical() bool { return d.Critical == nil || *d.Critical }

var targetTypes = strings.Fields("agent multi-agent a2a mcp tool model llm http api vector-store database runtime gateway custom")
var dimensions = strings.Fields("reachability protocol authentication capability functional dependency latency configuration")

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// LoadConfig accepts exactly one YAML document and rejects unknown fields.
// Parser details are deliberately suppressed because they may contain secrets.
func LoadConfig(r io.Reader) (Config, error) {
	var config Config
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return Config{}, errors.New("invalid YAML configuration")
	}
	if len(document.Content) != 1 || !validYAMLNode(document.Content[0], "config", 0) {
		return Config{}, errors.New("invalid configuration field types")
	}
	// Re-decode with KnownFields so unknown keys are rejected, including nested keys.
	encoded, err := yaml.Marshal(&document)
	if err != nil {
		return Config{}, errors.New("invalid YAML configuration")
	}
	strict := yaml.NewDecoder(strings.NewReader(string(encoded)))
	strict.KnownFields(true)
	if err := strict.Decode(&config); err != nil {
		return Config{}, errors.New("invalid YAML configuration")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Config{}, errors.New("configuration must contain exactly one document")
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}
func LoadConfigFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, errors.New("cannot open configuration file")
	}
	defer f.Close()
	return LoadConfig(f)
}
func (c Config) Validate() error {
	if c.Version != "v1" {
		return errors.New("unsupported configuration version; expected v1")
	}
	if len(c.Targets) == 0 {
		return errors.New("configuration requires at least one target")
	}
	for i, target := range c.Targets {
		if err := validateTarget(target, fmt.Sprintf("targets[%d]", i), 0); err != nil {
			return err
		}
	}
	return nil
}
func validateTarget(t Target, path string, depth int) error {
	if depth > 64 {
		return fmt.Errorf("%s: dependencies exceed maximum depth 64", path)
	}
	if strings.TrimSpace(t.Name) == "" || strings.TrimSpace(t.Endpoint) == "" {
		return fmt.Errorf("%s: name and endpoint are required", path)
	}
	if !contains(targetTypes, t.Type) {
		return fmt.Errorf("%s: unsupported target type", path)
	}
	for _, check := range t.Checks {
		if !contains(dimensions, check) {
			return fmt.Errorf("%s: unsupported check dimension", path)
		}
	}
	if t.TimeoutMS != nil && (*t.TimeoutMS < 1 || *t.TimeoutMS > 300000) {
		return fmt.Errorf("%s: timeout_ms must be between 1 and 300000", path)
	}
	if t.Retries != nil && (*t.Retries < 0 || *t.Retries > 3) {
		return fmt.Errorf("%s: retries must be between 0 and 3", path)
	}
	if t.RetryDelayMS != nil && (*t.RetryDelayMS < 0 || *t.RetryDelayMS > 60000) {
		return fmt.Errorf("%s: retry_delay_ms must be between 0 and 60000", path)
	}
	for dimension, ms := range t.CheckTimeouts {
		if !contains(dimensions, dimension) || dimension == "dependency" || ms < 1 || ms > 300000 {
			return fmt.Errorf("%s: invalid check timeout", path)
		}
	}
	if t.Auth != nil && (strings.TrimSpace(t.Auth.BearerEnv) == "" || strings.ContainsAny(t.Auth.BearerEnv, "=\x00")) {
		return fmt.Errorf("%s: auth requires a valid bearer_env reference", path)
	}
	if t.Thresholds != nil && t.Thresholds.LatencyMS != nil {
		n := *t.Thresholds.LatencyMS
		if n < 0 || math.IsNaN(n) || math.IsInf(n, 0) {
			return fmt.Errorf("%s: latency threshold must be finite and non-negative", path)
		}
	}
	for i, dep := range t.Dependencies {
		if err := validateTarget(dep.Target, fmt.Sprintf("%s.dependencies[%d]", path, i), depth+1); err != nil {
			return err
		}
	}
	return nil
}

// YAML's typed decoder permits scalar coercion and null values. Reject both
// where the JSON Schema requires a concrete type, before decoding Go structs.
func validYAMLNode(n *yaml.Node, kind string, depth int) bool {
	if depth > 140 || n.Kind != yaml.MappingNode {
		return false
	}
	seen := map[string]bool{}
	for i := 0; i < len(n.Content); i += 2 {
		key, value := n.Content[i], n.Content[i+1]
		if key.Tag != "!!str" || seen[key.Value] {
			return false
		}
		seen[key.Value] = true
		switch key.Value {
		case "version", "name", "type", "endpoint", "bearer_env":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
				return false
			}
		case "critical":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
				return false
			}
		case "timeout_ms", "retries", "retry_delay_ms":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
				return false
			}
		case "auth":
			if !validYAMLNode(value, "auth", depth+1) {
				return false
			}
		case "check_timeouts_ms":
			if value.Kind != yaml.MappingNode {
				return false
			}
			timeoutsSeen := map[string]bool{}
			for j := 0; j < len(value.Content); j += 2 {
				k, v := value.Content[j], value.Content[j+1]
				if k.Tag != "!!str" || timeoutsSeen[k.Value] || !contains(dimensions, k.Value) || k.Value == "dependency" || v.Tag != "!!int" || v.Kind != yaml.ScalarNode {
					return false
				}
				timeoutsSeen[k.Value] = true
			}
		case "latency_ms":
			if value.Kind != yaml.ScalarNode || (value.Tag != "!!int" && value.Tag != "!!float") {
				return false
			}
		case "checks":
			if value.Kind != yaml.SequenceNode {
				return false
			}
			for _, child := range value.Content {
				if child.Kind != yaml.ScalarNode || child.Tag != "!!str" {
					return false
				}
			}
		case "targets", "dependencies":
			if value.Kind != yaml.SequenceNode {
				return false
			}
			for _, child := range value.Content {
				if !validYAMLNode(child, "target", depth+1) {
					return false
				}
			}
		case "thresholds":
			if !validYAMLNode(value, "thresholds", depth+1) {
				return false
			}
		default:
			return false
		}
	}
	// Required fields must be present even if a Go zero value might be usable.
	if kind == "auth" {
		return seen["bearer_env"]
	}
	if kind == "config" {
		return seen["version"] && seen["targets"]
	}
	if kind == "target" {
		return seen["name"] && seen["type"] && seen["endpoint"]
	}
	return true
}

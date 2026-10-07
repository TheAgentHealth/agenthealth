package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
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
	HTTP          *HTTPOptions   `yaml:"http"`
	MCP           *MCPOptions    `yaml:"mcp"`
	A2A           *A2AOptions    `yaml:"a2a"`
}

// A2AOptions describes passive expectations and an explicitly safe interaction.
type A2AOptions struct {
	CardURL              string          `yaml:"card_url"`
	ProtocolVersion      string          `yaml:"protocol_version"`
	RequiredSkills       []string        `yaml:"required_skills"`
	RequiredCapabilities []string        `yaml:"required_capabilities"`
	Functional           *A2AInteraction `yaml:"functional"`
}
type A2AInteraction struct {
	Safe bool   `yaml:"safe"`
	Text string `yaml:"text"`
}

// MCPOptions configures Streamable HTTP discovery and an explicitly safe probe.
type MCPOptions struct {
	Transport         string         `yaml:"transport"`
	Stdio             *MCPStdio      `yaml:"stdio"`
	OAuth             *MCPOAuth      `yaml:"oauth"`
	ProtocolVersion   string         `yaml:"protocol_version"`
	RequiredTools     []string       `yaml:"required_tools"`
	RequiredResources []string       `yaml:"required_resources"`
	RequiredPrompts   []string       `yaml:"required_prompts"`
	Functional        *MCPInvocation `yaml:"functional"`
}
type MCPStdio struct {
	Command   string            `yaml:"command"`
	Args      []string          `yaml:"args"`
	Directory string            `yaml:"directory"`
	Env       map[string]string `yaml:"env"`
}

// MCPOAuth uses a pinned issuer and environment/file references, never secrets.
type MCPOAuth struct {
	Issuer          string   `yaml:"issuer"`
	ClientID        string   `yaml:"client_id"`
	Grant           string   `yaml:"grant"`
	ClientSecretEnv string   `yaml:"client_secret_env"`
	RefreshTokenEnv string   `yaml:"refresh_token_env"`
	TokenFile       string   `yaml:"token_file"`
	Scopes          []string `yaml:"scopes"`
	RedirectPort    *int     `yaml:"redirect_port"`
}

// CredentialReferences lists only references used by this target.
func (t Target) CredentialReferences() []string {
	refs := []string{}
	if t.Auth != nil {
		refs = append(refs, t.Auth.BearerEnv)
	}
	if t.MCP != nil {
		if t.MCP.Stdio != nil {
			for _, ref := range t.MCP.Stdio.Env {
				refs = append(refs, ref)
			}
		}
		if o := t.MCP.OAuth; o != nil {
			for _, ref := range []string{o.ClientSecretEnv, o.RefreshTokenEnv} {
				if ref != "" {
					refs = append(refs, ref)
				}
			}
		}
	}
	return refs
}

type MCPInvocation struct {
	Tool          string `yaml:"tool"`
	Safe          bool   `yaml:"safe"`
	ArgumentsJSON string `yaml:"arguments_json"`
}

func SupportedMCPVersion(v string) bool {
	return v == "2026-07-28" || v == "2025-03-26" || v == "2025-06-18" || v == "2025-11-25"
}

// HTTPOptions specifies response expectations, never outgoing credentials.
type HTTPOptions struct {
	ExpectedStatus []int             `yaml:"expected_status"`
	Headers        map[string]string `yaml:"headers"`
	BodyContains   string            `yaml:"body_contains"`
	MaxBodyBytes   *int              `yaml:"max_body_bytes"`
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
	if t.A2A != nil {
		o := t.A2A
		if t.Type != "a2a" || (o.ProtocolVersion != "" && o.ProtocolVersion != "0.3.0") {
			return fmt.Errorf("%s: A2A options require a2a target and supported protocol version", path)
		}
		for _, names := range [][]string{o.RequiredSkills, o.RequiredCapabilities} {
			seen := map[string]bool{}
			for _, name := range names {
				if strings.TrimSpace(name) == "" || seen[name] {
					return fmt.Errorf("%s: A2A requirements must be nonempty and unique", path)
				}
				seen[name] = true
			}
		}
		for _, name := range o.RequiredCapabilities {
			if !contains([]string{"streaming", "pushNotifications", "stateTransitionHistory"}, name) {
				return fmt.Errorf("%s: unsupported A2A capability expectation", path)
			}
		}
		if t.Checks != nil && !contains(t.Checks, "capability") && len(o.RequiredSkills)+len(o.RequiredCapabilities) > 0 {
			return fmt.Errorf("%s: A2A expectations require capability check", path)
		}
		if f := o.Functional; f != nil {
			if !contains(t.Checks, "functional") || !f.Safe || strings.TrimSpace(f.Text) == "" || len(f.Text) > 65536 {
				return fmt.Errorf("%s: A2A interaction requires functional opt-in, safe: true and text of at most 64 KiB", path)
			}
		}
	}
	if t.Type == "a2a" && contains(t.Checks, "functional") && (t.A2A == nil || t.A2A.Functional == nil) {
		return fmt.Errorf("%s: A2A functional check requires interaction", path)
	}
	if t.MCP != nil {
		if t.Type != "mcp" {
			return fmt.Errorf("%s: mcp options require mcp target", path)
		}
		if t.MCP.Transport != "" && t.MCP.Transport != "http" && t.MCP.Transport != "stdio" {
			return fmt.Errorf("%s: invalid MCP transport", path)
		}
		if t.MCP.Transport == "stdio" {
			if t.MCP.Stdio == nil || strings.TrimSpace(t.MCP.Stdio.Command) == "" || strings.ContainsRune(t.MCP.Stdio.Command, 0) || t.Auth != nil || t.MCP.OAuth != nil {
				return fmt.Errorf("%s: stdio requires a command and environment credentials", path)
			}
		} else if t.MCP.Stdio != nil {
			return fmt.Errorf("%s: stdio options require stdio transport", path)
		}
		if s := t.MCP.Stdio; s != nil {
			if strings.ContainsRune(s.Directory, 0) {
				return fmt.Errorf("%s: invalid stdio directory", path)
			}
			for _, arg := range s.Args {
				if strings.ContainsRune(arg, 0) {
					return fmt.Errorf("%s: invalid stdio argument", path)
				}
			}
			for key, ref := range s.Env {
				if !validEnvReference(key) || !validEnvReference(ref) {
					return fmt.Errorf("%s: invalid stdio environment reference", path)
				}
			}
		}
		if o := t.MCP.OAuth; o != nil {
			if t.Auth != nil || strings.TrimSpace(o.Issuer) == "" || strings.TrimSpace(o.ClientID) == "" || !contains([]string{"client_credentials", "refresh_token", "authorization_code"}, o.Grant) {
				return fmt.Errorf("%s: OAuth requires issuer, client ID and supported grant; cannot combine with bearer auth", path)
			}
			if o.Grant == "client_credentials" && o.ClientSecretEnv == "" || o.Grant == "refresh_token" && o.RefreshTokenEnv == "" && o.TokenFile == "" || o.Grant == "authorization_code" && o.TokenFile == "" {
				return fmt.Errorf("%s: OAuth grant requires credential or token file reference", path)
			}
			if o.TokenFile != "" && (strings.TrimSpace(o.TokenFile) == "" || strings.ContainsRune(o.TokenFile, 0)) {
				return fmt.Errorf("%s: invalid OAuth token file", path)
			}
			if o.RedirectPort != nil && (*o.RedirectPort < 0 || *o.RedirectPort > 65535) {
				return fmt.Errorf("%s: invalid OAuth redirect port", path)
			}
			for _, ref := range []string{o.ClientSecretEnv, o.RefreshTokenEnv} {
				if ref != "" && !validEnvReference(ref) {
					return fmt.Errorf("%s: invalid OAuth credential reference", path)
				}
			}
			seen := map[string]bool{}
			for _, scope := range o.Scopes {
				if !validOAuthScope(scope) || seen[scope] {
					return fmt.Errorf("%s: invalid OAuth scopes", path)
				}
				seen[scope] = true
			}
		}
		if t.MCP.ProtocolVersion != "" && !SupportedMCPVersion(t.MCP.ProtocolVersion) {
			return fmt.Errorf("%s: unsupported MCP protocol version", path)
		}
		for _, names := range [][]string{t.MCP.RequiredTools, t.MCP.RequiredResources, t.MCP.RequiredPrompts} {
			seen := map[string]bool{}
			for _, name := range names {
				if strings.TrimSpace(name) == "" || seen[name] {
					return fmt.Errorf("%s: MCP requirements must be nonempty and unique", path)
				}
				seen[name] = true
			}
		}
		if t.Checks != nil && !contains(t.Checks, "capability") && (len(t.MCP.RequiredTools)+len(t.MCP.RequiredResources)+len(t.MCP.RequiredPrompts) > 0) {
			return fmt.Errorf("%s: MCP requirements require capability check", path)
		}
		if f := t.MCP.Functional; f != nil {
			if !contains(t.Checks, "functional") || !f.Safe || strings.TrimSpace(f.Tool) == "" {
				return fmt.Errorf("%s: MCP invocation requires functional opt-in, tool and safe: true", path)
			}
			if f.ArgumentsJSON != "" {
				var args map[string]json.RawMessage
				if len(f.ArgumentsJSON) > 65536 || json.Unmarshal([]byte(f.ArgumentsJSON), &args) != nil || args == nil {
					return fmt.Errorf("%s: arguments_json must be a JSON object of at most 64 KiB", path)
				}
			}
		}
	}
	if t.Type == "mcp" && contains(t.Checks, "functional") && (t.MCP == nil || t.MCP.Functional == nil) {
		return fmt.Errorf("%s: MCP functional check requires invocation", path)
	}
	if t.HTTP != nil {
		if t.Type != "http" && t.Type != "api" {
			return fmt.Errorf("%s: http options require http or api target", path)
		}
		if t.HTTP.ExpectedStatus != nil && len(t.HTTP.ExpectedStatus) == 0 {
			return fmt.Errorf("%s: expected_status requires at least one status", path)
		}
		for _, status := range t.HTTP.ExpectedStatus {
			if status < 200 || status > 599 {
				return fmt.Errorf("%s: expected_status must be between 200 and 599", path)
			}
		}
		if t.HTTP.MaxBodyBytes != nil && (*t.HTTP.MaxBodyBytes < 1 || *t.HTTP.MaxBodyBytes > 1048576) {
			return fmt.Errorf("%s: max_body_bytes must be between 1 and 1048576", path)
		}
		if t.HTTP.BodyContains != "" && !contains(t.Checks, "functional") {
			return fmt.Errorf("%s: body_contains requires explicit functional check", path)
		}
		if t.Checks != nil && !contains(t.Checks, "protocol") && !contains(t.Checks, "functional") && (t.HTTP.ExpectedStatus != nil || len(t.HTTP.Headers) > 0) {
			return fmt.Errorf("%s: HTTP expectations require protocol or functional check", path)
		}
		seenHeaders := map[string]bool{}
		for name, value := range t.HTTP.Headers {
			canonical := http.CanonicalHeaderKey(name)
			if !validHTTPHeaderName(name) || !validHTTPHeaderValue(value) || seenHeaders[canonical] {
				return fmt.Errorf("%s: invalid or duplicate HTTP header expectation", path)
			}
			seenHeaders[canonical] = true
		}
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

func validHTTPHeaderName(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	return true
}
func validHTTPHeaderValue(s string) bool {
	for _, c := range s {
		if (c < 32 && c != '\t') || c == 127 {
			return false
		}
	}
	return true
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
		case "client_secret_env", "refresh_token_env", "token_file":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.TrimSpace(value.Value) == "" {
				return false
			}
		case "version", "name", "type", "endpoint", "bearer_env", "body_contains", "protocol_version", "tool", "arguments_json", "transport", "command", "directory", "issuer", "client_id", "grant", "card_url", "text":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
				return false
			}
		case "critical", "safe":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
				return false
			}
		case "timeout_ms", "retries", "retry_delay_ms", "max_body_bytes", "redirect_port":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
				return false
			}
		case "auth":
			if !validYAMLNode(value, "auth", depth+1) {
				return false
			}
		case "a2a", "mcp", "functional", "stdio", "oauth":
			if !validYAMLNode(value, key.Value, depth+1) {
				return false
			}
		case "http":
			if !validYAMLNode(value, "http", depth+1) {
				return false
			}
		case "expected_status":
			if value.Kind != yaml.SequenceNode {
				return false
			}
			for _, child := range value.Content {
				if child.Kind != yaml.ScalarNode || child.Tag != "!!int" {
					return false
				}
			}
		case "headers", "env":
			if value.Kind != yaml.MappingNode {
				return false
			}
			seenHeaders := map[string]bool{}
			for j := 0; j < len(value.Content); j += 2 {
				k, v := value.Content[j], value.Content[j+1]
				if k.Tag != "!!str" || v.Kind != yaml.ScalarNode || v.Tag != "!!str" || seenHeaders[mapKey(k.Value, key.Value)] {
					return false
				}
				seenHeaders[mapKey(k.Value, key.Value)] = true
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
		case "checks", "required_tools", "required_resources", "required_prompts", "args", "scopes", "required_skills", "required_capabilities":
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

func validEnvReference(ref string) bool {
	return strings.TrimSpace(ref) != "" && !strings.ContainsAny(ref, "=\x00")
}

func validOAuthScope(scope string) bool {
	if scope == "" {
		return false
	}
	for i := 0; i < len(scope); i++ {
		if scope[i] < 33 || scope[i] > 126 || scope[i] == '"' || scope[i] == '\\' {
			return false
		}
	}
	return true
}

func mapKey(value, field string) string {
	if field == "headers" {
		return strings.ToLower(value)
	}
	return value
}

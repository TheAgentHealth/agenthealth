package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/TheAgentHealth/agenthealth/core"
)

type oauthMetadata struct {
	Issuer                string   `json:"issuer"`
	TokenEndpoint         string   `json:"token_endpoint"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	AuthMethods           []string `json:"token_endpoint_auth_methods_supported"`
	PKCEMethods           []string `json:"code_challenge_methods_supported"`
	IssuerResponse        bool     `json:"authorization_response_iss_parameter_supported"`
}
type oauthDiscovery struct {
	metadata        oauthMetadata
	resource, scope string
}
type oauthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Issuer       string    `json:"issuer"`
	Resource     string    `json:"resource"`
	ClientID     string    `json:"client_id"`
	Scope        string    `json:"scope"`
}
type oauthCache struct {
	mu    sync.Mutex
	token oauthToken
}

func secureURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	return u.Scheme == "https" || (u.Scheme == "http" && (u.Hostname() == "localhost" || net.ParseIP(u.Hostname()) != nil && net.ParseIP(u.Hostname()).IsLoopback()))
}
func oauthError() error { return fail(core.Misconfigured, "mcp_oauth") }
func remember(r core.Request, values ...string) {
	if r.RememberSecret != nil {
		for _, value := range values {
			r.RememberSecret(value)
		}
	}
}
func oauthJSON(ctx context.Context, r core.Request, endpoint string, out any) error {
	if !secureURL(endpoint) {
		return oauthError()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return oauthError()
	}
	req.Header.Set("Accept", "application/json")
	response, err := oauthDo(r, req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return oauthError()
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 65537))
	if err != nil {
		return err
	}
	if len(body) > 65536 || json.Unmarshal(body, out) != nil {
		return oauthError()
	}
	return nil
}

var metadataChallenge = regexp.MustCompile(`(?i)(?:^|[, ]+)resource_metadata="([^"\\]+)"`)
var scopeChallenge = regexp.MustCompile(`(?i)(?:^|[, ]+)scope="([^"\\]+)"`)

func discoverOAuth(ctx context.Context, r core.Request) (oauthDiscovery, error) {
	options := r.Target.MCP.OAuth
	empty := oauthDiscovery{}
	issuer, err := url.Parse(options.Issuer)
	if err != nil || !secureURL(options.Issuer) || issuer.RawQuery != "" {
		return empty, oauthError()
	}
	endpoint, err := url.Parse(r.Target.Endpoint)
	if err != nil || !secureURL(r.Target.Endpoint) {
		return empty, oauthError()
	}
	resource := *endpoint
	resource.Scheme = strings.ToLower(resource.Scheme)
	resource.Host = strings.ToLower(resource.Host)
	if resource.Path == "/" {
		resource.Path = ""
	}
	resource.Fragment = ""
	resourceURI := resource.String()
	// Metadata requests never carry bearer tokens or client credentials.
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, r.Target.Endpoint, nil)
	if err != nil {
		return empty, oauthError()
	}
	response, err := oauthDo(r, req)
	if err != nil {
		return empty, err
	}
	challenge := response.Header.Get("WWW-Authenticate")
	response.Body.Close()
	scope := ""
	if m := scopeChallenge.FindStringSubmatch(challenge); len(m) > 1 {
		scope = m[1]
	}
	protectedURLs := []string{}
	if match := metadataChallenge.FindStringSubmatch(challenge); len(match) > 1 {
		u, err := url.Parse(match[1])
		if err != nil || u.Scheme != endpoint.Scheme || !strings.EqualFold(u.Host, endpoint.Host) {
			return empty, oauthError()
		}
		protectedURLs = append(protectedURLs, u.String())
	} else {
		base := endpoint.Scheme + "://" + endpoint.Host
		if endpoint.Path != "" && endpoint.Path != "/" {
			protectedURLs = append(protectedURLs, base+"/.well-known/oauth-protected-resource"+endpoint.EscapedPath())
		}
		protectedURLs = append(protectedURLs, base+"/.well-known/oauth-protected-resource")
	}
	var protected struct {
		Resource string   `json:"resource"`
		Servers  []string `json:"authorization_servers"`
		Scopes   []string `json:"scopes_supported"`
	}
	found := false
	for _, address := range protectedURLs {
		if oauthJSON(ctx, r, address, &protected) == nil {
			found = true
			break
		}
		if ctx.Err() != nil {
			return empty, oauthError()
		}
	}
	if !found || protected.Resource != resourceURI {
		return empty, oauthError()
	}
	trusted := false
	for _, server := range protected.Servers {
		if server == options.Issuer {
			trusted = true
		}
	}
	if !trusted {
		return empty, oauthError()
	}
	issuerBase := issuer.Scheme + "://" + issuer.Host
	path := strings.TrimSuffix(issuer.EscapedPath(), "/")
	metadataURLs := []string{issuerBase + "/.well-known/oauth-authorization-server" + path, strings.TrimSuffix(options.Issuer, "/") + "/.well-known/openid-configuration"}
	var metadata oauthMetadata
	found = false
	for _, address := range metadataURLs {
		candidate := oauthMetadata{}
		if oauthJSON(ctx, r, address, &candidate) == nil && candidate.Issuer == options.Issuer {
			metadata = candidate
			found = true
			break
		}
		if ctx.Err() != nil {
			return empty, oauthError()
		}
	}
	if !found || !secureURL(metadata.TokenEndpoint) {
		return empty, oauthError()
	}
	if issuer.Scheme == "https" {
		u, _ := url.Parse(metadata.TokenEndpoint)
		if u.Scheme != "https" {
			return empty, oauthError()
		}
	}
	if len(options.Scopes) > 0 {
		scope = strings.Join(options.Scopes, " ")
	} else if scope == "" {
		scope = strings.Join(protected.Scopes, " ")
	}
	return oauthDiscovery{metadata: metadata, resource: resourceURI, scope: scope}, nil
}
func exchangeToken(ctx context.Context, r core.Request, d oauthDiscovery, values url.Values) (oauthToken, error) {
	options := r.Target.MCP.OAuth
	empty := oauthToken{}
	values.Set("client_id", options.ClientID)
	values.Set("resource", d.resource)
	if d.scope != "" {
		values.Set("scope", d.scope)
	}
	secret := r.Environment[options.ClientSecretEnv]
	if options.ClientSecretEnv != "" && secret == "" {
		return empty, oauthError()
	}
	methods := d.metadata.AuthMethods
	if len(methods) == 0 {
		methods = []string{"client_secret_basic"}
	}
	method := "none"
	if secret != "" {
		method = ""
		for _, candidate := range []string{"client_secret_basic", "client_secret_post"} {
			for _, available := range methods {
				if available == candidate {
					method = candidate
					break
				}
			}
			if method != "" {
				break
			}
		}
		if method == "" {
			return empty, oauthError()
		}
	}
	if method == "client_secret_post" {
		values.Set("client_secret", secret)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.metadata.TokenEndpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return empty, oauthError()
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	if method == "client_secret_basic" {
		req.SetBasicAuth(url.QueryEscape(options.ClientID), url.QueryEscape(secret))
	}
	response, err := oauthDo(r, req)
	if err != nil {
		return empty, oauthError()
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return empty, oauthError()
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 65537))
	if err != nil {
		return empty, err
	}
	if len(body) > 65536 {
		return empty, oauthError()
	}
	var wire struct {
		Access  string `json:"access_token"`
		Refresh string `json:"refresh_token"`
		Type    string `json:"token_type"`
		Expires *int64 `json:"expires_in"`
		Scope   string `json:"scope"`
	}
	if json.Unmarshal(body, &wire) != nil {
		return empty, oauthError()
	}
	remember(r, wire.Access, wire.Refresh)
	if wire.Access == "" || !validBearerToken(wire.Access) || !strings.EqualFold(wire.Type, "Bearer") || wire.Expires != nil && (*wire.Expires <= 0 || *wire.Expires > 315360000) {
		return empty, oauthError()
	}
	if wire.Scope != "" {
		d.scope = wire.Scope
	}
	expires := int64(300)
	if wire.Expires != nil {
		expires = *wire.Expires
	}
	return oauthToken{AccessToken: wire.Access, RefreshToken: wire.Refresh, ExpiresAt: time.Now().Add(time.Duration(expires) * time.Second), Issuer: options.Issuer, Resource: d.resource, ClientID: options.ClientID, Scope: d.scope}, nil
}
func acquireToken(ctx context.Context, r core.Request) (string, error) {
	cache := &oauthCache{}
	if r.RunState != nil {
		value, _ := r.RunState.LoadOrStore("mcp.oauth", cache)
		cache = value.(*oauthCache)
	}
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if usableToken(cache.token, r.Target.MCP.OAuth) {
		return cache.token.AccessToken, nil
	}
	options := r.Target.MCP.OAuth
	if options.TokenFile != "" && cache.token.AccessToken == "" {
		token, err := readToken(options.TokenFile)
		if err == nil {
			if token.Issuer != options.Issuer || token.ClientID != options.ClientID || !sameResource(token.Resource, r.Target.Endpoint) {
				return "", oauthError()
			}
			remember(r, token.AccessToken, token.RefreshToken)
			cache.token = token
			if usableToken(token, options) {
				return token.AccessToken, nil
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", oauthError()
		}
	}
	if options.Grant == "authorization_code" && cache.token.RefreshToken == "" {
		return "", fail(core.Misconfigured, "mcp_login")
	}
	discovery, err := discoverOAuth(ctx, r)
	if err != nil {
		return "", err
	}
	values := url.Values{}
	if options.Grant == "client_credentials" {
		values.Set("grant_type", "client_credentials")
	} else {
		refresh := cache.token.RefreshToken
		if refresh == "" {
			refresh = r.Environment[options.RefreshTokenEnv]
		}
		if refresh == "" {
			return "", oauthError()
		}
		values.Set("grant_type", "refresh_token")
		values.Set("refresh_token", refresh)
	}
	token, err := exchangeToken(ctx, r, discovery, values)
	if err != nil {
		return "", err
	}
	if token.RefreshToken == "" {
		token.RefreshToken = cache.token.RefreshToken
	}
	if options.TokenFile != "" {
		if err := writeToken(options.TokenFile, token); err != nil {
			return "", oauthError()
		}
	}
	cache.token = token
	return token.AccessToken, nil
}
func sameResource(stored, endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	if u.Path == "/" {
		u.Path = ""
	}
	return stored == u.String()
}
func readToken(path string) (oauthToken, error) {
	var token oauthToken
	info, err := os.Lstat(path)
	if err != nil {
		return token, err
	}
	if !info.Mode().IsRegular() || runtime.GOOS != "windows" && info.Mode().Perm()&0077 != 0 || info.Size() > 65536 {
		return token, oauthError()
	}
	file, err := os.Open(path)
	if err != nil {
		return token, err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 65537))
	if err != nil || len(data) > 65536 || json.Unmarshal(data, &token) != nil {
		return token, oauthError()
	}
	return token, nil
}
func writeToken(path string, token oauthToken) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0700); err != nil {
		return err
	}
	dirInfo, err := os.Stat(directory)
	if err != nil || runtime.GOOS != "windows" && dirInfo.Mode().Perm()&0022 != 0 {
		return oauthError()
	}
	if info, err := os.Lstat(path); err == nil && !info.Mode().IsRegular() {
		return oauthError()
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	file, err := os.CreateTemp(directory, ".agenthealth-token-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if err := json.NewEncoder(file).Encode(token); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
func randomString() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// Login obtains a user-delegated OAuth token with PKCE and a loopback callback.
// The writer receives the browser URL and completion message, never tokens.
func Login(ctx context.Context, target core.Target, out io.Writer) error {
	if target.MCP == nil || target.MCP.OAuth == nil || target.MCP.OAuth.Grant != "authorization_code" {
		return oauthError()
	}
	if err := (core.Config{Version: "v1", Targets: []core.Target{target}}).Validate(); err != nil {
		return oauthError()
	}
	client := core.NewHTTPClient()
	defer client.CloseIdleConnections()
	r := core.Request{Target: target, Client: client, Environment: map[string]string{}}
	for _, ref := range target.CredentialReferences() {
		r.Environment[ref] = os.Getenv(ref)
		if r.Environment[ref] == "" {
			return oauthError()
		}
	}
	discovery, err := discoverOAuth(ctx, r)
	if err != nil {
		return err
	}
	if !secureURL(discovery.metadata.AuthorizationEndpoint) {
		return oauthError()
	}
	found := false
	for _, method := range discovery.metadata.PKCEMethods {
		if method == "S256" {
			found = true
		}
	}
	if !found {
		return oauthError()
	}
	issuer, _ := url.Parse(discovery.metadata.Issuer)
	authorization, _ := url.Parse(discovery.metadata.AuthorizationEndpoint)
	if issuer.Scheme == "https" && authorization.Scheme != "https" {
		return oauthError()
	}
	verifier, err := randomString()
	if err != nil {
		return oauthError()
	}
	state, err := randomString()
	if err != nil {
		return oauthError()
	}
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])
	port := 0
	if target.MCP.OAuth.RedirectPort != nil {
		port = *target.MCP.OAuth.RedirectPort
	}
	listener, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return oauthError()
	}
	defer listener.Close()
	redirect := "http://" + listener.Addr().String() + "/callback"
	type callback struct {
		code string
		err  error
	}
	callbacks := make(chan callback, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" || req.URL.Path != "/callback" || req.Host != listener.Addr().String() {
			http.NotFound(w, req)
			return
		}
		q, err := url.ParseQuery(req.URL.RawQuery)
		if err != nil || len(q["state"]) != 1 || q.Get("state") != state {
			http.Error(w, "Invalid OAuth callback", 400)
			return
		}
		validIssuer := q.Get("iss") == discovery.metadata.Issuer
		if len(q["iss"]) > 1 || discovery.metadata.IssuerResponse && !validIssuer || q.Get("iss") != "" && !validIssuer {
			http.Error(w, "Invalid OAuth issuer", 400)
			select {
			case callbacks <- callback{err: oauthError()}:
			default:
			}
			return
		}
		if len(q["code"]) != 1 || q.Get("code") == "" || q.Get("error") != "" {
			http.Error(w, "OAuth authorization failed", 400)
			select {
			case callbacks <- callback{err: oauthError()}:
			default:
			}
			return
		}
		select {
		case callbacks <- callback{code: q.Get("code")}:
		default:
		}
		fmt.Fprint(w, "Authorization received. You can close this window.")
	})
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second}
	defer server.Close()
	go server.Serve(listener)
	query := authorization.Query()
	query.Set("response_type", "code")
	query.Set("client_id", target.MCP.OAuth.ClientID)
	query.Set("redirect_uri", redirect)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	query.Set("state", state)
	query.Set("resource", discovery.resource)
	if discovery.scope != "" {
		query.Set("scope", discovery.scope)
	}
	authorization.RawQuery = query.Encode()
	if _, err := fmt.Fprintf(out, "Open this URL to authorize AgentHealth:\n%s\n", authorization.String()); err != nil {
		return err
	}
	var response callback
	select {
	case <-ctx.Done():
		return ctx.Err()
	case response = <-callbacks:
	}
	if response.err != nil {
		return response.err
	}
	token, err := exchangeToken(ctx, r, discovery, url.Values{"grant_type": {"authorization_code"}, "code": {response.code}, "code_verifier": {verifier}, "redirect_uri": {redirect}})
	if err != nil {
		return err
	}
	if err := writeToken(target.MCP.OAuth.TokenFile, token); err != nil {
		return oauthError()
	}
	_, err = fmt.Fprintln(out, "OAuth login complete; credentials saved to the configured token file.")
	return err
}

func oauthDo(r core.Request, req *http.Request) (*http.Response, error) {
	if r.MarkResponse != nil {
		req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{GotFirstResponseByte: r.MarkResponse}))
	}
	resp, err := r.Client.Do(req)
	if resp != nil && r.MarkResponse != nil {
		r.MarkResponse()
	}
	return resp, err
}

func validBearerToken(value string) bool {
	if value == "" {
		return false
	}
	padding := false
	for _, c := range value {
		if c == '=' {
			padding = true
			continue
		}
		if padding || !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || strings.ContainsRune("-._~+/", c)) {
			return false
		}
	}
	return true
}

func usableToken(token oauthToken, options *core.MCPOAuth) bool {
	if !validBearerToken(token.AccessToken) || !time.Now().Add(5*time.Second).Before(token.ExpiresAt) {
		return false
	}
	scopes := map[string]bool{}
	for _, scope := range strings.Fields(token.Scope) {
		scopes[scope] = true
	}
	for _, scope := range options.Scopes {
		if !scopes[scope] {
			return false
		}
	}
	return true
}

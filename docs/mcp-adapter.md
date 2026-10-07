# MCP health adapter

MCP targets support Streamable HTTP (JSON and SSE responses) and explicitly
configured stdio subprocesses. Supported revisions are `2025-03-26`,
`2025-06-18`, `2025-11-25`, and the stateless `2026-07-28` protocol.

```bash
agenthealth ping mcp http://localhost:3000/mcp
agenthealth doctor mcp http://localhost:3000/mcp
agenthealth doctor examples/mcp-check/agenthealth.yaml
```

HTTP endpoints must be the server's MCP URL; the adapter does not append
`/mcp`. The deprecated separate HTTP+SSE transport is not implemented.

## Protocol selection

With no version pin, checks probe `server/discover` using `2026-07-28`.
Modern requests carry per-request version, client identity, and capabilities,
plus matching `MCP-Protocol-Version` and `Mcp-Method` HTTP headers. Modern
functional calls also mirror `Mcp-Name` and valid `x-mcp-header` parameters,
encoding unsafe values with the protocol's Base64 sentinel. Invalid header
annotations exclude a tool from the discovered inventory. Results must be
complete; input-required results are inconclusive and never cause additional
functional calls.

A legacy HTTP rejection without a recognized modern error triggers
initialization-based negotiation. On stdio, an unrecognized RPC error or an
unanswered 500 ms discovery probe triggers legacy initialization. Recognized
modern version/header errors never silently downgrade. Legacy negotiation
offers `2025-11-25` and accepts any supported legacy revision. Set
`mcp.protocol_version` to pin a revision and disable cross-era fallback.
A version mismatch is `MISCONFIGURED`.

Modern checks do not send `initialize`, `notifications/initialized`, session
IDs, or DELETE. Legacy checks initialize independent sessions, validate server
identity and capabilities, send the initialized notification, and forward any
server-assigned session ID and negotiated version. Each HTTP session receives
best-effort DELETE cleanup within a maximum 250 ms budget and the check
deadline; cleanup errors do not change observations. Canceled sessions expire
on the server. Sessions are not recreated within a check.

These behaviors follow the upstream [versioning contract](https://modelcontextprotocol.io/specification/2026-07-28/basic/lifecycle),
[modern HTTP transport](https://modelcontextprotocol.io/specification/2026-07-28/basic/transports/streamable-http),
and [legacy lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle).
Only the listed revisions are supported; future revisions require explicit implementation.

## Passive checks and expectations

Defaults include configuration, reachability, authentication, protocol,
capability, and latency. HTTP reachability sends HEAD and accepts any HTTP
status, including 405, as connectivity evidence. HTTP latency measures HEAD,
excluding discovery and tool execution. DNS, TCP, TLS, and HTTP stages appear
when observed. Stdio reachability measures subprocess startup and protocol
establishment; it has no HTTP transport stages.

Authentication and protocol establish server compatibility. Capability
additionally lists every advertised tools, resources, and prompts inventory.
Unadvertised inventories are skipped. Lists follow pagination, validate
identity fields and tools' object input schemas, and reject repeated cursors
and duplicate identities. Passive checks do not read resources, retrieve
prompts, or invoke tools. Each check owns its session/process; targets and
dependencies do not share credentials or capabilities.

Required tool and prompt names match exactly; resources match by URI. Missing
requirements, including an unadvertised inventory, yield `DEGRADED`.
Requirements need `capability` when `checks` is explicitly specified.
Doctor displays dimension-level evidence and canonical messages, without
printing inventories or raw server responses.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: search-mcp
    type: mcp
    endpoint: https://example.com/mcp
    auth:
      bearer_env: MCP_TOKEN
    mcp:
      protocol_version: '2026-07-28'
      required_tools: [search]
      required_resources: ['health://ready']
      required_prompts: [summary]
```

Each response is capped at 1 MiB, each inventory at 100 pages and 10,000
entries, within the engine's deadline. SSE parsing stops at the response
instead of waiting for the stream to close. Notifications are ignored and
unsupported server requests are rejected. Over-limit responses/discovery are
`UNKNOWN`; malformed protocol and rejected RPCs are `UNHEALTHY`. HTTP 401/403
is `MISCONFIGURED`; other unexpected HTTP statuses are `UNHEALTHY`.
Connectivity failures and partial-response timeouts use the engine's normal
classifications. Functional calls never retry or resume.

## Stdio subprocesses

Set `mcp.transport: stdio` and supply a command and argument array. The
`stdio://` endpoint identifies the local target; it is not interpreted as a
shell command. Only explicitly configured executables run, without a shell.
Use trusted server commands, whose startup behavior the operator has reviewed.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: local-server
    type: mcp
    endpoint: stdio://local-server
    mcp:
      transport: stdio
      stdio:
        command: python3
        args: [server.py]
        directory: /path/to/server
        env:
          API_TOKEN: LOCAL_API_TOKEN
```

`env` maps child environment names to host environment references. Missing
values fail configuration. The child otherwise inherits only basic runtime
variables (`PATH`, `HOME`, `USER`, temporary-directory variables, and Windows
system-directory variables). Other parent environment variables are excluded.
Stdio cannot combine with HTTP bearer or OAuth options. Stderr is discarded;
stdout must contain bounded newline-delimited JSON-RPC messages.

Input closes at cleanup, with a short grace period before forced termination.
Cancellation closes streams and terminates the child. On POSIX platforms,
children in the server's process group are terminated too; servers must not
escape that group. Other platforms terminate the direct child. Linux subprocess
behavior is tested; other release platforms are cross-compiled, not runtime-tested.

## OAuth acquisition and login

For HTTP targets, choose either `auth.bearer_env` for an existing token or
`mcp.oauth` for token acquisition. OAuth requires a pre-registered `client_id`
and an explicitly trusted `issuer`. The adapter discovers protected resource
metadata and authorization-server/OIDC metadata, validates the issuer and
resource binding, and requests tokens for that MCP resource. HTTPS is required
except for explicit loopback URLs. Redirects are never followed. Challenge
metadata must share the MCP endpoint's origin; authorization-server endpoints
come from the pinned issuer's validated metadata.

For automated checks, use the [client credentials extension](https://modelcontextprotocol.io/extensions/auth/oauth-client-credentials)
with a secret reference:

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: automated-mcp
    type: mcp
    endpoint: https://example.com/mcp
    mcp:
      oauth:
        issuer: https://auth.example.com
        client_id: agenthealth-service
        grant: client_credentials
        client_secret_env: MCP_CLIENT_SECRET
        scopes: [read]
```

Client authentication uses `client_secret_basic` or `client_secret_post`,
selected from server metadata. Access tokens are reused only within the target's
run, with reacquisition before expiry. Optional `token_file` persists tokens.
The client credentials extension is declared in modern request capabilities.

For servers requiring user approval, configure authorization code login:

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: user-mcp
    type: mcp
    endpoint: https://example.com/mcp
    mcp:
      oauth:
        issuer: https://auth.example.com
        client_id: registered-public-client
        grant: authorization_code
        token_file: .credentials/user-mcp.json
        scopes: [read]
        redirect_port: 8765
```

```bash
agenthealth login agenthealth.yaml user-mcp
agenthealth check agenthealth.yaml
```

Login prints a browser URL and waits up to five minutes for approval via a
127.0.0.1 callback. The client registration must allow that loopback callback
(`http://127.0.0.1:8765/callback` in this example); omitting `redirect_port` selects
an ephemeral port. Login uses PKCE S256, unpredictable state, and validates the
callback issuer when present or advertised as required. A confidential client
can additionally reference `client_secret_env`.

Health checks never launch interactive login. They use saved access tokens and
refresh expired tokens, persisting rotated refresh tokens. Missing login
credentials yield `MISCONFIGURED` with login guidance. An externally supplied
refresh token can instead use `grant: refresh_token` and `refresh_token_env`.
Configured scopes override challenge/metadata scopes; otherwise challenged
scopes take priority over metadata scopes. Tokens without `expires_in` receive
a conservative five-minute lifetime. No token is refreshed by retrying a
functional call after a server rejection.

Token files bind credentials to issuer, resource, and client ID. Files use
0600 permissions; new parent directories use 0700. Symlink token files and
files accessible to other users are rejected on POSIX. Saving requires a
parent directory that is not writable by other users. Keep token directories
outside version control. Tokens, refresh tokens, environment credentials, and
server content never appear in health output; login prints no tokens.

OAuth dynamic client registration, JWT client assertions, and automatic scope
escalation are not implemented. Existing client registration is required.
See the upstream [authorization contract](https://modelcontextprotocol.io/specification/2026-07-28/basic/authorization).

## Functional invocation

Functional checks require explicit `checks: [functional]` opt-in, a named tool,
and `safe: true`. The operator must verify that the tool and arguments are
non-destructive. The discovered tool must additionally declare
`readOnlyHint: true` and `destructiveHint: false`; annotations alone cannot
prove safety. The adapter never chooses a tool automatically.

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: search-probe
    type: mcp
    endpoint: http://localhost:3000/mcp
    checks: [protocol, capability, functional, latency]
    mcp:
      required_tools: [search]
      functional:
        tool: search
        safe: true
        arguments_json: '{"query":"health"}'
```

`arguments_json` is an optional JSON object encoded as a YAML string, defaulting
to `{}`. The Go loader parses it and enforces a 64 KiB byte limit in addition to
schema validation. Missing tools are `DEGRADED`, unsafe annotations are
`MISCONFIGURED`, and `isError: true` is `UNHEALTHY`. A complete valid success is
`HEALTHY`; the adapter does not assess content quality. Discovery and the one
functional invocation share the same check-owned transport.

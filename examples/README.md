# Example configurations

These configurations are templates for your own running services. Replace
placeholder endpoints, issuer URLs, client IDs, command paths, and required
inventories before using them. They do not start an example server or provision
an OAuth client. Build the CLI as described in [CLI usage](../docs/cli.md).

| Example | What it demonstrates |
|---|---|
| [HTTP readiness](core-check/agenthealth.yaml) | HTTP status checks and latency thresholds |
| [MCP discovery](mcp-check/agenthealth.yaml) | Required tool names and resource URIs, with automatic protocol selection |
| [Modern MCP](mcp-check/modern.yaml) | Pinning stateless MCP `2026-07-28` |
| [MCP functional probe](mcp-check/functional.yaml) | One explicitly safe tool invocation |
| [MCP stdio](mcp-stdio/agenthealth.yaml) | A configured local subprocess |
| [OAuth browser login](mcp-oauth/agenthealth.yaml) | Authorization code + PKCE, private token storage, and subsequent refresh |
| [OAuth client credentials](mcp-oauth/client-credentials.yaml) | Noninteractive token acquisition using `MCP_CLIENT_SECRET` |
| [OAuth refresh token](mcp-oauth/refresh-token.yaml) | Acquisition using `MCP_REFRESH_TOKEN`, with persisted token rotation |
| [A2A passive peer](a2a-check/agenthealth.yaml) | A2A 0.3.0 card discovery, required skill, passive protocol/authentication, and latency |
| [A2A bearer credentials](a2a-check/bearer.yaml) | Passive checks using `AGENT_TOKEN` |
| [A2A custom card](a2a-check/custom-card.yaml) | A custom card path on the target origin |
| [A2A minimal interaction](a2a-check/functional.yaml) | One explicitly safe text interaction |

## Passive MCP checks

Point the HTTP examples at the actual MCP endpoint, including `/mcp` only if
that is the server's configured route. Default selection probes the modern
protocol and falls back to legacy initialization when appropriate. A pinned
modern target requires a server supporting `2026-07-28`.

```bash
agenthealth check examples/mcp-check/agenthealth.yaml
agenthealth check examples/mcp-check/modern.yaml --format json
agenthealth doctor examples/mcp-check/agenthealth.yaml
```

For stdio, replace the Python command, server script, and working directory with
your trusted server's values. Add `mcp.stdio.env` references if it needs
credentials; see [stdio setup](../docs/mcp-adapter.md#stdio-subprocesses).

```bash
agenthealth check examples/mcp-stdio/agenthealth.yaml
```

## OAuth checks

Register a client with the authorization server first, then configure the
matching issuer, client ID, scopes, and MCP endpoint. For the browser example,
allow the callback `http://127.0.0.1:8765/callback` in the client registration.
Login prints the authorization URL for you to open and saves credentials
under `.credentials/`. Health checks use saved tokens and refresh them without
prompting for browser approval.

```bash
agenthealth login examples/mcp-oauth/agenthealth.yaml user-mcp
agenthealth check examples/mcp-oauth/agenthealth.yaml
```

For automated acquisition, set `MCP_CLIENT_SECRET` or `MCP_REFRESH_TOKEN` in
your environment as appropriate; keep the actual values out of these files.
The refresh-token example persists rotated tokens in its private token file.

```bash
agenthealth check examples/mcp-oauth/client-credentials.yaml
agenthealth check examples/mcp-oauth/refresh-token.yaml
```

See [OAuth configuration and security](../docs/mcp-adapter.md#oauth-acquisition-and-login).
Existing client registration is required; dynamic registration and JWT client
assertions are not implemented.

## Functional probes

Review the configured tool and arguments for non-destructive behavior before
using the functional example. Its `safe: true` declaration is an operator
assertion. The server must also advertise `readOnlyHint: true` and
`destructiveHint: false`. Each probe calls the selected tool once, without
retries; the example expects a tool named `search` accepting a `query` string.

```bash
agenthealth check examples/mcp-check/functional.yaml
```

See [functional invocation](../docs/mcp-adapter.md#functional-invocation).

## A2A peer checks

[A2A configuration](a2a-check/agenthealth.yaml) checks a local A2A 0.3.0 JSON-RPC
peer at port 9000 with a required `health` skill. Start your own peer and adjust
the endpoint and skill expectation to match its card.

```bash
agenthealth check examples/a2a-check/agenthealth.yaml --format json
```

See the [A2A example setup guide](a2a-check/README.md) and
[A2A adapter guide](../docs/a2a-adapter.md) for bearer credentials,
capability expectations, custom card URLs, and opt-in interactions.

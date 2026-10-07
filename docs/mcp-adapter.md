# MCP health adapter

The MCP adapter supports the initialization-based Streamable HTTP transport at
an explicit HTTP/HTTPS MCP endpoint. It supports protocol revisions 2025-03-26,
2025-06-18, and 2025-11-25. It offers 2025-11-25 by default and accepts any of
these revisions negotiated by the server. Set `mcp.protocol_version` to pin one
revision; a different negotiated revision is `MISCONFIGURED`.

```bash
agenthealth ping mcp http://localhost:3000/mcp
agenthealth doctor mcp http://localhost:3000/mcp
agenthealth doctor examples/mcp-check/agenthealth.yaml
```

The endpoint must be the server's MCP URL; the adapter does not append `/mcp`.
Stdio, the older separate HTTP+SSE transport, and the newer stateless protocol
are not implemented. See the upstream [2025-11-25 transport contract](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
and [initialization lifecycle](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle).

## Passive checks

Defaults include configuration, reachability, authentication, protocol,
capability, and latency. Reachability sends HEAD and accepts any HTTP status,
including 405, as connectivity evidence. Latency measures that HEAD request;
it does not measure discovery or tool execution. DNS, TCP, TLS, and HTTP stages
appear when observed, using the same engine diagnostics as the HTTP adapter.

Authentication and protocol each initialize a fresh session, validate the
JSON-RPC envelope, server identity, capabilities and negotiated version, then
send `notifications/initialized`. Capability additionally lists every
advertised tools, resources, and prompts inventory. Unadvertised inventories
are skipped. Lists follow pagination, validate identity fields, reject
repeated cursors and duplicate identities, and inspect tools' object input
schemas. These checks do not read resources, retrieve prompts, or invoke tools.

Checks use independent sessions, including for duplicate endpoints and nested
dependencies, so one check cannot reuse another target's credentials or stale
capabilities. A server-assigned session ID is forwarded with the negotiated
version on subsequent requests. Session termination uses best-effort DELETE
with a maximum 250 ms budget within the check deadline; unsupported termination
or cleanup failure does not change the health observation. Canceled sessions
are left to server expiry. No expired session is recreated within a check.

JSON and SSE replies are supported. SSE parsing stops at the response instead
of waiting for a persistent stream to close. Server notifications are ignored;
server requests are rejected because the client advertises no optional client
capabilities. Each response stream is capped at 1 MiB, each inventory at 100
pages and 10,000 entries, all within the engine's per-check deadline. Over-limit
responses or discovery are `UNKNOWN`; malformed protocol and rejected JSON-RPC
requests are `UNHEALTHY`. HTTP 401/403 is `MISCONFIGURED`. Other unexpected HTTP
statuses are `UNHEALTHY`. Connectivity failures and partial response timeouts
follow the engine's normal classifications and retry policy.

## Expectations and credentials

Required tool and prompt names match exactly; resources match by URI. Missing
requirements, including an unadvertised required inventory, yield `DEGRADED`.
Requirements need a capability check when `checks` is explicitly specified.
The doctor output shows dimension-level evidence and canonical troubleshooting
messages; it does not print the server's inventory or raw responses.

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
      protocol_version: '2025-11-25'
      required_tools: [search]
      required_resources: ['health://ready']
      required_prompts: [summary]
```

Bearer tokens are resolved from environment references. OAuth discovery,
interactive login, refresh, and token acquisition are not implemented. TLS is
verified, redirects are not followed, and only the configured endpoint receives
credentials. Server messages, content, session IDs and invocation arguments
never appear in results or diagnostics.

## Functional invocation

Functional checks require explicit `checks: [functional]` opt-in, a named tool,
and `safe: true`. The operator must verify that the tool and arguments are
non-destructive. The adapter additionally requires the discovered tool to have
`readOnlyHint: true` and `destructiveHint: false`; annotations alone cannot prove
safety. It never chooses a tool automatically and never retries a tool call.

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
schema validation. An absent tool is `DEGRADED`, unsafe annotations are
`MISCONFIGURED`, and a tool result with `isError: true` is `UNHEALTHY`. A valid
success result is `HEALTHY`; the adapter does not verify the semantic quality of
returned content. Functional discovery and invocation share one isolated session.

# Agent Router integration

`router` targets use GET on explicitly configured read-only HTTP health,
liveness or readiness signals. Protocol accepts 2xx by default; configured
status/header expectations override the defaults. Authentication failures
(401/403) are MISCONFIGURED; other unexpected statuses are UNHEALTHY.
Reachability establishes connectivity regardless of HTTP status.

```bash
agenthealth ping router http://localhost:8081/ready
agenthealth check examples/router-check/agenthealth.yaml --format json
agenthealth doctor examples/router-check/agenthealth.yaml
```

## Route and backend evidence

Use separate named dependencies for router signals, direct backends and each
configured routed endpoint. HTTP/API, MCP and A2A routes retain their own adapter
contracts, credentials and functional opt-in. Multiple routes to one backend
execute independently. [Phase 10](dependency-graph.md) supports a shared backend `id` referenced by multiple edges, with each route kept as a distinct node.
Critical dependencies propagate failure; optional failures degrade the aggregate.
Backends and routes execute even when the router signal fails. Inspect checks
and dependency results to distinguish evidence from aggregate status.

A bounded functional GET through a read-only configured route requires
`checks: [functional]`. HTTP body expectations require that opt-in and use
`max_body_bytes` (default 64 KiB, maximum 1 MiB). Active probes never retry. Reachability and authentication preflight also issue
GET requests, so every configured signal/probe endpoint must be read-only.
MCP tool calls and A2A tasks use their existing explicit safe-probe contracts.
A CLI route probe establishes the CLI-to-route path; use the Phase 7 agent
safe downstream handler to establish communication from the direct agent runtime.

## Supported interfaces and limits

This is a vendor-neutral HTTP interface integration, with no product/version
certification. Configure actual deployment endpoints; no default admin paths,
ports, route inventory or backend URLs are inferred. Availability means observed
responses from configured endpoints, not router membership or failover state.
A failed route alone cannot establish whether the router or backend caused it.
Separate signals can provide additional evidence without asserting causality.

Bearer credentials use environment references. Verified TLS, redirect rejection,
deadlines, cancellation, response bounds and allowlisted diagnostics reuse the
[HTTP contract](http-adapter.md). Names should avoid sensitive topology; neither
response bodies nor expected values are emitted as diagnostics. No discovery,
model generation, arbitrary mutation or automatic failover probe is performed.
Local tests cover signal/backend/path failure isolation and aggregation;
real product interoperability remains Phase 25 work.

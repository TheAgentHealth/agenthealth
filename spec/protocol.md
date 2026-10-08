# Agent Health Protocol (AHP)

Status: experimental v1 HTTP binding, implemented in source. This is a pre-1.0
reference contract, not an independently ratified industry standard.
See the [Phase 11 design and impact review](../docs/rfcs/phase-11-ahp.md).

## Version and transport

Use HTTP GET with JSON responses. HTTPS termination is required for bearer
credentials outside a trusted local connection. Responses include `AHP-Version:
v1`, `Content-Type: application/json`, `Cache-Control: no-store` and
`X-Content-Type-Options: nosniff`. Clients may send `AHP-Version: v1`; another
version returns 400. AHP version and AHS `spec_version` are independent.
The [envelope schema](schemas/ahp.schema.json) defines the binding.
Other transports and vendor extension fields are not defined in v1.

## Operations and semantics

| GET path | Access | Meaning |
|---|---|---|
| `/health` | Public | Aggregate health summary |
| `/ready` | Public | Same aggregate evidence, readiness decision |
| `/live` | Public | Refresh loop running; independent of target health |
| `/health/dependencies` | Bearer | Full recursive result evidence with topology |
| `/health/capabilities` | Bearer | Same evidence, including requested capability and functional dimensions |

Successful envelopes contain `ahp_version`, `spec_version`, `status` and, for
completed checks, UTC `observed_at`. Detailed endpoints add `evidence` containing
the existing [result document](result-schema.md), including graph IDs,
relationships, critical edge policy, dimension checks and recursive dependencies.
They do not infer path success from endpoint, gateway or router health. Capability
and functional evidence remain separate. Unrequested checks remain absent.

Health/readiness returns 200 for HEALTHY or DEGRADED, 503 for other states.
Optional dependency degradation therefore remains ready. Liveness returns 200
while the refresh loop runs and 503 otherwise. Before the first snapshot, after
shutdown, or when the snapshot exceeds its maximum age, health is UNKNOWN/503.
Detailed stale responses contain only an error, never old evidence.

Errors contain `ahp_version` and a stable `error`: `not_found` (404),
`method_not_allowed` (405, Allow: GET), `unsupported_version` (400),
`unauthorized` (401, WWW-Authenticate: Bearer), or `snapshot_unavailable` (503).
Clients must inspect HTTP status and body; no redirects or protocol negotiation
fallback are specified. Endpoint availability and the version header provide
explicit discovery; no automatic discovery advertisement is implemented.

## Execution and security

Requests read snapshots and never execute checks. The reference implementation
refreshes serially every 30 seconds after each completed run, using the existing
60-second engine budget; snapshots expire after two minutes. API callers can
configure bounded refresh and age policies. Existing check opt-in, timeout,
retry, concurrency and credential rules apply. Explicit functional checks in a
serving configuration run on every refresh; review these before starting.

Public responses disclose only aggregate status and observation time. Detailed
endpoints require a configured bearer token and grant access to all configured
result topology. The token is an environment reference in the CLI, compared in
constant time, and never emitted. Result output uses the engine's existing
validation and redaction. Target names and IDs are visible to authorized clients;
use a dedicated configuration if different audiences need separate scopes.

Bind to loopback by default. For remote access use a TLS proxy with rate limits
and network restrictions. The server bounds header sizes and read/write/idle
timeouts. Probe frequency is independent of request volume. Per-client rate
limiting belongs at the proxy; multi-tenant authorization is future work.

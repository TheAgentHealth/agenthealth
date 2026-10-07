# Agent Health Protocol (AHP)

> **Status: Proposed / experimental.** AHP is a direction, not a shipped standard. See [Phase 9 — Agent Health Protocol (AHP)](../ROADMAP.md#phase-9--agent-health-protocol-ahp) in the roadmap. Nothing in this document is normative yet.

## Goal

Define a vendor-neutral protocol through which agentic systems can **expose and exchange** standardized health and readiness information. AHP implements the semantics defined by the [Agent Health Specification](health-model.md) — it does not redefine health, only how health is transmitted.

## Why this is not defined yet

Running health checks with the `agenthealth` CLI does not, by itself, constitute a protocol. A protocol requires a defined, versioned, interoperable wire contract that multiple independent implementations can agree on. Per the roadmap, AHP work begins only after the HTTP, MCP, A2A, Agent, and Dependency Graph phases prove the AHS model works in practice (Phases 4–8).

## Proposed core operations (non-normative)

```text
GET /health
GET /ready
GET /live
GET /health/dependencies
GET /health/capabilities
```

HTTP is a likely first binding, not necessarily the only one. AHP may eventually need to map onto other transports (e.g. MCP or A2A native mechanisms) rather than assuming HTTP is universal.

## Proposed response envelope (non-normative)

<!-- spec-example: skip reason="illustrative AHP sketch only; intentionally omits target.name (not yet a defined requirement for AHP) so it does not validate against the normative Result schema" -->
```json
{
  "spec_version": "v1",
  "status": "DEGRADED",
  "target": {
    "type": "agent"
  },
  "checks": {},
  "dependencies": []
}
```

This reuses the [recursive result shape](result-schema.md#recursive-shape) that the CLI is designed to produce (see [Phase 3](../ROADMAP.md#phase-3--universal-cli)); AHP would formalize it as a network-exchanged contract rather than a CLI output format. Each entry in `dependencies` would follow the same recursive shape, not a flattened summary.

## Open questions to resolve before this becomes normative

- Protocol versioning and negotiation
- Readiness vs liveness semantics (distinct from each other)
- Dependency representation over the wire (inline vs reference)
- Capability health representation
- Error representation (vs a `200` with a `DEGRADED` body)
- Authentication and authorization of the health endpoint itself
- Content types (`application/json` only, or others)
- Caching behavior (can `/health` responses be cached, and for how long)
- Timeout semantics for slow dependency checks
- An extension mechanism for vendor-specific fields
- Discovery: how a system advertises AHP support
- Security: information disclosure, dependency redaction, public vs authenticated health detail, rate limiting

## Security

Any eventual AHP implementation must, at minimum:

- never expose secrets or credentials in responses,
- distinguish public (unauthenticated) health summaries from authenticated/detailed health,
- support rate limiting to prevent health endpoints from becoming a denial-of-service vector,
- redact sensitive dependency metadata unless explicitly authorized.

## Conformance

AHP implementations should eventually be testable independently of the AgentHealth reference implementation (CLI, SDKs). No conformance suite exists yet.

## Relationship to other spec documents

AHP is one possible transport/exchange binding for the semantics defined in [health-model.md](health-model.md) and the shape defined in [result-schema.md](result-schema.md). It does not replace either.

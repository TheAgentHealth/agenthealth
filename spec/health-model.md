# Agent Health Specification (AHS)

> Status: Draft. This is the most mature layer of the project (see [Specification → Protocol → Implementation](../README.md#specification--protocol--implementation)).

The Agent Health Specification defines the vocabulary and health model used by every AgentHealth target type, adapter, and output format. It defines **what health means**, independent of any transport, language, or vendor.

## Health States

AgentHealth defines a common health vocabulary:

| Status | Meaning |
|---|---|
| `HEALTHY` | Target is operating normally |
| `DEGRADED` | Target works but one or more non-fatal conditions are impaired |
| `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| `UNREACHABLE` | Target cannot be contacted |
| `MISCONFIGURED` | Configuration prevents a valid health determination or operation |
| `UNKNOWN` | Health cannot currently be determined |

Adapters may expose additional diagnostic detail, but MUST map their overall result onto one of these six states.

## Health Dimensions

Each dimension answers a narrower question about a target.

### Reachability

Can AgentHealth establish communication with the target? (TCP/HTTP connectivity, MCP transport connectivity, A2A endpoint availability, model API availability, database connectivity.)

### Protocol Health

Does the target correctly support the expected protocol? (MCP initialization, protocol version compatibility, A2A discovery, required endpoint validation, capability negotiation.)

### Authentication

Can the configured identity authenticate successfully? (API keys, OAuth credentials, service credentials, workload identity, token validity.) Secrets MUST NOT be exposed in AgentHealth output.

### Capability Health

Does the target expose the capabilities required by the workload? (MCP tools/resources, agent skills, model availability, expected API operations, required runtime features.)

### Functional Health

Can the target successfully perform a minimal, safe operation? (MCP `tools/list`, minimal model request, agent discovery, lightweight API operation, database `SELECT 1`, vector store metadata/read.) Functional checks SHOULD be safe and non-destructive by default.

### Dependency Health

Are the target's required dependencies healthy? An agent may be reachable while a dependency is not; overall status reflects dependency propagation rules (critical vs optional — see [Status Aggregation](#status-aggregation)).

### Latency Health

Response latency compared against configurable thresholds (e.g. `latency_ms` vs `thresholds.latency_ms`).

### Configuration Health

Detects configuration problems (missing endpoint/credentials, invalid protocol selection, unavailable model, missing required capability, malformed configuration, incompatible protocol version) and reports them as `MISCONFIGURED` rather than misclassifying the target as unreachable.

## Target Model

AgentHealth targets are typed (`agent`, `mcp`, `a2a`, `model`, `http`, `database`, `vector-store`, and others). See [target-model.md](target-model.md) for the full list, required fields, and which health dimensions apply to which target type.

## Status Aggregation

When a target has dependencies, overall status MUST be derived from:

- the target's own check results across all dimensions,
- its dependencies' statuses,
- whether each dependency is marked `critical`,
- configured thresholds (e.g. latency).

A critical dependency in `UNREACHABLE` or `UNHEALTHY` state SHOULD propagate to the parent target's overall status. Non-critical (optional) dependency failures SHOULD be reflected without necessarily failing the parent target outright (e.g. resulting in `DEGRADED` rather than `UNHEALTHY`).

## Relationship to other spec documents

- Target types and per-type applicable dimensions: [target-model.md](target-model.md)
- Wire-level result representation: [result-schema.md](result-schema.md)
- Declarative target/check/dependency configuration: [configuration.md](configuration.md)
- How adapters must implement these semantics for a given technology: [adapter-spec.md](adapter-spec.md)
- CLI exit code mapping: [exit-codes.md](exit-codes.md)
- Security requirements for any conforming implementation: [security.md](security.md)
- How these semantics could be exposed/exchanged over a network: [protocol.md](protocol.md) (experimental)

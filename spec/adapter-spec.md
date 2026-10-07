# Adapter Contract

> Status: Draft. Defines what a technology-specific adapter (MCP, A2A, HTTP, model, database, vector store, community adapters) must implement to plug into the AgentHealth core without modifying it. See [Adapter Architecture](../README.md#adapter-architecture) and [Phase 21 — Plugin / Adapter Ecosystem](../ROADMAP.md#phase-21--plugin--adapter-ecosystem).

## Responsibilities of an adapter

An adapter translates technology-specific behavior into the common [Agent Health Specification](health-model.md) vocabulary. It must:

1. Declare **metadata**: adapter name, version, supported target type(s), compatibility version.
2. Accept **configuration**: the subset of the [configuration spec](configuration.md) relevant to its target type.
3. Implement applicable **health checks**: reachability, protocol, authentication, capability, functional, latency, configuration — an adapter is not required to implement every dimension if a dimension does not apply to its technology.
4. Report **capabilities**: what the target exposes (tools, resources, skills, operations, features).
5. Produce a **normalized result**: output conforming to the [result schema](result-schema.md), never a technology-specific shape.
6. Normalize **errors**: internal/transport errors must be mapped to `UNREACHABLE` or `MISCONFIGURED` rather than leaking raw exceptions.
7. Declare **security requirements**: what credentials/permissions the adapter's checks require, and which of its checks are active vs passive (see [Passive vs Active Checks](../README.md#passive-vs-active-checks)).

## Required metadata

| Field | Description |
|---|---|
| `name` | Adapter identifier (e.g. `mcp`, `a2a`, `http`) |
| `version` | Adapter version |
| `target_types` | One or more [target types](target-model.md) this adapter handles |
| `compatibility_version` | Core engine / adapter interface version this adapter was built against |

## Safety requirements

- Functional checks MUST be non-destructive by default.
- Any check capable of side effects (invoking a tool, running inference, writing data) MUST be explicit, opt-in (active check) per [Passive vs Active Checks](../README.md#passive-vs-active-checks).
- Adapters MUST NOT log or emit secrets, tokens, or credentials in results or diagnostics.
- Adapters SHOULD support configurable timeouts and fail safely (`UNKNOWN`/`UNREACHABLE`) on timeout rather than hanging.

## Compatibility

Adapters declare a `compatibility_version` against the core engine's adapter interface so the engine can reject or warn about incompatible adapters rather than failing unpredictably. The exact versioning scheme will be finalized alongside [Phase 21 — Plugin / Adapter Ecosystem](../ROADMAP.md#phase-21--plugin--adapter-ecosystem).

## Conformance

Independent conformance testing for adapters is tracked under [Phase 22 — AgentHealth Conformance](../ROADMAP.md#phase-22--agenthealth-conformance).

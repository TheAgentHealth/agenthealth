# Adapter Contract

> Status: Draft. Defines what a technology-specific adapter (MCP, A2A, HTTP, model, database, vector store, community adapters) must implement to plug into the AgentHealth core without modifying it. See [Adapter Architecture](../README.md#adapter-architecture) and [Phase 21 — Plugin / Adapter Ecosystem](../ROADMAP.md#phase-21--plugin--adapter-ecosystem).

## Responsibilities of an adapter

An adapter translates technology-specific behavior into the common [Agent Health Specification](health-model.md) vocabulary. It must:

1. Declare **metadata**: adapter name, version, supported target type(s), compatibility version.
2. Accept **configuration**: the subset of the [configuration spec](configuration.md) relevant to its target type.
3. Implement applicable **health checks**: reachability, protocol, authentication, capability, functional, latency, configuration — an adapter is not required to implement every dimension if a dimension does not apply to its technology. Respect [check prerequisites](health-model.md#check-prerequisites): a dimension blocked by a failed prerequisite (e.g. Protocol/Authentication/Capability after a failed Reachability check) MUST be omitted from `checks` entirely, never recorded as `UNKNOWN`.
4. Report **capabilities**: what the target exposes (tools, resources, skills, operations, features).
5. Produce a **normalized result**: output conforming to the [result schema](result-schema.md), never a technology-specific shape.
6. Normalize **errors**: classify failures per the shared [Error Classification table](health-model.md#error-classification) rather than inventing adapter-specific mappings or leaking raw exceptions.
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
- Functional checks are **active** checks (see [Passive vs Active Checks](../README.md#passive-vs-active-checks)) and MUST NOT execute unless explicitly requested via `checks` (see [configuration.md § Default Checks](configuration.md#default-checks)).
- Any check capable of side effects (invoking a tool, running inference, writing data) MUST be explicit, opt-in (active check) per [Passive vs Active Checks](../README.md#passive-vs-active-checks).
- Adapters MUST NOT log or emit secrets, tokens, or credentials in results or diagnostics.
- Adapters SHOULD support configurable timeouts. A timeout with no response received MUST be classified `UNREACHABLE`; an ambiguous or partial response received before the timeout MUST be classified `UNKNOWN` — per the [Error Classification table](health-model.md#error-classification), never an adapter-specific choice between the two.

## Compatibility

Adapters declare a `compatibility_version` against the core engine's adapter interface so the engine can reject or warn about incompatible adapters rather than failing unpredictably. The exact versioning scheme will be finalized alongside [Phase 21 — Plugin / Adapter Ecosystem](../ROADMAP.md#phase-21--plugin--adapter-ecosystem).

## Conformance

Independent conformance testing for adapters is tracked under [Phase 22 — AgentHealth Conformance](../ROADMAP.md#phase-22--agenthealth-conformance).

## Reference Go interface (Phase 2 draft)

The compiled-in interface is implemented in [core/adapter.go](../core/adapter.go). `Registry.Register` requires `compatibility_version: v1`, rejects duplicate target types, and requires Configuration and Reachability gates. The core owns Latency and Dependency; adapters declare their other supported dimensions and active dimensions. Functional must be declared active. Active dimensions never run by default.

`Adapter.Check` receives a context and a request containing the target, resolved credential, a TLS-verifying HTTP client, and `MarkResponse`. Call `MarkResponse` immediately when any response arrives so a deadline while reading a partial response can be classified `UNKNOWN`. Return `Observation.ResponseReceived` as well when a call completes. Use `Failure` for semantic errors with a known health status, rather than raw diagnostic text.

Adapters must honor cancellation, treat request data and metadata as immutable, use the supplied client for HTTP requests, and keep active probes non-destructive. The engine bounds calls and returns at deadlines even if an adapter ignores cancellation, but Go cannot forcibly terminate arbitrary in-process code; an uncooperative call occupies a pool slot until it returns. Adapter panics and invalid check statuses normalize to `UNKNOWN`.

The engine emits canonical diagnostics instead of forwarding adapter messages or panic/error strings. Resolved credentials are redacted throughout the result tree. Adapters must also avoid logging raw requests, credentials, responses, and errors themselves. Adapter implementation remains a trusted-code boundary.

### Run-scoped credentials and state (Phase 5 draft)

The Go request exposes `Environment`, a per-target snapshot of configured
credential references, and `RunState`, a concurrent map owned by one target
execution. Adapters can reuse acquired credentials between that target's
checks without sharing state with another target or run. `RememberSecret`
registers dynamically obtained access/refresh tokens for engine redaction;
adapters must call it before exposing any observations involving those tokens.
Raw credentials and state never become result fields. This extends the request
context without changing the required Adapter methods or result contract.

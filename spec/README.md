# AgentHealth Specification

This directory contains the specification documents for AgentHealth, introduced in the main [README](../README.md#specification--protocol--implementation).

| Layer | Document | Status |
|---|---|---|
| Agent Health Specification (AHS) | [health-model.md](health-model.md) | Draft — most mature layer |
| Target Model | [target-model.md](target-model.md) | Draft |
| Agent Health Protocol (AHP) | [protocol.md](protocol.md) | Experimental HTTP v1, implemented in source |
| Configuration | [configuration.md](configuration.md) | Draft |
| Result Schema | [result-schema.md](result-schema.md) | Draft |
| Exit Codes | [exit-codes.md](exit-codes.md) | Draft |
| Adapter Contract | [adapter-spec.md](adapter-spec.md) | Draft |
| Security Requirements | [security.md](security.md) | Draft |
| JSON Schemas | [schemas/](schemas/) | Draft — `result.schema.json`, `configuration.schema.json`, experimental `ahp.schema.json` |

## Relationship between documents

```text
health-model.md   -> defines WHAT health means (states, dimensions)
target-model.md   -> defines target types and per-type applicable dimensions
result-schema.md  -> defines how a single check result is represented
configuration.md  -> defines how targets/checks/dependencies are declared
exit-codes.md     -> defines the CLI exit code contract for automation
adapter-spec.md   -> defines the contract a technology-specific adapter must satisfy
security.md       -> defines security requirements any conforming implementation must satisfy
schemas/          -> machine-enforceable JSON Schemas for result-schema.md and configuration.md
protocol.md       -> (experimental) defines HOW health is exposed/exchanged over the wire
```

## Implemented adapters

Reference implementations currently support [Qdrant/Pinecone vector stores](../docs/vector-store-adapter.md), [native databases](../docs/database-adapter.md), [model/LLM APIs](../docs/model-adapter.md), [Agent Router HTTP health signals](../docs/agent-router.md), [Agentgateway HTTP health signals](../docs/agentgateway.md), [Agent/multi-agent](../docs/agent-adapter.md), [HTTP/API](../docs/http-adapter.md),
[MCP](../docs/mcp-adapter.md), and [A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility](../docs/a2a-adapter.md).
See [A2A configuration](configuration.md#a2a-expectations-phase-6-draft) and
[Phase 6 examples](../examples/a2a-check/README.md) for discovery, credentials,
expected skills/capabilities, and opt-in interaction. The target vocabulary
also includes future adapters; implementation availability follows the roadmap.

[Direct and composite agent health](target-model.md#direct-and-composite-agent-health)
covers the user's first agent and the downstream agents it communicates with,
directly or through A2A. Models, tools, and data services remain supporting
dependencies. `multi-agent` describes the cooperating system as a whole.

## Versioning

Specification documents use a `spec_version` field (e.g. `v1`) independent of the AgentHealth software release version. Breaking changes to any document here require a version bump and a changelog entry in this directory.

## Proposing changes

Specification changes are more sensitive than implementation changes because downstream adapters, SDKs, and (eventually) AHP conformance tests depend on them. See [GOVERNANCE.md](../GOVERNANCE.md) for the RFC-style process used for spec changes, and [CONTRIBUTING.md](../CONTRIBUTING.md) for the general contribution workflow.

Phase 8 Agentgateway integration is implemented in source. See [the gateway guide](../docs/agentgateway.md). Gateway checks use explicitly configured read-only HTTP signals and separate backend/path dependencies.

Phase 9 Agent Router HTTP integration is released in v0.7.0.
See [the router guide](../docs/agent-router.md). Named router signals, direct backends and configured paths retain separate evidence.

## Dependency graph extension

Phase 10 is included in v0.8.0 as an additive v1 configuration/result extension. See the [graph contract](../docs/rfcs/phase-10-dependency-graph.md) and [phase alignment audit](../docs/phase-10-alignment.md). Older v0.7.0 binaries do not include it.

## Experimental AHP HTTP v1

[Protocol](protocol.md) defines a separately versioned envelope around existing
AHS results. Serving adds no configuration fields, health states, dimensions,
target types or changes to CLI health exit codes. See the
[alignment audit](../docs/phase-11-alignment.md).

## Kubernetes consumers

[Phase 14 deployment tooling](../docs/kubernetes.md) consumes the existing v1
configuration, AHS results, CLI exit mapping and experimental AHP HTTP binding.
It adds no specification fields or adapter contracts. The
[alignment audit](../docs/phase-14-alignment.md) records each earlier-phase
impact decision; Kubernetes examples validate embedded configuration against
these schemas. A chart version remains independent of `spec_version`.

The [Python SDK](../docs/python-sdk.md) consumes existing v1 configuration and result schemas; schema equality is checked in SDK tests.

The [JavaScript / TypeScript SDK](../docs/javascript-sdk.md) also wraps the
exact-version CLI and preserves the existing configuration, result and health
exit-code contracts. Its bundled schemas are parity-tested against the specification.

Phase 17 implements [model/LLM health](../docs/model-adapter.md) for OpenAI-compatible and Anthropic APIs.
Passive model listings and explicit minimal inference preserve existing health
semantics. See the [Phase 17 alignment audit](../docs/phase-17-alignment.md) for cross-phase impacts.

Phase 18 adds [native database health](../docs/database-adapter.md) for PostgreSQL, MySQL, Redis and MongoDB.
Fixed SQL queries require explicit functional opt-in. See the
[alignment audit](../docs/phase-18-alignment.md) and [disposable examples](../examples/database-check/README.md).

Phase 19 implements [Qdrant and Pinecone vector-store checks](../docs/vector-store-adapter.md)
with explicit read-only query opt-in. See the [phase alignment audit](../docs/phase-19-alignment.md).

[Phase 20 runtime integrations](../docs/framework-integrations.md) implement the
existing agent resource contract; no v1 schema or exit-code migration is required.

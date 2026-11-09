# Documentation guide

Start with [installation](installation.md), then choose an adapter and deployment
mode. [Capability status](phase-status.md) is generated from [status.json](status.json).
Published release notes describe their tagged versions; source-only improvements
require a build containing the change.

| Area | Guides |
|---|---|
| Commands and output | [CLI](cli.md), [Go engine](../core/README.md), [result contract](../spec/result-schema.md), [exit codes](../spec/exit-codes.md) |
| HTTP/API | [HTTP health checks](http-adapter.md) |
| MCP | [HTTP/stdio, OAuth and safe tools](mcp-adapter.md) |
| A2A | [Cards, skills and safe tasks](a2a-adapter.md) |
| Direct/composite agents | [Application contract](agent-adapter.md), [reference middleware](reference-middleware.md) |
| Gateway/router | [Gateway HTTP health signals](agentgateway.md), [router HTTP health signals](agent-router.md) |
| Dependency graphs | [Graph guide](dependency-graph.md) |
| AHP serving | [Serving](ahp.md), [wire contract](../spec/protocol.md) |
| Distribution | [Docker](docker.md), [packages and OCI](distribution.md), [Kubernetes/Helm](kubernetes.md) |
| Safety | [Security policy](../SECURITY.md), [threat model](threat-model.md) |
| Validation | [Stabilization](stabilization.md), [interoperability](interoperability.md) |
| Design and releases | [Specification](../spec/README.md), [roadmap](../ROADMAP.md), [release notes](releases/), [detailed phase designs](roadmap-details.md) |

Examples are templates requiring your services and credentials. Schema/example
validation proves structure, not connectivity or product interoperability.
See [examples](../examples/README.md) and [testing](stabilization.md) for repeatable
local commands and remaining acceptance criteria.

The [Python SDK](python-sdk.md) provides typed sync/async CLI access; see its
[impact review](phase-15-alignment.md) and [example](../examples/python-sdk/README.md).

- [JavaScript / TypeScript SDK](javascript-sdk.md) — Node ESM wrapper and npm distribution.
- [Phase 16 alignment](phase-16-alignment.md) — earlier/later phase impact review.

Phase 17 implements [model/LLM health](model-adapter.md) for OpenAI-compatible and Anthropic APIs.
Passive model listings and explicit minimal inference preserve existing health
semantics. See the [Phase 17 alignment audit](phase-17-alignment.md) for cross-phase impacts.

Phase 18 adds [native database health](database-adapter.md) for PostgreSQL, MySQL, Redis and MongoDB.
Fixed SQL queries require explicit functional opt-in. See the
[alignment audit](phase-18-alignment.md) and [disposable examples](../examples/database-check/README.md).

Phase 19 implements [Qdrant and Pinecone vector-store checks](vector-store-adapter.md)
with explicit read-only query opt-in. See the [phase alignment audit](phase-19-alignment.md).

Phase 20 supplies the supported [LangGraph/FastAPI runtime integration](framework-integrations.md)
with safe peer hooks and an SDK consumer node. Other framework products remain planned.

Phase 21: [CI/CD deployment gates](ci-cd.md), [provider examples](../examples/ci/README.md)
and [cross-phase alignment](phase-21-alignment.md).

## Observability export (Phase 22)

Released in v0.20.0: `--format prometheus|logs|otlp|webhook` produces
redacted snapshots. See the [export guide](../docs/observability.md) and
[cross-phase alignment](../docs/phase-22-alignment.md). Delivery is explicit
and external; SDK JSON, AHP, health codes and adapter checks remain unchanged.

## Executable adapter ecosystem (Phase 23)

Released in v0.21.0: explicit local manifests and SHA-256-pinned executable
adapters with bounded JSON transport, cancellation and normalized engine output.
See the [guide](plugins.md) and [all-phase alignment](phase-23-alignment.md).
Processes retain the invoking user’s OS permissions; no sandbox or certification
is implied. Built-in adapters and AHS/AHP v1 contracts remain compatible.

## Conformance evidence

[Phase 24](conformance.md) validates existing v1 result evidence and runs mandatory
reference behavioral cases. It changes no AHS/AHP contracts and awards no
certification or product compatibility status.

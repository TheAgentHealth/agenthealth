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

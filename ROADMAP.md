# AgentHealth Roadmap

## Current: stabilization and security

Before expanding adapters and SDKs, reconcile current security boundaries,
make passive HTTP method selection explicit, reduce documentation drift, and
build repeatable coverage, fuzzing and graph profiling evidence. See the
[stabilization plan](docs/stabilization.md), [threat model](docs/threat-model.md)
and [impact review](docs/rfcs/stabilization.md).

## Next: real interoperability

Expand controlled official MCP/A2A SDK coverage to independently maintained
servers and frameworks. Test gateway/router products and a Kubernetes version
matrix, plus released-configuration upgrade and rollback. Product testing must
record exact versions, configuration, expected failures and results. Existing
SDK checks are useful evidence, but do not complete Phase 25.

## Later: capability expansion and production hardening

Python/TypeScript SDKs, model/database/vector adapters, framework integrations,
scoped authorization and isolated third-party adapters follow stabilization.
No blanket release freeze is imposed: fixes may ship under existing release
review, CI and publication policies. Do not claim a security audit, tenant
isolation or product compatibility without corresponding evidence.

## Phase Impact Review

For each substantive phase revision, assess earlier/later contracts, schemas,
code, examples, distribution and validation. Record required extensions or an
explicit no-impact rationale. Revisit before marking the phase complete.
See [contribution policy](CONTRIBUTING.md) and [detailed designs](docs/roadmap-details.md).

## Component Release Policy

Implemented distributions share the software version and immutable source tag.
See [release policy](RELEASING.md). Contract versions remain independent.

## Guiding Strategy

Health evidence must distinguish direct agents, composite agents, communication
paths and supporting dependencies. Implemented capability and measured
interoperability are separate claims. [Status](docs/phase-status.md) is generated
from [status.json](docs/status.json); historical release notes retain their scope.

# Phase 0 — Project Foundation

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-0--project-foundation).

# Phase 1 — Agent Health Specification

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-1--agent-health-specification).

# Phase 2 — Core Engine

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-2--core-engine).

# Phase 3 — Universal CLI

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-3--universal-cli).

# Phase 4 — Generic HTTP Health Adapter

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-4--generic-http-health-adapter).

# Phase 5 — MCP Health Adapter

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-5--mcp-health-adapter).

# Phase 6 — A2A Health Adapter

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-6--a2a-health-adapter).

# Phase 7 — Agent Health

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-7--agent-health).

# Phase 8 — Agentgateway Integration

Implemented scope: **gateway HTTP health signals**, configured backend/path evidence; product interoperability remains unverified.

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-8--agentgateway-integration).

# Phase 9 — Agent Router Integration

Implemented scope: **router HTTP health signals**, configured backend/path evidence; route decisions are not independently verified.

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-9--agent-router-integration).

# Phase 10 — Dependency Graph

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-10--dependency-graph).

# Phase 11 — Agent Health Protocol (AHP)

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-11--agent-health-protocol-ahp).

# Phase 12 — Docker Distribution

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-12--docker-distribution).

# Phase 13 — Standalone Binaries

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-13--standalone-binaries).

# Phase 14 — Kubernetes Integration

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-14--kubernetes-integration).

# Phase 15 — Python SDK / PyPI

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-15--python-sdk--pypi).

# Phase 16 — JavaScript / TypeScript SDK

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-16--javascript--typescript-sdk).

# Phase 17 — Model / LLM Adapters

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-17--model--llm-adapters).

# Phase 18 — Database Adapters

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-18--database-adapters).

# Phase 19 — Vector Store Adapters

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-19--vector-store-adapters).

# Phase 20 — Agent Runtime / Framework Integrations

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-20--agent-runtime--framework-integrations).

# Phase 21 — CI/CD Integrations

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-21--cicd-integrations).

# Phase 22 — Observability Export

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-22--observability-export).

# Phase 23 — Plugin / Adapter Ecosystem

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-23--plugin--adapter-ecosystem).

# Phase 24 — AgentHealth Conformance

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-24--agenthealth-conformance).

# Phase 25 — Ecosystem Interoperability Testing

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-25--ecosystem-interoperability-testing).

# Phase 26 — Production Hardening

See the [status index](docs/phase-status.md) and [phase design](docs/roadmap-details.md#phase-26--production-hardening).

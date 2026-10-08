# AgentHealth Roadmap

This roadmap describes the proposed development path for AgentHealth.

AgentHealth is intended to become a universal, vendor-neutral health, readiness, reachability, and diagnostics layer for AI agents and agentic infrastructure.

The roadmap is intentionally capability-driven rather than tied to fixed calendar dates.

See the [documentation guide](docs/README.md) for phase 0–10 references and examples.

## Foundation stabilization before Phase 7

**Status:** Released in [v0.4.0](https://github.com/TheAgentHealth/agenthealth/releases/tag/v0.4.0),
implemented in [PR #9](https://github.com/TheAgentHealth/agenthealth/pull/9).
The full capability roadmap remains the objective.

- [x] AHS required-capability and server-version classification alignment
- [x] Optional stable diagnostic codes in machine output and matching schema
- [x] Default A2A 1.0 JSON-RPC with explicit 0.3.0 compatibility mode
- [x] Shared per-target MCP sessions/inventories and A2A passive evidence
- [x] Official current/legacy SDK interoperability suite and nightly matrix
- [x] Vulnerability scanning, SHA-pinned Actions, deterministic archives,
      per-platform CycloneDX SBOMs and signed provenance/SBOM attestations

See the [contract RFC](docs/rfcs/foundation-health-contract.md),
[RFC issue](https://github.com/TheAgentHealth/agenthealth/issues/8),
[SDK interoperability matrix](docs/interoperability.md) and
[release notes](docs/releases/v0.4.0.md) for scope, impact and migration.

Earlier phases: AHS, schemas, core lifecycle/output, CLI defaults, MCP/A2A and
binary packaging receive foundation extensions; their historical baselines stay
complete. Later phases: agent health, graph, AHP, SDKs and conformance consume
status/code and shared per-target evidence; no topology or agent-role fields are
introduced here. v0.4.0 changes the A2A default and failure exit codes as documented.

Phase 7 is now implemented in source after this milestone. Phase 8 gateway integration is implemented in source; router integration and dependency graph execution are implemented in source; experimental AHP HTTP serving is implemented in source;
full distribution, SDK/ecosystem and observability/conformance remain later
milestones. Basic Docker/Kubernetes examples may proceed alongside agent health.
No additional adapter is declared implemented by this foundation milestone.

---

# Guiding Strategy

AgentHealth development follows five principles:

1. **Specification before ecosystem sprawl**
2. **Small, reliable core**
3. **Protocol-first interoperability**
4. **Universal distribution**
5. **Community-extensible adapters**

The project should avoid attempting to integrate every AI vendor during its initial development.

The first milestone is to prove that a common health model works across multiple agentic protocols and infrastructure types.

---

# Phase Impact Review

Every new phase or substantive revision to an existing phase must assess its
impact on both earlier and later phases before implementation. This applies
to all phases, not only agent, gateway, or router integrations.

Record the assessment in the phase proposal or roadmap section:

- **Earlier phases:** identify affected completed or planned capabilities, including AHS, schemas, core execution, CLI, adapters, and distribution. Add explicit extension tasks within the new phase where changes are required; completed phases retain their baseline status.
- **Later phases:** identify affected designs, dependencies, interfaces, examples, conformance requirements, and validation work. Update their planned deliverables and cross-references accordingly.
- **Compatibility and sequencing:** document any migration/versioning needs, prerequisite work, and changes to implementation order. Contract changes follow the specification process.
- **No impact:** explicitly record when no earlier or later phase changes are required, with a brief reason. Do not assume every phase must change AHS, the core, or the CLI.

Revisit the assessment when implementation reveals new effects. A phase is not
complete until its required updates to affected layers, documentation, and
validation are handled and affected future deliverables reflect the outcome.

---

# Impact of Agent Roles and Gateway/Router Integrations

The additions in Phases 7–9 expand planned coverage. Phase 1 remains complete
as an implemented AHS draft; Phases 7–9 explicitly include incremental AHS
updates alongside their implementation. Completed Phases 0–6 remain the
baseline, with compatibility assessed before changing their contracts.

- **AHS (Phase 1):** build on the implemented draft through explicit update items in Phases 7–9; retain the common health states and dimensions and review any required aggregation extensions. Direct/composite are agent roles, not new target types. Any new role, route, or relationship fields require specification review, matching schemas, and a compatibility assessment.
- **Core and CLI (Phases 2–3):** Phases 7–9 explicitly track required execution, validation, and output extensions; Phase 10 adds graph scheduling and diagnostics, and Phase 11 adds AHP serving support.
- **HTTP, MCP, and A2A (Phases 4–6):** reuse implemented endpoint checks where applicable. Endpoint health alone does not verify communication from the first agent to a downstream agent.
- **Dependency Graph (Phase 10):** distinguish agents, supporting dependencies, gateways/routers, and communication paths, including shared backends and multiple routes.
- **AHP (Phase 11):** exchange the evidence and relationships established by AHS and Phase 10. It must not assume a healthy gateway means a healthy backend or a working route.
- **Kubernetes, SDKs, runtime integrations, CI/CD, and observability:** consume the agreed results and distinguish first-agent, downstream-agent, path, and dependency failures. Final interfaces depend on the graph and AHP designs.
- **Conformance, interoperability, and hardening:** verify these combinations and failure cases before declaring the corresponding support stable.

The new phases add implementation and validation work before AHP. Phase numbers
express the capability sequence, not a promise that all unrelated distribution
work must wait. Any contract extensions remain proposals until approved through
the specification process; no schema or exit-code change is made by this roadmap.

---

# Phase 0 — Project Foundation

**Status:** Complete

Goal:

Establish the project identity, architecture, governance foundation, and technical direction.

## Deliverables

- [x] Project repository
- [x] README
- [x] Roadmap
- [x] Open-source license ([LICENSE](LICENSE), Apache-2.0)
- [x] CONTRIBUTING.md
- [x] CODE_OF_CONDUCT.md
- [x] SECURITY.md
- [x] Initial governance model ([GOVERNANCE.md](GOVERNANCE.md))
- [x] Architecture documentation ([docs/architecture.md](docs/architecture.md))
- [x] Design principles (see [README Design Principles](README.md#design-principles))
- [x] Issue templates (`.github/ISSUE_TEMPLATE/`)
- [x] Pull request templates (`.github/PULL_REQUEST_TEMPLATE.md`)
- [x] Release process ([RELEASING.md](RELEASING.md))
- [x] Semantic versioning policy ([RELEASING.md](RELEASING.md))
- [x] CI pipeline ([.github/workflows/ci.yml](.github/workflows/ci.yml)) — spec schema tests ([tests/spec/](tests/spec/)), a markdown link checker ([scripts/check_markdown_links.py](scripts/check_markdown_links.py)), and doc-example validation ([scripts/validate_doc_examples.py](scripts/validate_doc_examples.py)) that checks marked README/ROADMAP/spec prose examples against the schemas, all required to pass before merge. The Go core test suite runs in an additional CI job alongside these checks.

## Architecture Decisions

Define:

- [x] Repository strategy: **single repository (monorepo)**. AHS and AHP specification docs, the core engine, CLI, adapters, SDKs (Python/PyPI, JavaScript/npm), and distribution integrations (Docker, Kubernetes, CI) all live in this one repository rather than being split across separate repos. See [Repository Structure](README.md#repository-structure).
- [x] Core implementation language: **Go** (see [docs/architecture.md](docs/architecture.md#core-implementation-language))
- [x] CLI architecture: single binary, subcommands (see [docs/architecture.md](docs/architecture.md#cli-architecture))
- [x] Adapter interface: in-process/compiled-in initially (see [docs/architecture.md](docs/architecture.md#adapter-interface))
- [x] configuration format: YAML (see [spec/configuration.md](spec/configuration.md))
- [x] result format: JSON (see [spec/result-schema.md](spec/result-schema.md))
- [x] error model: normalized onto health states (see [docs/architecture.md](docs/architecture.md#error-model))
- [x] exit codes: normative mapping implemented in the engine and CLI (see [docs/architecture.md](docs/architecture.md#exit-codes))
- [x] plugin strategy: in-repo adapters initially (see [docs/architecture.md](docs/architecture.md#plugin-strategy))
- [x] SDK strategy: thin wrappers over the core engine (see [docs/architecture.md](docs/architecture.md#sdk-strategy))

---

# Phase 1 — Agent Health Specification

**Status:** Complete as a draft — vocabulary, target model, result schema, error classification, aggregation, active opt-in syntax, environment credential references, timeouts, and retries are defined in [spec/](spec/README.md) and implemented by Phase 2. The contract remains pre-1.0; additional authentication schemes and secret-manager integration remain future extensions.

Goal:

Define a language-neutral health specification before expanding integrations.

## Health States

Standardized in [spec/health-model.md](spec/health-model.md#health-states):

```text
HEALTHY
DEGRADED
UNHEALTHY
UNREACHABLE
MISCONFIGURED
UNKNOWN
```

## Health Dimensions

Defined in [spec/health-model.md](spec/health-model.md#health-dimensions):

- [x] reachability
- [x] protocol
- [x] authentication
- [x] capability
- [x] functional
- [x] dependency
- [x] latency
- [x] configuration

## Target Model

Defined in [spec/target-model.md](spec/target-model.md):

```text
agent
multi-agent
a2a
mcp
tool
model
llm
http
api
vector-store
database
runtime
gateway
custom
```

## Result Schema

Defined in [spec/result-schema.md](spec/result-schema.md).

Example:

<!-- spec-example: result -->
```json
{
  "spec_version": "v1",
  "target": {
    "name": "example",
    "type": "mcp"
  },
  "status": "HEALTHY",
  "latency_ms": 82,
  "checks": {
    "reachability": { "status": "HEALTHY" }
  },
  "dependencies": []
}
```

## Configuration Specification

Defined in [spec/configuration.md](spec/configuration.md):

- [x] target definitions
- [x] checks
- [x] thresholds
- [x] authentication references
- [x] dependencies
- [x] critical vs optional dependencies
- [x] timeout policies
- [x] retry policies
- [x] active/passive check behavior

Phase 2 resolves the baseline policy contract in [configuration execution policies](spec/configuration.md#execution-policies-phase-2-draft). Further extensions are listed under [Still to be defined](spec/configuration.md#still-to-be-defined).

## Specification Documents

Structure (matches the current [spec/](spec/README.md) directory; `README.md` and `protocol.md` were added beyond the original target list as an index and the experimental AHP document, respectively):

```text
spec/
├── README.md
├── health-model.md
├── target-model.md
├── configuration.md
├── result-schema.md
├── exit-codes.md
├── adapter-spec.md
├── security.md
├── protocol.md
└── schemas/
    ├── result.schema.json
    └── configuration.schema.json
```

---

# Phase 2 — Core Engine

**Status:** Complete — the [Go core engine](core/README.md) implements execution, policies, normalization, dependency evaluation, output, and safety controls. A small HTTP reference adapter exercises the engine; the CLI and HTTP diagnostics are implemented in Phases 3 and 4.

Goal:

Implement the reference health engine.

## Core Capabilities

- [x] Target loading
- [x] Configuration validation
- [x] Check execution
- [x] Timeouts
- [x] Retries
- [x] Latency measurement
- [x] Result normalization
- [x] Error normalization
- [x] Health aggregation
- [x] Dependency evaluation
- [x] Human-readable output
- [x] JSON output

## Safety

- [x] Secret redaction
- [x] Safe logging
- [x] TLS verification
- [x] Non-destructive defaults
- [x] Active-check controls
- [x] Timeout protection

---

# Phase 3 — Universal CLI

**Status:** Implemented — `ping`, `check`, `doctor`, and `version` are available with terminal, JSON, YAML, and specification exit codes. Agent/multi-agent, gateway/router HTTP health signals, HTTP/API, MCP, and A2A 1.0 JSON-RPC adapters with explicit 0.3.0 compatibility are available. See [CLI usage](docs/cli.md).

Goal:

Deliver the first usable AgentHealth experience.

## Commands

### Ping

```bash
agenthealth ping <type> <target>
```

Examples:

```bash
agenthealth ping http https://example.com
agenthealth ping mcp http://localhost:3000
```

A2A usage is implemented under Phase 6, extended in v0.4.0 for 1.0 JSON-RPC peers with explicit 0.3.0 compatibility.

### Check

```bash
agenthealth check agenthealth.yaml
```

### Doctor

```bash
agenthealth doctor agenthealth.yaml
```

### OAuth login

Added in Phase 5; see [CLI login usage](docs/cli.md#mcp-oauth-login).

```bash
agenthealth login examples/mcp-oauth/agenthealth.yaml user-mcp
```

### Version

```bash
agenthealth version
```

## Output Formats

- [x] terminal
- [x] JSON
- [x] YAML

Potential later formats:

- [ ] JUnit
- [ ] SARIF
- [ ] Prometheus

## Exit Codes

Standardize exit behavior for automation. Normative mapping defined in [spec/exit-codes.md](spec/exit-codes.md):

```text
0 = healthy
1 = degraded
2 = unhealthy
3 = unreachable
4 = misconfigured
5 = unknown / inconclusive result
6 = internal error (no result produced)
```

---

# Phase 4 — Generic HTTP Health Adapter

**Status:** Implemented — transport-stage diagnostics, passive status/header validation, bearer authentication, and bounded opt-in body matching are available. See [HTTP adapter](docs/http-adapter.md).

Goal:

Establish a universal baseline adapter.

## Checks

- [x] DNS resolution
- [x] TCP connectivity
- [x] TLS validation
- [x] HTTP connectivity
- [x] expected HTTP status
- [x] headers
- [x] authentication
- [x] response matching
- [x] latency
- [x] timeout

Example:

```bash
agenthealth ping http https://api.example.com/health
```

This adapter provides a foundation for services that do not implement agent-specific protocols.

---

# Phase 5 — MCP Health Adapter

**Status:** Implemented for Streamable HTTP (JSON/SSE), configured stdio, legacy initialization and modern stateless MCP 2026-07-28. OAuth client credentials, refresh, and PKCE login are available. See [MCP adapter](docs/mcp-adapter.md) for configuration and safety. Deprecated separate HTTP+SSE, dynamic OAuth registration, and JWT assertions remain outside current scope.

Goal:

Make MCP a first-class AgentHealth target.

## Capabilities

- [x] MCP Streamable HTTP transport connectivity (JSON/SSE)
- [x] bounded legacy SSE response resumption without repeating tool calls
- [x] tool content type and required-field validation
- [x] configured stdio subprocesses
- [x] modern stateless MCP `2026-07-28` discovery and request metadata
- [x] initialization
- [x] protocol negotiation
- [x] protocol version validation
- [x] capability discovery
- [x] bearer authentication
- [x] OAuth client credentials acquisition
- [x] OAuth refresh tokens and rotation
- [x] OAuth authorization code + PKCE login
- [x] private token storage and credential redaction
- [x] tools discovery
- [x] resources discovery
- [x] prompts discovery
- [x] required tool validation
- [x] required resource validation
- [x] latency measurement
- [x] safe functional invocation

Example:

```bash
agenthealth ping mcp http://localhost:3000/mcp
```

## MCP Doctor

```bash
agenthealth doctor mcp http://localhost:3000/mcp
```

Doctor displays dimension-level health evidence and canonical troubleshooting
messages. Required inventory failures are `DEGRADED`; use a configured target
to declare the expected tools/resources/prompts:

```bash
agenthealth doctor examples/mcp-check/agenthealth.yaml
```

See the [examples guide](examples/README.md) for stdio, modern protocol,
client credentials, browser login, refresh tokens, and functional probes.
See [CLI login usage](docs/cli.md#mcp-oauth-login) for the separate OAuth login
command; health checks remain noninteractive.

---

# Phase 6 — A2A Health Adapter

Goal:

Support Agent-to-Agent health validation.

**Status:** Implemented in v0.3.0 for A2A 0.3.0 JSON-RPC and extended in v0.4.0
with default A2A 1.0 JSON-RPC and explicit 0.3.0 compatibility.
See the [A2A adapter guide](docs/a2a-adapter.md) for configuration and limits.
Other protocol versions, REST/gRPC, streaming execution, extended cards, and
OAuth acquisition remain future work.

## Capabilities

- [x] endpoint discovery
- [x] agent metadata
- [x] protocol validation
- [x] authentication through environment-referenced bearer credentials
- [x] capability discovery
- [x] expected skill validation
- [x] minimal interaction
- [x] latency
- [x] error classification

Example:

```bash
agenthealth ping a2a https://agent.example.com
agenthealth doctor a2a https://agent.example.com
agenthealth check examples/a2a-check/agenthealth.yaml --format json
```

See [Phase 6 examples](examples/a2a-check/README.md) for passive discovery,
bearer credentials, a custom card URL, and opt-in safe interaction. Functional
probes never retry or poll; pending work is inconclusive.

---

# Phase 7 — Agent Health

**Status:** Released as a draft interface in [v0.5.0](docs/releases/v0.5.0.md),
implemented in [PR #19](https://github.com/TheAgentHealth/agenthealth/pull/19). See the
[adapter interface](docs/agent-adapter.md), [example](examples/agent-check/agenthealth.yaml),
and [contract/impact proposal](docs/rfcs/phase-7-agent-health.md).

Goal:

Check the user's first, user-facing agent itself, as well as the dependencies
used by the first agent and its downstream agents.

## Runtime prerequisite

The agent application must implement the [health resource and safe probe handler](docs/agent-adapter.md).
For first-to-downstream checks, that handler must actually contact the selected
peer and report completion. AgentHealth implements the probe client and result
classification; it does not install runtime handlers. Framework-specific handler
integrations remain Phase 20 work. See the [setup example](examples/agent-check/README.md).

## Direct Agent Health

The `agent` target includes the first agent that receives the user's request.
It may answer directly, use tools or a model, or delegate to other agents.
Calling another agent is not a requirement for checking its own health.

Implemented checks cover the agent's endpoint, authentication, metadata,
capabilities, liveness, and readiness to accept work. An explicitly opted-in,
bounded, non-destructive minimal task will provide functional evidence that
the agent can respond. Dependency health alone does not establish that the
agent itself works; checks should distinguish direct agent evidence from
dependency results and report inconclusive evidence as such.

## Composite Agent Health

Composite agents are the downstream agents that the user's first agent
communicates with, either directly or through A2A. Planned health checks cover
each downstream agent's endpoint, authentication, capabilities, readiness,
and optional safe functional task, plus the communication path from the first
agent. Results should distinguish a downstream agent failure from a failure
in that communication path.

Direct and composite describe agent roles in this roadmap, not new target
types. A `multi-agent` target describes the cooperating system as a whole.
Models, tools, MCP servers, data services, gateways, and routers remain
supporting dependencies and are covered by dependency health.

An agent health check may combine:

```text
User-facing agent endpoint
      │
      ├── Composite agents (direct communication or A2A)
      ├── Agentgateway / Agent Router (when used)
      ├── Model
      ├── MCP
      ├── Tools
      └── Data
```

## Capabilities

- [x] Update AHS for first (direct) and downstream (composite) agent roles, direct/A2A communication evidence, and agent liveness/readiness semantics.
- [x] Extend the Phase 1 AHS baseline as required, including configuration/result schemas, adapter contracts, documentation, and compatibility/versioning review.
- [x] Extend the Phase 2 core baseline as required to validate and execute the new checks, aggregate their results, and preserve distinct agent/path/dependency evidence; add meaningful fixtures and regression tests.
- [x] Extend the Phase 3 CLI baseline as required to expose the new checks through `ping`, `check`, and `doctor`, with consistent terminal/JSON/YAML output and the existing exit-code contract.
- [x] agent endpoint health
- [x] direct user-facing agent checks without requiring downstream agents
- [x] composite downstream agent health via direct integrations or A2A
- [x] first-to-downstream agent communication-path validation
- [x] agent authentication, liveness, and readiness
- [x] metadata validation
- [x] capability validation
- [x] dependency discovery
- [x] dependency execution
- [x] explicit opt-in, bounded, non-destructive minimal agent task
- [x] distinguish direct agent failures from dependency failures
- [x] dependency aggregation
- [x] critical dependency propagation

## Impact Review

Earlier: add optional v1 agent options/schema fixtures, strict Go validation,
agent diagnostics, per-target document sharing and CLI registration/tests;
reuse HTTP transport safety, A2A peer checks and independent critical/optional
dependency aggregation. Result shape, states and exit codes are unchanged.

Later: Phases 8–11 and their SDK/runtime/conformance consumers must preserve
named peer/path/dependency evidence, names-only configured discovery and the
distinction between runtime-declared readiness and completed task evidence.
Path handlers contact the selected peer directly or through A2A; AgentHealth
trusts their completion report and does not independently trace remote code.
Phase 8 gateway and Phase 9 router HTTP signals are implemented in source; graph scheduling is implemented in source; experimental AHP HTTP serving is implemented in source.

Compatibility: this additive v1 draft turns previously unsupported agent types
into executable targets. Existing supported targets keep their behavior.
Maintainer review is recorded by the PR #19 merge and RFC #18. Discovery requires explicit
configuration; arbitrary framework task APIs are outside this adapter contract.

---

# Phase 8 — Agentgateway Integration

**Status:** Released in [v0.6.0](docs/releases/v0.6.0.md) — HTTP health signals and explicitly configured protocol paths, implemented in [PR #21](https://github.com/TheAgentHealth/agenthealth/pull/21); product-version interoperability remains Phase 25 work.

Goal:

Check Agentgateway independently or as a dependency of the user-facing agent,
with separate evidence for gateway health and the health of its backends.

## Capabilities

- [x] Update AHS for gateway health, backend health, and gateway communication-path evidence, including their aggregation and redaction requirements.
- [x] Extend the Phase 1 AHS baseline as required, including configuration/result schemas, adapter contracts, documentation, and compatibility/versioning review.
- [x] Extend the Phase 2 core baseline as required to validate and execute the new checks, aggregate their results, and preserve distinct agent/path/dependency evidence; add meaningful fixtures and regression tests.
- [x] Extend the Phase 3 CLI baseline as required to expose the new checks through `ping`, `check`, and `doctor`, with consistent terminal/JSON/YAML output and the existing exit-code contract.
- [x] endpoint reachability and authentication
- [x] exposed health, liveness, and readiness signals
- [x] configured backend availability validation where supported
- [x] explicitly opted-in, bounded, non-destructive functional probes through the gateway
- [x] distinguish gateway failures from downstream agent, model, or tool failures
- [x] normalize results into the shared health model
- [x] document supported versions, transports, credentials, and diagnostic limits
- [x] examples connecting the user-facing agent, Agentgateway, and downstream targets

Start with existing HTTP/API, MCP, or A2A checks where Agentgateway exposes
those interfaces. The gateway adapter uses the existing `gateway` type and HTTP options; no new target type is introduced.

- [x] Add gateway-specific path evidence while preserving the Phase 7 configured peer/path results and safe probe contract. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: extend AHS/core/CLI as required and build on Phase 7 agent roles. Later: supply gateway/backend evidence to router, graph, AHP, and validation phases.

---

See the [Phase 8 RFC](docs/rfcs/phase-8-agentgateway.md) and
[integration guide](docs/agentgateway.md) for contracts, impact, and limitations.

# Phase 9 — Agent Router Integration

**Status:** Released in [v0.7.0](docs/releases/v0.7.0.md) — vendor-neutral HTTP signals and configured route/backend evidence, implemented in [PR #24](https://github.com/TheAgentHealth/agenthealth/pull/24); product interoperability remains Phase 25 work.

Goal:

Check an Agent Router independently or as a dependency of the user-facing
agent, with separate evidence for router health and the health of routed targets.

## Capabilities

- [x] Update AHS for router health, configured routes, routed backend health, and route-specific failures, including their aggregation and redaction requirements.
- [x] Extend the Phase 1 AHS baseline as required, including configuration/result schemas, adapter contracts, documentation, and compatibility/versioning review.
- [x] Extend the Phase 2 core baseline as required to validate and execute the new checks, aggregate their results, and preserve distinct agent/path/dependency evidence; add meaningful fixtures and regression tests.
- [x] Extend the Phase 3 CLI baseline as required to expose the new checks through `ping`, `check`, and `doctor`, with consistent terminal/JSON/YAML output and the existing exit-code contract.
- [x] endpoint reachability and authentication
- [x] exposed health, liveness, and readiness signals
- [x] configured route and backend availability validation where supported
- [x] explicitly opted-in, bounded, non-destructive functional probes through a configured route
- [x] distinguish router failures from downstream agent, model, or tool failures
- [x] normalize results into the shared health model
- [x] document supported products, versions, transports, credentials, and diagnostic limits
- [x] examples connecting the user-facing agent, Agent Router, and downstream targets

The `router` adapter uses explicitly configured read-only HTTP signals. Named
HTTP/MCP/A2A dependencies represent direct backends and configured routes;
functional probes require explicit opt-in. Product-specific discovery is
unsupported. See the [Phase 9 RFC](docs/rfcs/phase-9-agent-router.md),
[router guide](docs/agent-router.md) and [example](examples/router-check/README.md).
[Phase 10](docs/dependency-graph.md) now extends graph diagnostics and shared backend identity in source.

- [x] Add route-specific evidence without conflating Phase 7 peer health with first-runtime communication probes. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: add `router` to configuration/result type schemas and Go validation,
register the CLI adapter, and reuse HTTP safety plus independent dependency
aggregation. Preserve HTTP HEAD, gateway GET and Phase 7 runtime path probes.
No result structure, health-state or exit-code changes are introduced.

Later: graph, AHP, SDKs, observability and conformance must preserve named router
signals, direct backends and individual routes. Phase 10 implements explicit shared backend identity;
product-specific discovery remains unimplemented. Phase 25 must validate actual
product interoperability rather than infer it from generic HTTP tests.

Compatibility: existing configurations remain compatible; older binaries reject
`router`. See the [Phase 9 impact review](docs/rfcs/phase-9-agent-router.md).

---

# Phase 10 — Dependency Graph

**Status:** Included in [v0.8.0](docs/releases/v0.8.0.md) — explicit shared-node references, edge evidence, bounded parallel execution and tree output. See the [graph guide](docs/dependency-graph.md) and [contract review](docs/rfcs/phase-10-dependency-graph.md).

Goal:

Make dependency-aware diagnostics a defining AgentHealth capability.

The graph must cover the first (direct) agent, downstream (composite) agents,
direct or A2A communication paths, and any intervening gateway/router.
Agent roles and communication relationships need to be distinguishable from
supporting model, tool, MCP, and data dependencies. A healthy downstream agent
does not establish that the first agent can reach it through a configured route.

The current nested dependency contract is a starting point. Shared backends,
multiple routes, and path-specific health may require a richer graph model;
its additive v1 contract is recorded in the [Phase 10 review](docs/rfcs/phase-10-dependency-graph.md).

## Features

- [x] Consolidate AHS updates from Phases 7–9 into the graph configuration and result contracts
- [x] Dependency graph construction
- [x] Extend existing nested dependencies to the graph model
- [x] Preserve existing critical dependency policies in graph execution
- [x] Preserve existing optional dependency policies in graph execution
- [x] Extend existing failure propagation to graph nodes and communication paths
- [x] Cycle detection
- [x] Parallel health execution
- [x] Configurable concurrency
- [x] Dependency timeout budgets
- [x] Tree output
- [x] first-agent and downstream-agent relationships via direct integrations or A2A
- [x] gateway/router communication-path evidence separate from backend health
- [x] shared backend identity and multiple-route representation, preserving Phase 9 named router/backend/route evidence
- [x] distinguish node failures from communication-path failures in propagation and output

The Phase 2 core already executes nested dependencies and aggregates critical
and optional failures. The items above cover graph extensions and
preservation of that behavior, rather than initial implementation.

Example:

```text
Research Agent
│
├── Model ....................... HEALTHY
│
├── MCP
│   ├── GitHub .................. HEALTHY
│   └── Database ................ DEGRADED
│
├── Search Tool ................. HEALTHY
│
└── Vector Store ................ UNREACHABLE

Overall: DEGRADED
```

- [x] Resolve Phase 7 advertised dependency names only through explicit configuration and preserve agent, path and supporting dependency identity. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: consolidate Phases 7–9 contracts, extend core graph execution, and add CLI graph diagnostics. The [alignment audit](docs/phase-10-alignment.md) records affected earlier contracts, guides and examples, plus no-impact rationales. Later: establish graph evidence consumed by AHP, SDKs, integrations, and conformance.

---

# Phase 11 — Agent Health Protocol (AHP)

**Status:** Included in [v0.9.0](docs/releases/v0.9.0.md) — experimental HTTP v1

Goal:

Define a vendor-neutral protocol through which agentic systems can expose and exchange standardized health and readiness information.

The Agent Health Protocol implements the Phase 1 AHS baseline and its approved extensions from Phases 7–10.

This phase intentionally comes **after** the HTTP, MCP, A2A, Agent Health, Agentgateway, Agent Router, and Dependency Graph phases rather than immediately following the specification. The health semantics need to be proven across those adapters in reality first; AHP then generalizes the lessons learned into an interoperable wire contract. Standardizing an exchange protocol before understanding the operational requirements would risk locking in the wrong contract.

See the [wire contract](spec/protocol.md) for the implemented choices and limits.

## Core Operations

Implemented HTTP operations:

```text
GET /health
GET /ready
GET /live
GET /health/dependencies
GET /health/capabilities
```

HTTP JSON is the experimental v1 binding. Other transports remain future work.

## Protocol Model

Define:

- [x] protocol versioning
- [x] health response envelope
- [x] readiness semantics
- [x] liveness semantics
- [x] dependency representation preserving Phase 10 node IDs and edge relationship/critical policy, with topology redaction
- [x] capability health
- [x] error representation
- [x] authentication
- [x] authorization
- [x] content types
- [x] caching behavior
- [x] timeout semantics
- [ ] vendor extension mechanism (deferred; v1 defines no extension fields)
- [x] represent first/downstream agent relationships, supporting dependencies, and communication-path evidence established in Phase 10
- [x] distinguish gateway/router health from routed backend health without exposing sensitive topology by default

## Service Mode

- [x] implement AHP serving in the core and expose the Phase 3 CLI extension `agenthealth serve`
- [x] validate serving against the updated AHS/configuration/result contracts and document compatibility

A reference "server mode" lets AgentHealth expose AHP instead of only consuming other systems' health:

```bash
agenthealth serve examples/ahp-check/agenthealth.yaml
```

```text
                   Kubernetes
                       │
                       │ health/readiness
                       ▼
                ┌─────────────┐
                │ AgentHealth │
                │    AHP      │
                └──────┬──────┘
                       │
             ┌─────────┼──────────┐
             ▼         ▼          ▼
           Agent      MCP       Model
             │
             ├──────────────┐
             ▼              ▼
         A2A Agent       Vector DB
```

Service mode is implemented in source as an experimental pre-1.0 deliverable.

## Discovery

Explicit endpoints and the AHP-Version header identify support. Automatic discovery remains deferred.

## Security

Define:

- [x] information disclosure requirements
- [x] authentication mechanisms
- [x] dependency redaction
- [x] sensitive metadata handling
- [x] public vs authenticated health information
- [x] rate limiting recommendations

## Conformance

The envelope schema and fixtures validate the wire shape independently. A separate
implementation has not yet been tested; broader conformance remains Phase 24 work.

- [x] Exchange Phase 7 declared readiness and completed functional/path evidence separately; do not infer communication from endpoint health. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: review AHS/graph contracts for wire exchange, add core serving support, and extend the CLI with `serve`. Later: provide a versioned exchange contract for Kubernetes, SDK consumers, and protocol conformance.

Implementation choices, proxy rate limiting, authorization scope and transport
limits are defined in the [contract](spec/protocol.md), [guide](docs/ahp.md) and
[impact review](docs/rfcs/phase-11-ahp.md). Independent interoperability remains
unverified. See the [coverage and alignment audit](docs/phase-11-alignment.md).

---

# Phase 12 — Docker Distribution

**Status:** Implemented in source — a multi-stage, digest-pinned, distroless,
non-root image build ([Dockerfile](Dockerfile)) producing Linux AMD64/ARM64
binaries that are byte-identical to the attested release archives, validated
in CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) and published by
the release workflow ([.github/workflows/release.yml](.github/workflows/release.yml))
with BuildKit provenance, SBOM, and a GitHub-signed attestation, starting with
the next tagged release. No image has been published yet. See the
[container usage guide](docs/installation.md#container-image).

Goal:

Allow AgentHealth to run without installing a programming language runtime.

## Deliverables

- [x] Official container image (`ghcr.io/theagenthealth/agenthealth`; publishing gated on the next tagged release)
- [x] Multi-stage build
- [x] Minimal runtime image (digest-pinned `distroless/static-debian12:nonroot`, non-root `65532:65532`, no shell)
- [x] Linux AMD64
- [x] Linux ARM64
- [x] Signed images (keyless `actions/attest` attestation pushed to the registry)
- [x] SBOM (`sbom: true` plus a pushed build-provenance/SBOM attestation)
- [x] Versioned tags (`vX.Y.Z`, floating `vX` and `latest`, pre-releases excluded from floating tags)
- [x] Docker Hub mirror (`theagenthealth/agenthealth`), implemented in source and gated on `DOCKERHUB_USERNAME`/`DOCKERHUB_TOKEN` repository secrets

Example:

```bash
docker run --rm ghcr.io/theagenthealth/agenthealth \
  ping mcp http://host.docker.internal:3000
```

## Impact Review

Earlier: packages the existing CLI unchanged; no AHS, schema, core, CLI or
adapter contract changes. The release workflow gains a container job that runs
only after the binary release and verifies image binaries against the attested
archives (Phase 13). MCP stdio targets need their server executable in a derived
image, and interactive `agenthealth login` is not supported in the minimal image.
GitHub creates a new organization package as private by default, so the first
publish needs a one-time manual visibility change; the release smoke test logs
out of GHCR first so this is caught as a release failure instead of silently
passing (see [RELEASING.md](RELEASING.md)).
Later: Phase 14 consumes the image and its tags for probes, init containers and
Jobs; CI integrations can use it directly; hardening should add image
vulnerability scanning. The Docker Hub mirror publishes the same tags once its
secrets exist, with build provenance/SBOM but no GitHub-signed attestation.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 13 — Standalone Binaries

**Status:** Initial GitHub binary distribution implemented early for the v0.1.0 preview. Five platform archives, checksums, installation docs, and a tag-triggered release workflow are available. Per-platform CycloneDX SBOMs and keyless signed provenance/SBOM attestations are implemented in v0.4.0. OS-native executable signing, macOS notarization, additional package channels and broader runtime validation remain future work.

Goal:

Make installation trivial.

## Platforms

- [x] Linux AMD64
- [x] Linux ARM64
- [x] macOS ARM64
- [x] macOS AMD64
- [x] Windows AMD64

## Distribution

Potential channels:

- [x] GitHub Releases
- [ ] Homebrew
- [x] Linux standalone binary archives
- [ ] Linux packages (DEB/RPM)
- [ ] Windows package manager
- [x] container registries (GHCR, optional Docker Hub mirror; Phase 12)

## Impact Review

Earlier: extend existing binary release tooling; no AHS/core/CLI health-contract change is expected solely for distribution. Later: supply verified binaries to SDK wrappers, CI, and supported-platform validation.

---

# Phase 14 — Kubernetes Integration

**Status:** Planned — not yet implemented.

Goal:

Make AgentHealth useful as a cloud-native operational primitive.

## Readiness

```yaml
readinessProbe:
  exec:
    command:
      - agenthealth
      - check
      - /etc/agenthealth/config.yaml
```

## Planned Integrations

- [ ] Readiness probes
- [ ] Startup probes
- [ ] Init containers
- [ ] Jobs
- [ ] CronJobs
- [ ] Sidecar mode
- [ ] ConfigMap configuration
- [ ] Secret integration
- [ ] Helm examples
- [ ] Deployment validation
- [ ] Readiness examples for first agents, downstream agents, and gateway/router paths using the updated AHS results

## Future Exploration

- [ ] Kubernetes Operator
- [ ] Custom Resource Definitions
- [ ] Admission policies
- [ ] Health-aware rollout integration

Potential conceptual resource:

```yaml
apiVersion: agenthealth.io/v1alpha1
kind: AgentHealthCheck

metadata:
  name: research-agent

spec:
  target:
    type: agent
    endpoint: http://research-agent:8080

  dependencies:
    enabled: true
```

CRDs/operators are intentionally deferred until the basic health model proves useful.

## Impact Review

Earlier: consume updated AHS results, CLI checks, Docker images, and AHP serving where used; review any required readiness contract extensions. Later: provide deployment scenarios for CI, interoperability, and hardening.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 15 — Python SDK / PyPI

**Status:** Planned — not yet implemented.

Goal:

Allow Python applications and agent frameworks to consume AgentHealth programmatically.

## Distribution

```bash
pip install agenthealth
```

## API

Conceptual:

```python
from agenthealth import check

result = check(
    target_type="mcp",
    endpoint="http://localhost:3000"
)

print(result.status)
```

## Capabilities

- [ ] Core health API
- [ ] Async API
- [ ] Typed results
- [ ] Configuration loader
- [ ] Adapter access
- [ ] Dependency results
- [ ] Typed access to agent-role, gateway/router, and communication-path evidence defined by the updated AHS/graph contracts
- [ ] Exceptions/error model

The Python SDK must follow the Agent Health Specification rather than defining it.

- [ ] Expose Phase 7 agent options, named peer/path results and stable agent diagnostic codes through typed APIs. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

- [ ] Preserve Phase 10 node IDs and edge relationship/critical metadata in Python results. See the [graph contract](docs/rfcs/phase-10-dependency-graph.md).

## Impact Review

Earlier: wrap the core/CLI and updated AHS graph results without duplicating health semantics. Later: provide Python APIs for runtime integrations and interoperability tests.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 16 — JavaScript / TypeScript SDK

**Status:** Planned — not yet implemented.

Goal:

Provide first-class integration for the JavaScript agent ecosystem.

## Distribution

```bash
npm install @agenthealth/sdk
```

## API

Conceptual:

```typescript
import { check } from "@agenthealth/sdk";

const result = await check({
  type: "mcp",
  endpoint: "http://localhost:3000"
});
```

## Capabilities

- [ ] TypeScript types
- [ ] Promise-based API
- [ ] configuration support
- [ ] normalized results
- [ ] dependency results
- [ ] typed access to agent-role, gateway/router, and communication-path evidence defined by the updated AHS/graph contracts
- [ ] browser/server scope definition

- [ ] Expose Phase 7 agent options, named peer/path results and stable agent diagnostic codes through typed APIs. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

- [ ] Preserve Phase 10 node IDs and edge relationship/critical metadata in JavaScript/TypeScript results. See the [graph contract](docs/rfcs/phase-10-dependency-graph.md).

## Impact Review

Earlier: wrap the core/CLI and updated AHS graph results without duplicating health semantics. Later: provide JavaScript/TypeScript APIs for runtime integrations and interoperability tests.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 17 — Model / LLM Adapters

**Status:** Planned — not yet implemented.

Goal:

Validate the model dependencies used by agents.

## Common Checks

- [ ] endpoint
- [ ] authentication
- [ ] model availability
- [ ] minimal inference
- [ ] latency
- [ ] timeout
- [ ] rate limiting
- [ ] provider errors

Potential adapters:

```text
OpenAI-compatible APIs
Anthropic
Amazon Bedrock
Azure AI
Google Vertex AI
Ollama
others
```

Vendor-specific integrations should remain outside the core whenever practical.

## Impact Review

Earlier: add model adapters under the shared contracts and expose them through core/CLI registration; review any specification extensions. Later: provide model dependency coverage for runtime, conformance, and interoperability tests.

---

# Phase 18 — Database Adapters

**Status:** Planned — not yet implemented.

Potential targets:

```text
PostgreSQL
MySQL
Redis
MongoDB
```

## Common Checks

- [ ] connectivity
- [ ] authentication
- [ ] minimal safe query
- [ ] read readiness
- [ ] latency

## Impact Review

Earlier: add database adapters under the shared contracts and expose them through core/CLI registration; review any specification extensions. Later: provide database dependency coverage for runtime, conformance, and interoperability tests.

---

# Phase 19 — Vector Store Adapters

**Status:** Planned — not yet implemented.

Potential targets:

```text
Pinecone
Qdrant
Weaviate
Milvus
pgvector
OpenSearch
Elasticsearch
```

## Common Checks

- [ ] connectivity
- [ ] authentication
- [ ] collection/index availability
- [ ] metadata access
- [ ] minimal query
- [ ] latency

## Impact Review

Earlier: add vector-store adapters under the shared contracts and expose them through core/CLI registration; review any specification extensions. Later: provide retrieval dependency coverage for runtime, conformance, and interoperability tests.

---

# Phase 20 — Agent Runtime / Framework Integrations

**Status:** Planned — not yet implemented.

Goal:

Allow agent frameworks to expose and consume AgentHealth.

Potential integrations may include:

```text
LangGraph
Semantic Kernel
CrewAI
AutoGen
other agent runtimes
```

Possible integration patterns:

```text
framework plugin
middleware
health endpoint
SDK integration
adapter
```

- [ ] expose first-agent, downstream-agent, gateway/router, and communication-path evidence through supported framework integration patterns

AgentHealth should remain framework-neutral.

- [ ] Implement the Phase 7 safe health-resource handler in each supported runtime; downstream probes must actually contact the selected peer. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: integrate the agent roles, gateway/router paths, adapters, SDKs, and AHP where applicable; review any contract extensions. Later: supply framework scenarios for CI, observability, and interoperability.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 21 — CI/CD Integrations

**Status:** Planned — not yet implemented.

Goal:

Make agent health part of deployment validation.

Potential integrations:

- [ ] GitHub Actions
- [ ] GitLab CI
- [ ] Jenkins
- [ ] Azure DevOps
- [ ] generic shell integration

Example:

```text
Build
  ↓
Test
  ↓
Deploy
  ↓
AgentHealth
  ↓
Promote
```

- [ ] deployment checks and failure reporting for first agents, downstream agents, and gateway/router paths

Potential CI output formats:

- [ ] JUnit
- [ ] annotations
- [ ] SARIF where appropriate

## Impact Review

Earlier: consume CLI/SDK results and add agreed automation output formats as required. Later: supply deployment validation examples for interoperability and hardening; CI integration alone does not require new health states.

---

# Phase 22 — Observability Export

**Status:** Planned — not yet implemented.

Goal:

Allow existing observability systems to consume AgentHealth results.

Potential integrations:

- [ ] OpenTelemetry
- [ ] Prometheus
- [ ] structured logs
- [ ] webhooks

- [ ] export distinct agent, gateway/router, path, and dependency evidence with sensitive topology redaction

AgentHealth will not attempt to become a complete observability backend.

Instead:

```text
AgentHealth
    │
    ├── OpenTelemetry
    ├── Prometheus
    ├── Logs
    └── Webhooks
```

- [ ] Export shared node evidence once per ID while retaining distinct edge policies and redacting sensitive topology. See the [graph contract](docs/rfcs/phase-10-dependency-graph.md).

## Impact Review

Earlier: map approved result evidence to exports and extend output interfaces as required; no new health semantics are expected solely for export. Later: validate export compatibility and redaction during conformance and hardening.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 23 — Plugin / Adapter Ecosystem

**Status:** Planned — not yet implemented.

Goal:

Allow the community to extend AgentHealth without modifying the core.

## Adapter Contract

Define:

- [ ] metadata
- [ ] configuration
- [ ] health checks
- [ ] capabilities
- [ ] result normalization
- [ ] error normalization
- [ ] security requirements
- [ ] compatibility version
- [ ] adapter evidence mapping for direct/composite agents, gateways, routers, and communication paths under the updated AHS
- [ ] adapter extension examples for the direct (first) agent and composite (downstream) agents using direct communication or A2A
- [ ] Agentgateway and Agent Router integration examples with gateway/router, backend, and communication-path evidence kept distinct

Potential architecture:

```text
AgentHealth Core
      │
      └── Adapter API
             ├── Agent checks (direct / composite roles)
             ├── HTTP / MCP / A2A
             ├── Agentgateway integrations
             ├── Agent Router integrations
             ├── Model / Tool / Data adapters
             └── Community adapters
```

Direct/composite describe agent roles; the diagram does not introduce new
target types or require a separate adapter for each role. Integration mechanisms
follow the approved contracts and supported interfaces from Phases 7–10.

## Impact Review

Earlier: formalize the existing adapter baseline, including compatibility and evidence mapping; update core/CLI registration if the extension mechanism requires it. Later: supply the contract tested by conformance and ecosystem validation.

---

# Phase 24 — AgentHealth Conformance

**Status:** Planned — not yet implemented.

Goal:

Allow implementations and integrations to verify compliance with the Agent Health Specification.

Potential capabilities:

```bash
agenthealth conformance test <target>
```

## Conformance Areas

- [ ] Result schema
- [ ] Status semantics
- [ ] Required fields
- [ ] Error behavior
- [ ] Dependency representation
- [ ] Security requirements
- [ ] Adapter behavior
- [ ] AHS updates from Phases 7–10, including agent roles, gateway/router paths, aggregation, and compatible result representation
- [ ] direct (first) agent health and composite (downstream) agent health with distinct evidence
- [ ] first-to-downstream communication via direct integrations or A2A
- [ ] Agentgateway and Agent Router health separate from backend and route/path health
- [ ] failure propagation, safe functional opt-in, credential redaction, and topology redaction across these integrations

Potential future levels:

```text
AgentHealth Core Compatible
AgentHealth Protocol Compatible
AgentHealth Full Compatible
```

Any certification/trademark program would require separate governance and community approval.

- [ ] Verify the Phase 7 document contract, names-only discovery, safe task bounds and actual first-runtime peer communication. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

- [ ] Validate Phase 10 shared execution, cycles, missing references, edge policies, independent node budgets and labeled tree projection. See the [graph contract](docs/rfcs/phase-10-dependency-graph.md).

## Impact Review

Earlier: validate the approved AHS, AHP, adapters, outputs, and integrations; feed defects back to their owning layers. Later: provide conformance checks for interoperability and production hardening.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 25 — Ecosystem Interoperability Testing

**Status:** Planned — not yet implemented.

Goal:

Test AgentHealth across heterogeneous real-world agent stacks.

Example matrix:

```text
Agent Framework A
      │
      ├── MCP Server X
      ├── Model Provider Y
      └── Vector DB Z

Agent Framework B
      │
      ├── MCP Server Q
      ├── Model Provider R
      └── Database S
```

AgentHealth should produce consistent health semantics regardless of implementation.

The test matrix must also include a first agent communicating with downstream
agents directly, through A2A, and through Agentgateway or an Agent Router.
Include healthy backends behind a failing path, failing backends behind a
healthy gateway/router, shared backends, and multiple configured routes.

## Integration Scenarios

- [ ] direct (first) agent checks with and without composite (downstream) agents
- [ ] first agent communicating with downstream agents directly or through A2A across supported frameworks
- [ ] Agentgateway and Agent Router checked independently and in the first-to-downstream communication path
- [ ] healthy downstream agent behind a failing gateway/router or route
- [ ] failing downstream agent behind a healthy gateway/router
- [ ] shared downstream agents and backends reached through multiple configured routes
- [ ] consistent results across CLI, SDKs, and AHP where supported

- [ ] Exercise Phase 7 handlers that actually contact peers directly and via A2A, including a healthy peer behind a failed communication path. See the [Phase 7 proposal](docs/rfcs/phase-7-agent-health.md).

## Impact Review

Earlier: test supported cross-phase combinations and feed contract or implementation defects back to their owning layers. Later: supply failure scenarios and evidence for production hardening.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# Phase 26 — Production Hardening

**Status:** Planned — not yet implemented.

Before a stable 1.0 release:

- [ ] Security audit
- [ ] Threat model
- [ ] Performance benchmarks
- [ ] Large dependency graph testing
- [ ] Retry behavior testing
- [ ] Failure injection testing
- [ ] first-agent, downstream-agent, and gateway/router path failure isolation
- [ ] Cross-platform testing
- [ ] Backward compatibility policy
- [ ] Stable configuration specification
- [ ] Stable result schema
- [ ] Stable adapter interface
- [ ] Stable CLI
- [ ] Documentation review

- [ ] Exercise Phase 10 DAG projection limits, shared active-check deduplication, cancellation and per-run concurrency retention for uncooperative adapters. See the [graph contract](docs/rfcs/phase-10-dependency-graph.md).

## Impact Review

Earlier: audit and fix supported contracts, implementations, distribution, and integrations; track compatibility for resulting changes. Later: gate 1.0 and inform post-1.0 work; this phase adds no new feature contract by itself.


Phase 11 integration requirements are recorded in the [AHP impact review](docs/rfcs/phase-11-ahp.md#later-phases).

---

# AgentHealth 1.0

The 1.0 milestone should represent a stable operational contract rather than simply a feature count.

Minimum expectations:

```text
✓ Stable Agent Health Specification
✓ Stable health states
✓ Stable result schema
✓ Stable CLI
✓ HTTP support
✓ MCP support
✓ A2A support
✓ Direct and composite agent health
✓ Agentgateway integration
✓ Agent Router integration
✓ Dependency graph
✓ Agent Health Protocol (AHP) first iteration
✓ Docker
✓ Standalone binaries
✓ Kubernetes integration
✓ Python SDK
✓ JavaScript/TypeScript SDK
✓ Adapter framework
✓ Security documentation
✓ Conformance tests
```

---

# Post-1.0 Exploration

Potential areas include:

## Health Policies

Declarative organizational health requirements.

```yaml
policy:
  require:
    - model
    - mcp

  max_latency_ms: 1000

  minimum_status: HEALTHY
```

## Health Profiles

Examples:

```text
development
ci
production
kubernetes-readiness
deep-diagnostics
```

## Fleet Health

Check multiple agents:

```bash
agenthealth fleet check agents.yaml
```

## Health Graph API

Expose dependency health programmatically.

## Historical Health

Optional integration with external time-series/observability systems.

AgentHealth itself should avoid becoming a monitoring database.

## Remote Health Service

Service mode (`agenthealth serve`) is now tracked under [Phase 11 — Agent Health Protocol (AHP)](#phase-11--agent-health-protocol-ahp) rather than purely as post-1.0 exploration, since it is the natural reference implementation of AHP. Further exploration here is limited to advanced serving behaviors beyond the core AHP contract, such as fleet-wide serving or multi-tenant service mode.

## Policy Engines

Explore integrations with external policy systems rather than embedding complex governance logic into the core.

---

# Long-Term Ecosystem Vision

AgentHealth should evolve from:

```text
CLI health checker
```

into:

```text
                     TheAgentHealth
                           │
              Agent Health Specification
                         (AHS)
                           │
                ┌──────────┴──────────┐
                │                     │
                ▼                     ▼
       Agent Health Protocol    Reference Engine
              (AHP)                   │
                │              ┌──────┼───────┐
                │              ▼      ▼       ▼
                │             CLI    SDKs    K8s
                │                     │
                └──────────┬──────────┘
                           │
                        Adapters
                           │
                           ├── Direct (first) agent
                           ├── Composite (downstream) agents
                           │     └── Direct communication / A2A
                           ├── Agentgateway
                           ├── Agent Router
                           └── HTTP / MCP / Models / Tools / Data
                                      │
                                      ▼
                             Agentic Ecosystem
```

The common health contract should cover the user's first agent, its downstream
agents, and communication between them, including paths through Agentgateway
or an Agent Router. Agent health, gateway/router health, path health, and
supporting dependency health remain distinguishable. Direct/composite are
agent roles rather than new target types. This diagram describes the planned
ecosystem; implementation availability follows the phase statuses above.

The goal is not to own every integration.

The goal is to provide the **common health contract** connecting them.

---

# Community Priorities

Community feedback should influence prioritization.

Particularly valuable contributions include:

- new protocol adapters,
- MCP interoperability testing,
- A2A interoperability testing,
- Kubernetes integrations,
- model adapters,
- database adapters,
- vector store adapters,
- SDKs,
- specification feedback,
- security reviews,
- production use cases.

---

# Project Success

AgentHealth succeeds if developers can eventually treat this:

```bash
agenthealth ping agent <endpoint>
```

as naturally as they currently treat:

```bash
ping <host>
curl <endpoint>
```

and if infrastructure platforms can consume a common health contract regardless of:

```text
agent framework
model provider
cloud
programming language
protocol
database
vector store
runtime
```

That is the long-term objective:

> **A universal health and readiness layer for the Agentic AI ecosystem.**

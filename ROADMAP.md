# AgentHealth Roadmap

This roadmap describes the proposed development path for AgentHealth.

AgentHealth is intended to become a universal, vendor-neutral health, readiness, reachability, and diagnostics layer for AI agents and agentic infrastructure.

The roadmap is intentionally capability-driven rather than tied to fixed calendar dates.

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

## Architecture Decisions

Define:

- [x] Repository strategy: **single repository (monorepo)**. AHS and AHP specification docs, the core engine, CLI, adapters, SDKs (Python/PyPI, JavaScript/npm), and distribution integrations (Docker, Kubernetes, CI) all live in this one repository rather than being split across separate repos. See [Repository Structure](README.md#repository-structure).
- [x] Core implementation language: **Go** (see [docs/architecture.md](docs/architecture.md#core-implementation-language))
- [x] CLI architecture: single binary, subcommands (see [docs/architecture.md](docs/architecture.md#cli-architecture))
- [x] Adapter interface: in-process/compiled-in initially (see [docs/architecture.md](docs/architecture.md#adapter-interface))
- [x] configuration format: YAML (see [spec/configuration.md](spec/configuration.md))
- [x] result format: JSON (see [spec/result-schema.md](spec/result-schema.md))
- [x] error model: normalized onto health states (see [docs/architecture.md](docs/architecture.md#error-model))
- [x] exit codes: conceptual mapping defined, final value TBD in spec process (see [docs/architecture.md](docs/architecture.md#exit-codes))
- [x] plugin strategy: in-repo adapters initially (see [docs/architecture.md](docs/architecture.md#plugin-strategy))
- [x] SDK strategy: thin wrappers over the core engine (see [docs/architecture.md](docs/architecture.md#sdk-strategy))

---

# Phase 1 — Agent Health Specification

**Status:** Partially complete (draft) — vocabulary, target model, result schema, error classification, and status aggregation semantics are drafted in [spec/](spec/README.md). Authentication references, timeout policies, retry policies, and active/passive opt-in syntax remain open and are deferred to Phase 2 (see [spec/configuration.md § Still to be defined](spec/configuration.md#still-to-be-defined)).

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

```json
{
  "spec_version": "v1",
  "target": {
    "name": "example",
    "type": "mcp"
  },
  "status": "HEALTHY",
  "latency_ms": 82,
  "checks": {}
}
```

## Configuration Specification

Defined in [spec/configuration.md](spec/configuration.md):

- [x] target definitions
- [x] checks
- [x] thresholds
- [ ] authentication references
- [x] dependencies
- [x] critical vs optional dependencies
- [ ] timeout policies
- [ ] retry policies
- [ ] active/passive check behavior

Remaining items are explicitly listed as open in [spec/configuration.md § Still to be defined](spec/configuration.md#still-to-be-defined) and will be resolved alongside the Phase 2 core engine implementation.

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

Goal:

Implement the reference health engine.

## Core Capabilities

- [ ] Target loading
- [ ] Configuration validation
- [ ] Check execution
- [ ] Timeouts
- [ ] Retries
- [ ] Latency measurement
- [ ] Result normalization
- [ ] Error normalization
- [ ] Health aggregation
- [ ] Dependency evaluation
- [ ] Human-readable output
- [ ] JSON output

## Safety

- [ ] Secret redaction
- [ ] Safe logging
- [ ] TLS verification
- [ ] Non-destructive defaults
- [ ] Active-check controls
- [ ] Timeout protection

---

# Phase 3 — Universal CLI

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
agenthealth ping a2a https://agent.example.com
```

### Check

```bash
agenthealth check agenthealth.yaml
```

### Doctor

```bash
agenthealth doctor agenthealth.yaml
```

### Version

```bash
agenthealth version
```

## Output Formats

- [ ] terminal
- [ ] JSON
- [ ] YAML

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

Goal:

Establish a universal baseline adapter.

## Checks

- [ ] DNS resolution
- [ ] TCP connectivity
- [ ] TLS validation
- [ ] HTTP connectivity
- [ ] expected HTTP status
- [ ] headers
- [ ] authentication
- [ ] response matching
- [ ] latency
- [ ] timeout

Example:

```bash
agenthealth ping http https://api.example.com/health
```

This adapter provides a foundation for services that do not implement agent-specific protocols.

---

# Phase 5 — MCP Health Adapter

Goal:

Make MCP a first-class AgentHealth target.

## Capabilities

- [ ] MCP transport connectivity
- [ ] initialization
- [ ] protocol negotiation
- [ ] protocol version validation
- [ ] capability discovery
- [ ] authentication
- [ ] tools discovery
- [ ] resources discovery
- [ ] prompts discovery
- [ ] required tool validation
- [ ] required resource validation
- [ ] latency measurement
- [ ] safe functional invocation

Example:

```bash
agenthealth ping mcp http://localhost:3000
```

## MCP Doctor

```bash
agenthealth doctor mcp http://localhost:3000
```

Potential diagnostics:

```text
✓ Transport
✓ Initialization
✓ Protocol
✓ Authentication
✓ tools/list
✓ resources/list
⚠ Expected tool "search" missing

Status: DEGRADED
```

---

# Phase 6 — A2A Health Adapter

Goal:

Support Agent-to-Agent health validation.

## Capabilities

- [ ] endpoint discovery
- [ ] agent metadata
- [ ] protocol validation
- [ ] authentication
- [ ] capability discovery
- [ ] expected skill validation
- [ ] minimal interaction
- [ ] latency
- [ ] error classification

Example:

```bash
agenthealth ping a2a https://agent.example.com
```

---

# Phase 7 — Agent Health

Goal:

Provide composite health checks for complete agents.

An agent health check may combine:

```text
Agent endpoint
      │
      ├── A2A
      ├── Model
      ├── MCP
      ├── Tools
      └── Data
```

## Capabilities

- [ ] agent endpoint health
- [ ] metadata validation
- [ ] capability validation
- [ ] dependency discovery
- [ ] dependency execution
- [ ] minimal agent task
- [ ] dependency aggregation
- [ ] critical dependency propagation

---

# Phase 8 — Dependency Graph

Goal:

Make dependency-aware diagnostics a defining AgentHealth capability.

## Features

- [ ] Dependency graph construction
- [ ] Nested dependencies
- [ ] Critical dependencies
- [ ] Optional dependencies
- [ ] Failure propagation
- [ ] Cycle detection
- [ ] Parallel health execution
- [ ] Configurable concurrency
- [ ] Dependency timeout budgets
- [ ] Tree output

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

---

# Phase 9 — Agent Health Protocol (AHP)

**Status:** Proposed / experimental

Goal:

Define a vendor-neutral protocol through which agentic systems can expose and exchange standardized health and readiness information.

The Agent Health Protocol implements the semantics defined by the Agent Health Specification (Phase 1).

This phase intentionally comes **after** the HTTP, MCP, A2A, Agent Health, and Dependency Graph phases rather than immediately following the specification. The health semantics need to be proven across those adapters in reality first; AHP then generalizes the lessons learned into an interoperable wire contract. Standardizing an exchange protocol before understanding the operational requirements would risk locking in the wrong contract.

## Core Operations

Potential protocol operations:

```text
GET /health
GET /ready
GET /live
GET /health/dependencies
GET /health/capabilities
```

HTTP is a likely first binding, not necessarily the only one. Exact transport bindings and endpoint conventions — including whether AHP maps onto other transports such as MCP or A2A mechanisms — will be defined through the specification process rather than locked in now.

## Protocol Model

Define:

- [ ] protocol versioning
- [ ] health response envelope
- [ ] readiness semantics
- [ ] liveness semantics
- [ ] dependency representation
- [ ] capability health
- [ ] error representation
- [ ] authentication
- [ ] authorization
- [ ] content types
- [ ] caching behavior
- [ ] timeout semantics
- [ ] extension mechanism

## Service Mode

A reference "server mode" lets AgentHealth expose AHP instead of only consuming other systems' health:

```bash
agenthealth serve
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

This generalizes the `agenthealth serve` concept previously scoped only to Post-1.0 Exploration. Once AHP's wire contract stabilizes, service mode should move into the pre-1.0 roadmap rather than remain purely exploratory — it is what makes AgentHealth infrastructure that other systems can query, not just a CLI that queries others.

## Discovery

Explore a standard mechanism through which systems can advertise Agent Health Protocol support.

## Security

Define:

- [ ] information disclosure requirements
- [ ] authentication mechanisms
- [ ] dependency redaction
- [ ] sensitive metadata handling
- [ ] public vs authenticated health information
- [ ] rate limiting recommendations

## Conformance

AHP implementations should be testable independently of the AgentHealth reference implementation.

---

# Phase 10 — Docker Distribution

Goal:

Allow AgentHealth to run without installing a programming language runtime.

## Deliverables

- [ ] Official container image
- [ ] Multi-stage build
- [ ] Minimal runtime image
- [ ] Linux AMD64
- [ ] Linux ARM64
- [ ] Signed images
- [ ] SBOM
- [ ] Versioned tags

Example:

```bash
docker run --rm agenthealth/agenthealth \
  ping mcp http://host.docker.internal:3000
```

---

# Phase 11 — Standalone Binaries

Goal:

Make installation trivial.

## Platforms

- [ ] Linux AMD64
- [ ] Linux ARM64
- [ ] macOS ARM64
- [ ] macOS AMD64
- [ ] Windows AMD64

## Distribution

Potential channels:

- [ ] GitHub Releases
- [ ] Homebrew
- [ ] Linux packages
- [ ] Windows package manager
- [ ] container registries

---

# Phase 12 — Kubernetes Integration

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

---

# Phase 13 — Python SDK / PyPI

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
- [ ] Exceptions/error model

The Python SDK must follow the Agent Health Specification rather than defining it.

---

# Phase 14 — JavaScript / TypeScript SDK

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
- [ ] browser/server scope definition

---

# Phase 15 — Model / LLM Adapters

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

---

# Phase 16 — Database Adapters

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

---

# Phase 17 — Vector Store Adapters

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

---

# Phase 18 — Agent Runtime / Framework Integrations

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

AgentHealth should remain framework-neutral.

---

# Phase 19 — CI/CD Integrations

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

Potential CI output formats:

- [ ] JUnit
- [ ] annotations
- [ ] SARIF where appropriate

---

# Phase 20 — Observability Export

Goal:

Allow existing observability systems to consume AgentHealth results.

Potential integrations:

- [ ] OpenTelemetry
- [ ] Prometheus
- [ ] structured logs
- [ ] webhooks

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

---

# Phase 21 — Plugin / Adapter Ecosystem

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

Potential architecture:

```text
AgentHealth Core
      │
      └── Adapter API
             │
       ┌─────┼─────┐
       │     │     │
      MCP   A2A   Community
                  Adapters
```

---

# Phase 22 — AgentHealth Conformance

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

Potential future levels:

```text
AgentHealth Core Compatible
AgentHealth Protocol Compatible
AgentHealth Full Compatible
```

Any certification/trademark program would require separate governance and community approval.

---

# Phase 23 — Ecosystem Interoperability Testing

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

---

# Phase 24 — Production Hardening

Before a stable 1.0 release:

- [ ] Security audit
- [ ] Threat model
- [ ] Performance benchmarks
- [ ] Large dependency graph testing
- [ ] Retry behavior testing
- [ ] Failure injection testing
- [ ] Cross-platform testing
- [ ] Backward compatibility policy
- [ ] Stable configuration specification
- [ ] Stable result schema
- [ ] Stable adapter interface
- [ ] Stable CLI
- [ ] Documentation review

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
✓ Agent health
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

Service mode (`agenthealth serve`) is now tracked under [Phase 9 — Agent Health Protocol (AHP)](#phase-9--agent-health-protocol-ahp) rather than purely as post-1.0 exploration, since it is the natural reference implementation of AHP. Further exploration here is limited to advanced serving behaviors beyond the core AHP contract, such as fleet-wide serving or multi-tenant service mode.

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
          ┌────────┬───────┼──────┬────────┐
          ▼        ▼       ▼      ▼        ▼
         MCP      A2A    Models  Tools    Data
                           │
                           ▼
                  Agentic Ecosystem
```

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
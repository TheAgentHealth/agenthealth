# TheAgentHealth

> Universal Health & Readiness Framework for Agentic Systems

**AgentHealth** is an open-source, vendor-neutral, framework-neutral, and language-neutral project for determining whether AI agents and the infrastructure they depend on are **reachable, correctly configured, responsive, ready, and operationally healthy**.

Think of AgentHealth as:

> **`ping` + `curl` + `doctor` + Kubernetes health probes for the Agentic AI ecosystem.**

AgentHealth is designed to work across:

- AI agents
- Multi-agent systems
- Agent-to-Agent (A2A) endpoints
- MCP servers
- MCP tools and resources
- LLM and model endpoints
- External tools
- APIs
- Vector databases
- Databases
- Agent runtimes
- Agent gateways
- Agent routers
- Supporting infrastructure

The project provides a common health model, a Go engine, a CLI, machine-readable health results, and HTTP/API, MCP, and A2A adapters. SDKs, container distribution, Kubernetes integrations, and additional adapters remain roadmap work.

---

# Why AgentHealth?

Traditional health checks answer questions such as:

```text
Is the process running?
Is port 8080 open?
Does /health return HTTP 200?
```

Those checks are necessary, but they are often insufficient for agentic AI systems.

An agent can return HTTP 200 while being unable to perform useful work because:

- its model endpoint is unavailable,
- an MCP server cannot initialize,
- authentication has expired,
- a required tool is unavailable,
- a vector database cannot be reached,
- an A2A peer is unreachable,
- required capabilities are missing,
- the agent runtime is misconfigured,
- dependencies are experiencing excessive latency,
- or a minimal agent execution fails.

AgentHealth is intended to answer the more useful operational question:

> **Can this agentic system actually perform the work it is expected to perform right now?**

---

# Project Goals

AgentHealth aims to provide:

1. A **vendor-neutral health model** for agentic systems.
2. A **language-neutral Agent Health Specification**.
3. An experimental **Agent Health Protocol (AHP)** proposal for standardizing how health is exposed and exchanged over the wire (future).
4. A universal command-line interface.
5. Health checks for common agent protocols and infrastructure.
6. Dependency-aware health diagnostics.
7. Machine-readable results for automation.
8. SDKs for application integration.
9. Docker/container distribution.
10. Kubernetes-native health and readiness integration.
11. CI/CD deployment validation.
12. An extensible adapter/plugin model.
13. A community-driven interoperability layer for agentic infrastructure.

---

# Specification → Protocol → Implementation

AgentHealth separates **what health means** from **how health is exposed** from **the tools intended to implement it** (see [Project Status](#project-status) for implemented and planned components).

```text
                    TheAgentHealth
                          │
          ┌───────────────┼────────────────┐
          │               │                │
          ▼               ▼                ▼
 Agent Health       Agent Health       Reference
 Specification      Protocol (AHP)    Implementation
 (AHS)              [proposed]
          │               │                │
          │               │         ┌──────┼────────┐
          │               │         │      │        │
          │               │        CLI    SDK     Docker
          │               │                       / Helm
          │               │
          ▼               ▼
 Defines WHAT       Defines HOW
 health means       health is exposed
                    and exchanged
```

## Agent Health Specification (AHS)

The specification defines the vocabulary and health model that the rest of the project builds on:

```text
HEALTHY
DEGRADED
UNHEALTHY
UNREACHABLE
MISCONFIGURED
UNKNOWN

Reachability
Protocol Health
Authentication
Capability Health
Functional Health
Dependency Health
Latency Health
Configuration Health
```

This is documented throughout this README (see [Core Concepts](#core-concepts) and [Standard Health States](#standard-health-states)) and is the most mature layer of the project today.

## Agent Health Protocol (AHP) — proposed / experimental

> **Status: experimental / future.** AHP is a direction, not a shipped standard. Running health checks with the CLI does not, by itself, constitute a protocol — a protocol requires a defined, versioned, interoperable wire contract that multiple independent implementations can agree on.

AHP is the proposed answer to a narrower question than the specification: once you know what health means, how should it be **exposed and exchanged** over the wire so that any client, gateway, or orchestrator can query it consistently, regardless of language or framework?

A future AHP could standardize endpoints such as:

```text
GET /health
GET /ready
GET /live
GET /dependencies
GET /capabilities
```

and a common response envelope such as:

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

The [Machine-Readable Health Format](#machine-readable-health-format) that AgentHealth's CLI produces (see [Phase 3 — Universal CLI](ROADMAP.md#phase-3--universal-cli)) is an early, implementation-specific version of this idea. Formalizing it into AHP, with a versioned schema, transport requirements, and conformance rules independent of this repository's CLI, is future work.

## Reference Implementation

The reference implementation is the concrete, versioned software that implements the specification (and, eventually, the protocol):

- the `agenthealth` CLI,
- language SDKs (Python, JavaScript/TypeScript),
- Docker images and Helm charts,
- Kubernetes integrations.

These are described in detail in [Universal CLI](#universal-cli), [Distribution](#distribution), and [Kubernetes](#kubernetes).

## Naming

| Concept | Name |
|---|---|
| Organization | TheAgentHealth |
| Project | AgentHealth |
| Specification | Agent Health Specification (AHS) |
| Protocol (proposed) | Agent Health Protocol (AHP) |
| CLI | `agenthealth` |
| Config file | `agenthealth.yaml` |
| Python package | `agenthealth` |
| npm package | `@agenthealth/sdk` |
| Docker image | `agenthealth/agenthealth` |
| Helm chart | `agenthealth` |

---

# What AgentHealth Is Not

AgentHealth is **not intended to replace**:

- observability platforms,
- distributed tracing systems,
- metrics platforms,
- logging systems,
- evaluation frameworks,
- security scanners,
- model monitoring platforms,
- API gateways,
- MCP gateways,
- or agent frameworks.

Those systems provide continuous visibility into application behavior.

AgentHealth focuses on a narrower operational problem:

> **Health, readiness, reachability, dependency validation, and diagnostics.**

An observability platform may tell you how a system behaved over the last 24 hours.

AgentHealth should be able to tell you:

```text
Can my agent work right now?

If not, what dependency is preventing it?
```

---

# Core Concepts

AgentHealth separates health into multiple dimensions instead of treating health as a single HTTP status.

## Reachability

Can AgentHealth establish communication with the target?

Examples:

- TCP/HTTP connectivity
- MCP transport connectivity
- A2A endpoint availability
- model API availability
- database connectivity

---

## Protocol Health

Does the target correctly support the expected protocol?

Examples:

- MCP initialization
- protocol version compatibility
- A2A discovery
- required endpoint validation
- capability negotiation

---

## Authentication

Can the configured identity authenticate successfully?

Examples:

- API key validation
- OAuth credentials
- service credentials
- workload identity
- token validity

Secrets must never be exposed in AgentHealth output.

---

## Capability Health

Does the target expose the capabilities required by the workload?

Examples:

- MCP tools
- MCP resources
- agent skills
- model availability
- expected API operations
- required runtime features

---

## Functional Health

Can the target successfully perform a minimal operation that actually exercises it (not merely list what it claims to support)?

Examples:

- minimal model inference request
- lightweight API operation
- database `SELECT 1`
- vector store read operation

Functional checks are classified **active** (see [Passive vs Active Checks](#passive-vs-active-checks)): they never run by default and require being explicitly listed in a target's `checks` (see [spec/configuration.md § Default Checks](spec/configuration.md#default-checks)). Even when opted in, functional checks should remain minimal and non-destructive.

---

## Dependency Health

Are the target's required dependencies healthy?

An agent may itself be reachable while one of its dependencies is not.

AgentHealth therefore supports dependency-aware health.

Example:

```text
Research Agent
│
├── Model Endpoint .............. HEALTHY
│
├── GitHub MCP .................. HEALTHY
│
├── PostgreSQL MCP .............. DEGRADED
│
├── Search Tool ................. HEALTHY
│
└── Vector Database ............. UNREACHABLE

Overall: DEGRADED
```

---

## Latency Health

AgentHealth can measure response latency and compare it against configurable thresholds.

Example:

```text
Latency: 1,842 ms
Threshold: 1,000 ms

Status: DEGRADED
```

---

## Configuration Health

AgentHealth can detect configuration problems that are identifiable *before* attempting to reach or operate the target, such as:

- missing endpoint,
- missing credentials,
- invalid or unsupported protocol selection,
- malformed configuration.

These are reported as:

```text
MISCONFIGURED
```

rather than incorrectly labeling the target as unreachable.

Problems that can only be discovered by actually contacting the target — such as an incompatible protocol version returned during negotiation, or a required capability/model the target doesn't expose — are Protocol Health and Capability Health concerns instead, and resolve to `UNHEALTHY` (see [spec/health-model.md § Error Classification](spec/health-model.md#error-classification)).

---

# Standard Health States

AgentHealth defines a common health vocabulary.

| Status | Meaning |
|---|---|
| `HEALTHY` | Target is operating normally |
| `DEGRADED` | Target works but one or more non-fatal conditions are impaired |
| `UNHEALTHY` | Target is reachable but cannot perform required functionality |
| `UNREACHABLE` | Target cannot be contacted |
| `MISCONFIGURED` | Configuration prevents a valid health determination or operation |
| `UNKNOWN` | Health cannot currently be determined |

Adapters may expose additional diagnostic details while mapping their overall result to the common health model.

---

# Universal CLI

The CLI supports `http`, `api`, and `mcp` targets. See [CLI usage](docs/cli.md) for the complete command reference.

```bash
agenthealth ping http https://service.example.com/health
agenthealth ping mcp http://localhost:3000/mcp
agenthealth check agenthealth.yaml --format json
agenthealth doctor agenthealth.yaml
agenthealth doctor mcp http://localhost:3000/mcp
agenthealth ping a2a http://localhost:9000
agenthealth doctor examples/a2a-check/agenthealth.yaml
agenthealth version
```

Configured dependencies are evaluated automatically; no `--dependencies` flag
is needed. Agent and model adapters remain planned.

---

# `ping`, `check`, and `doctor`

AgentHealth exposes three primary operational concepts.

## `agenthealth ping`

Quickly determine whether a target is reachable and responding correctly.

```bash
agenthealth ping mcp http://localhost:3000
```

Example:

Conceptual display; see [CLI usage](docs/cli.md) for the implemented output and health states.

```text
Target: github-mcp
Type: MCP

Reachability       PASS
Protocol           PASS
Authentication     PASS
Capabilities       PASS
Latency            84 ms

Status: HEALTHY
```

---

## `agenthealth check`

Run standardized health policies.

```bash
agenthealth check agenthealth.yaml
```

This is designed for:

- automation,
- CI/CD,
- production validation,
- Kubernetes,
- scheduled checks,
- platform engineering workflows.

---

## `agenthealth doctor`

Perform deeper diagnostics and explain why a system is unhealthy.

```bash
agenthealth doctor agenthealth.yaml
```

Conceptual future output (current doctor output is documented in [CLI usage](docs/cli.md)):

```text
AgentHealth Doctor

✓ Agent endpoint reachable
✓ A2A discovery successful
✓ Model endpoint reachable
✓ Authentication valid
✓ GitHub MCP healthy
⚠ PostgreSQL MCP latency: 1,832 ms
✗ Search MCP unreachable
✓ Vector store healthy

2 issues detected.

Overall Status: DEGRADED
```

---

# Supported Target Types

The long-term AgentHealth architecture is intended to support:

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

Currently implemented target types are `http`, `api`, `mcp`, and `a2a` (A2A 0.3.0 JSON-RPC). The remaining types are specification vocabulary for future adapters.

See [ROADMAP.md](ROADMAP.md).

---

# MCP Health

The [MCP adapter](docs/mcp-adapter.md) supports Streamable HTTP JSON/SSE,
configured stdio subprocesses, legacy initialization, and stateless MCP
`2026-07-28`. It checks authentication, paginated tools/resources/prompts,
required inventories, and latency. Functional tool calls require explicit
safe opt-in and never retry.

| Feature | Configuration example | Documentation |
|---|---|---|
| Passive discovery and required inventories | [MCP discovery](examples/mcp-check/agenthealth.yaml) | [Checks and expectations](docs/mcp-adapter.md#passive-checks-and-expectations) |
| Modern stateless protocol | [Version pin](examples/mcp-check/modern.yaml) | [Protocol selection](docs/mcp-adapter.md#protocol-selection) |
| Local stdio server | [Subprocess configuration](examples/mcp-stdio/agenthealth.yaml) | [Stdio setup](docs/mcp-adapter.md#stdio-subprocesses) |
| OAuth client credentials | [Automated authentication](examples/mcp-oauth/client-credentials.yaml) | [OAuth setup](docs/mcp-adapter.md#oauth-acquisition-and-login) |
| OAuth browser login + PKCE | [Interactive login configuration](examples/mcp-oauth/agenthealth.yaml) | [Login command](docs/cli.md#mcp-oauth-login) |
| OAuth refresh tokens and rotation | [Refresh configuration](examples/mcp-oauth/refresh-token.yaml) | [Token refresh and storage](docs/mcp-adapter.md#oauth-acquisition-and-login) |
| Explicit safe functional invocation | [Tool probe](examples/mcp-check/functional.yaml) | [Functional safety](docs/mcp-adapter.md#functional-invocation) |

Point these commands at your server's actual MCP endpoint:

```bash
agenthealth ping mcp http://localhost:3000/mcp
agenthealth doctor mcp http://localhost:3000/mcp
```

For OAuth browser login, first replace the example's endpoint, issuer, and
client ID with your registered server/client settings, then run:

```bash
agenthealth login examples/mcp-oauth/agenthealth.yaml user-mcp
agenthealth check examples/mcp-oauth/agenthealth.yaml
```

Health checks stay noninteractive; only `login` asks for browser approval.
Tokens are stored privately and redacted from health output. Client credentials
and refresh-token examples reference environment variables instead of literal
secrets. See the [examples guide](examples/README.md) for setup instructions and
[Phase 5 status](ROADMAP.md#phase-5--mcp-health-adapter) for implemented scope.

Deprecated separate HTTP+SSE, dynamic OAuth client registration, and JWT client
assertions are not implemented. The adapter documentation lists remaining
protocol, transport, and platform limits.

---

# A2A Health

The [A2A adapter](docs/a2a-adapter.md) validates A2A 0.3.0 JSON-RPC peers:
agent card discovery and metadata, passive protocol/authentication checks,
expected skills and capabilities, latency, and explicit safe interaction probes.

```bash
agenthealth ping a2a https://research-agent.example.com
agenthealth check examples/a2a-check/agenthealth.yaml --format json
```

| Example | Behavior |
|---|---|
| [Passive peer](examples/a2a-check/agenthealth.yaml) | Default discovery and read-only protocol/authentication checks, required skill, latency |
| [Bearer credentials](examples/a2a-check/bearer.yaml) | Environment-referenced token for passive checks |
| [Custom card URL](examples/a2a-check/custom-card.yaml) | Discovery from a different path on the configured origin |
| [Minimal interaction](examples/a2a-check/functional.yaml) | One explicitly enabled, operator-declared safe text probe |

Examples require your own running peer and matching skill expectations. See the
[example setup guide](examples/a2a-check/README.md) and
[adapter guide](docs/a2a-adapter.md) for configuration, classifications,
and supported scope. Other A2A versions and transports remain future work.

---

# Agent Health

**Planned in [Phase 7 — Agent Health](ROADMAP.md#phase-7--agent-health).**
This includes the user's first, user-facing agent, whether it answers directly
or calls other agents. Checks will cover its own endpoint, authentication,
liveness, readiness, and capabilities, with an explicitly opted-in, bounded,
non-destructive minimal task for functional validation.

**Direct agent health** checks the user's first, user-facing agent.

**Composite agent health** checks the downstream agents that the first agent
communicates with, either through a direct integration or through A2A.
Each downstream agent's own health and the communication path from the first
agent should be checked separately.

Models, tools, MCP servers, data services, gateways, and routers are supporting
dependencies; their checks are described as dependency health. Here, direct
and composite describe agent roles in the interaction, not new target types.
A `multi-agent` target describes the cooperating system as a whole. Results
should distinguish first-agent, downstream-agent, communication-path, and
supporting-dependency failures.

Conceptual command (the `agent` adapter is not yet implemented):

```bash
agenthealth ping agent https://research-agent.example.com
```

Potential checks:

```text
Endpoint
Agent metadata
Agent capabilities
A2A
Model dependency
MCP dependencies
Tool dependencies
Data dependencies
Minimal execution
Latency
```

Agentgateway and Agent Router health are planned separately in
[Phase 8 — Agentgateway Integration](ROADMAP.md#phase-8--agentgateway-integration)
and [Phase 9 — Agent Router Integration](ROADMAP.md#phase-9--agent-router-integration).
They may be checked independently or as agent dependencies, covering exposed
health/readiness signals, authentication, configured backend or route
availability, and optional safe functional probes. Integration results will
distinguish gateway/router failures from downstream failures; specific
products, versions, and transports will be documented during design.

---

# Model / LLM Health

AgentHealth can validate model dependencies without being tied to one provider.

Checks may include:

- endpoint reachability,
- authentication,
- model availability,
- minimal inference,
- latency,
- rate limiting,
- provider errors,
- malformed responses.

Provider-specific integrations should be implemented through adapters.

---

# Tool and API Health

Generic HTTP/API checks enable AgentHealth to validate dependencies that do not implement an agent-specific protocol.

Checks can include:

```text
DNS
TCP
TLS
HTTP status
Authentication
Response schema
Expected content
Latency
```

---

# Vector Store Health

Adapters may provide checks for vector infrastructure.

Examples:

- connectivity,
- authentication,
- collection/index existence,
- metadata access,
- lightweight query,
- latency.

Potential ecosystems include:

```text
Pinecone
Qdrant
Weaviate
Milvus
pgvector
OpenSearch
Elasticsearch
and others
```

Provider integrations are expected to grow through community adapters.

---

# Database Health

Database checks may include:

- connection establishment,
- authentication,
- minimal safe query,
- connection latency,
- read readiness.

Examples:

```text
PostgreSQL
MySQL
Redis
MongoDB
and others
```

---

# Dependency Graph

Dependency-aware diagnostics are a core design principle.

Example:

```text
                    Research Agent
                          │
           ┌──────────────┼───────────────┐
           │              │               │
         Model           MCP            Vector DB
           │              │
       HEALTHY      ┌─────┴─────┐
                    │           │
                 GitHub       Database
                 HEALTHY      DEGRADED
                                  │
                              PostgreSQL
                              HIGH LATENCY
```

AgentHealth should be able to calculate an overall status based on:

- critical dependencies,
- optional dependencies,
- health policies,
- failure propagation rules,
- latency thresholds.

---

# AgentHealth Configuration

A configuration file could look like:

<!-- spec-example: configuration -->
```yaml
version: v1

targets:
  - name: research-agent
    type: agent
    endpoint: https://agent.example.com

    checks:
      - reachability
      - protocol
      - authentication
      - capability
      - functional
      - latency

    thresholds:
      latency_ms: 1000

    dependencies:
      - name: github-mcp
        type: mcp
        endpoint: http://github-mcp:3000
        critical: true

      - name: vector-store
        type: vector-store
        endpoint: http://vector-db:6333
        critical: true

      - name: analytics-api
        type: http
        endpoint: https://analytics.example.com/health
        critical: false
```

The configuration specification will evolve independently of any particular implementation language. See [spec/configuration.md](spec/configuration.md) for the normative field definitions and defaults.

---

# Machine-Readable Health Format

AgentHealth is designed for both humans and machines.

Example JSON:

<!-- spec-example: result -->
```json
{
  "spec_version": "v1",
  "target": { "name": "research-agent", "type": "agent" },
  "status": "UNHEALTHY",
  "latency_ms": 182,
  "checks": {
    "reachability": { "status": "HEALTHY" },
    "authentication": { "status": "HEALTHY" },
    "protocol": { "status": "HEALTHY" },
    "capability": { "status": "HEALTHY" },
    "functional": { "status": "HEALTHY" },
    "latency": { "status": "HEALTHY" },
    "dependency": {
      "status": "UNHEALTHY",
      "message": "critical dependency vector-store is UNREACHABLE"
    }
  },
  "dependencies": [
    {
      "target": { "name": "github-mcp", "type": "mcp" },
      "status": "HEALTHY",
      "latency_ms": 84,
      "checks": { "reachability": { "status": "HEALTHY" } },
      "dependencies": []
    },
    {
      "target": { "name": "vector-store", "type": "vector-store" },
      "status": "UNREACHABLE",
      "latency_ms": null,
      "checks": { "reachability": { "status": "UNREACHABLE", "message": "connection refused" } },
      "dependencies": []
    }
  ]
}
```

The parent's own status is `UNHEALTHY`, not a literal copy of its `UNREACHABLE` dependency — see [spec/health-model.md § Status Aggregation](spec/health-model.md#status-aggregation) for why, and [spec/result-schema.md](spec/result-schema.md) for the full normative schema (including multi-target batch output).

---

# Architecture

AgentHealth is designed around a small core and extensible adapters.

```text
                         AgentHealth
                              │
                 Agent Health Specification
                              │
                       Health Engine
                              │
             ┌────────────────┼────────────────┐
             │                │                │
       Policy Engine    Dependency Engine   Output Engine
             │                │                │
             └────────────────┼────────────────┘
                              │
                       Adapter Interface
                              │
                              ├── Agent (direct / composite roles)
                              ├── HTTP / MCP / A2A
                              ├── Agentgateway / Agent Router integrations
                              └── Model / Tool / Data adapters
```

The Health Engine is designed to implement the Agent Health Specification. Its [Go core engine](core/README.md) now provides configuration loading, check execution, timeout/retry policies, result normalization, dependency evaluation, output formatting, and safety controls (see [Phase 2 — Core Engine](ROADMAP.md#phase-2--core-engine)). The experimental Agent Health Protocol (AHP) (see [Specification → Protocol → Implementation](#specification--protocol--implementation)), once finalized, would standardize how the Output Engine exposes and exchanges these results over the wire.

---

# Adapter Architecture

AgentHealth should not hard-code every vendor into its core.

Adapters provide technology-specific health logic.

Conceptually:

```text
AgentHealth Core
      │
      ├── MCP Adapter
      ├── A2A Adapter
      ├── HTTP Adapter
      ├── Agent checks (direct / composite roles)
      ├── Agentgateway integrations
      ├── Agent Router integrations
      ├── Model Adapter
      ├── Database Adapter
      ├── Vector Store Adapter
      └── Community Adapters
```

The adapter interface should allow third parties to implement checks without modifying the core engine.

The planned [Phase 23 — Plugin / Adapter Ecosystem](ROADMAP.md#phase-23--plugin--adapter-ecosystem)
includes extension examples for the first (direct) agent, downstream
(composite) agents, Agentgateway, and Agent Router. Direct/composite describe
agent roles, not separate target types or a requirement for separate adapters.
The diagrams show intended coverage; HTTP/API, MCP, and A2A are the currently
implemented adapters. Gateway/router integrations will use the supported
interfaces and contracts established in Phases 7–10.

---

# Distribution

Standalone binaries and source builds are available; see [installation](docs/installation.md). The Docker, SDK, and Kubernetes examples below describe planned integrations, not published packages or images.

## Standalone CLI

Pre-built binaries for:

```text
Linux x86_64
Linux ARM64
macOS Apple Silicon
macOS Intel
Windows x86_64
```

Example:

```bash
agenthealth --version
```

---

## Docker

AgentHealth can run without installing language runtimes.

Example:

```bash
docker run --rm agenthealth/agenthealth \
  ping mcp http://host.docker.internal:3000
```

Container images should support common CPU architectures where practical.

---

## Python / PyPI

Python applications can integrate AgentHealth through an SDK.

```bash
pip install agenthealth
```

Conceptual usage:

```python
from agenthealth import check

result = check(
    target_type="mcp",
    endpoint="http://localhost:3000"
)

print(result.status)
```

The Python package is an SDK/distribution mechanism.

**AgentHealth itself is not a Python-specific standard.**

---

# JavaScript / TypeScript / npm

Node.js and TypeScript applications can integrate AgentHealth through an SDK.

```bash
npm install @agenthealth/sdk
```

Conceptual usage:

```typescript
import { check } from "@agenthealth/sdk";

const result = await check({
  type: "mcp",
  endpoint: "http://localhost:3000"
});

console.log(result.status);
```

---

# Kubernetes

AgentHealth is designed for cloud-native environments.

A traditional readiness probe may verify only that an HTTP endpoint responds.

AgentHealth can validate whether the workload's critical AI dependencies are actually ready.

Example:

```yaml
readinessProbe:
  exec:
    command:
      - agenthealth
      - check
      - /etc/agenthealth/config.yaml
```

Potential Kubernetes integrations include:

- readiness probes,
- startup probes,
- Jobs,
- CronJobs,
- init containers,
- sidecars,
- deployment validation,
- admission/policy integrations in future releases.

---

# CI/CD

AgentHealth can act as a deployment validation gate.

Example workflow:

```text
Build
  │
  ↓
Deploy
  │
  ↓
AgentHealth
  │
  ├── Agent
  ├── A2A
  ├── MCP
  ├── Model
  ├── Tools
  └── Data
  │
  ↓
Promote / Rollback
```

Example:

```bash
agenthealth check production.yaml

if [ $? -ne 0 ]; then
  echo "Agentic system failed readiness validation"
  exit 1
fi
```

Exit-code semantics are defined in [the specification](spec/exit-codes.md).

---

# Security Principles

Health tooling frequently interacts with credentials and production infrastructure.

AgentHealth should therefore follow several principles:

- never print secrets,
- redact authentication material,
- avoid destructive checks,
- use minimal privileges,
- make active/functional probes explicit,
- support configurable timeouts,
- prevent accidental credential leakage in logs,
- clearly distinguish passive and active checks,
- provide secure defaults,
- support TLS verification by default.

Adapters should document the permissions required for their checks.

---

# Passive vs Active Checks

AgentHealth distinguishes between checks that inspect state and checks that execute functionality.

## Passive

Examples:

```text
Reach endpoint
Negotiate protocol
List capabilities
Read metadata
Measure connection latency
```

## Active

Examples:

```text
Invoke a tool
Run minimal model inference
Execute a database query
Send a test task to an agent
```

Active checks should be opt-in when they could create cost, side effects, or workload activity.

---

# Observability Integration

AgentHealth itself is not an observability platform.

However, health results may eventually be exported to observability ecosystems through:

- JSON,
- OpenTelemetry,
- Prometheus metrics,
- structured logs,
- webhooks,
- CI/CD annotations.

This allows existing monitoring platforms to consume AgentHealth results.

---

# Example Use Cases

### Developer debugging

```bash
agenthealth doctor local-agent.yaml
```

Find why a local agent cannot function.

### Platform engineering

Validate agent workloads before production deployment.

### SRE

Determine which dependency is causing an agent to become degraded.

### CI/CD

Prevent deployment promotion when critical agent dependencies are unavailable.

### Kubernetes

Use agent-aware readiness instead of simple HTTP-only readiness.

### MCP development

Validate that an MCP server initializes correctly and exposes expected capabilities.

### Multi-agent systems

Check whether required peer agents are available and compatible.

### Enterprise AI

Provide a consistent operational health interface across heterogeneous agent stacks.

---

# Design Principles

AgentHealth follows these principles:

### Universal

No dependency on one programming language, cloud, model provider, or agent framework.

### Open

Specification and reference implementation developed openly.

### Extensible

New protocols and technologies can be supported through adapters.

### Safe

Health checks should be non-destructive by default.

### Machine-readable

Every human-readable result should have a structured equivalent.

### Cloud-native

Designed for containers, Kubernetes, CI/CD, and automation.

### Composable

AgentHealth should complement existing platforms rather than replace them.

### Interoperable

The same health vocabulary should work across heterogeneous agent infrastructure.

---

# Repository Structure

AgentHealth is developed as a **single repository (monorepo)**. The specification (AHS), the proposed protocol (AHP), the core engine, CLI, adapters, SDKs, and distribution integrations are not split across separate repos — they version and release together from here:

```text
agenthealth/
│
├── README.md
├── ROADMAP.md
├── LICENSE
├── CONTRIBUTING.md
├── GOVERNANCE.md
├── SECURITY.md
├── CODE_OF_CONDUCT.md
│
├── spec/
│   ├── README.md
│   ├── health-model.md        # Agent Health Specification (AHS)
│   ├── protocol.md            # Agent Health Protocol (AHP) — proposed/experimental
│   ├── configuration.md
│   ├── result-schema.md
│   └── adapter-spec.md
│
├── cmd/
│   └── agenthealth/            # CLI entrypoint
│
├── core/                       # reference health engine
│
├── adapters/
│   ├── http/
│   ├── mcp/
│   ├── a2a/
│   ├── agent/
│   └── model/
│
├── sdk/                       # planned
│   ├── python/                 # planned PyPI SDK
│   └── javascript/             # planned npm SDK
│
├── integrations/
│   ├── docker/
│   ├── kubernetes/
│   └── ci/
│
├── examples/
│
└── docs/
```

This tree includes planned directories for future adapters, SDKs, and integrations. The exact repository structure may change as the implementation evolves, but the monorepo strategy itself is a settled decision (see [Phase 0 — Project Foundation](ROADMAP.md#phase-0--project-foundation) in the roadmap).

---

# Run the CLI

Download a standalone binary from [GitHub Releases](https://github.com/TheAgentHealth/agenthealth/releases); see [installation instructions](docs/installation.md). Downloaded binaries do not require Go.

Build from source with Go 1.23 or newer:

```bash
go build -o agenthealth ./cmd/agenthealth
./agenthealth ping http https://example.com
./agenthealth check examples/core-check/agenthealth.yaml --format json
./agenthealth doctor examples/core-check/agenthealth.yaml
./agenthealth version
```

See [CLI usage](docs/cli.md) for formats, safety defaults, adapter availability, and exit codes.

---

# Project Status

> **Early-stage / Pre-1.0**

Coverage for phases 0–6 is indexed in [the documentation guide](docs/README.md).

AgentHealth is under active design and development. The Go engine, CLI, HTTP/API adapter, MCP HTTP/stdio adapter, and A2A JSON-RPC adapter are implemented. Other adapters, SDKs, containers, and Kubernetes integrations remain roadmap work.

Interfaces, schemas, commands, and configuration formats may change before the 1.0 release.

The current priority is establishing:

- the health model,
- specification,
- CLI,
- core engine,
- MCP support,
- A2A support,
- generic HTTP checks,
- dependency-aware health,
- and universal distribution.

---

# Roadmap

See [ROADMAP.md](ROADMAP.md) for the detailed development roadmap.

High-level direction:

```text
Specification
      ↓
Core Engine
      ↓
CLI
      ↓
HTTP + MCP + A2A
      ↓
Direct + Composite Agent Health (Phase 7)
      ↓
Agentgateway Integration (Phase 8)
      ↓
Agent Router Integration (Phase 9)
      ↓
Dependency Graph
      ↓
Agent Health Protocol (AHP)
      ↓
Docker
      ↓
Kubernetes
      ↓
Python SDK
      ↓
JavaScript/TypeScript SDK
      ↓
Model + Data Adapters
      ↓
Observability Integrations
      ↓
Community Adapter Ecosystem
      ↓
Conformance + Ecosystem Interoperability Testing
```

---

# Governance

AgentHealth is intended to evolve as a community-driven open-source project.

The long-term objective is neutral, transparent governance suitable for broad ecosystem collaboration.

The project welcomes participation from:

- AI agent developers,
- MCP developers,
- A2A developers,
- platform engineers,
- SREs,
- DevOps engineers,
- AI infrastructure engineers,
- cloud-native developers,
- framework maintainers,
- model providers,
- enterprises,
- researchers,
- standards communities.

---

# Contributing

Contributions are welcome.

Useful contribution areas include:

- core implementation,
- protocol adapters,
- provider adapters,
- SDKs,
- Kubernetes integrations,
- CI/CD integrations,
- specification design,
- interoperability testing,
- direct/composite agent, Agentgateway, and Agent Router health integrations,
- conformance and communication-path failure testing,
- documentation,
- examples,
- security reviews,
- testing.

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

---

# Community Adapter Ideas

Potential future adapters include:

```text
Agent checks (direct / composite roles)
Agentgateway integrations
Agent Router integrations

MCP
A2A
OpenAI-compatible APIs
Anthropic
Amazon Bedrock
Azure AI
Google Vertex AI
Ollama

PostgreSQL
MySQL
MongoDB
Redis

Pinecone
Qdrant
Weaviate
Milvus
pgvector
OpenSearch
Elasticsearch

LangGraph
Semantic Kernel
CrewAI
AutoGen
and other agent runtimes
```

Listing a technology here does not imply endorsement, partnership, or current support.

---

# Interoperability Vision

The long-term goal is bigger than a CLI.

Planned [Phase 24 — AgentHealth Conformance](ROADMAP.md#phase-24--agenthealth-conformance)
will validate the health of the first (direct) agent, downstream (composite)
agents, Agentgateway, Agent Router, and their communication paths against the
approved AHS and AHP contracts. Supporting dependency results remain distinct.

[Phase 25 — Ecosystem Interoperability Testing](ROADMAP.md#phase-25--ecosystem-interoperability-testing)
will test first-to-downstream communication directly, through A2A, and through
gateways or routers across supported stacks. Scenarios include a healthy
downstream agent behind a failing path, a failing agent behind a healthy
gateway/router, shared backends, and multiple routes. No conformance program
or certification is currently available.

AgentHealth aims to establish a common operational contract for agentic systems:

```text
How do I determine whether an agent is healthy?

How does an agent expose readiness?

Can my first agent communicate with its downstream agents,
directly or through A2A, Agentgateway, or an Agent Router?

How do I represent dependency health?

How do I distinguish degraded from unhealthy?

How do platforms consume health consistently?

How can health work across MCP, A2A, models,
tools, databases, runtimes, and vendors?
```

A common health specification allows different implementations to answer those questions consistently.

---

# Long-Term Vision

AgentHealth aims to become a simple, universal operational primitive for agentic infrastructure.

The planned ecosystem covers the user's first (direct) agent, the downstream
(composite) agents it communicates with directly or via A2A, Agentgateway,
Agent Router, and supporting model, tool, MCP, and data dependencies. The
common health contract should distinguish each agent's health from gateway,
router, and communication-path health. This follows the
[roadmap ecosystem vision](ROADMAP.md#long-term-ecosystem-vision); direct/composite
are agent roles, and implementation availability follows the phase statuses.

Just as developers commonly use:

```text
ping
curl
health endpoints
readiness probes
```

for conventional distributed systems, the goal is for agent developers and operators to be able to use:

```text
agenthealth ping
agenthealth check
agenthealth doctor
```

across the agentic AI ecosystem.

---

# License

AgentHealth is released under the [Apache License 2.0](LICENSE).

The Apache-2.0 license was chosen for its explicit patent grant, which fits a project whose long-term ambition includes contribution toward broader interoperability standards (see [Interoperability Vision](#interoperability-vision)).

---

# Disclaimer

AgentHealth is an independent open-source project.

References to third-party projects, protocols, companies, products, or trademarks are for interoperability and descriptive purposes only and do not imply affiliation, sponsorship, or endorsement.

# Target Model

> Status: Draft. Defines the set of target types AgentHealth can check, referenced by [health-model.md](health-model.md#health-dimensions), [configuration.md](configuration.md), and [result-schema.md](result-schema.md).

## Target types

| Type | Description |
|---|---|
| `agent` | A single AI agent endpoint, including the user's first, user-facing agent; may answer directly or use dependencies and downstream agents |
| `multi-agent` | A system composed of multiple cooperating agents |
| `a2a` | An Agent-to-Agent protocol endpoint |
| `mcp` | A Model Context Protocol server |
| `tool` | A discrete tool invoked by an agent (not itself an MCP/A2A endpoint) |
| `model` | A model/LLM inference endpoint |
| `llm` | Specialization of `model` scoped specifically to large language models |
| `http` | A generic HTTP/API endpoint without an agent-specific protocol |
| `api` | A generic API dependency, synonymous with `http` for most checks |
| `vector-store` | A vector database used for retrieval |
| `gateway` | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ |
| `database` | A relational, document, or key-value database |
| `runtime` | An agent runtime/framework process |
| `gateway` | An agent or MCP gateway sitting in front of one or more targets |
| `custom` | Any target type not covered above, implemented via a community adapter |

Not every target type is required to be implemented by a given release; see
[ROADMAP.md](../ROADMAP.md) for adapter and integration phasing, including
Agent Health, Agentgateway, and Agent Router before the Dependency Graph and AHP phases.

The reference CLI currently implements `agent`, `multi-agent`, `http`, `api`, `mcp`, `a2a`, and `gateway`.
A2A defaults to 1.0 JSON-RPC with explicit 0.3.0 compatibility; see the [adapter guide](../docs/a2a-adapter.md).

## Direct and composite agent health

**Direct agent** means the user's first agent; **composite agents** are downstream
agents it communicates with directly or through A2A. Individual agents use
`agent`; A2A endpoints use `a2a`; a cooperating-system endpoint uses `multi-agent`.
Supporting models, tools and data services remain dependencies.

The [agent adapter](../docs/agent-adapter.md) validates runtime metadata,
liveness/readiness, required capabilities and named dependency discovery.
Explicit safe functional probes validate completed responses. Configure each
peer independently and a separate first-runtime path probe for communication
evidence; endpoint health alone does not establish that path. Roles and paths
use named nested results without new role or topology fields. The existing
[aggregation contract](health-model.md#status-aggregation) applies.

## Required fields per target type

Every target, regardless of type, MUST supply:

- `name` — human-readable identifier
- `type` — one of the values above
- `endpoint` — address used to reach the target

See [configuration.md](configuration.md#target-fields) for the full field list, and [adapter-spec.md](adapter-spec.md) for what an adapter must implement per target type.

## Applicable dimensions per target type

Not all [health dimensions](health-model.md#health-dimensions) apply to every target type. For example:

| Target type | Reachability | Protocol | Auth | Capability | Functional | Dependency | Latency | Configuration |
|---|---|---|---|---|---|---|---|---|
| `agent` | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| `mcp` | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ | ✓ |
| `a2a` | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ | ✓ |
| `model` / `llm` | ✓ | — | ✓ | ✓ | ✓ | — | ✓ | ✓ |
| `http` / `api` | ✓ | ✓ | ✓ | — | ✓ | — | ✓ | ✓ |
| `gateway` | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ |
| `database` | ✓ | — | ✓ | — | ✓ | — | ✓ | ✓ |
| `vector-store` | ✓ | — | ✓ | ✓ | ✓ | — | ✓ | ✓ |

This table is indicative, not exhaustive — individual adapters define exactly which dimensions they implement (see [adapter-spec.md](adapter-spec.md)).

A ✓ in the Functional column means the dimension is *applicable* to that target type, not that it runs by default: `functional` is always an active, opt-in-only check regardless of target type (see [configuration.md § Default Checks](configuration.md#default-checks)).

The Dependency column indicates only whether a `checks.dependency` summary entry is part of that type's *default* checks — it does not restrict whether a target of that type may declare nested `dependencies:`. Any target type may declare dependencies, which always execute and aggregate regardless of this column (see [configuration.md § Dependency Execution](configuration.md#dependency-execution)).

Gateway health, independently configured backends and proxied paths use separate
named targets/dependencies. The gateway adapter supports HTTP read-only signals;
MCP and A2A paths use their own adapters. See [Phase 8](../docs/agentgateway.md).

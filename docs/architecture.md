# Architecture Decisions

This document records the architecture decisions made in [Phase 0 — Project Foundation](../ROADMAP.md#phase-0--project-foundation). It complements the narrative [Architecture](../README.md#architecture) and [Adapter Architecture](../README.md#adapter-architecture) sections in the README, and the normative [spec/](../spec/README.md) documents.

## Repository strategy

**Decision:** single repository (monorepo). See [Repository Structure](../README.md#repository-structure).

## Core implementation language

**Decision: Go.**

Rationale:

- Compiles to a single static binary per platform, matching the project's "standalone binary + Docker + Kubernetes" distribution model (see [Distribution](../README.md#distribution)) without requiring a language runtime on the target machine.
- Well-suited to CLIs that talk to many network protocols concurrently (HTTP, MCP, A2A, databases) — mirrors the implementation language of comparable infrastructure CLIs (`kubectl`, `terraform`, `docker`).
- Straightforward cross-compilation for Linux/macOS/Windows and AMD64/ARM64, needed for [Phase 13 — Standalone Binaries](../ROADMAP.md#phase-13--standalone-binaries).
- Mature container/Kubernetes ecosystem tooling, relevant to [Phase 14 — Kubernetes Integration](../ROADMAP.md#phase-14--kubernetes-integration).

Python and JavaScript/TypeScript are not used for the core engine; they remain **SDKs that wrap the reference implementation** (see [Phase 15](../ROADMAP.md#phase-15--python-sdk--pypi) and [Phase 16](../ROADMAP.md#phase-16--javascript--typescript-sdk)), consistent with ["AgentHealth itself is not a Python-specific standard."](../README.md#python--pypi)

## CLI architecture

**Decision:** a single `agenthealth` binary with subcommands (`ping`, `check`, `doctor`, `version`, and later `serve`), following the same pattern as `git`, `kubectl`, and `docker`. Subcommands share a common core engine rather than being separate binaries.

## Adapter interface

**Decision:** adapters are compiled into the core engine initially (in-repo, in-process), rather than loaded as external dynamic plugins. This keeps the trust boundary simple and avoids a plugin ABI before the model is proven. A true out-of-process/dynamic plugin mechanism is deferred to [Phase 23 — Plugin / Adapter Ecosystem](../ROADMAP.md#phase-23--plugin--adapter-ecosystem). All adapters, in-process or future-external, must satisfy the [Adapter Contract](../spec/adapter-spec.md).

The CLI currently registers HTTP/API, MCP, and A2A 0.3.0 JSON-RPC adapters.
A2A uses the shared HTTP client and per-target run state for bounded card
retrieval, then validates metadata and the advertised same-origin endpoint.
The engine owns deadlines, retries, aggregation, latency, and output; the
adapter owns protocol checks and the opt-in interaction. See the
[A2A guide](a2a-adapter.md) for the Phase 6 behavior.

## Planned agent health scopes

[Phase 7](../ROADMAP.md#phase-7--agent-health) covers the direct agent
(the user's first agent) and composite agents (the downstream agents it
communicates with directly or through A2A). Planned checks distinguish each
agent's own health from the health of the communication path between agents.
Models, tools, MCP servers, data services, gateways, and routers are supporting
dependencies, covered separately by dependency health. These agent roles use
the existing target vocabulary; `multi-agent` describes the cooperating system
as a whole. Agent-specific checks are not yet implemented.
See the [target model](../spec/target-model.md#direct-and-composite-agent-health-planned).

## Configuration format

**Decision:** YAML, as specified in [spec/configuration.md](../spec/configuration.md).

## Result format

**Decision:** JSON, as specified in [spec/result-schema.md](../spec/result-schema.md). Human-readable terminal output is a presentation layer over the same underlying result, not a separate data model.

## Error model

**Decision:** errors are normalized into the existing health states rather than a separate exception/error taxonomy, per the shared [Error Classification table](../spec/health-model.md#error-classification). An adapter/engine error that still allows a result to be produced maps to `UNKNOWN`; an error that prevents producing any result at all surfaces only as CLI [exit code 6](../spec/exit-codes.md), never as a target status.

## Exit codes

**Decision:** see the normative mapping in [spec/exit-codes.md](../spec/exit-codes.md).

```text
0 = healthy
1 = degraded
2 = unhealthy
3 = unreachable
4 = misconfigured
5 = unknown / inconclusive result
6 = internal error (no result produced)
```

Implemented in [Phase 3 — Universal CLI](../ROADMAP.md#phase-3--universal-cli).

## Plugin strategy

**Decision:** no external plugin loading in the initial implementation (see Adapter interface above). Community adapters initially contribute via pull request into this repository's `adapters/` directory. A formal out-of-tree plugin strategy is evaluated in [Phase 23](../ROADMAP.md#phase-23--plugin--adapter-ecosystem).

## SDK strategy

**Decision:** SDKs (Python, JavaScript/TypeScript) are thin wrappers that invoke the `agenthealth` binary or call the same result/configuration schemas, rather than reimplementing check logic per language. This keeps one source of truth for health semantics (the Go core engine and the [spec/](../spec/README.md) documents) instead of N parallel implementations drifting apart.

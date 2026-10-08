# Phase 11 coverage and earlier-phase alignment

Experimental AHP HTTP v1 is included in v0.9.0. It is
not included in v0.8.0. The
[design and impact review](rfcs/phase-11-ahp.md) records scope and later requirements.

## Coverage checklist

| Area | Coverage |
|---|---|
| README and roadmap | Source status, HTTP operations, serving usage, deferred extensions and later-phase requirements |
| Specification | [HTTP binding](../spec/protocol.md), [envelope schema](../spec/schemas/ahp.schema.json), result/security references, index and changelog |
| Engine and CLI | [Snapshot server](../core/serve.go), [serve command](../cmd/agenthealth/serve.go), shared adapter registry and graceful shutdown |
| Documentation | [Serving guide](ahp.md), architecture, CLI/core guides, documentation and phase indexes |
| Examples | [Serving configuration and setup](../examples/ahp-check/README.md), linked from the example index; existing graph/nested examples remain valid |
| Tests | [HTTP contract tests](../core/serve_test.go), [CLI invocation tests](../cmd/agenthealth/serve_test.go), [schema fixtures](../tests/spec/test_ahp.py), existing regression suite |
| Validation | Go vet/race tests, Python suite, markdown links, marked document examples, all standalone YAML examples and diff checks |

## Earlier phases

The [Phase 10 alignment audit](phase-10-alignment.md) remains the baseline for
Phases 0–9. Phase 11 adds the following changes or explicit no-impact decisions.

| Phase | Alignment or no-impact rationale |
|---|---|
| 0: Foundation | Architecture, contribution/agent guidance and indexes reflect source availability. Governance, licensing and security reporting remain unchanged. |
| 1: AHS | AHP adds a separate envelope schema. Existing health states, dimensions, target vocabulary, configuration/result schemas and spec_version remain v1. Security and result documents link to the implemented binding. |
| 2: Engine | Background runs retain existing bounds, cancellation, credential redaction and adapter lifecycles. HTTP requests read snapshots and do not execute probes. |
| 3: CLI | Adds serve with loopback default, listen address and token environment reference. Existing health commands/formats/exit codes remain compatible; serving exits 0 on shutdown and 6 on tool failure. |
| 4: HTTP/API | Existing HEAD/GET checks and expectations remain unchanged. The AHP binding uses GET; generic HTTP checks remain separate from an AHP-specific consumer, which is not implemented. Existing examples remain valid. |
| 5: MCP | Existing HTTP/stdio/OAuth and inventory behavior remains unchanged. Each refresh starts a new engine run; no session is shared across refreshes. Existing examples require no migration. |
| 6: A2A | Existing protocol versions, card/task checks and safety remain unchanged. Endpoint checks still do not establish direct-agent communication. Existing examples remain valid. |
| 7: Agent health | Capability, declared readiness and completed functional/path evidence remain separate in authenticated results. Serving does not implement the application-specific runtime probe handler. |
| 8: Agentgateway | Gateway signals, direct backends and path nodes remain separate evidence. Public summaries conceal topology; authenticated evidence retains the result tree. |
| 9: Agent Router | Named router checks, direct backends and configured route evidence remain unchanged. Sharing a backend does not combine route nodes. |
| 10: Dependency graph | Detailed snapshots retain IDs, edge relationships/critical policies and recursive dependencies. Public responses omit these fields. No graph configuration or execution contract changes. |
| 13: Binary distribution | The existing builder compiles the same CLI and needs no packaging interface change. Published v0.8.0 binaries and historical release notes retain their actual capabilities. |

Earlier adapter guides already describe Phase 10 shared lifecycle and edge
semantics. Phase 11 does not change adapter protocols, so their setup examples
need no rewrite. Active checks explicitly configured for serve repeat per
refresh; this operational difference is documented in the protocol and guide.

## Remaining limits

Vendor extension fields, automatic discovery and independent implementation
interoperability are deferred. Remote TLS and per-client rate limiting require
a proxy. Authorization grants one whole-configuration scope; finer scopes and
multi-tenancy remain future work. Schema fixtures and reference tests do not
claim completion of Phase 24 conformance or Phase 25 interoperability.

Historical release notes and RFC proposals retain their historical scope;
current behavior follows the specifications, guides and phase-status index.

# Phase 10 coverage and earlier-phase alignment

Phase 10 is included in v0.8.0. The
[graph contract](rfcs/phase-10-dependency-graph.md) records the additive v1
extension and its impact review. See [release notes](releases/v0.8.0.md) for installation and migration.

## Phase 10 coverage

| Area | Coverage |
|---|---|
| README and roadmap | Source availability, graph capabilities, compatibility, and future-phase dependencies |
| Specification | Configuration/result fields, schemas, aggregation, adapter lifecycle, security, and changelog |
| Engine and CLI | Explicit graph validation, shared execution, parallel bounded calls, node budgets, edge aggregation, terminal/JSON/YAML evidence |
| Documentation | [Graph guide](dependency-graph.md), [RFC](rfcs/phase-10-dependency-graph.md), phase indexes and affected adapter guides |
| Examples | [Standalone graph configuration](../examples/graph-check/README.md); existing nested examples remain valid |
| Tests | [Engine graph tests](../core/graph_test.go), [CLI integration tests](../cmd/agenthealth/main_test.go), valid/invalid configuration and result fixtures, legacy regression suite |
| Validation tooling | Schema validation covers marked documentation and all standalone YAML examples, including the graph example |

## Earlier phases

Earlier completed phases retain their historical baselines. Phase 10 adds the
extensions below; it does not retroactively change older released binaries.

| Phase | Alignment or explicit no-impact rationale |
|---|---|
| 0: Foundation | Documentation indexes, contribution guidance and example validation coverage updated. Repository identity, governance, license and security reporting policy need no change. |
| 1: AHS | Configuration/result contracts, schemas, fixtures, aggregation context, adapter lifecycle and graph security aligned. Health states, dimensions, target vocabulary, specification version and exit codes remain unchanged. |
| 2: Engine | Graph validation and scheduling extend nested execution. Shared IDs execute once per run, edges apply independent policies, budgets remain independent, and results retain declaration order. Cleanup stays bounded. |
| 3: CLI | Existing `check`/`doctor` and terminal/JSON/YAML formats support graph configuration and labeled evidence. No new command is needed. Exit-code semantics remain unchanged. |
| 4: HTTP/API | Existing request methods, status/header/body expectations, credentials, TLS, response bounds, retries and active opt-in remain in force for each node. Guide updated; legacy HTTP examples remain valid. |
| 5: MCP | One referenced ID shares one target session/process and inventories within a run. Distinct nodes/runs remain isolated. OAuth and stdio configuration need no migration. Guide and adapter contract updated; existing examples remain valid. |
| 6: A2A | One referenced ID shares card/passive evidence and explicit task execution. Route nodes remain separate. A2A protocol versions and origin/credential safety remain unchanged. Guide and example compatibility guidance updated. |
| 7: Agent health | Advertised names match explicit resolved dependency identities. Peer endpoint health and first-runtime communication probes stay separate. Guide/example updated and CLI reference/path regression covers the distinction. |
| 8: Agentgateway | Explicit shared backend IDs and separate path nodes preserve signal/backend/path evidence. Guide/example updated; HTTP signal safety and functional opt-in remain unchanged. |
| 9: Agent Router | Shared backend IDs can be referenced by multiple router/route nodes. Stale future-work claims corrected in current guide/roadmap; named signals and independent route evidence remain intact. Guide/example updated. |

Phase 13’s existing standalone binary baseline needs no new packaging interface:
its build/release commands compile the same CLI. Installation versions and
published release notes continue to describe the released artifacts. Docker,
Kubernetes, SDKs and broader conformance/interoperability remain planned.

## Validation and limits

Run `go vet ./...`, `go test -race ./...`, the Python test suite, the markdown
link checker, documentation example validation and `git diff --check`.
Existing adapter regressions cover earlier behavior; graph-specific tests cover
shared identity, policies, cycles, bounds, active deduplication and path evidence.
The optional SDK interoperability cases require their integration environment;
skipped cases do not establish live product interoperability.

The standalone YAML files are service configuration templates. Their schema
validation does not start services or certify vendor deployments. Historical
RFCs and release notes retain their original proposal/release scope; current
behavior follows the specifications, guides and phase-status index.

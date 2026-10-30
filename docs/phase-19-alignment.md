# Phase 19 alignment audit

Scope: Qdrant REST and Pinecone dense-vector data-plane REST. The
[design](rfcs/phase-19-vector-store.md) and
[RFC #57](https://github.com/TheAgentHealth/agenthealth/issues/57) record the
additive v1 extension authorized by the user. This records implementation
scope; publication requires the synchronized release verification checklist.

| Phase | Alignment and impact |
|---|---|
| 0 foundation | README, contribution/agent guidance, documentation and example indexes link the implemented scope and explicit provider limits. |
| 1 specification | Existing vector-store vocabulary reused; optional vector_store block, strict YAML/Go validation, canonical schemas and valid/invalid fixtures. Placeholder targets remain valid configuration but fail adapter Configuration until options are supplied. Health states, result v1 and exits unchanged. |
| 2 engine | Existing credential snapshot, redaction, response tracking, transport, budgets, pools, passive retry and active no-retry rules reused. Metadata cache is per target/run. Canonical vector diagnostics added. |
| 3 CLI | Shared registry supports check, doctor and serve; help lists vector-store. Ping cannot supply provider/collection; documented config-file use. |
| 4 HTTP | Uses the supplied verified-TLS, redirect-denying client. Read-only metadata POST is passive by semantics. Generic HTTP method/body behavior unchanged. |
| 5 MCP | No protocol change. Vector readiness does not establish MCP tool availability or safety. |
| 6 A2A | No protocol change. Vector health does not establish peer communication. |
| 7 agent | Retrieval services are supporting dependencies, never composite agents. Direct/composite and communication-path evidence remains distinct. |
| 8–9 gateway/router | No changes to HTTP signals or product claims. Retrieval remains a separate dependency. |
| 10 graph | Existing nested dependencies/shared nodes carry vector options; critical/optional aggregation and independent credentials remain unchanged. Cache never crosses nodes or runs. |
| 11 AHP | Shared registry exposes vector health through snapshots. Authorization, freshness, readiness and independent /live behavior unchanged. |
| 12 Docker | Adapter compiles into the ordinary image; no external vector client executable needed. Shared version/image metadata advances with release. |
| 13 binaries/packages | Pure HTTP adapter ships in five-platform binaries and existing package/OCI formats. Release checks remain mandatory. |
| 14 Kubernetes | Existing Secret environment injection supports vector API keys. Readiness reflects dependency health; /live remains independent. Check budgets must fit exec/Job deadlines; chart, appVersion and images stay synchronized. No chart API extension needed. |
| 15 Python | Bundled canonical configuration schema updated; matching CLI owns checks and safety. No Python provider client added. |
| 16 JavaScript | Bundled canonical schema updated and parity-tested; typed extensible configuration retains options. Installed matching CLI owns health semantics. |
| 17 models | Retrieval and embedding/inference are independent evidence. Queries never call an embedding model. Existing model opt-in and token bounds unchanged. |
| 18 databases | Native database probes unchanged. SQL login/SELECT 1 does not verify pgvector; it remains a future vector dialect. |
| 19 vectors | Qdrant collection metadata, Pinecone index metadata, authentication, dense-vector availability, explicit minimal query and latency implemented. Other listed providers and vector dialects remain planned. |
| 20 runtime | Future integrations must preserve retrieval dependency, direct/composite agent and path evidence separately, and propagate cancellation. |
| 21 CI/CD | Add real Qdrant 1.15.5 disposable metadata/query/missing-collection/auth-denial matrix alongside platform/race/schema/SDK/deployment gates. |
| 22 observability | Stable canonical codes and existing result fields only. No retrieved objects, metadata, query values or exporter added. |
| 23 plugins | Trusted compiled adapter only; no dynamic plugins or tenant isolation claim. |
| 24 conformance | New fixtures and response/transport safety regressions extend existing contracts. Full conformance remains planned. |
| 25 interoperability | Controlled Qdrant 1.15.5 matrix; Pinecone official API fixtures only. Pinecone cloud accounts, all other products, multi-node/failover and version matrices remain future work. |
| 26 hardening | Bound response/vector sizes, finite nonzero vectors, fixed read-only requests, top-1/no data returns, no query retries, verified TLS/redirect denial and credential-safe diagnostics. No security audit claim. |

Validation: Go vet/race, strict schema fixtures, documentation examples/links,
SDK schema parity and installed-package checks, real Qdrant matrix, existing
native database tests and Kubernetes/Helm runtime gates. Release verification
must record actual artifacts, provenance, checksums and anonymous package access
before updating published-status claims. Historical release notes retain their
original scope.

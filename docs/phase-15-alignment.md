# Phase 15 coverage and alignment

Python SDK source targets v0.13.1; publication remains a separate verification
step. Python 3.10+ uses an explicitly installed matching CLI. No binary is
installed or downloaded automatically.

## Coverage

| Deliverable | Implementation |
|---|---|
| Core/adapter access | Sync `check` and `Client.check` invoke the existing Go CLI |
| Async API | Native async subprocess execution with cancellation cleanup |
| Typed results | AHS statuses, checks, diagnostic codes, transport steps, roots and dependencies |
| Configuration loader | Bounded JSON/YAML, duplicate-key rejection, schema validation; Go semantic validation |
| Agent/path and graph evidence | Target IDs, incoming relationship/critical metadata, separate check evidence |
| Error model | Health exits return results; fixed exceptions for tool, compatibility, deadline and output failures |
| Distribution | Shared-version wheel/sdist, GitHub Release assets/checksums/provenance and PyPI OIDC/attestations |
| Validation | Real CLI integration, built-wheel/source-archive installation, OS/Python CI matrix |

## Earlier phases

| Phases | Impact and alignment |
|---|---|
| 0: Foundation | README, roadmap, status, contribution/release documentation and CI discover the SDK. |
| 1: Specification | Vendored schemas are byte-identical and tested. AHS/AHP v1 remain unchanged. |
| 2: Engine | Go owns health aggregation, safety, retries, budgets, graph validation and credential resolution. |
| 3: CLI | Exact matching software version; JSON check results and exits 0–6 remain authoritative. |
| 4–6: HTTP/MCP/A2A | Existing configuration selects adapters. No new protocol or OAuth contracts; stdio remains trusted executable input. |
| 7: Agent health | Typed diagnostics and separate direct/composite/path evidence; no invented role fields. |
| 8–9: Gateway/router | Configured HTTP signals remain distinct from independently verified routing behavior. |
| 10: Graph | Ordered roots, node IDs, incoming relationships/critical metadata and shared projections preserved. |
| 11: AHP | No endpoint or serving contract change; SDK is a CLI client, not an AHP server. |
| 12–13: Distribution | No archive/install behavior changes; manual CLI installation retains exact platform/checksum/provenance verification. New Python assets join the complete release set. |
| 14: Kubernetes | Probes/Jobs/Helm still consume the CLI. SDK deployment requires a matching CLI; no Operator/CRD prerequisite. Version-matrix evidence and broader authenticated graph upgrade/rollback extend validation. |

Phase 14's [alignment audit](phase-14-alignment.md) remains authoritative for
its historical implementation and publication. No physical edit to every older
example is needed: existing configurations and health semantics are consumed
unchanged and are tested against released/current/released binaries.

## Later phases

Phase 16 should reuse the subprocess/version/error contract and synchronized
publication model. Phases 17–19 remain Go adapter expansion; listing a target
type in Python does not implement it. Phase 20 can consume sync/async reports;
Python reference middleware does not require the SDK. Phase 21 can use report
exit codes and reuse the new package/compatibility CI. Phase 22 may consume
structured reports without introducing telemetry in this SDK. Phase 23 still
requires third-party adapter isolation rather than loading Python plugins in
process. Phase 24 should cover SDK round-trips and schema conformance. Phase 25
must keep product/framework interoperability distinct from SDK fixture coverage.
Phase 26 retains scoped authorization, external audit, native signing, broader
fault/resource tests and Windows descendant lifecycle controls.

## Validation and release acceptance

Run repository Go/Python/documentation checks, strict Helm lint and Kubernetes
runtime scenarios, plus the dedicated installed-wheel CI matrix. Release builds
must verify SDK metadata and vendored schema versions; the wheel rebuilt from
the sdist must install independently. Post-publication acceptance includes
GitHub signatures/checksums, archive contents and Linux behavior, PyPI hashes,
attestations, clean-environment installation and exact-version CLI execution.
Do not mark publication complete until those external checks pass.

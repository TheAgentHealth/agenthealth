# Phase 16 impact review and alignment

Phase 16 implements the Node.js ESM JavaScript/TypeScript SDK, synchronized at
v0.14.0 with CLI, Python, chart and all existing distribution formats. Publication
is a separate verification milestone. The existing npm 0.0.0 placeholder remains
historical and is not the SDK described here.

| Phases | Assessment and required alignment |
|---|---|
| 0–1 foundation/spec | No health/schema changes. Canonical configuration/result schemas are bundled and parity-tested. |
| 2–3 engine/CLI | Go retains aggregation, policy, diagnostics and exit semantics. Exact CLI version probe precedes checks; failures are sanitized. |
| 4 HTTP | Real local HTTP success, failure and batch checks exercise the wrapper. |
| 5 MCP | Config accepts existing protocol options and opt-ins. POSIX groups are stopped on cancellation; Windows descendant cleanup remains limited. |
| 6 A2A | Existing A2A 1.0 and explicit 0.3 compatibility configuration passes through; no new protocol claims. |
| 7 agents | Direct/composite terminology and named peer/path evidence retained; supporting dependencies are not composite agents. |
| 8–9 gateway/router | Existing read-only signals and separate backend/path evidence preserved without new product support claims. |
| 10 graph | Typed IDs, relationship/critical metadata and recursive dependencies preserve CLI evidence. No SDK graph reaggregation. |
| 11 AHP | Experimental serving stays a CLI capability; no HTTP server/browser SDK added. |
| 12–13 distribution | Signed/checksummed reproducible npm tarball joins release assets and OCI bundle; npm registry bytes must match. |
| 14 Kubernetes | Deployment pins/chart metadata advance to the shared version; existing Kubernetes CI/runtime requirements remain. |
| 15 Python | Python metadata advances to the shared version; canonical schemas and subprocess/version contracts stay aligned. |
| 17–19 adapters | Planned adapters require Go implementation first; SDK configuration can follow canonical schema additions. |
| 20 frameworks | Future integrations consume installed SDK, propagate AbortSignal and preserve health reports versus tool exceptions. |
| 21 CI/CD | Six Node platform/version jobs and exact-source pre-release gate; `release.yml` with environment `npm` is the npm trusted OIDC identity. |
| 22 observability | Reports retain diagnostics; no exporter implemented or implied. |
| 23 plugins | No plugin loading or new adapter registration introduced. |
| 24 conformance | Canonical schema parity and installed-package tests added; full conformance remains planned. |
| 25 interoperability | Controlled Node/CLI tests added; real JavaScript framework/product interoperability remains future work. |
| 26 hardening | Bounded output/configuration, private temporary files, deadline/cancellation and sanitized errors included. Browser support, OS signing and Windows descendant termination remain limitations. |

Completed extensions: source SDK/API declarations, configuration loading,
normalized report parsing, error/cancellation coverage, schema parity, deterministic
packaging, installed-tarball smoke tests, six-job CI, shared-version metadata,
OIDC publication using the configured workflow, registry integrity/signature
verification, guides/examples and roadmap/status alignment. Historical release
records are retained. Later phase design includes the Phase 20/25 integration work.

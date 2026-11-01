# Phase 21 impact and alignment review

Released in v0.19.0: CI/CD deployment gates, JUnit/GitHub reports, verified
binary installation and provider examples. Synchronized publication and public installation
checks passed; see the [verification record](releases/verification-v0.19.0.md).
See the [guide](ci-cd.md) and [examples](../examples/ci/README.md).

| Phase | Assessment and completed alignment |
|---|---|
| 0 foundation | Focused integration directory, standard contribution/CI/release process; no new service. |
| 1 specification | Existing v1 schemas and exit codes unchanged; CI reports are projections, not new AHS wire formats. |
| 2 engine | Reporting consumes validated results; no execution, aggregation, retry or active-check changes. |
| 3 CLI | Additive `junit`/`github` formats for check/ping/doctor; output failures still return 6. |
| 4 HTTP | Passive readiness example; no hidden GET, redirects or functional activation. |
| 5 MCP | Existing trusted configuration and subprocess boundary; runner process timeout is not child-process isolation. |
| 6 A2A | Endpoint evidence remains distinct from verified peer exchange; no protocol change. |
| 7 agents | Direct/composite/path checks use existing reviewed examples; supporting dependencies retain their role. |
| 8 gateway | Configured gateway/backend/path evidence retained; no product certification. |
| 9 router | Configured router signals retain existing scope; no inferred route decision. |
| 10 graph | Each dependency occurrence gets a distinct positional report identity; JSON retains IDs, relationships and edge policy. |
| 11 AHP | No protocol or freshness changes; deployment checks do not introduce an AHP authorization mechanism. |
| 12 Docker | Same CLI formats in image; all implemented distributions remain mandatory at release. |
| 13 binaries | Exact OS/architecture/version download verifies checksum plus source-tag/workflow provenance before install. |
| 14 Kubernetes | Reuse schema/Helm tests and explicit-context runtime suite. Job completion, failure and timeout determine gates. Preserve live/ready separation and secret references. |
| 15 Python SDK | No SDK API changes; health remains CLI-owned. Python integration helper uses only stdlib. |
| 16 JS/TS SDK | No SDK API changes or report parsers; SDK callers continue requesting JSON. |
| 17 model | Supporting dependency only; no automatic inference. |
| 18 database | Existing login/PING and explicit SQL opt-in; no implicit query. |
| 19 vector store | Existing provider metadata and explicit read-only query opt-in; no query activation by CI. |
| 20 runtime | Safe graph/peer callbacks unchanged; pipeline results cannot claim additional runtime interoperability. |
| 22 observability | CI projections deliberately omit sensitive topology; future exporters must separately preserve evidence and redact topology. |
| 23 plugins | No third-party code loading or sandbox guarantee added. |
| 24 conformance | Report and gate regression coverage establishes behavior, not certification. |
| 25 interoperability | Provider files are templates; hosted-provider execution and real product/version matrix remain future validation work. |
| 26 hardening | No secret exposure to untrusted PR configuration; runner isolation, scoped credentials and bounded execution remain required. |

Validation includes all six health statuses, dependency occurrence identities,
empty/invalid/cyclic evidence, writer errors, CLI health/tool exits, Action version
mismatch/timeouts/stale reports/path guards and installer checksum/provenance
failures. Existing Go/schema/SDK/distribution/Kubernetes checks remain in place.
SARIF is explicitly deferred because health evidence is not a security finding.

# AgentHealth conformance

Phase 24 provides offline AHS v1 result validation and a reference behavioral
suite. These are two distinct forms of evidence. Neither awards certification,
a compatibility badge, a security audit, or independently measured product
interoperability. See the [all-phase impact review](phase-24-alignment.md).

## Validate a result document

```bash
agenthealth conformance test examples/conformance/healthy.json
agenthealth conformance test examples/conformance/dependency.json
```

The argument is a regular local UTF-8 JSON file containing a single result or a
nonempty batch. The command does not contact endpoints, execute adapters, load
credentials, or run configurations. It accepts at most 4 MiB and 64 dependency
levels, rejects duplicate keys and trailing documents, and tolerates extension
fields permitted by the v1 schema. Only `spec_version: v1` and the current AHS
dimensions are supported. Unknown versions or dimensions need a future profile.
YAML, configuration files, AHP envelopes and observability exports are not inputs
to this result profile.

Validation covers required fields, types, health states, target types, graph IDs,
relationships, diagnostic codes, transport steps, latency and dependency shape.
It also recomputes status aggregation: dependency diagnostic entries do not
contribute to own health, critical failures contribute `UNHEALTHY`, optional
failures contribute `DEGRADED`, and omitted criticality defaults to true.
An empty-check `MISCONFIGURED` result represents a pre-check configuration failure.

Output is a JSON report with `conformance_version`, `profile`, `passed`, `results`
(top-level result count on success), and a stable `code`: `conformant`,
`invalid_document`, `document_too_large`, `unsupported_spec_version`,
`invalid_result`, or `inconsistent_status`. Reports omit input names, paths,
messages and parser errors. Read/open/write errors produce a generic stderr
diagnostic instead. Exit 0 means the evidence passed; exit 6 means invalid
evidence, invocation or tool failure. A conformant `UNHEALTHY` or `UNREACHABLE`
document still exits 0. Health commands retain their existing exits 0–5.

A document cannot prove that its producer actually performed a check, honored
active opt-in, verified TLS, or removed every secret. Do not treat a passing
document as evidence of those runtime properties or of current service health.

## Reference behavioral suite

From a trusted source checkout with Go 1.26+ and Python installed:

```bash
python3 scripts/test_conformance.py --report conformance-report.json
```

The runner executes the reference engine, CLI and adapter Go tests with the race
detector, a fresh test run and a 15-minute limit. A C compiler is needed for race
testing. It requires explicit passing evidence for every case in its versioned
catalog; missing, failed or skipped mandatory cases, including skipped subcases,
fail the run. Other test skips are counted in the report and do not imply tested
external services. Raw subprocess output is not included in the report; run
`go test -race ./...` for local diagnostics when the suite fails.

| Area | Mandatory evidence |
|---|---|
| Result contract and semantics | Offline positive/negative cases, required fields, bounded parsing, severity and dependency aggregation |
| Safety and error behavior | Prerequisite gates, passive defaults, active no-retry, timeout/partial-response distinction, panic normalization, secret suppression and bounded calls |
| HTTP, MCP and A2A | Passive readiness, header/body separation, credential controls, protocol versions, OAuth redaction, no credential forwarding on redirects |
| Direct/composite agents | Declared health, bounded safe tasks, separate peer/path evidence, actual first-runtime peer callback in controlled fixtures |
| Gateway/router | Separate signal, backend and path evidence |
| Dependency graph | Invalid graphs, shared execution, edge policies, concurrency, independent budgets and topology redaction |
| AHP | HTTP binding, authorization, status mapping and graph evidence |
| Model/database/vector | Transport controls, endpoint safety, cancellation and functional opt-in/no-retry |
| CI/observability | Safe report projections, redaction, status mapping, one-hot gauges and duplicate-edge uniqueness |
| Executable adapters | Versioned manifest, hash pin, bounded streams, response markers, normalized output and cancellation |

The catalog reuses tests in their owning packages so fixes remain coupled to
the implementation. This is reference-implementation evidence, not a portable
black-box certification of arbitrary third-party software. Independent products,
real database/provider matrices and deployed Kubernetes configurations require
their separate test suites. Framework/SDK and distribution checks remain in CI;
their results are not silently inferred from the Go catalog. Phase 25 records
exact product versions and interoperability; Phase 26 adds remaining hardening.

The [examples](../examples/conformance/README.md) include a schema-valid document
with deliberately inconsistent status, showing why shape validation alone is
insufficient. CI runs the catalog, retains its JSON evidence, and separately runs
schema, SDK, framework, packaging and deployment gates.

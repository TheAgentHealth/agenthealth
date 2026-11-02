# Phase 22 impact and alignment review

Released in v0.20.0; see [publication verification](releases/verification-v0.20.0.md). Phase 14 was
already released: its original chart remains historical, and current charts
follow the shared software version policy. This review covers all phases.

| Phases | Assessment and completed alignment |
|---|---|
| 0–1 foundation/spec | Reduced export contract documented separately; AHS/AHP schemas and fixtures unchanged because no canonical field or health semantics change. |
| 2–3 engine/CLI | Validated, buffered writers and four CLI formats; status codes and write failures tested. |
| 4–9 adapters/agents/gateway/router | Allowlisted target types/check dimensions and relationships retain evidence distinctions; names, messages and endpoints omitted. No adapter network behavior changes. |
| 10 graph | Explicit-ID nodes/checks deduplicated; incoming edge policies and repeated edge occurrences retained; anonymous nodes remain distinct. Aliases are snapshot-local. |
| 11 AHP | Serving/probes retain canonical JSON; no new HTTP endpoints or instrumentation of incoming requests. |
| 12–14 distributions/Kubernetes | Export code ships in every CLI format; deployment modes and chart values remain compatible. Textfile/log delivery is external; no unimplemented ServiceMonitor claim. Chart versions align during synchronized release preparation. |
| 15–16 SDKs | Existing typed JSON handling unchanged; shared release version updated with all formats. No invented SDK exporter API. |
| 17–19 model/database/vector | Status/check/latency evidence passes through the common projection; provider identifiers and query details omitted. No active checks added. |
| 20 runtime | Runtime hooks and peer callbacks unchanged; callers can collect CLI snapshots externally. |
| 21 CI/CD | JUnit/GitHub projections remain separate; observability formats preserve health exit codes. Delivery examples retain nonhealthy evidence and distinguish tool/delivery failures. |
| 22 exports | Prometheus text, JSON Lines, OTLP/HTTP JSON log payloads and webhook v1 payloads implemented with explicit external delivery examples. No retry/background service. |
| 23 plugins | Future adapters must normalize to the existing evidence model before export and keep arbitrary labels out. |
| 24 conformance | Future suite must validate export mapping, redaction, duplicate-edge uniqueness and writer failures; no certification claimed. |
| 25 interoperability | Actual collector/hosted receiver matrices remain unverified; payload generation alone is not product compatibility. |
| 26 hardening | Future work includes collector delivery backpressure/freshness, large-graph cardinality and restricted telemetry access. No audit/isolation claim. |

Validation: [core tests](../core/output_observability_test.go),
[CLI tests](../cmd/agenthealth/main_test.go),
[delivery examples](../examples/observability/README.md), existing full Go and
Python/schema/documentation checks. Historical release notes are preserved.

# Phase 17 alignment audit

Scope: model/LLM OpenAI-compatible and Anthropic dialects, implemented in source
for v0.15.0. Based on [RFC #14](https://github.com/TheAgentHealth/agenthealth/issues/14)
and [PR #17](https://github.com/TheAgentHealth/agenthealth/pull/17) by Sharath K.
The contributor commit retains its original author; current integration and
additional validation are separate commits. Hosted provider/product testing is
not established by controlled HTTP fixtures or the contributor's Ollama report.

| Phases | Impact and required extension |
|---|---|
| 0 foundation | Contribution, status, README and guide indexes include model scope and credit. |
| 1 specification | Additive optional `model` block, schema fixtures, Go validation and stable diagnostics; v1 states/result shapes/exit codes unchanged. |
| 2 engine | Existing credential resolution, retry/cancellation, aggregation, redaction and bounded HTTP client reused; canonical model diagnostics added. |
| 3 CLI | Both target aliases registered in shared check/doctor/ping/serve registry; passive ping regression. |
| 4 HTTP | No HTTP adapter changes; model transport retains verified TLS, no redirects, bounded bodies and no error-body disclosure. |
| 5 MCP | No protocol/transport changes; full regression suite retained. |
| 6 A2A | No protocol changes; model availability does not prove agent communication. |
| 7 agents | Models remain supporting dependencies; graph example separates direct/composite agents, path and model nodes. |
| 8–9 gateway/router | Existing signals unchanged; no new product interoperability claims. |
| 10 graph | Shared model IDs execute once per run, independent runs refetch; optional/critical edge policies unchanged. |
| 11 AHP | Shared registry supports model targets in snapshots; freshness/auth/readiness behavior unchanged. |
| 12–13 distribution | Compile adapter into all CLI/image/package formats; synchronized v0.15.0 metadata and signed publication required. |
| 14 Kubernetes | Chart/image pins advance together; configuration can reference model credentials supplied as secrets. No active default probes. Existing runtime CI remains required. |
| 15–16 SDKs | Both bundled configuration schemas updated and parity-tested; wrappers preserve engine-owned classifications. SDK/CLI versions advance together. |
| 18–19 data adapters | Still planned; reuse scoped configuration, explicit active checks and bounded responses without implying SQL/vector support. |
| 20 frameworks | Future integrations distinguish advertised models from successful inference and agent/path evidence. |
| 21 CI/CD | Existing platform, schema, SDK and distribution jobs cover new source; new safety regressions included. |
| 22 observability | Stable diagnostic codes available; no exporter or usage/billing telemetry added. |
| 23 plugins | Compiled trusted adapter only; no dynamic loading or tenant isolation introduced. |
| 24 conformance | Configuration fixtures and malformed-response tests added; full conformance remains planned. |
| 25 interoperability | Controlled dialect fixtures plus attributed contributor Ollama report; hosted OpenAI/Anthropic, vLLM/LiteLLM and cloud-signing matrices remain future work. |
| 26 hardening | TLS/redirect/deadline/partial-response/no-retry/secret regressions; token bound relies on remote enforcement and truthful usage. No security-audit claim. |

Validation: Go vet/race suite, Python/schema/example suite, SDK schema parity,
Node tests/typecheck, markdown links and documentation example validation.
Release publication and native multi-platform/Kubernetes checks use the existing
synchronized release gate. See [release notes](releases/v0.15.0.md).

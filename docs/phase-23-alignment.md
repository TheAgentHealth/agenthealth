# Phase 23 impact and alignment review

Implemented in source for the proposed v0.21.0 synchronized release. Published
availability remains recorded separately in [capability status](phase-status.md).
Phase 14 is already released; its original chart tag remains historical.

| Phases | Assessment and completed alignment |
|---|---|
| 0–1 foundation/spec | Versioned executable protocol and RFC #68 supplement the existing adapter contract. No AHS/AHP field, target type or configuration schema change. |
| 2 engine | Existing registry metadata validation, prerequisite gates, normalized errors, active opt-in, deadlines and aggregation apply. Metadata has explicit JSON keys. |
| 3 CLI | Repeatable explicit `--adapter` option for check/ping/doctor; invalid manifests and duplicate types fail with code 6. No automatic execution from target YAML. |
| 4–9 adapters/agents/gateway/router | Built-ins retain ownership. Direct/composite role examples keep separately observed peer, path, gateway/router and backend evidence distinct. Rich built-in options are not forwarded. |
| 10 graph | Engine remains the sole owner of node scheduling, independent dependencies and shared-node execution. Children receive no graph/topology. |
| 11 AHP | No remote plugin submission or serving flags. The same canonical JSON response contract remains. |
| 12–14 distribution/Kubernetes | Host transport ships in every binary/container. External executables are not bundled or installed by chart/manifests. No new mount/command or RBAC permission. Shared release preparation updates all pins; historical chart remains unchanged. |
| 15–16 SDKs | No new SDK manifest API; existing typed JSON and exact CLI version requirements remain. Packages adopt the shared version for release. |
| 17–19 model/database/vector | Native adapters unchanged; only explicitly configured bearer credentials reach an executable. Native protocol settings and additional secrets do not cross this transport. |
| 20 runtime | Framework integrations unchanged. Executable adapters do not establish framework/product interoperability. |
| 21 CI/CD | Health codes and JUnit/GitHub projections remain unchanged. No untrusted text is emitted as workflow commands. |
| 22 observability | Only normalized status enters existing projections. Plugins cannot introduce labels, messages, endpoints or arbitrary result fields. |
| 23 ecosystem | Local manifests, pinned executable bytes, metadata validation, bounded strict JSON events, no environment inheritance and cancellation implemented; documented runnable fixture and trust limits. |
| 24 conformance | Future independent suite must test manifest compatibility, response markers, strict bounds, active opt-in and normalized output. No certification claim. |
| 25 interoperability | Controlled local protocol tests are not real community adapter/product testing. Retain exact-version external matrix work. |
| 26 hardening | OS sandboxing, egress restrictions, interpreter/dependency pinning, Windows descendant termination and external security audit remain future work. Process lifecycle separation does not establish tenant isolation. |

Validation includes native process protocol, hash mismatch, manifest compatibility,
metadata immutability, empty environment, explicit credentials, malformed/oversized
output and cancellation/partial-response tests; CLI registration tests; existing
Go race/vet, schema/docs, SDK and deployment gates. Historical phase alignment and
release notes retain their original scope; this review supersedes their future
Phase 23 statements for the current source.

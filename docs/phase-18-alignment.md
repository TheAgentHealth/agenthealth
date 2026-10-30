# Phase 18 alignment audit

Scope: PostgreSQL, MySQL, Redis and MongoDB native database health under the
existing `database` target vocabulary. [RFC #54](https://github.com/TheAgentHealth/agenthealth/issues/54)
records the authorized additive configuration and safety design. The synchronized v0.16.0 publication is verified; see the
[verification record](releases/verification-v0.16.0.md).

| Phases | Impact and completed alignment |
|---|---|
| 0 foundation | README, documentation/example indexes and contribution guidance describe the implemented database scope. |
| 1 specification | Additive optional `database` block, canonical/bundled schemas, valid/invalid fixtures and strict Go validation. Existing states, result shape, contract v1 and exits remain unchanged. |
| 2 engine | Existing credential references/environment snapshot/redaction, deadlines, aggregation and pools reused. No raw driver errors escape; native response tracking preserves no-response retry classification. |
| 3 CLI | Shared check/doctor/ping/serve registry registers database targets; SQL ping requires a configuration file with username/database options. |
| 4 HTTP | No transport changes; native drivers do not use the engine HTTP client. TLS verification remains mandatory by default. |
| 5 MCP | No protocol changes; database evidence does not establish tool availability. Existing regressions retained. |
| 6 A2A | No protocol changes; database health does not establish peer communication. Existing regressions retained. |
| 7 agents | Data services remain supporting dependencies. Direct/composite agents and communication paths retain separate evidence. |
| 8–9 gateway/router | No changes to configured HTTP signals or product compatibility claims. A database is a separate dependency. |
| 10 graph | Existing database vocabulary can be used in nested dependencies/shared graph nodes; per-run native evidence cache does not cross node/run boundaries. Critical/optional edge aggregation is unchanged. |
| 11 AHP | Same shared registry serves database results; endpoint authorization, snapshot freshness and readiness/liveness remain unchanged. |
| 12–13 distribution | Native drivers compile into the ordinary five-platform CLI, images and package formats. No external database executables required by installed CLI. Shared release metadata advances together. |
| 14 Kubernetes | Existing configuration/Secret injection supports database passwords. Explicit plaintext is only for loopback test services. Keep AHP/probe/Job deadlines aligned and dependency readiness separate from `/live`. Chart/image/source versions remain synchronized. |
| 15–16 SDKs | Both bundled canonical configuration schemas updated. Wrappers continue to use matching CLI and preserve engine results; no language-specific driver implementation added. |
| 17 models | Model/provider configuration and active inference remain independent. SQL query probes likewise require functional opt-in. Historical v0.15.0 scope remains accurate. |
| 19 vectors | Still planned. Database login/SELECT 1 does not implement pgvector, collection/index discovery or retrieval health. |
| 20 frameworks | Future integrations must keep data health separate from agent and communication-path evidence. |
| 21 CI/CD | Versioned disposable four-service native matrix adds healthy/authentication-denial checks, alongside schema, race, platform, SDK and deployment jobs. |
| 22 observability | Existing health dimensions and output only; no query metrics exporter or database content collection. |
| 23 plugins | Trusted compiled drivers only, no dynamic plugin or tenant isolation claim. |
| 24 conformance | Add schema rejection and native protocol safety regression coverage; full conformance remains planned. |
| 25 interoperability | Controlled PostgreSQL 17.6, MySQL 8.4.6, Redis 7.4.5 and MongoDB 8.0.13 matrix; cloud products, version matrices, TLS database deployment matrices and cluster/failover compatibility remain future work. |
| 26 hardening | Verified TLS default, no insecure fallback, fixed query constants, environment passwords, context-bound connections and no native operation retries. Go source minimum advances to 1.26 for security-patched dependencies; minimum-version CI and build guides align. No security audit claim. |

Validation includes Go vet/race, canonical fixtures and examples, SDK schema
parity/type declarations, markdown links and real database healthy/wrong-password
runs. Full distribution and disposable Kubernetes runtime validation remain
required CI/release gates. See the [database guide](database-adapter.md) for
limitations and [release notes](releases/v0.16.0.md) for synchronized artifacts.

Current cross-phase follow-up: [Phase 19 alignment](phase-19-alignment.md)
adds Qdrant/Pinecone retrieval dependency checks. Historical release scope above
is unchanged; database login remains distinct from vector retrieval.

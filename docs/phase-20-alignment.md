# Phase 20 impact and alignment review

Implemented in source: optional LangGraph/FastAPI SDK integration, allowlisted
peer callbacks and typed health consumer nodes. Other framework products remain
planned. Release publication is tracked separately from source completion.

| Phases | Assessment and completed alignment |
|---|---|
| 0–3 foundation/spec/core/CLI | Existing v1 configuration, result and exit codes preserved; no new target type or schema. SDK uses CLI-owned aggregation. |
| 4–6 HTTP/MCP/A2A | Existing transport contracts unchanged; peer callbacks use application-selected transports, not discovered URLs. A2A endpoint health remains distinct from communication evidence. |
| 7 agent health | Supported resource uses existing HEAD/GET/POST contract, explicit safe opt-in and bounded task parsing. Unknown peers fail closed. |
| 8–10 gateways/router/graph | Preserve separate backend, route and peer evidence; no fabricated route verification. Existing graph examples remain canonical. |
| 11 AHP | No wire changes; application health resource is distinct from AHP result serving. |
| 12–14 distribution/Kubernetes | Optional Python dependencies only; CLI/container and chart behavior unchanged. Passive readiness and explicit functional repetition documented. Historical chart tags retained. |
| 15–16 SDKs | Ship integration in Python wheel/sdist; optional extras avoid framework imports for core users. Node ESM cancellation/deadline contract unchanged; native JS framework plugins deferred. |
| 17–19 supporting adapters | Model/database/vector results remain independent dependencies; no inference, SQL or vector query is implicitly enabled. |
| 21 CI/CD | Reuse installed-SDK framework tests; future deployment gates must preserve unhealthy reports and tool failures distinctly. |
| 22 observability | Future exports must preserve direct/peer/path distinction; no raw task/error export added. |
| 23 plugins | Trusted in-process callbacks are not sandboxed plugins or tenant isolation. |
| 24 conformance | Contract tests establish a baseline, not certification. |
| 25 interoperability | Pin exact framework versions and record controlled scope; broader product matrix remains required. |
| 26 hardening | Authentication, process limits and cooperative cancellation remain application responsibilities. |

Validation: compiled graph and FastAPI contract tests, selected-peer dispatch,
strict boolean success, sanitized errors, timeout, consumer error/cancellation
propagation; existing full repository checks and release distribution gates.
See the [integration guide](framework-integrations.md).

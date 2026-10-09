# Reference application middleware

The agent adapter requires an application contract; it does not automatically
inspect arbitrary agent frameworks. These source examples implement the existing
[health document and safe-task contract](agent-adapter.md). They are not published
SDKs. Integrate them with your application authentication and resource limits.

| Runtime | Source | Integration |
|---|---|---|
| Go net/http | [Handler](../examples/reference/go/handler.go) | Mount `Handler(snapshot, probe)` at `/health`; use nil probe for passive-only health |
| Python/FastAPI | [Router](../examples/reference/python/agenthealth_ref.py) | `app.include_router(router(snapshot, probe))`; probe must be async |
| TypeScript/Express | [Router](../examples/reference/typescript/agenthealth-ref.ts) | `app.use(authMiddleware, agentHealth(snapshot, probe))` |
| LangGraph | [Safe graph bridge](../examples/reference/python/langgraph_ref.py) | Pass a dedicated compiled graph with async `ainvoke`, producing `healthy: bool` |

`snapshot` returns a nonempty name, current `live` and `ready` booleans, and unique
nonempty capability/dependency names. Advertised dependency names must match
explicit CLI configuration. Do not return URLs, credentials or arbitrary runtime
state. Go normalizes empty slices; Python/TypeScript callers supply arrays.

The optional probe receives `{safe: true, text, downstream?}`. Implement only
allowlisted health operations. Return `completed` and `success`; when a downstream
selector is supplied, return the matching downstream only after actually
performing the authorized peer exchange. Never echo the selector as fabricated
proof. The LangGraph bridge rejects downstream probes until you implement that
exchange; use a dedicated safe graph, not an unrestricted production agent.

Reference POST handlers bound request bodies and set a one-second probe deadline.
Cancellation is cooperative: Go cannot terminate uncooperative callbacks, Python
callbacks can suppress cancellation, and JavaScript work can ignore AbortSignal.
Apply process limits, ingress timeouts and authentication. These examples do not
provide tenant isolation, automatic authorization or framework product certification.

## Run the framework examples

Install the pinned Python requirements in a dedicated virtual environment, then
run `python -m pytest examples/reference/python -q`. For Express, run
`npm ci --prefix examples/reference/typescript`,
`npm --prefix examples/reference/typescript run typecheck` and
`npm --prefix examples/reference/typescript test`. Go contract tests run with
`go test ./examples/reference/go`. CI runs these checks explicitly.

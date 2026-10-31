# Agent runtime integrations

Released in v0.18.0, Phase 20 provides a supported Python SDK integration for LangGraph and FastAPI.
Install `agenthealth[langgraph]` with the matching verified CLI for consumer nodes.
The server handler does not require a CLI. Semantic Kernel, CrewAI, AutoGen and
JavaScript framework plugins remain planned; Express and Go remain reference middleware.

## Expose health

```python
from agenthealth.integrations.langgraph import langgraph_router
app.include_router(langgraph_router(safe_graph, snapshot, peers={"peer": probe_peer}))
```

`safe_graph` is a dedicated compiled LangGraph with async `ainvoke`. Its state
accepts `health_probe` and returns `healthy: bool`. Use no production tools,
checkpointer, thread memory, model inference or mutating nodes in this graph.
`snapshot()` returns the Phase 7 name/live/ready/capabilities/dependencies document;
see the [contract](agent-adapter.md). Passive HEAD/GET never invoke the graph.
The optional `peers` mapping copies operator-owned selectors to async callbacks.
Each callback receives text, must actually contact that selected peer through the
application's authenticated, read-only transport, and return exactly `True` on
success. Unknown selectors fail without executing a callback; they never resolve
URLs from input. No automatic discovery or retries occur. The mapping is trusted
application code, not independent proof of communication.

POST requires explicit `safe: true`; text is bounded to 64 KiB UTF-8, selectors
to 1024 Unicode characters, envelopes to 512 KiB. The whole probe has a one-second
cooperative deadline. Exceptions/timeouts produce generic 503 responses without
backend details. Cancellation cannot isolate uncooperative callbacks. Wrap routes
with application authentication, ingress limits and process resource limits.

## Consume health in a graph

```python
from agenthealth import Client
from agenthealth.integrations.langgraph import health_node
builder.add_node("dependencies", health_node(Client(timeout=5), "agenthealth.yaml"))
```

Declare an `agenthealth: CheckReport` field in the graph state. The node returns
the complete typed SDK report, including unhealthy results and dependency trees;
only SDK/tool failures raise exceptions. Configuration is fixed by the operator,
not read from graph state. Client deadlines are seconds and native async
cancellation stops CLI work. Store reports only in trusted state; application
serialization/checkpoint policy is the caller's responsibility. There is no
JavaScript API change: the existing Node SDK retains AbortSignal and millisecond
deadlines. [JavaScript examples](../examples/javascript-sdk/README.md) remain valid.

## Evidence and deployment

Keep direct agent, composite peer endpoint, and first-to-peer path targets
separate. Models, tools, gateways, routers, databases and vector stores remain
supporting dependencies with independent policies. Vector health does not prove
retrieval relevance. AHP serving remains separate from this application endpoint.
Use the [agent configuration](../examples/agent-check/agenthealth.yaml) and
[graph example](../examples/graph-check/README.md) without schema migration.
Kubernetes may deploy the application and existing CLI sidecar independently;
readiness should stay passive. Scheduled functional checks repeat authorized
work. No framework runtime is bundled into the CLI image or Helm chart.

## Validation

[Runnable example](../examples/runtime-check/README.md) and
[contract tests](../examples/reference/python/) use FastAPI 0.143.0,
LangGraph 1.2.14 and httpx 0.28.1. CI installs the SDK before testing the compiled
graph, safe task bounds, selected peer callback, failures and SDK consumer.
This controlled baseline does not certify arbitrary production agents or complete
Phase 25. See [alignment](phase-20-alignment.md) and the upstream
[LangGraph graph API](https://reference.langchain.com/python/langgraph/graph).

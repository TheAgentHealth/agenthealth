# LangGraph runtime health

Use a virtual environment and the matching CLI. From the repository root:

```bash
pip install -e './sdk/python[langgraph]' uvicorn
uvicorn app:app --app-dir examples/runtime-check --port 8080
```

In another terminal:

```bash
agenthealth check examples/runtime-check/agenthealth.yaml --format json
```

Expected: HEALTHY, exit 0. This dedicated graph performs no production work.
The configuration explicitly opts into a single functional POST; removing it
keeps passive metadata checks. The local example has no authentication and binds
to loopback by default. Add application authentication before exposing it.

For separate direct/composite/path evidence, start with the existing
[agent example](../agent-check/README.md), adding allowlisted `peers` callbacks
only when each performs the selected authorized peer exchange. HTTP peer success
and failure are exercised in [runtime tests](../reference/python/test_runtime.py).
The [guide](../../docs/framework-integrations.md) covers consumer nodes,
deadlines, cancellation, Kubernetes and remaining framework scope.

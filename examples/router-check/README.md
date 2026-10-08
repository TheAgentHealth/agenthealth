# Router, backend and route checks

[agenthealth.yaml](agenthealth.yaml) connects a direct agent to a router readiness
signal, a direct A2A backend, an A2A routed path and an explicitly opted-in
read-only route probe. Replace every endpoint with your deployment's configured
interfaces. The sample ports/paths are illustrative, not product defaults.
The functional route must be non-destructive and return the fixed `ready` marker.

```bash
agenthealth doctor examples/router-check/agenthealth.yaml
```

Router, direct-backend and routed-path checks preserve separate evidence.
The CLI route probe does not prove communication from the direct agent runtime.
See the [router guide](../../docs/agent-router.md) for safety and limitations.

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

## Phase 10 compatibility

This nested example remains valid. To share a backend across paths, give that node an explicit `id` and use `ref` edges. Each communication path keeps a distinct node and its own checks; edge policy remains independent. See the [graph example](../graph-check/README.md). Graph fields require v0.8.0 or newer; v0.7.0 binaries reject them.

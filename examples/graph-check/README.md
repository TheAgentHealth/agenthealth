# Shared dependency graph

The [configuration](agenthealth.yaml) checks a first agent, one shared peer, a communication probe and a router.
Run compatible handlers on localhost ports 8080–8082, then run:

```bash
agenthealth check examples/graph-check/agenthealth.yaml --format json
```

The first agent handler must implement the [agent interface](../../docs/agent-adapter.md),
including an explicitly safe probe that contacts the selected peer.
The router endpoint supplies a read-only HTTP health signal.
See the [graph guide](../../docs/dependency-graph.md) for policies, budgets and evidence semantics.

## Kubernetes compatibility

Mount this graph only after replacing localhost addresses and reviewing its active communication probe. Shared-node execution and independent edge policies remain unchanged.
See [deployment scenarios](../kubernetes/README.md) and the
[deployment guide](../../docs/kubernetes.md).

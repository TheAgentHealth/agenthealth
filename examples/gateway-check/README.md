# Gateway checks

The [integration guide](../../docs/agentgateway.md) explains the configured
direct agent, gateway readiness signal, direct downstream and proxied A2A
path. Replace the local endpoints with your deployment addresses. No servers
are started by this example; functional probes require explicit safe configuration.

## Phase 10 compatibility

This nested example remains valid. To share a backend across paths, give that node an explicit `id` and use `ref` edges. Each communication path keeps a distinct node and its own checks; edge policy remains independent. See the [graph example](../graph-check/README.md). Graph fields require v0.8.0 or newer; v0.7.0 binaries reject them.

## Kubernetes compatibility

Replace localhost with your in-cluster addresses when mounting this configuration. Gateway and routed peer checks retain separate evidence; the Kubernetes fixture is not product certification.
See [deployment scenarios](../kubernetes/README.md) and the
[deployment guide](../../docs/kubernetes.md).

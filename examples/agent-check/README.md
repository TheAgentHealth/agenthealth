# Agent, peer and communication-path checks

This Phase 7 configuration is a template for your running services. Phase 7
requires v0.5.0 or newer; see [installation](../../docs/installation.md).

Before running [agenthealth.yaml](agenthealth.yaml):

- Expose the first agent's HEAD/GET health resource at port 8080, implementing
  the [agent health document](../../docs/agent-adapter.md#passive-interface).
- Expose the A2A peer at port 8081 with an agent card advertising skill `answer`.
- Implement a safe POST probe handler in the first agent. When it receives
  `downstream: peer`, it must contact that peer using its configured direct or
  A2A integration and report completed success/failure with the same peer ID.
- Replace the optional model-health HTTP URL with your supporting service's
  health endpoint. This checks HTTP health, not model inference.

AgentHealth implements sending and evaluating probes. This example does not
start servers or install the receiving handler. Peer endpoint health alone
cannot prove the first agent can communicate with the peer.

```bash
go run ./cmd/agenthealth check examples/agent-check/agenthealth.yaml --format json
go run ./cmd/agenthealth doctor examples/agent-check/agenthealth.yaml
```

The result tree separates `first`, `peer`, `first-to-peer`, and `model-health`.
The peer and path are critical; the model-health dependency is optional.
The path probe is explicitly active and executes once. Configure credentials
independently for each target when your services require authentication.

## Phase 10 compatibility

This nested example remains valid. To share a backend across paths, give that node an explicit `id` and use `ref` edges. Each communication path keeps a distinct node and its own checks; edge policy remains independent. See the [graph example](../graph-check/README.md). Graph fields require v0.8.0 or newer; v0.7.0 binaries reject them.

# Agentgateway integration

`gateway` targets check explicitly configured read-only HTTP health, liveness
or readiness endpoints using GET. A 2xx response passes protocol; 401/403
is MISCONFIGURED, other unexpected statuses are UNHEALTHY. Reachability alone
records connectivity regardless of status. HTTP status/header expectations,
bearer environment references, deadlines, latency thresholds, verified TLS,
redirect rejection and allowlisted transport diagnostics use the HTTP adapter
contract. Functional body matching is bounded and requires explicit opt-in;
active checks never retry. Passive GET is appropriate only for read-only signals.

```bash
agenthealth ping gateway http://localhost:15021/healthz/ready
agenthealth doctor examples/gateway-check/agenthealth.yaml
```

## Gateway, backend and path evidence

Configure separate named dependencies for direct backends and gateway-facing
HTTP, MCP or A2A routes. Those protocol adapters retain their own credentials,
expectations and safe functional probes. A direct agent can include this gateway
as a dependency and use the Phase 7 safe downstream handler to test communication
from its own runtime. An external CLI route probe does not prove that runtime path.

The gateway's protocol check describes only its configured health signal.
Its overall status also includes dependencies: critical failures propagate and
optional failures degrade. Backends execute even when the gateway fails.
Use separate gateway targets for separate health/liveness/readiness signals;
no paths or ports are inferred. Names identify evidence and should avoid sensitive
topology; credentials and diagnostic details follow the shared redaction contract.

## Supported interfaces and limits

This integration supports Agentgateway deployments exposing read-only HTTP
signals, MCP Streamable HTTP, or A2A 1.0 JSON-RPC (explicit 0.3.0 compatibility).
It does not bind to a product version or parse version-specific admin inventory.
The readiness example uses `/healthz/ready` on port 15021 as described by the
[upstream project](https://github.com/agentgateway/agentgateway/issues/1567).
Verify the endpoint for your deployment. Liveness is checked only when exposed
and explicitly configured. Backend availability is direct protocol evidence,
not gateway load-balancer membership, eviction state or route configuration.

A failed proxied request is path evidence; without independent health signals
its cause is inconclusive. No automatic discovery, admin configuration dumping,
model generation or mutating health operation is supported. Real product-version
interoperability testing remains Phase 25 work; local tests exercise HTTP behavior
and independent failure evidence.

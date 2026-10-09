# Serving AHP

Use v0.9.0 or newer. The example path requires a v0.9.0 or newer repository
checkout and your application running on port 9000. In one terminal run:

```bash
agenthealth serve examples/ahp-check/agenthealth.yaml
```

In a second terminal query:

```bash
curl http://127.0.0.1:8080/ready
curl http://127.0.0.1:8080/live
```

Set AHP_TOKEN to a private random token of at least 16 characters through your
secret manager or environment. Bearer-authorized GET requests to
`/health/dependencies` and `/health/capabilities` return detailed results.
To enable authenticated endpoints, stop the server and restart it in the first
terminal after setting AHP_TOKEN:

```bash
agenthealth serve examples/ahp-check/agenthealth.yaml --listen 127.0.0.1:8080 --token-env AHP_TOKEN
```

Without a token these endpoints remain inaccessible. Public `/health` and
`/ready` disclose only aggregate status and observation time.

The initial health response is UNKNOWN/503 until the first run completes.
Checks repeat 30 seconds after each run and expire after two minutes. Explicit
active probes also repeat; use a passive configuration for routine monitoring.
SIGINT/SIGTERM cancel checks and gracefully shut down HTTP within five seconds.
Successful shutdown returns 0; invocation, binding and server failures return 6.
A health failure does not exit the serving process.

Remote deployments should terminate HTTPS and enforce rate limits at a proxy.
See the [protocol](../spec/protocol.md) and [impact review](rfcs/phase-11-ahp.md).
This experimental implementation is included in v0.9.0; v0.8.0 does not include it.

## Kubernetes serving

[Phase 14 scenarios](../examples/kubernetes/README.md) use this server in a
Deployment and conventional sidecar. `/ready` follows snapshot freshness and
aggregate readiness; `/live` drives process liveness/startup. Unlike CLI exec
probes, AHP readiness can accept DEGRADED. Refresh timing, protected endpoint
authorization and graceful shutdown retain this guide's existing behavior.
See the [deployment guide](kubernetes.md) for Secrets and network exposure.

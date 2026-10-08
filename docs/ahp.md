# Serving AHP

Build the current source and run:

```bash
agenthealth serve examples/ahp-check/agenthealth.yaml
agenthealth serve examples/ahp-check/agenthealth.yaml --listen 127.0.0.1:8080 --token-env AHP_TOKEN
curl http://127.0.0.1:8080/ready
curl http://127.0.0.1:8080/live
```

Set AHP_TOKEN to a private random token of at least 16 characters through your
secret manager or environment. Bearer-authorized GET requests to
`/health/dependencies` and `/health/capabilities` return detailed results.
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
This experimental implementation is available in source; v0.8.0 does not include it.

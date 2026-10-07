# HTTP health adapter

The HTTP adapter supports `http` and `api` targets. Default checks use HEAD: connectivity, authentication, HTTP status/header expectations, configuration, and measured reachability latency. A default response must be 2xx to be healthy. A reachable 404 or 500 is `UNHEALTHY`; 401/403 is `MISCONFIGURED`. Explicit `checks: [reachability]` measures connectivity alone, independent of status.

```bash
./agenthealth ping http https://example.com
./agenthealth doctor examples/core-check/agenthealth.yaml
```

The example configuration requires `http://localhost:8080/health` to return 200. To try it locally, create a temporary directory with a `health` file and serve that directory:

```bash
mkdir -p /tmp/agenthealth-demo
echo ready > /tmp/agenthealth-demo/health
python3 -m http.server 8080 --directory /tmp/agenthealth-demo
```

In another terminal, run doctor. Removing the `health` file makes the protocol check fail with `UNHEALTHY` while reachability remains healthy.

## Response expectations

Use `http.expected_status` to accept specific statuses and `http.headers` to require response header values. Header names are case-insensitive and values match exactly. For example:

<!-- spec-example: configuration -->
```yaml
version: v1
targets:
  - name: service
    type: http
    endpoint: https://example.com/health
    http:
      expected_status: [200, 204]
      headers:
        X-Service-Ready: 'true'
```

Body inspection uses GET and requires `checks` to include `functional`. Use `http.body_contains` for literal text matching and `http.max_body_bytes` to cap the body (default 64 KiB, maximum 1 MiB). Body matching requires the complete body to fit within the limit; oversized responses are inconclusive. GET probes remain non-destructive and never retry. See the [configuration specification](../spec/configuration.md#http-response-expectations-phase-4-draft) for a full example.

## Transport diagnostics

Reachability traces the actual HTTP request using Go's [HTTP tracing API](https://pkg.go.dev/net/http/httptrace). It records DNS resolution, TCP connection, TLS handshake/verification, and HTTP response headers. Only observed stages appear. IP literals do not require DNS, plain HTTP has no TLS, and connection reuse can omit earlier stages. With proxy settings, these observations describe the route used by the HTTP transport, which may include a proxy rather than direct target connectivity.

Transport stages appear under `checks.reachability.steps` in JSON/YAML and beneath reachability in terminal output. `HEALTHY` means a stage completed, `UNREACHABLE` means it failed, and `UNKNOWN` means it had started but had not completed at the observation boundary. Multiple connection address attempts can occur; a successful connection takes precedence over failed attempts. Retry diagnostics describe the final attempt.

Connectivity latency measures the successful HEAD reachability attempt, including observed DNS/TCP/TLS work, rather than GET body reading or all checks combined. Threshold violations are `DEGRADED`. Per-check deadlines cover network work and body reading; retries apply only to passive failures before any response bytes arrive.

## Safety

TLS certificate and hostname verification remain mandatory. Redirects are never followed, and credentials use environment references (`auth.bearer_env`). Response bodies, response headers, expected values, addresses, and raw transport errors are not emitted. HEAD checks do not fall back automatically to GET. A service that rejects HEAD requires an explicit check policy.

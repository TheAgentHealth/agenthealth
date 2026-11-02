# Observability snapshots and explicit delivery

Use a matching CLI built from this source (Phase 22). Run from the repository
root. Existing HTTP/graph configurations work without new configuration fields.

```bash
# Keep unhealthy evidence; fail only on a tool error before attempting delivery.
health_code=0
agenthealth check examples/core-check/agenthealth.yaml --format webhook > snapshot.json || health_code=$?
if [ "$health_code" -gt 5 ]; then exit "$health_code"; fi
# Set WEBHOOK_URL to an approved HTTPS receiver. curl verifies TLS, does not
# follow redirects, makes one bounded request and does not print response bodies.
http_code=$(curl --silent --show-error --fail --output /dev/null --proto '=https' \
  --max-time 10 --max-redirs 0 -H 'Content-Type: application/json' \
  --write-out '%{http_code}' --data-binary @snapshot.json "$WEBHOOK_URL") || exit 6
case "$http_code" in 2??) ;; *) exit 6 ;; esac
exit "$health_code"
```

For OTLP, select `--format otlp` and send the generated file to an approved
HTTPS collector's `/v1/logs` endpoint using the same request. The receiver must
accept OTLP/HTTP JSON logs; webhook receivers must accept AgentHealth's v1
envelope. Configure collector authentication through your organization's secret
mechanism; never embed tokens in committed examples or health configuration.
The relay owns delivery errors and response handling. A successful HTTP response
does not establish durable retention; inspect collector acceptance separately.

For an existing node exporter textfile directory on the same filesystem:

```bash
health_code=0
# Use a unique temporary file so a scraper never sees a partial snapshot.
snapshot=$(mktemp /var/lib/node_exporter/textfile_collector/agenthealth.XXXXXX)
agenthealth check examples/core-check/agenthealth.yaml --format prometheus > "$snapshot" || health_code=$?
if [ "$health_code" -gt 5 ]; then rm -f "$snapshot"; exit "$health_code"; fi
mv "$snapshot" /var/lib/node_exporter/textfile_collector/agenthealth.prom
exit "$health_code"
```

Schedule with your existing scheduler; monitor freshness separately and remove
stale files when retiring a configuration. For structured logs, select
`--format logs` and configure your collector to parse JSON Lines from stdout.
See [mapping, redaction and limitations](../../docs/observability.md).

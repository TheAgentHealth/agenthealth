# Observability export

Phase 22 provides finite snapshots through `ping`, `check` and `doctor`:

| Format | Output | Delivery |
|---|---|---|
| `prometheus` | Prometheus text 0.0.4 gauges | Atomic textfile publication to an existing exporter |
| `logs` | One reduced JSON object per line | Existing log collector reads stdout or a file |
| `otlp` | OTLP/HTTP JSON `ExportLogsServiceRequest` | POST with `application/json` to a collector's `/v1/logs` |
| `webhook` | JSON envelope with `export_version: v1` and `records` | Explicit HTTPS POST to a compatible receiver |

```bash
agenthealth check examples/core-check/agenthealth.yaml --format prometheus
agenthealth check examples/graph-check/agenthealth.yaml --format logs
agenthealth check examples/core-check/agenthealth.yaml --format otlp
agenthealth check examples/core-check/agenthealth.yaml --format webhook
```

Health exit codes remain 0–5; invalid configuration, invalid output and write
failure use 6. A nonzero health code still produces a valid snapshot. SDKs
continue to use canonical JSON. Exports are reduced projections, not replacements
for the AHS result schema. See the [delivery example](../examples/observability/README.md).

## Evidence and privacy

Each record has `kind` (`node`, `check` or `edge`), `node`, target `type` and
`status`. Check records add `dimension`; node records include `latency_ms` when
measured. Edge records add a unique snapshot-local `edge`, `parent`, configured
`relationship` and effective `critical` (default true). Empty relationship means
unspecified. Node health is not replaced by its contribution to a parent.
All six statuses remain distinct. Transport steps and diagnostic codes are
omitted along with arbitrary messages.

Explicit graph IDs deduplicate node/check evidence. Every incoming edge retains
its own policy; a shared node's outgoing edges export once. Nodes without IDs
remain distinct. The exporter expects coherent engine-generated snapshots;
it does not reconcile conflicting caller-supplied results with the same ID.
Aliases (`node-1`, `edge-1`) are assigned in configuration traversal order and
reset for every snapshot. Do not correlate them across reordered configurations.
Direct and composite agents both retain type `agent`; downstream/path relationships
carry the available distinction. No role is inferred from names.

Configured names, IDs, URLs, credentials and diagnostic text are excluded.
Topology shape, counts, target types, relationships, status and timing remain
visible: this is data minimization, not topology anonymization or tenant isolation.
Restrict collector/file access accordingly. No original identifier opt-in exists.

## Metric mapping

`agenthealth_status` is a one-hot gauge for each of the six `status` values.
Other labels are `kind`, `node`, `parent`, `type`, `dimension`, `relationship`,
`critical` and `edge`; inapplicable labels are empty strings. Edge labels prevent
repeated relationships from producing duplicate series. `agenthealth_latency_seconds`
converts measured node latency from milliseconds, with `node` and `type` labels.
Missing latency produces no sample. No counter, histogram or fabricated zero
latency is emitted. Cardinality grows with configured nodes, checks and edges;
use bounded configurations and replace the whole textfile on every run.

OTLP emits logs with scope `agenthealth`, a shared `observedTimeUnixNano`, and
JSON-encoded reduced records in `body.stringValue`. It does not fabricate trace
IDs, spans, resource identity, or severity mapping. The timestamp records export
time, not the start of each check. Webhook records use the same evidence mapping.

The formats follow the [Prometheus exposition specification](https://prometheus.io/docs/instrumenting/exposition_formats/)
and [OTLP JSON specification](https://opentelemetry.io/docs/specs/otlp/).
Collector ingestion and hosted-provider compatibility require separate measured
interoperability evidence. This baseline has no scrape server, remote-write,
gRPC exporter, durable queue, authentication manager, automatic delivery or retry.
AHP endpoints, health schemas and Kubernetes probes are unchanged.

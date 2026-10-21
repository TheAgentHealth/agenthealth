# AgentHealth Helm chart

Chart 0.15.0 pins CLI v0.15.0 and defaults to serving experimental AHP.
`mode` accepts `serve`, `job`, or `cronjob`. Supply inline YAML with
`--set-file config=<path>` or an existing ConfigMap with key `config.yaml`.
`image.digest` overrides the explicit image tag.

The chart and CLI now share the software release version and `v*` source tag.
The release workflow attaches the chart to the common GitHub Release and
publishes `oci://ghcr.io/theagenthealth/charts/agenthealth`. v0.14.0 is published and verified; see the
[verification record](../../../docs/releases/verification-v0.14.0.md).
See [distribution](../../../docs/distribution.md),
[the guide](../../../docs/kubernetes.md) and
[scenarios](../../../examples/kubernetes/README.md).

Current source targets synchronized v0.15.0; the linked v0.14.0 verification
record is historical. See [Phase 17 notes](../../../docs/releases/v0.15.0.md).

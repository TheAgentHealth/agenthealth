# AgentHealth Helm chart

Chart 0.12.0 pins CLI v0.12.0 and defaults to serving experimental AHP.
`mode` accepts `serve`, `job`, or `cronjob`. Supply inline YAML with
`--set-file config=<path>` or an existing ConfigMap with key `config.yaml`.
`image.digest` overrides the explicit image tag.

The chart and CLI now share the software release version and `v*` source tag.
The release workflow attaches the chart to the common GitHub Release and
publishes `oci://ghcr.io/theagenthealth/charts/agenthealth`. v0.12.0 is public and verified; see
[publication verification](../../../docs/releases/verification-v0.12.0.md).
See [distribution](../../../docs/distribution.md),
[the guide](../../../docs/kubernetes.md) and
[scenarios](../../../examples/kubernetes/README.md).

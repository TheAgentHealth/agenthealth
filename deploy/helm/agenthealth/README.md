# AgentHealth Helm chart

Chart 0.12.1 pins CLI v0.12.1 and defaults to serving experimental AHP.
`mode` accepts `serve`, `job`, or `cronjob`. Supply inline YAML with
`--set-file config=<path>` or an existing ConfigMap with key `config.yaml`.
`image.digest` overrides the explicit image tag.

The chart and CLI now share the software release version and `v*` source tag.
The release workflow attaches the chart to the common GitHub Release and
publishes `oci://ghcr.io/theagenthealth/charts/agenthealth`. Publication requires
an authorized tag; source preparation alone does not make v0.12.1 available.
See [distribution](../../../docs/distribution.md),
[the guide](../../../docs/kubernetes.md) and
[scenarios](../../../examples/kubernetes/README.md).

# AgentHealth Helm chart

Chart 0.13.2 pins CLI v0.13.2 and defaults to serving experimental AHP.
`mode` accepts `serve`, `job`, or `cronjob`. Supply inline YAML with
`--set-file config=<path>` or an existing ConfigMap with key `config.yaml`.
`image.digest` overrides the explicit image tag.

The chart and CLI now share the software release version and `v*` source tag.
The release workflow attaches the chart to the common GitHub Release and
publishes `oci://ghcr.io/theagenthealth/charts/agenthealth`. Source now targets v0.13.2; the previous v0.12.0 publication is recorded in
[its verification record](../../../docs/releases/verification-v0.12.0.md).
See [distribution](../../../docs/distribution.md),
[the guide](../../../docs/kubernetes.md) and
[scenarios](../../../examples/kubernetes/README.md).

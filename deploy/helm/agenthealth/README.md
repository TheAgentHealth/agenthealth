# AgentHealth Helm chart

Chart 0.1.0 pins CLI v0.11.1 and defaults to serving experimental AHP.
`mode` accepts `serve`, `job`, or `cronjob`. Configure inline YAML through
`--set-file config=<path>`, or mount an `existingConfigMap` with key
`config.yaml`. `image.digest` overrides the explicit image tag.

See the [guide](../../../docs/kubernetes.md) and
[scenarios](../../../examples/kubernetes/README.md) for both registry references,
installation, Secrets, limitations and validation. Chart 0.1.0 was published
under historical helm-v0.1.0; the existing workflow still uses that trigger.
The next release must synchronize chart/CLI version 0.12.0 and publish the
chart to GHCR and GitHub Releases. See the
[distribution plan](../../../docs/distribution.md); OCI publishing is not implemented yet.

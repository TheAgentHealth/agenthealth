# AgentHealth Helm chart 0.1.0

Prepared component release notes; this chart has not been published yet.

Chart 0.1.0 pins CLI `v0.11.0` (`appVersion`) and supports AHP Deployment,
one-shot Job and CronJob modes, read-only ConfigMap configuration and existing
Secret references. Kubernetes runtime scenarios were validated on kind 1.32.2
with Helm 3.17.3 on Linux AMD64. The declared manifest floor is Kubernetes 1.29;
other cluster versions have not been runtime validated.

Both public CLI image references were verified on October 8, 2026:

- Default: `ghcr.io/theagenthealth/agenthealth:v0.11.0`
- Optional mirror: `docker.io/theagenthealth/agenthealth:v0.11.0`

Kustomize source examples and an optional application-image exec-probe pattern
are also included. Selected operational patterns credit Sharath K (`sharath568`)
from [PR #15](https://github.com/TheAgentHealth/agenthealth/pull/15).

Source paths at the intended component release revision:

- [Manifests](https://github.com/TheAgentHealth/agenthealth/tree/helm-v0.1.0/deploy/kubernetes)
- [Chart source](https://github.com/TheAgentHealth/agenthealth/tree/helm-v0.1.0/deploy/helm/agenthealth)
- [Usage examples](https://github.com/TheAgentHealth/agenthealth/tree/helm-v0.1.0/examples/kubernetes)
- [Packaged chart](https://github.com/TheAgentHealth/agenthealth/releases/download/helm-v0.1.0/agenthealth-0.1.0.tgz)

These tag-pinned links become available only after the authorized chart release.
Verify that they resolve before announcing publication. This release does not
republish the CLI or container images.

After publication, download the chart and `checksums.txt` from the component
release, verify the checksum, and use a reviewed configuration:

```bash
sha256sum --check checksums.txt
helm upgrade --install agenthealth ./agenthealth-0.1.0.tgz \
  --set-file config=./agenthealth.yaml
```

For a runnable fixture scenario, check out the component tag and follow
[the usage examples](../../examples/kubernetes/README.md). Set
`image.repository=docker.io/theagenthealth/agenthealth` to choose the mirror.

Only CLI exit 0 passes exec probes and validation Jobs. AHP readiness uses its
existing aggregate contract and snapshot interval. The chart exposes no Service
or Ingress; configure network isolation and HTTPS when needed. Completed Jobs
need a new release name for reruns or a deliberate workload deletion for updates.
No CRDs, operator, admission policy, real product interoperability certification,
or chart provenance signatures are included.

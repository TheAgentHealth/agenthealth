# Kubernetes integration

Phase 14 is implemented in source. The chart is version **0.1.0**, with
`appVersion: v0.11.0`. Chart publication is independent of CLI releases.
The deployment files require Kubernetes 1.29 or newer; the runtime smoke suite
uses Kubernetes 1.32.2. Only versions actually exercised are listed in the
[alignment audit](phase-14-alignment.md).

Start with the [runnable scenarios](../examples/kubernetes/README.md),
[manifests](../deploy/kubernetes/), or [Helm chart](../deploy/helm/agenthealth/).
They consume the existing executable without changing AHS, AHP or CLI contracts.

## Images and configuration

Both public references for the pinned CLI version are:

- GHCR: `ghcr.io/theagenthealth/agenthealth:v0.11.0`
- Docker Hub: `docker.io/theagenthealth/agenthealth:v0.11.0`

Anonymous registry manifest retrieval verified both on October 8, 2026.
The chart defaults to GHCR. Set `image.repository` to the Docker Hub reference
without its tag to select the mirror. For plain YAML, replace the full image
reference in each AgentHealth container, including init containers. Registry
index digests differ; use the digest verified for the chosen registry.

```bash
helm lint deploy/helm/agenthealth --strict
helm template agenthealth deploy/helm/agenthealth
helm upgrade --install agenthealth deploy/helm/agenthealth \
  --set image.repository=docker.io/theagenthealth/agenthealth \
  --set image.tag=v0.11.0 \
  --set-file config=examples/kubernetes/agenthealth.yaml
```

Review configuration before applying it. ConfigMaps contain trusted executable
configuration, never credential values. Replace the demonstration endpoints with
your application's read-only health resources. Configuration is mounted read-only;
chart changes to inline configuration trigger a rollout. An `existingConfigMap`
can supply `config.yaml`; changes to that external map need an explicit rollout.

## Probe and workload behavior

The serving Deployment and conventional sidecar expose experimental AHP on
port 8081. `/ready` gates traffic using current dependency health; `/live` checks
the monitor process independently. Startup uses `/live`, avoiding dependency-driven
restart loops. Initial UNKNOWN and stale snapshots fail readiness. Checks repeat
30 seconds after completion and snapshots expire after two minutes; readiness
changes are therefore delayed by the observation interval. This is not an
instantaneous application probe. See [AHP serving](ahp.md).

The exec example uses the absolute binary path with no shell. The CLI must be
present in the probed container; Kubernetes cannot run a sidecar's executable
inside the application container. Its 10-second probe timeout exceeds the sample
five-second check budget. Only HEALTHY/exit 0 passes. DEGRADED/1 and all other
nonzero exits fail, even when an optional dependency caused degradation. AHP
readiness follows its separately defined aggregate contract, so it can accept
DEGRADED. Choose the behavior deliberately. Kubernetes may restart a container
after repeated startup failures; dependency readiness alone should normally
use readiness rather than startup/liveness. See the
[Kubernetes probe documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/).

Init checks gate startup against an already running external dependency. Never
check an application container that cannot start until the init check finishes.
The engine performs no retries in these configurations; Kubernetes can rerun
failed init containers under the Pod restart policy. Jobs use `restartPolicy:
Never`, `backoffLimit: 0`, and a 30-second deadline. CronJobs additionally forbid
overlap. Each schedule is a new check, not an engine retry. Inspect Job status
and logs rather than treating resource creation as health success.

The sidecar example uses a regular container in a Deployment, so it needs no
native sidecar feature gate. Its readiness contributes to the Pod's readiness;
a separate monitor Deployment cannot gate a different application's Pod.
Sidecar port 8081 must be free in the shared Pod network namespace.

## Kustomize and application-container exec probes

Kustomize is an alternative to Helm. The reusable base is
`deploy/kubernetes/kustomization.yaml`; the runnable fixture overlay is
`examples/kubernetes/kustomization.yaml`. The overlay generates a ConfigMap
from `agenthealth.yaml` and rewrites workload references to its content hash.
Set the kubectl namespace explicitly. ConfigMap changes can change immutable Job
pod templates; delete a completed validation Job deliberately before reapplying.

```bash
kubectl kustomize examples/kubernetes
kubectl -n agenthealth-demo apply -k examples/kubernetes
```

Edit the Kustomize `images` entry's `newName` to choose either registry listed
above and keep `newTag: v0.11.0`. Do not combine these examples with separately
installed same-name resources in the same namespace.

For an application-container exec probe, the optional
[derived-image Dockerfile](../examples/kubernetes/Dockerfile.exec-probe) copies
the binary into an nginx fixture image. In production, replace the final stage
with your application image. Build from the repository root:

```bash
docker build -f examples/kubernetes/Dockerfile.exec-probe \
  -t agenthealth-demo-exec:local .
docker run --rm --entrypoint /usr/local/bin/agenthealth agenthealth-demo-exec:local version
```

Supply `--build-arg AGENTHEALTH_IMAGE=docker.io/theagenthealth/agenthealth:v0.11.0`
to use the mirror. Mount the reviewed configuration and Secret references in
the application container; use the absolute executable path shown in the exec
probe example. Startup should normally check the application's own process
rather than its supporting dependencies.

Shared-dependency readiness affects every application replica that uses the
same policy. A common gateway outage can remove every replica from Service
endpoints. Choose this only when serving without that dependency is undesirable;
otherwise prefer Jobs/CronJobs for evidence. Request load also grows with probe
frequency, target count and replica count.

## Secrets and exposure

Create `agenthealth-credentials` through your secret manager with keys
`agent-token` and, when enabling protected AHP endpoints, `ahp-token` (at least
16 characters). The [Secret Job](../examples/kubernetes/secret-job.yaml) uses
`secretKeyRef` and `auth.bearer_env`. The fixture accepts the synthetic token `fixture-token` at `/auth` and rejects
other values; use your authenticated service and secret manager in production.

```bash
helm upgrade --install agenthealth deploy/helm/agenthealth \
  --set-file config=examples/kubernetes/agenthealth.yaml \
  -f examples/kubernetes/secret-values.yaml
```

The values file references existing Secrets; it never creates or embeds them.
`ahpTokenEnv` must name one of `secretEnv`'s entries. Missing credentials fail
configuration before authenticated checks run. Environment Secret updates require
a rollout. Avoid putting tokens in Helm values, command lines or ConfigMaps.

Pods run without API credentials, as non-root, with a read-only root filesystem,
dropped capabilities, resource bounds and RuntimeDefault seccomp. There is no
Service or Ingress for AHP by default. Binding on `0.0.0.0` permits kubelet probes
and Pod-network access; apply namespace NetworkPolicy for your topology and
terminate HTTPS at a trusted proxy if exposing health outside the cluster.
Public AHP routes disclose aggregate status; detailed routes require a token.

## Validation and packaging

```bash
python3 -m pytest tests/kubernetes -q
python3 scripts/test_kubernetes.py --context kind-agenthealth-phase14
helm package deploy/helm/agenthealth --destination /tmp/agenthealth-chart
```

The smoke script requires an explicitly selected disposable cluster context,
creates a dedicated namespace, and removes that namespace afterwards. It tests
healthy and failed Jobs, AHP and exec readiness outage/recovery without
restarts, init blocking/recovery, sidecar readiness, CronJobs and chart modes.
CI builds and loads the current source image into kind and passes `--image`.
Omit `--image` to exercise the pinned released image. It is not real agent/gateway/router interoperability
certification: the scenario serves deterministic HTTP fixtures.

A chart release uses `helm-v0.1.0`, attaches `agenthealth-0.1.0.tgz` and
`checksums.txt`, and links to manifests, chart source and examples at that exact
tag. See [chart releases](../RELEASING.md#helm-chart-releases). Source changes
and local packaging do not publish a release.

## Contributor credit

Kustomize packaging, the application-image binary-copy pattern, shared-dependency
guidance, current-source image loading, and init outage/recovery scenarios adapt
ideas from [sharath568's PR #15](https://github.com/TheAgentHealth/agenthealth/pull/15)
and [issue #13](https://github.com/TheAgentHealth/agenthealth/issues/13).
They are integrated into the Phase 14 layout, existing exit contracts and
independent Helm release policy. Superseded Phase 12 changes were not imported.

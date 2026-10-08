# Kubernetes integration examples

These manifests run the [AgentHealth container image](../../docs/installation.md#container-image)
to gate and observe a workload's supporting dependencies in Kubernetes
([Phase 14](../../ROADMAP.md#phase-14--kubernetes-integration)). They are
templates: replace the `research-agent` workload, service names and
credentials with your own.

| File | Pattern | Use it to |
|---|---|---|
| [agenthealth.yaml](agenthealth.yaml) | AgentHealth configuration, mounted from a ConfigMap | Declare the MCP server and model gateway the agent depends on |
| [init-container.yaml](init-container.yaml) | Init container | Start the agent only after its dependencies are `HEALTHY` |
| [readiness-probe.yaml](readiness-probe.yaml) | Startup and readiness probes | Remove the agent from Service endpoints while dependencies are unhealthy |
| [deployment-validation-job.yaml](deployment-validation-job.yaml) | Job | Gate a CI/CD pipeline after a rollout |
| [cronjob.yaml](cronjob.yaml) | CronJob | Record periodic JSON health evidence in logs |
| [secret.yaml](secret.yaml) | Secret | Supply `bearer_env` credentials as environment variables |
| [kustomization.yaml](kustomization.yaml) | Kustomize | Generate the ConfigMap and pin the image tag |

```bash
kubectl create namespace agents
kubectl apply -k integrations/kubernetes
kubectl -n agents wait --for=condition=complete --timeout=120s job/agenthealth-validate
```

The checks are passive. Do not add `functional` checks to probes or frequent
CronJobs without considering their cost and side effects.

## Choosing a pattern

- **Init container** is the simplest gate and needs no change to the agent
  image. It checks only at start-up; Kubernetes retries a failing init
  container with back-off and the pod stays in `Init` until it succeeds.
- **Readiness probe** checks continuously. Exec probes run inside the agent
  container, so the agent image must contain the binary. Copy it from the
  release image, as shown at the top of
  [readiness-probe.yaml](readiness-probe.yaml). To use it, replace
  `init-container.yaml` with `readiness-probe.yaml` in
  [kustomization.yaml](kustomization.yaml) so the generated ConfigMap name
  resolves.
- **Job** and **CronJob** do not affect traffic. Use them for deployment
  validation and monitoring.

## Exit codes and probe semantics

Kubernetes treats every non-zero exit as a failed check. AgentHealth exits `0`
only for `HEALTHY`, so `DEGRADED` (exit `1`) also fails a probe, init container
or Job. Mark non-essential dependencies `critical: false` when their failure
should not block readiness. See [exit codes](../../spec/exit-codes.md).

Set `timeoutSeconds` above the worst case for one run: per-target
`timeout_ms` multiplied by attempts, plus `retry_delay_ms` between them.

## Operational cautions

- **Never use dependency checks in a `livenessProbe`.** Restarting the agent
  cannot repair a dependency and turns an outage into a restart loop.
- **Readiness on a shared dependency affects every replica at once.** If the
  model gateway fails, all agent pods become unready and the Service has no
  endpoints. Prefer this when serving without the dependency is worse than
  not serving; otherwise, use the Job or CronJob for visibility instead.
- Keep probe intervals moderate. Every probe run sends requests to each
  dependency, multiplied by the replica count.
- The image runs as UID 65532 with no shell. The manifests use a read-only root
  filesystem and drop all capabilities, which satisfies the `restricted` Pod
  Security Standard for the AgentHealth containers.

## Validation

These manifests were tested on kind (Kubernetes v1.37) with a locally built
image and an nginx stand-in for the model gateway:

- with the dependency absent, the init container and Job reported
  `UNREACHABLE` and blocked start-up;
- after the dependency started, the agent rolled out and the Job and a
  CronJob-created Job completed `HEALTHY`;
- with the readiness probe, scaling the dependency to zero made all agent pods
  unready without restarts, and they became ready again when it returned.

The repository tests validate [agenthealth.yaml](agenthealth.yaml) against the
configuration schema and check the manifests' image and security settings.
Helm charts, sidecar mode and an operator remain planned.

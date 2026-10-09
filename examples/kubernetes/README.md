# Kubernetes scenarios

Run from the repository root with kubectl, Helm 3 and a disposable Kubernetes
cluster. These scenarios use a deterministic nginx HTTP fixture, not a deployed
agent framework, Agentgateway or Agent Router product. Read the
[Kubernetes guide](../../docs/kubernetes.md) for image references, safety,
Secret setup and probe semantics.

## Healthy deployment and validation Job

```bash
kubectl create namespace agenthealth-demo
kubectl -n agenthealth-demo apply -f examples/kubernetes/demo.yaml
kubectl -n agenthealth-demo rollout status deployment/demo --timeout=90s
kubectl -n agenthealth-demo apply -f examples/kubernetes/configmap.yaml
kubectl -n agenthealth-demo apply -f deploy/kubernetes/deployment.yaml
kubectl -n agenthealth-demo rollout status deployment/agenthealth --timeout=90s
kubectl -n agenthealth-demo apply -f deploy/kubernetes/job.yaml
kubectl -n agenthealth-demo wait --for=condition=complete job/agenthealth-check --timeout=90s
kubectl -n agenthealth-demo logs job/agenthealth-check
```

The result distinguishes direct agent, composite agent, gateway/router signals
and a configured HTTP route. Passive routed endpoint checks establish the
checker's path, not communication from the first agent to its peer. To test
first-to-peer communication use your runtime's explicitly authorized safe handler
as described in [agent checks](../../docs/agent-adapter.md).

## Init, sidecar, exec probes and scheduled checks

```bash
kubectl -n agenthealth-demo apply -f examples/kubernetes/init.yaml
kubectl -n agenthealth-demo rollout status deployment/demo-init --timeout=90s
kubectl -n agenthealth-demo apply -f examples/kubernetes/sidecar.yaml
kubectl -n agenthealth-demo rollout status deployment/demo-sidecar --timeout=90s
kubectl -n agenthealth-demo apply -f examples/kubernetes/exec-probes.yaml
kubectl -n agenthealth-demo rollout status deployment/agenthealth-exec --timeout=90s
kubectl -n agenthealth-demo apply -f deploy/kubernetes/cronjob.yaml
kubectl -n agenthealth-demo create job scheduled-check --from=cronjob/agenthealth-check
kubectl -n agenthealth-demo wait --for=condition=complete job/scheduled-check --timeout=90s
```

Init gates against the external demo Service. The sidecar checks its application
on localhost and gates the same Pod's readiness. The exec example runs checks
inside the AgentHealth serving container; it does not install the executable
into your application image.

## Helm

```bash
helm upgrade --install monitor deploy/helm/agenthealth -n agenthealth-demo \
  --set-file config=examples/kubernetes/agenthealth.yaml
helm upgrade --install validation deploy/helm/agenthealth -n agenthealth-demo \
  --set mode=job --set-file config=examples/kubernetes/agenthealth.yaml
helm upgrade --install scheduled deploy/helm/agenthealth -n agenthealth-demo \
  --set mode=cronjob --set-file config=examples/kubernetes/agenthealth.yaml
```

A completed Job is a one-shot gate; use a new release name to rerun it. Kubernetes
Job pod templates are immutable, so updating configuration or switching between
workload modes in an existing Helm release can require deleting the old workload.
Use separate release names for modes. The chart intentionally creates no Helm
hooks or automatic destructive replacement.

## Secret injection

Create the existing `agenthealth-credentials` Secret with an `agent-token` key
through your secret manager. The fixture at `/auth` accepts the synthetic test
value `fixture-token`; other values produce MISCONFIGURED. Then run:

```bash
kubectl -n agenthealth-demo apply -f examples/kubernetes/secret-job.yaml
kubectl -n agenthealth-demo wait --for=condition=complete job/agenthealth-auth --timeout=90s
kubectl -n agenthealth-demo logs job/agenthealth-auth
```

Production endpoints must enforce their own authentication. The runtime smoke
suite creates an ephemeral fixture Secret, checks success, replaces it with a
rejected value, and verifies failure without credential disclosure.

## Failure and cleanup

```bash
kubectl -n agenthealth-demo scale deployment/demo --replicas=0
kubectl -n agenthealth-demo delete job agenthealth-check
kubectl -n agenthealth-demo apply -f deploy/kubernetes/job.yaml
kubectl -n agenthealth-demo wait --for=condition=failed job/agenthealth-check --timeout=90s
kubectl -n agenthealth-demo logs job/agenthealth-check
kubectl -n agenthealth-demo delete namespace agenthealth-demo
```

A nonzero health exit fails the Job without retry. The monitor loses readiness
after its next completed observation, but dependency failure does not restart
its AHP process. The local sidecar remains healthy because its own fixture is
still running. For automated checks, run
`python3 scripts/test_kubernetes.py --context <disposable-context>`.

## Kustomize alternative

Use a fresh namespace for this alternative to the individual apply commands:

```bash
kubectl create namespace agenthealth-kustomize
kubectl -n agenthealth-kustomize apply -k examples/kubernetes
kubectl -n agenthealth-kustomize rollout status deployment/demo --timeout=90s
kubectl -n agenthealth-kustomize rollout status deployment/agenthealth --timeout=90s
kubectl -n agenthealth-kustomize wait --for=condition=complete job/agenthealth-check --timeout=90s
kubectl -n agenthealth-kustomize delete namespace agenthealth-kustomize
```

The generated configuration is content-addressed and workload references are
rewritten automatically. The base at `deploy/kubernetes` can also be consumed
by your own overlay. See the [guide](../../docs/kubernetes.md#kustomize-and-application-container-exec-probes)
for registry selection and the optional application-image binary-copy example.
Kustomize and these additional operational scenarios credit
[sharath568's PR #15](https://github.com/TheAgentHealth/agenthealth/pull/15).

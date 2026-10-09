# Reusable manifests

Apply `configmap.yaml` after editing its target endpoint, then select
`deployment.yaml` for AHP serving, `job.yaml` for a one-shot validation gate, or
`cronjob.yaml` for scheduled checks. All resources use the current kubectl
namespace. The serving monitor gates its own Pod; use the sidecar example to
gate an application's Pod. See the [guide](../../docs/kubernetes.md) and
[runnable scenarios](../../examples/kubernetes/README.md).

`kubectl kustomize deploy/kubernetes` renders the reusable base. Build an overlay
with a reviewed ConfigMap and your namespace/image choices, or use the
[runnable Kustomize overlay](../../examples/kubernetes/kustomization.yaml).

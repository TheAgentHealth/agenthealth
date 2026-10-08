#!/usr/bin/env bash
# End-to-end test of integrations/kubernetes on an existing cluster (kind in CI).
# Usage: scripts/kubernetes_smoke.sh <agenthealth-image> [namespace]
# The image must already be available to the cluster (for kind: kind load).
set -euo pipefail

image=${1:?usage: kubernetes_smoke.sh <agenthealth-image> [namespace]}
namespace=${2:-agenthealth-smoke}
root=$(cd "$(dirname "$0")/.." && pwd)
work=$(mktemp -d)
trap 'rm -rf "$work"; kubectl delete namespace "$namespace" --ignore-not-found --wait=false >/dev/null' EXIT

cp "$root"/integrations/kubernetes/*.yaml "$work/"
# Point the examples at an nginx stand-in for the model gateway and a runnable agent.
cat >"$work/agenthealth.yaml" <<YAML
version: v1
targets:
  - name: model-gateway
    type: http
    endpoint: http://model-gateway.$namespace.svc.cluster.local/
    auth:
      bearer_env: MODEL_GATEWAY_TOKEN
    http:
      expected_status: [200]
    timeout_ms: 2000
YAML
sed -i.bak "s#^namespace: .*#namespace: $namespace#" "$work/kustomization.yaml"
sed -i.bak "s#registry.example.com/research-agent:1.0.0#nginxinc/nginx-unprivileged:1.27-alpine#" "$work/init-container.yaml"
name=${image%:*} tag=${image##*:}
awk -v name="$name" -v tag="$tag" '/^    newTag:/ { print "    newName: " name; print "    newTag: \"" tag "\""; next } { print }' \
  "$work/kustomization.yaml" >"$work/kustomization.new" && mv "$work/kustomization.new" "$work/kustomization.yaml"

kubectl create namespace "$namespace" >/dev/null
kubectl apply -k "$work" >/dev/null

echo "Dependency absent: validation must fail"
kubectl -n "$namespace" wait --for=condition=failed --timeout=120s job/agenthealth-validate >/dev/null
kubectl -n "$namespace" logs job/agenthealth-validate | grep -q 'model-gateway (http): UNREACHABLE'

echo "Dependency present: workload starts and checks pass"
kubectl -n "$namespace" create deployment model-gateway --image=nginx:1.27-alpine --port=80 >/dev/null
kubectl -n "$namespace" expose deployment model-gateway --port=80 >/dev/null
kubectl -n "$namespace" rollout status deployment/model-gateway --timeout=120s >/dev/null
kubectl -n "$namespace" rollout status deployment/research-agent --timeout=300s >/dev/null
kubectl -n "$namespace" delete job agenthealth-validate >/dev/null
kubectl apply -k "$work" >/dev/null
kubectl -n "$namespace" wait --for=condition=complete --timeout=120s job/agenthealth-validate >/dev/null
kubectl -n "$namespace" create job --from=cronjob/agenthealth-periodic periodic-smoke >/dev/null
kubectl -n "$namespace" wait --for=condition=complete --timeout=120s job/periodic-smoke >/dev/null
kubectl -n "$namespace" logs job/periodic-smoke | python3 -c 'import json, sys; assert json.load(sys.stdin)["status"] == "HEALTHY"'
echo "Kubernetes examples passed"

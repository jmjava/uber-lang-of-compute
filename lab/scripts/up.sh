#!/usr/bin/env bash
# Bring up the KBL Kind lab: cluster, images, CRDs, operator, TSDB, sample workflow.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
IMAGE_TAG="${KBL_LAB_IMAGE_TAG:-lab}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: $1 is required" >&2
    exit 1
  }
}

need kind
need docker
need kubectl
need kustomize

mkdir -p /tmp/kbl-lab

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  echo "Creating Kind cluster $CLUSTER_NAME..."
  kind create cluster --name "$CLUSTER_NAME" --config "$ROOT/lab/kind/kind-config.yaml"
else
  echo "Kind cluster $CLUSTER_NAME already exists"
fi

echo "Building lab images..."
docker build -f "$ROOT/controller/docker/kbl-controller/Dockerfile" -t "kbl-controller:${IMAGE_TAG}" "$ROOT"
docker build -f "$ROOT/controller/docker/kbl-tsdb/Dockerfile" -t "kbl-tsdb:${IMAGE_TAG}" "$ROOT"
docker build -f "$ROOT/controller/docker/domino-runner/Dockerfile" -t "kbl-domino-runner:${IMAGE_TAG}" "$ROOT"
docker build -f "$ROOT/controller/docker/domino-runner-julia/Dockerfile" -t "kbl-domino-runner-julia:${IMAGE_TAG}" "$ROOT"

echo "Loading images into Kind..."
kind load docker-image "kbl-controller:${IMAGE_TAG}" --name "$CLUSTER_NAME"
kind load docker-image "kbl-tsdb:${IMAGE_TAG}" --name "$CLUSTER_NAME"
kind load docker-image "kbl-domino-runner:${IMAGE_TAG}" --name "$CLUSTER_NAME"
kind load docker-image "kbl-domino-runner-julia:${IMAGE_TAG}" --name "$CLUSTER_NAME"

echo "Installing CRDs..."
kubectl apply -f "$ROOT/crds/"

echo "Deploying KBL platform (controller + TSDB)..."
kustomize build "$ROOT/lab/kustomize/overlays/kind" | kubectl apply -f -

echo "Waiting for deployments..."
kubectl -n kbl-system rollout status deployment/kbl-controller --timeout=120s
kubectl -n kbl-system rollout status deployment/kbl-tsdb --timeout=120s

echo "Applying lab ComputeContext + Workflow..."
kubectl apply -f "$ROOT/lab/manifests/computecontext-lab.yaml"
kubectl apply -f "$ROOT/lab/manifests/workflow-lab.yaml"

echo ""
echo "Lab is up. Useful commands:"
echo "  kubectl get workflows -o wide"
echo "  kubectl -n kbl-system get pods"
echo "  kubectl logs -n kbl-system deployment/kbl-controller -f"
echo "  kubectl get configmap finance-lab-replay -o yaml   # after workflow completes"

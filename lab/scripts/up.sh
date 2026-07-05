#!/usr/bin/env bash
# Bring up the KBL Kind lab: cluster, images, CRDs, operator, TSDB, Volcano, sample workflow.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
IMAGE_TAG="${KBL_LAB_IMAGE_TAG:-lab}"
KBL_LAB_PROFILE="${KBL_LAB_PROFILE:-home}"
INSTALL_VOLCANO="${KBL_LAB_VOLCANO:-1}"
INSTALL_OPENKURISE="${KBL_LAB_OPENKURISE:-1}"

case "$KBL_LAB_PROFILE" in
  compact)
    KIND_CONFIG="$ROOT/lab/kind/kind-config-compact.yaml"
    QUEUE_MANIFEST="$ROOT/lab/manifests/volcano/queue-compact.yaml"
    WHEEL_MANIFEST="$ROOT/lab/manifests/volcano/computewheel-julia-finance-compact.yaml"
    APPLY_VOLCANO_CONTEXTS=0
    APPLY_VOLCANO_BURST=0
    INSTALL_OPENKURISE="${KBL_LAB_OPENKURISE:-0}"
    EXPECTED_NODES=2
    ;;
  home)
    KIND_CONFIG="$ROOT/lab/kind/kind-config-home.yaml"
    QUEUE_MANIFEST="$ROOT/lab/manifests/volcano/queue-home.yaml"
    WHEEL_MANIFEST="$ROOT/lab/manifests/volcano/computewheel-julia-finance.yaml"
    APPLY_VOLCANO_CONTEXTS=1
    APPLY_VOLCANO_BURST=1
    EXPECTED_NODES=3
    ;;
  default|*)
    KIND_CONFIG="$ROOT/lab/kind/kind-config.yaml"
    QUEUE_MANIFEST="$ROOT/lab/manifests/volcano/queue.yaml"
    WHEEL_MANIFEST="$ROOT/lab/manifests/volcano/computewheel-julia-finance.yaml"
    APPLY_VOLCANO_CONTEXTS=1
    APPLY_VOLCANO_BURST=1
    EXPECTED_NODES=3
    ;;
esac

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

mkdir -p /tmp/kbl-lab/cp /tmp/kbl-lab/w1 /tmp/kbl-lab/w2

echo "Lab profile: ${KBL_LAB_PROFILE} (Kind config: ${KIND_CONFIG##*/})"

if ! kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  echo "Creating Kind cluster $CLUSTER_NAME..."
  kind create cluster --name "$CLUSTER_NAME" --config "$KIND_CONFIG"
else
  echo "Kind cluster $CLUSTER_NAME already exists"
  NODE_COUNT="$(kubectl get nodes --no-headers 2>/dev/null | wc -l | tr -d ' ')"
  if [[ "${NODE_COUNT}" != "${EXPECTED_NODES}" ]]; then
    echo "warning: profile ${KBL_LAB_PROFILE} expects ${EXPECTED_NODES} nodes, found ${NODE_COUNT}" >&2
    echo "         recreate: ./lab/scripts/down.sh && KBL_LAB_PROFILE=${KBL_LAB_PROFILE} ./lab/scripts/up.sh" >&2
  fi
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

if [[ "${INSTALL_VOLCANO}" != "0" ]]; then
  chmod +x "$ROOT/lab/scripts/install-volcano.sh"
  "$ROOT/lab/scripts/install-volcano.sh"
fi

if [[ "${INSTALL_OPENKURISE}" != "0" ]]; then
  chmod +x "$ROOT/lab/scripts/install-openkruise.sh"
  "$ROOT/lab/scripts/install-openkruise.sh"
fi

echo "Deploying KBL platform (controller + TSDB)..."
kustomize build "$ROOT/lab/kustomize/overlays/kind" | kubectl apply -f -

echo "Waiting for deployments..."
kubectl -n kbl-system rollout status deployment/kbl-controller --timeout=120s
kubectl -n kbl-system rollout status deployment/kbl-tsdb --timeout=120s

echo "Applying lab ComputeContext + Workflow..."
kubectl apply -f "$ROOT/lab/manifests/computecontext-lab.yaml"
kubectl apply -f "$ROOT/lab/manifests/workflow-lab.yaml"

if [[ "${INSTALL_VOLCANO}" != "0" ]]; then
  echo "Applying Volcano demo (queue + ComputeWheel volcano-init)..."
  kubectl apply -f "$QUEUE_MANIFEST"
  if [[ "${APPLY_VOLCANO_CONTEXTS}" == "1" ]]; then
    kubectl apply -f "$ROOT/lab/manifests/volcano/computecontexts-volcano.yaml"
  fi
  kubectl apply -f "$WHEEL_MANIFEST"
  echo "Waiting for ComputeWheel julia-finance-wheel (Workflow → DominoChain → VCJob)..."
  kubectl wait --for=jsonpath='{.status.phase}'=Idle \
    computewheel/julia-finance-wheel --timeout=600s 2>/dev/null || {
    echo "Wheel still processing — check: ./lab/scripts/verify-volcano.sh"
  }
  if [[ "${APPLY_VOLCANO_BURST}" == "1" ]]; then
    chmod +x "$ROOT/lab/scripts/apply-volcano-burst.sh"
    "$ROOT/lab/scripts/apply-volcano-burst.sh"
  fi
fi

if [[ "${INSTALL_OPENKURISE}" != "0" ]]; then
  echo "Applying OpenKruise demo (Julia hot-swap DominoChain)..."
  kubectl apply -k "$ROOT/lab/manifests/openkruise/"
  echo "Waiting for DominoChain julia-finance-openkruise..."
  kubectl wait --for=jsonpath='{.status.phase}'=Completed \
    dominochain/julia-finance-openkruise --timeout=300s 2>/dev/null || {
    echo "OpenKruise chain still running — check: kubectl get dchain,pods,crr -l kbl.io/openkruise-demo=true"
  }
fi

echo ""
echo "Lab is up (profile=${KBL_LAB_PROFILE}). Useful commands:"
echo "  ./lab/scripts/verify-volcano.sh          # Volcano queue, VCJobs, pod placement"
echo "  kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node,kbl.io/gpu"
echo "  kubectl get workflows -o wide"
echo "  kubectl -n kbl-system get pods -o wide"
if [[ "${INSTALL_VOLCANO}" != "0" ]]; then
  echo "  kubectl get wheel julia-finance-wheel -o wide"
  echo "  kubectl get wf -l kbl.io/computewheel=julia-finance-wheel"
  echo "  kubectl get dchain,vcjob -l kbl.io/volcano-demo=true"
  echo "  kubectl get pods -l kbl.io/volcano-demo=true -o wide"
  echo "  kubectl -n volcano-system get pods"
fi
if [[ "${INSTALL_OPENKURISE}" != "0" ]]; then
  echo "  kubectl get dchain julia-finance-openkruise -o wide"
  echo "  kubectl get pods -l kbl.io/openkruise-demo=true"
fi
echo "  kubectl logs -n kbl-system deployment/kbl-controller -f"
echo "  kubectl get configmap finance-lab-replay -o yaml   # after workflow completes"

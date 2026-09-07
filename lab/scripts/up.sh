#!/usr/bin/env bash
# Bring up the KBL Kind lab: cluster, images, CRDs, operator, TSDB, Volcano, sample workflow.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
IMAGE_TAG="${KBL_LAB_IMAGE_TAG:-lab}"
KBL_LAB_PROFILE="${KBL_LAB_PROFILE:-home}"
INSTALL_VOLCANO="${KBL_LAB_VOLCANO:-1}"
INSTALL_OPENKURISE="${KBL_LAB_OPENKURISE:-1}"
INSTALL_JULIA="${KBL_LAB_JULIA:-1}"
WHEEL_NAME="julia-finance-wheel"
APPLY_DESK_DAY=0
JULIA_WHEEL_MANIFEST=""
JULIA_WHEEL_NAME=""

case "$KBL_LAB_PROFILE" in
  compact)
    KIND_CONFIG="$ROOT/lab/kind/kind-config-compact.yaml"
    QUEUE_MANIFEST="$ROOT/lab/manifests/volcano/queue-compact.yaml"
    WHEEL_MANIFEST="$ROOT/lab/manifests/volcano/computewheel-desk-day-compact.yaml"
    WHEEL_NAME="rates-desk-wheel"
    APPLY_DESK_DAY=1
    if [[ "${INSTALL_JULIA}" != "0" ]]; then
      JULIA_WHEEL_MANIFEST="$ROOT/lab/manifests/volcano/computewheel-julia-finance-compact.yaml"
      JULIA_WHEEL_NAME="julia-finance-wheel"
    fi
    APPLY_VOLCANO_CONTEXTS=0
    APPLY_VOLCANO_BURST=0
    INSTALL_OPENKURISE="${KBL_LAB_OPENKURISE:-1}"
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

echo "Lab profile: ${KBL_LAB_PROFILE} volcano=${INSTALL_VOLCANO} julia=${INSTALL_JULIA} openkruise=${INSTALL_OPENKURISE}"

if findmnt -n -o FSTYPE / 2>/dev/null | grep -q '^overlay'; then
  export KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER="${KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER:-native}"
  echo "nested overlay detected; Kind containerd snapshotter=${KIND_EXPERIMENTAL_CONTAINERD_SNAPSHOTTER}"
fi

# Nested Docker: iptables-legacy FORWARD DROP blocks Kind node ICC.
if command -v iptables-legacy >/dev/null 2>&1; then
  if [[ "$(id -u)" -eq 0 ]]; then
    iptables-legacy -P FORWARD ACCEPT 2>/dev/null || true
    iptables -P FORWARD ACCEPT 2>/dev/null || true
  else
    sudo iptables-legacy -P FORWARD ACCEPT 2>/dev/null || true
    sudo iptables -P FORWARD ACCEPT 2>/dev/null || true
  fi
fi

mkdir -p /tmp/kbl-lab/cp /tmp/kbl-lab/w1 /tmp/kbl-lab/w2

echo "Lab profile: ${KBL_LAB_PROFILE} (Kind config: ${KIND_CONFIG##*/})"

ensure_kube_proxy_nftables() {
  # Existing clusters created with iptables mode keep that ConfigMap; nested
  # kernels then fail kube-proxy sync (missing xt_statistic).
  local conf
  conf="$(kubectl -n kube-system get cm kube-proxy -o jsonpath='{.data.config\.conf}' 2>/dev/null || true)"
  if [[ -z "$conf" ]]; then
    return
  fi
  if echo "$conf" | grep -q '^mode: nftables'; then
    return
  fi
  echo "switching kube-proxy to nftables (nested Kind ClusterIP)..."
  local patch
  patch="$(kubectl -n kube-system get cm kube-proxy -o json | python3 -c '
import json, sys
obj = json.load(sys.stdin)
obj["data"]["config.conf"] = obj["data"]["config.conf"].replace("mode: iptables", "mode: nftables")
print(json.dumps({"data": obj["data"]}))
')"
  kubectl -n kube-system patch cm kube-proxy --type merge -p "$patch"
  kubectl -n kube-system rollout restart ds/kube-proxy
  kubectl -n kube-system rollout status ds/kube-proxy --timeout=90s || true
}

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

ensure_kube_proxy_nftables

echo "Building lab images..."
docker build -f "$ROOT/controller/docker/kbl-controller/Dockerfile" -t "kbl-controller:${IMAGE_TAG}" "$ROOT"
docker build -f "$ROOT/controller/docker/kbl-tsdb/Dockerfile" -t "kbl-tsdb:${IMAGE_TAG}" "$ROOT"
docker build -f "$ROOT/controller/docker/domino-runner/Dockerfile" -t "kbl-domino-runner:${IMAGE_TAG}" "$ROOT"
if [[ "${INSTALL_JULIA}" != "0" ]]; then
  docker build -f "$ROOT/controller/docker/domino-runner-julia/Dockerfile" -t "kbl-domino-runner-julia:${IMAGE_TAG}" "$ROOT"
fi

echo "Loading images into Kind..."
docker pull registry.k8s.io/pause:3.9
kind load docker-image registry.k8s.io/pause:3.9 --name "$CLUSTER_NAME"
kind load docker-image "kbl-controller:${IMAGE_TAG}" --name "$CLUSTER_NAME"
kind load docker-image "kbl-tsdb:${IMAGE_TAG}" --name "$CLUSTER_NAME"
kind load docker-image "kbl-domino-runner:${IMAGE_TAG}" --name "$CLUSTER_NAME"
if [[ "${INSTALL_JULIA}" != "0" ]]; then
  kind load docker-image "kbl-domino-runner-julia:${IMAGE_TAG}" --name "$CLUSTER_NAME"
fi

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

echo "Applying lab ComputeContext + catalog Snapshot/Domino/Universe CRs..."
WORKER="$(kubectl get nodes -l kbl.io/lab-role=compute -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)"
if [[ -z "$WORKER" ]]; then
  WORKER="$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}')"
fi
sed "s/nodeName: .*/nodeName: ${WORKER}/" "$ROOT/lab/manifests/computecontext-lab.yaml" | kubectl apply -f -
kubectl apply -k "$ROOT/lab/manifests/catalog/"
echo "Waiting for catalog snapshots to seal..."
kubectl wait --for=jsonpath='{.status.phase}'=Sealed snapshot/julia-curve-2025-04-15 --timeout=120s
kubectl wait --for=jsonpath='{.status.phase}'=Sealed snapshot/ust-par-2025-04-15 --timeout=120s

echo "Applying lab Workflow..."
kubectl apply -f "$ROOT/lab/manifests/workflow-lab.yaml"

if [[ "${INSTALL_VOLCANO}" != "0" ]]; then
  echo "Applying Volcano demo (queue + ComputeWheel volcano-init)..."
  kubectl apply -f "$QUEUE_MANIFEST"
  if [[ "${APPLY_VOLCANO_CONTEXTS}" == "1" ]]; then
    kubectl apply -f "$ROOT/lab/manifests/volcano/computecontexts-volcano.yaml"
  fi
  if [[ "${APPLY_DESK_DAY}" == "1" ]]; then
    echo "Applying rates desk-day book (ny-rates + ln-rates on ${WORKER})..."
    sed "s/nodeName: .*/nodeName: ${WORKER}/" "$ROOT/lab/manifests/volcano/computecontexts-desk-day.yaml" | kubectl apply -f -
  fi
  kubectl apply -f "$WHEEL_MANIFEST"
  echo "Waiting for ComputeWheel ${WHEEL_NAME} (Workflow → DominoChain → VCJob)..."
  kubectl wait --for=jsonpath='{.status.phase}'=Idle \
    "computewheel/${WHEEL_NAME}" --timeout=600s 2>/dev/null || {
    echo "Wheel still processing — check: ./lab/scripts/verify-volcano.sh"
  }
  if [[ -n "${JULIA_WHEEL_MANIFEST}" ]]; then
    echo "Applying Julia finance wheel ${JULIA_WHEEL_NAME}..."
    kubectl apply -f "$JULIA_WHEEL_MANIFEST"
    echo "Waiting for ComputeWheel ${JULIA_WHEEL_NAME}..."
    kubectl wait --for=jsonpath='{.status.phase}'=Idle \
      "computewheel/${JULIA_WHEEL_NAME}" --timeout=600s 2>/dev/null || {
      echo "Julia wheel still processing — check: ./lab/scripts/verify-volcano.sh"
    }
  fi
  if [[ "${APPLY_VOLCANO_BURST}" == "1" ]]; then
    chmod +x "$ROOT/lab/scripts/apply-volcano-burst.sh"
    "$ROOT/lab/scripts/apply-volcano-burst.sh"
  fi
fi

if [[ "${INSTALL_OPENKURISE}" != "0" ]]; then
  echo "Applying OpenKruise demo (Julia catalog Workflow → DominoChain)..."
  kubectl apply -k "$ROOT/lab/manifests/openkruise/"
  echo "Waiting for Workflow julia-finance-openkruise..."
  kubectl wait --for=jsonpath='{.status.phase}'=Completed \
    workflow/julia-finance-openkruise --timeout=600s 2>/dev/null || {
    echo "OpenKruise workflow still running — check: kubectl get wf,dchain,pods -l kbl.io/openkruise-demo=true"
  }
fi

echo ""
echo "Lab is up (profile=${KBL_LAB_PROFILE}). Useful commands:"
echo "  ./lab/scripts/verify-volcano.sh          # Volcano queue, VCJobs, pod placement"
echo "  kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node,kbl.io/gpu"
echo "  kubectl get workflows -o wide"
echo "  kubectl -n kbl-system get pods -o wide"
if [[ "${INSTALL_VOLCANO}" != "0" ]]; then
  echo "  kubectl get wheel -l kbl.io/volcano-demo=true -o wide"
  echo "  kubectl get wf -l kbl.io/computewheel"
  echo "  kubectl get dchain,vcjob -l kbl.io/volcano-demo=true"
  echo "  kubectl get pods -l kbl.io/volcano-demo=true -o wide"
  echo "  kubectl -n volcano-system get pods"
fi
if [[ "${INSTALL_OPENKURISE}" != "0" ]]; then
  echo "  kubectl get wf,dchain -l kbl.io/openkruise-demo=true"
  echo "  kubectl get pods -l kbl.io/openkruise-demo=true"
fi
echo "  kubectl logs -n kbl-system deployment/kbl-controller -f"
echo "  kubectl get configmap finance-lab-replay -o yaml   # after workflow completes"

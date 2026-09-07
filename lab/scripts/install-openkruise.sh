#!/usr/bin/env bash
# Install OpenKruise into the current kubectl context (Helm chart).
set -euo pipefail

OPENKURISE_VERSION="${KBL_OPENKURISE_VERSION:-1.6.4}"
OPENKURISE_IMAGE_TAG="${KBL_OPENKURISE_IMAGE_TAG:-v${OPENKURISE_VERSION#v}}"
RELEASE_NAME="${KBL_OPENKURISE_RELEASE:-kruise}"
NAMESPACE="${KBL_OPENKURISE_NAMESPACE:-kruise-system}"
CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
KRUISE_IMAGE="openkruise/kruise-manager:${OPENKURISE_IMAGE_TAG}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: $1 is required" >&2
    exit 1
  }
}

need kubectl

if ! command -v helm >/dev/null 2>&1; then
  echo "Helm not found — installing Helm 3..."
  curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
fi

# Nested Kind (native snapshotter) cannot unpack some images' overlay whiteouts.
# Flatten to a single layer and ctr-import onto every node.
kind_load_kruise_image() {
  echo "Preloading ${KRUISE_IMAGE} into Kind..."
  docker pull "${KRUISE_IMAGE}"
  if kind load docker-image "${KRUISE_IMAGE}" --name "$CLUSTER_NAME"; then
    return
  fi
  echo "kind load failed (nested snapshotter whiteouts); flattening ${KRUISE_IMAGE}..."
  local tmp="kbl-kruise-flat-$$"
  local flat="kbl-kruise-manager:lab"
  docker rm -f "$tmp" >/dev/null 2>&1 || true
  docker create --name "$tmp" "${KRUISE_IMAGE}" >/dev/null
  docker export "$tmp" | docker import --change 'ENTRYPOINT ["/manager"]' - "$flat"
  docker rm "$tmp" >/dev/null
  local node
  for node in $(kind get nodes --name "$CLUSTER_NAME"); do
    docker save "$flat" | docker exec -i "$node" ctr --namespace=k8s.io images import -
    docker exec "$node" ctr --namespace=k8s.io images tag \
      "docker.io/library/${flat}" "docker.io/${KRUISE_IMAGE}" >/dev/null 2>&1 || \
      docker exec "$node" ctr --namespace=k8s.io images tag \
        "docker.io/library/${flat}" "${KRUISE_IMAGE}" >/dev/null 2>&1 || true
  done
}

KIND_CLUSTER=0
if command -v docker >/dev/null 2>&1 && command -v kind >/dev/null 2>&1 \
  && kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  KIND_CLUSTER=1
  kind_load_kruise_image
fi

helm repo add openkruise https://openkruise.github.io/charts/ 2>/dev/null || true
helm repo update openkruise

kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f - >/dev/null

HELM_ARGS=(
  --version "${OPENKURISE_VERSION}"
  --namespace "${NAMESPACE}"
  --wait
  --timeout 5m
  --set manager.replicas=1
  --set featureGates="KruiseDaemon=true\,ImagePullJobGate=true"
)
if [[ "$KIND_CLUSTER" == "1" ]]; then
  HELM_ARGS+=(--set manager.image.pullPolicy=IfNotPresent)
fi

echo "Installing OpenKruise ${OPENKURISE_VERSION} (KruiseDaemon + ImagePullJobGate on)..."
helm upgrade --install "${RELEASE_NAME}" openkruise/kruise "${HELM_ARGS[@]}"

echo "Waiting for kruise-controller-manager..."
kubectl -n "${NAMESPACE}" rollout status deployment/kruise-controller-manager --timeout=180s

if [[ "$KIND_CLUSTER" == "1" ]]; then
  for wh in $(kubectl get mutatingwebhookconfiguration,validatingwebhookconfiguration -o name | grep kruise || true); do
    kubectl patch "$wh" --type json \
      -p '[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"}]' >/dev/null 2>&1 || true
  done
fi

echo "OpenKruise is ready."

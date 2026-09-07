#!/usr/bin/env bash
# Install Volcano batch scheduler into the current kubectl context.
set -euo pipefail

VOLCANO_VERSION="${KBL_VOLCANO_VERSION:-v1.9.0}"
MANIFEST_URL="https://raw.githubusercontent.com/volcano-sh/volcano/${VOLCANO_VERSION}/installer/volcano-development.yaml"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: $1 is required" >&2
    exit 1
  }
}

need kubectl

CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
if command -v docker >/dev/null 2>&1 && command -v kind >/dev/null 2>&1 \
  && kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  for img in \
    "volcanosh/vc-webhook-manager:${VOLCANO_VERSION}" \
    "volcanosh/vc-controller-manager:${VOLCANO_VERSION}" \
    "volcanosh/vc-scheduler:${VOLCANO_VERSION}"; do
    echo "Preloading $img into Kind..."
    docker pull "$img"
    kind load docker-image "$img" --name "$CLUSTER_NAME"
  done
fi

echo "Installing Volcano ${VOLCANO_VERSION} from ${MANIFEST_URL}..."
kubectl apply -f "${MANIFEST_URL}"

echo "Waiting for Volcano admission..."
kubectl -n volcano-system rollout status deployment/volcano-admission --timeout=180s

echo "Waiting for Volcano controllers..."
kubectl -n volcano-system rollout status deployment/volcano-controllers --timeout=180s

echo "Waiting for Volcano scheduler..."
kubectl -n volcano-system rollout status deployment/volcano-scheduler --timeout=180s

echo "Volcano is ready."

#!/usr/bin/env bash
# Install OpenKruise into the current kubectl context (Helm chart).
set -euo pipefail

OPENKURISE_VERSION="${KBL_OPENKURISE_VERSION:-1.6.4}"
RELEASE_NAME="${KBL_OPENKURISE_RELEASE:-kruise}"
NAMESPACE="${KBL_OPENKURISE_NAMESPACE:-kruise-system}"

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

helm repo add openkruise https://openkruise.github.io/charts/ 2>/dev/null || true
helm repo update openkruise

echo "Installing OpenKruise ${OPENKURISE_VERSION} (KruiseDaemon disabled for Kind lab)..."
helm upgrade --install "${RELEASE_NAME}" openkruise/kruise \
  --version "${OPENKURISE_VERSION}" \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  --wait \
  --timeout 5m \
  --set manager.replicas=1 \
  --set featureGates="KruiseDaemon=false"

echo "Waiting for kruise-controller-manager..."
kubectl -n "${NAMESPACE}" rollout status deployment/kruise-controller-manager --timeout=180s

echo "OpenKruise is ready."

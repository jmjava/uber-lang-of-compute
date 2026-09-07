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

ensure_kind_admission_secret() {
  # Nested Kind often cannot run gen-admission-secret.sh in-cluster (API timeouts).
  # Mint the webhook certs on the host and patch caBundle, matching Volcano's secret keys.
  if kubectl -n volcano-system get secret volcano-admission-secret >/dev/null 2>&1; then
    return
  fi
  echo "Creating volcano-admission-secret on the host..."
  local certdir service namespace secret cabundle wh
  certdir="$(mktemp -d)"
  service="volcano-admission-service"
  namespace="volcano-system"
  secret="volcano-admission-secret"
  cat > "${certdir}/csr.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${service}
DNS.2 = ${service}.${namespace}
DNS.3 = ${service}.${namespace}.svc
EOF
  openssl genrsa -out "${certdir}/ca.key" 2048 >/dev/null 2>&1
  openssl req -x509 -new -nodes -key "${certdir}/ca.key" -subj "/CN=${service}.${namespace}.svc" -days 3650 -out "${certdir}/ca.crt" >/dev/null 2>&1
  openssl genrsa -out "${certdir}/server.key" 2048 >/dev/null 2>&1
  openssl req -new -key "${certdir}/server.key" -subj "/CN=${service}.${namespace}.svc" -out "${certdir}/server.csr" -config "${certdir}/csr.conf" >/dev/null 2>&1
  openssl x509 -req -in "${certdir}/server.csr" -CA "${certdir}/ca.crt" -CAkey "${certdir}/ca.key" -CAcreateserial \
    -out "${certdir}/server.crt" -extensions v3_req -extfile "${certdir}/csr.conf" -days 3650 >/dev/null 2>&1
  kubectl -n "$namespace" create secret generic "$secret" \
    --from-file=tls.key="${certdir}/server.key" \
    --from-file=tls.crt="${certdir}/server.crt" \
    --from-file=ca.crt="${certdir}/ca.crt"
  cabundle="$(base64 -w0 "${certdir}/ca.crt")"
  for wh in $(kubectl get mutatingwebhookconfiguration,validatingwebhookconfiguration -o name | grep volcano || true); do
    kubectl get "$wh" -o json | CABUNDLE="$cabundle" python3 -c '
import json, sys, os
obj = json.load(sys.stdin)
bundle = os.environ["CABUNDLE"]
for hook in obj.get("webhooks", []):
    hook.setdefault("clientConfig", {})["caBundle"] = bundle
json.dump(obj, sys.stdout)
' | kubectl apply -f -
  done
  rm -rf "$certdir"
}

CLUSTER_NAME="${KBL_KIND_CLUSTER:-kbl-lab}"
KIND_CLUSTER=0
if command -v docker >/dev/null 2>&1 && command -v kind >/dev/null 2>&1 \
  && kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  KIND_CLUSTER=1
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
if [[ "$KIND_CLUSTER" == "1" ]]; then
  # Kind already has the images from preload; Always would re-pull from Docker Hub and stall.
  curl -fsSL "${MANIFEST_URL}" | sed 's/imagePullPolicy: Always/imagePullPolicy: IfNotPresent/g' | kubectl apply -f -
  ensure_kind_admission_secret
  echo "Kind lab: drop cluster-wide pod webhooks (admission has no endpoints until certs exist; they block kbl-system pods)."
  kubectl delete mutatingwebhookconfiguration volcano-admission-service-pods-mutate --ignore-not-found >/dev/null
  kubectl delete validatingwebhookconfiguration volcano-admission-service-pods-validate --ignore-not-found >/dev/null
  for wh in volcano-admission-service-jobs-mutate volcano-admission-service-podgroups-mutate volcano-admission-service-queues-mutate; do
    kubectl patch mutatingwebhookconfiguration "$wh" --type json \
      -p '[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"}]' >/dev/null 2>&1 || true
  done
  for wh in volcano-admission-service-jobs-validate volcano-admission-service-queues-validate; do
    kubectl patch validatingwebhookconfiguration "$wh" --type json \
      -p '[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"}]' >/dev/null 2>&1 || true
  done
  # Compact/nested Kind: admission CrashLoops on API timeouts and is unused after
  # pod webhooks are dropped. Controllers + scheduler are enough for VCJobs.
  kubectl -n volcano-system scale deploy/volcano-admission --replicas=0 >/dev/null 2>&1 || true
else
  kubectl apply -f "${MANIFEST_URL}"
fi

if [[ "$KIND_CLUSTER" != "1" ]]; then
  echo "Waiting for Volcano admission..."
  kubectl -n volcano-system rollout status deployment/volcano-admission --timeout=180s
fi

echo "Waiting for Volcano controllers..."
kubectl -n volcano-system rollout status deployment/volcano-controllers --timeout=180s

echo "Waiting for Volcano scheduler..."
kubectl -n volcano-system rollout status deployment/volcano-scheduler --timeout=180s

echo "Volcano is ready."

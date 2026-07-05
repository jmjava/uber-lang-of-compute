#!/usr/bin/env bash
# Apply parallel Volcano burst demo (two VCJobs pinned to different workers).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TEMPLATE="$ROOT/lab/manifests/volcano/dominochain-volcano-burst.template.yaml"

WORKER_A="${WORKER_A:-$(kubectl get nodes -l 'kbl.io/lab-role=compute' -o jsonpath='{.items[0].metadata.name}')}"
WORKER_B="${WORKER_B:-$(kubectl get nodes -l 'kbl.io/lab-role=compute' -o jsonpath='{.items[1].metadata.name}')}"

if [[ -z "$WORKER_A" ]]; then
  echo "error: no compute worker nodes found" >&2
  exit 1
fi

if [[ -z "$WORKER_B" || "$WORKER_B" == "$WORKER_A" ]]; then
  echo "warning: only one compute worker — burst-b will share node with burst-a" >&2
  WORKER_B="$WORKER_A"
fi

echo "Volcano burst: VCJob A → ${WORKER_A}, VCJob B → ${WORKER_B}"
sed -e "s/__WORKER_A__/${WORKER_A}/g" -e "s/__WORKER_B__/${WORKER_B}/g" "$TEMPLATE" | kubectl apply -f -

echo "Waiting for burst DominoChains..."
kubectl wait --for=jsonpath='{.status.phase}'=Completed \
  dominochain/volcano-burst-a dominochain/volcano-burst-b --timeout=600s 2>/dev/null || {
  echo "Burst still running — check: kubectl get dchain,vcjob,pods -l kbl.io/volcano-burst=true"
}

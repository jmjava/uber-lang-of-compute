#!/usr/bin/env bash
# Show Volcano queue, VCJobs, and pod placement for the lab demo.
set -euo pipefail

section() {
  echo ""
  echo "=== $1 ==="
}

section "Volcano system"
kubectl -n volcano-system get pods 2>/dev/null || echo "(volcano-system not found — set KBL_LAB_VOLCANO=1)"

section "Queue kbl-lab"
if kubectl get queue kbl-lab &>/dev/null; then
  kubectl get queue kbl-lab -o custom-columns=NAME:.metadata.name,STATE:.status.state,INQUEUE:.status.inqueue,RUNNING:.status.running
  echo ""
  kubectl get queue kbl-lab -o jsonpath='  capability: cpu={.spec.capability.cpu} memory={.spec.capability.memory}{"\n"}' 2>/dev/null || true
else
  echo "  (queue kbl-lab not found)"
fi

section "ComputeWheel (Ferris wheel → VCJob pipeline)"
kubectl get computewheel julia-finance-wheel -o wide 2>/dev/null || echo "  (wheel not found)"
kubectl get wf -l kbl.io/computewheel=julia-finance-wheel -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,CONTEXT:.spec.routing.computeContextRef 2>/dev/null || true

section "Volcano Jobs + PodGroups"
kubectl get vcjob -l kbl.io/volcano-demo=true -o wide 2>/dev/null || kubectl get jobs.batch.volcano.sh -l kbl.io/volcano-demo=true 2>/dev/null || echo "  (no VCJobs yet)"
kubectl get podgroups.scheduling.volcano.sh -l kbl.io/volcano-demo=true 2>/dev/null || true

section "DominoChains (volcano-init)"
kubectl get dominochain -l kbl.io/volcano-demo=true -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,QUEUE:.spec.volcanoQueue 2>/dev/null || true

section "Pods (scheduler + node placement)"
kubectl get pods -l kbl.io/volcano-demo=true -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,VCJOB:.metadata.labels.batch\\.volcano\\.sh/job-name,NODE:.spec.nodeName,SCHEDULER:.spec.schedulerName 2>/dev/null || true

section "Init-chain logs (last domino slot)"
POD=$(kubectl get pods -l kbl.io/volcano-demo=true --field-selector=status.phase=Succeeded -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null || true)
if [[ -n "$POD" ]]; then
  CONTAINER=$(kubectl get pod "$POD" -o jsonpath='{.spec.initContainers[-1].name}' 2>/dev/null || true)
  if [[ -n "$CONTAINER" ]]; then
    echo "  pod=$POD container=$CONTAINER"
    kubectl logs "$POD" -c "$CONTAINER" --tail=15 2>/dev/null || true
  fi
else
  echo "  (no succeeded volcano demo pods yet)"
fi

section "Quick checks"
echo "  scheduler=volcano on pods above confirms Volcano batch path"
echo "  two contexts on wheel → sequential VCJobs through queue kbl-lab"
echo "  burst demo: kubectl get dchain -l kbl.io/volcano-burst=true"

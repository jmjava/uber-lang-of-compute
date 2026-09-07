#!/usr/bin/env bash
# Show Volcano queue, VCJobs, and pod placement for the lab demo.
# Pass --strict (or KBL_VERIFY_STRICT=1) to fail if Volcano or a VCJob is missing.
set -euo pipefail

STRICT=0
if [[ "${1:-}" == "--strict" || "${KBL_VERIFY_STRICT:-0}" == "1" ]]; then
  STRICT=1
fi

section() {
  echo ""
  echo "=== $1 ==="
}

fail() {
  echo "error: $*" >&2
  exit 1
}

section "Volcano system"
if kubectl -n volcano-system get pods >/dev/null 2>&1; then
  kubectl -n volcano-system get pods
else
  echo "(volcano-system not found — set KBL_LAB_VOLCANO=1)"
  if [[ "$STRICT" == "1" ]]; then
    fail "volcano-system namespace missing"
  fi
fi

section "Queue kbl-lab"
if kubectl get queue kbl-lab &>/dev/null; then
  kubectl get queue kbl-lab -o custom-columns=NAME:.metadata.name,STATE:.status.state,INQUEUE:.status.inqueue,RUNNING:.status.running
  echo ""
  kubectl get queue kbl-lab -o jsonpath='  capability: cpu={.spec.capability.cpu} memory={.spec.capability.memory}{"\n"}' 2>/dev/null || true
else
  echo "  (queue kbl-lab not found)"
  if [[ "$STRICT" == "1" ]]; then
    fail "queue kbl-lab missing"
  fi
fi

section "ComputeWheel (Ferris wheel → VCJob pipeline)"
kubectl get computewheel -l kbl.io/volcano-demo=true -o wide 2>/dev/null || kubectl get computewheel -o wide 2>/dev/null || echo "  (wheel not found)"
kubectl get wf -l kbl.io/computewheel -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,CONTEXT:.spec.routing.computeContextRef 2>/dev/null || true

section "Volcano Jobs + PodGroups"
if kubectl get jobs.batch.volcano.sh >/dev/null 2>&1; then
  kubectl get jobs.batch.volcano.sh -o wide
else
  echo "  (Volcano Job CRD not installed)"
fi
kubectl get podgroups.scheduling.volcano.sh 2>/dev/null || true

section "DominoChains (volcano-init)"
kubectl get dominochain -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,QUEUE:.spec.volcanoQueue,RUNTIME:.spec.runtime 2>/dev/null || true

section "Pods (scheduler + node placement)"
kubectl get pods -l kbl.io/dominochain -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,VCJOB:.metadata.labels.batch\\.volcano\\.sh/job-name,NODE:.spec.nodeName,SCHEDULER:.spec.schedulerName 2>/dev/null || \
  kubectl get pods -l kbl.io/volcano-demo=true -o custom-columns=NAME:.metadata.name,STATUS:.status.phase,VCJOB:.metadata.labels.batch\\.volcano\\.sh/job-name,NODE:.spec.nodeName,SCHEDULER:.spec.schedulerName 2>/dev/null || true

section "Init-chain logs (last domino slot)"
POD=$(kubectl get pods -l kbl.io/dominochain --field-selector=status.phase=Succeeded -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null || true)
if [[ -z "$POD" ]]; then
  POD=$(kubectl get pods -l kbl.io/volcano-demo=true --field-selector=status.phase=Succeeded -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null || true)
fi
if [[ -z "$POD" ]]; then
  # volcano-init keeps /pause running; inits still finished.
  POD=$(kubectl get pods -l kbl.io/dominochain -o jsonpath='{.items[-1].metadata.name}' 2>/dev/null || true)
fi
if [[ -n "$POD" ]]; then
  CONTAINER=$(kubectl get pod "$POD" -o jsonpath='{.spec.initContainers[-1].name}' 2>/dev/null || true)
  if [[ -n "$CONTAINER" ]]; then
    echo "  pod=$POD container=$CONTAINER"
    kubectl logs "$POD" -c "$CONTAINER" --tail=15 2>/dev/null || true
  fi
else
  echo "  (no volcano demo pods yet)"
fi

section "Quick checks"
echo "  scheduler=volcano on pods above confirms Volcano batch path"
echo "  compact desk-day book: rates-desk-wheel (ny-rates → ln-rates) → Workflow → DominoChain → VCJob"
echo "  compact Julia wheel: julia-finance-wheel (default-context, julia:greeks) on the same worker"
echo "  OpenKruise runtime: julia-finance-openkruise (runner slots on kruise-managed cluster)"
echo "  burst demo: kubectl get dchain -l kbl.io/volcano-burst=true"

section "Julia finance wheel"
if kubectl get computewheel julia-finance-wheel &>/dev/null; then
  kubectl get computewheel julia-finance-wheel -o wide
  kubectl get wf -l kbl.io/computewheel=julia-finance-wheel -o custom-columns=NAME:.metadata.name,PHASE:.status.phase,CONTEXT:.spec.routing.computeContextRef 2>/dev/null || true
else
  echo "  (julia-finance-wheel not applied — set KBL_LAB_JULIA=1)"
fi

section "OpenKruise"
if kubectl -n kruise-system get deploy kruise-controller-manager &>/dev/null; then
  kubectl -n kruise-system get pods
  kubectl get wf julia-finance-openkruise -o wide 2>/dev/null || echo "  (julia-finance-openkruise workflow not applied)"
  kubectl get dchain julia-finance-openkruise-dchain -o wide 2>/dev/null || \
    kubectl get dchain julia-finance-openkruise -o wide 2>/dev/null || echo "  (openkruise dchain not applied)"
  kubectl get pods -l kbl.io/openkruise-demo=true 2>/dev/null || true
else
  echo "  (kruise-system not found — set KBL_LAB_OPENKURISE=1)"
fi

if [[ "$STRICT" == "1" ]]; then
  kubectl -n volcano-system get deploy volcano-scheduler -o jsonpath='{.status.readyReplicas}' 2>/dev/null | grep -qx '1' \
    || fail "volcano-scheduler is not ready"
  kubectl get computewheel -l kbl.io/volcano-demo=true --no-headers 2>/dev/null | grep -q . \
    || fail "no ComputeWheel with kbl.io/volcano-demo=true"
  kubectl get jobs.batch.volcano.sh --no-headers 2>/dev/null | grep -q . \
    || fail "no Volcano Jobs created"
  if kubectl get computewheel julia-finance-wheel &>/dev/null; then
    kubectl get computewheel julia-finance-wheel -o jsonpath='{.status.phase}' 2>/dev/null | grep -Eq 'Idle|Completed' \
      || fail "julia-finance-wheel is not Idle"
  fi
  if kubectl -n kruise-system get deploy kruise-controller-manager &>/dev/null; then
    kubectl -n kruise-system get deploy kruise-controller-manager -o jsonpath='{.status.readyReplicas}' 2>/dev/null | grep -qx '1' \
      || fail "kruise-controller-manager is not ready"
    phase=""
    phase="$(kubectl get dchain julia-finance-openkruise-dchain -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    if [[ -z "$phase" ]]; then
      phase="$(kubectl get dchain julia-finance-openkruise -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    fi
    if [[ -z "$phase" ]]; then
      phase="$(kubectl get wf julia-finance-openkruise -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    fi
    echo "$phase" | grep -Eq 'Completed' \
      || fail "julia-finance-openkruise is not Completed"
  fi
  echo ""
  echo "strict Volcano checks passed"
fi

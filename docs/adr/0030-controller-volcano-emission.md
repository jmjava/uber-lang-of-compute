# ADR 0030: Controller Volcano Job Emission

## Status

Accepted

## Context

Phase 25 installed Volcano in the Kind lab and applied a **static** `batch.volcano.sh/Job` manifest mirroring the DominoChain init-container handoff layout. The blog vision places Volcano in the **Provisioning** layer: the controller should emit scheduled batch work, not operators hand-writing VCJobs.

DominoChain already supports:

- `kubernetes-init` — standard Pod with init containers
- `openkruise` — placeholder Pod + ContainerRecreateRequest hot-swap

There was no runtime path for Volcano emission from the reconciler.

## Decision

### 1. New runtime: `volcano-init`

Extend `DominoChain.spec.runtime` with `volcano-init`. When selected, the DominoChain reconciler:

1. Ensures the snapshot ConfigMap (unchanged)
2. Creates a Volcano `Job` named `{chainName}-chain`
3. Polls `status.state.phase` until `Completed` or a terminal failure
4. Runs the same engine finalization as init-chain completion

### 2. Volcano Job shape

`dominochain.Builder.BuildVolcanoJob()` reuses `BuildInitChainPod()` pod template semantics:

- Single task `domino-chain` with init containers + pause container
- `schedulerName: volcano` on Job and task pod template
- `spec.queue` from optional `DominoChain.spec.volcanoQueue` (default `default`)
- Gang/complete policies: `TaskCompleted` → `CompleteJob`
- `nodeSelector`, runner image, Julia env parity inherited from init-chain builder

### 3. RBAC

Grant the controller `batch.volcano.sh/jobs` verbs in `lab/kustomize/base/rbac.yaml`.

### 4. Lab wiring

Replace static VCJob + snapshot ConfigMap under `lab/manifests/volcano/` with:

- `Queue` `kbl-lab` (unchanged)
- `DominoChain` `julia-finance-volcano` with `runtime: volcano-init`, `volcanoQueue: kbl-lab`

The controller creates the snapshot ConfigMap and Volcano Job on reconcile.

## Consequences

- Lab and production can drive Volcano from DominoChain CRs — no duplicate manifest maintenance
- Volcano CRD must be installed (`lab/scripts/install-volcano.sh` or Helm on EKS)
- `status.podName` stores the Volcano Job name for volcano-init chains (same field as Pod name for other runtimes)
- ComputeWheel → Volcano queue assignment implemented in Phase 27 (ADR 0031)

## References

- ADR 0007 — Hot-Swapped Dominos Implementation
- ADR 0029 — Volcano Kind Lab
- `controller/pkg/dominochain/volcano.go`
- `lab/manifests/volcano/dominochain-julia-finance.yaml`

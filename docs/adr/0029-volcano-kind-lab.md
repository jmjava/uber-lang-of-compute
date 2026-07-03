# ADR 0029: Volcano Batch Scheduler in Kind Lab

## Status

Accepted

## Context

The [Uber Language of Compute blog series](https://jmenke.blogspot.com/) describes Volcano in the **Provisioning** layer: SyncSets map to Volcano Jobs, Generations to time slices, and the Data Pond drives node priority for gang-scheduled domino chains via PodGroup semantics.

Through Phase 24 the repo shipped:

- **ComputeWheel** for application-level time-slice rotation (ADR 0006)
- **DominoChain** reconciler emitting standard Pods with init-container chains (ADR 0007)
- **Kind lab** with a single-node cluster (ADR 0026)

There was no Volcano install, no `batch.volcano.sh/Job`, no `PodGroup`, and no `schedulerName: volcano` in the lab path. Operators on a 64 GiB / 20 CPU host could not exercise the blog's batch-scheduling model locally.

## Decision

### 1. Multi-node Kind cluster

Extend `lab/kind/kind-config.yaml` to **1 control-plane + 2 workers**:

| Node | Labels | Mount |
|------|--------|-------|
| control-plane | `kbl.io/lab-role=control-plane` | `/tmp/kbl-lab/cp` → `/var/kbl` |
| worker 1 | `kbl.io/lab-role=compute` | `/tmp/kbl-lab/w1` → `/var/kbl` |
| worker 2 | `kbl.io/lab-role=compute`, `kbl.io/tsdb-node=true` | `/tmp/kbl-lab/w2` → `/var/kbl` |

The Kind overlay pins **kbl-tsdb** to `kbl.io/tsdb-node=true`, modeling node-local Data Pond placement.

### 2. Volcano install script

`lab/scripts/install-volcano.sh` applies the official Volcano manifest (default `KBL_VOLCANO_VERSION=v1.9.0`) and waits for `volcano-system` Deployments. `lab/scripts/up.sh` invokes it by default; set `KBL_LAB_VOLCANO=0` to skip.

### 3. Volcano demo manifests (`lab/manifests/volcano/`)

| Resource | Purpose |
|----------|---------|
| `Queue` `kbl-lab` | Lab queue with 20 CPU / 64 GiB capability |
| `ConfigMap` snapshot | Sealed inline instruments JSON (same shape as DominoChain ConfigMap) |
| `Job` `julia-finance-volcano` | Volcano Job with init-container Julia chain (`julia:identity` → `julia:interpolate` → `julia:greeks`), `schedulerName: volcano`, `nodeSelector: kbl.io/lab-role=compute` |

The Job mirrors `BuildInitChainPod` env/volume layout from `controller/pkg/dominochain/builder.go` but is applied directly in the lab (controller Volcano emission deferred).

### 4. Scope boundary

This phase is **lab-only**. The DominoChain reconciler continues to emit standard Pods. Future work may translate SyncSet/Generation CRs into Volcano Jobs and PodGroups from the controller.

## Consequences

- Lab operators can validate Volcano gang scheduling and queue semantics on modest hardware
- Multi-node Kind increases local resource use; recreate the cluster after upgrading from the single-node config
- Volcano install pulls a remote manifest; pin version via `KBL_VOLCANO_VERSION`
- Production EKS may adopt Volcano via Helm; Kind lab proves manifest and domino-chain parity first

## References

- ADR 0006 — Compute Wheel Rotation
- ADR 0007 — Hot-Swapped Dominos Implementation
- ADR 0026 — Kind Lab and AWS CDK
- [Volcano documentation](https://volcano.sh/)
- [Blog: Volcano search](https://jmenke.blogspot.com/search?q=volcano)

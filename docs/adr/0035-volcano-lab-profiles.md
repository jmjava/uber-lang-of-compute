# ADR 0035: Volcano Lab Profiles and Home Workstation Setup

## Status

Accepted

## Context

Phase 25–27 shipped Volcano in the Kind lab (queue + ComputeWheel → VCJob), but operators reported:

- The Volcano demo was hard to **see in action** (single context, no verification script).
- Lab docs assumed **64 GiB / 20 CPU** — too heavy for i7 laptops used on the road.
- The intended **home i9 + GPU** workstation had no profile, kubeconfig remote-access notes, or GPU node label for future dominos.

## Decision

### 1. Lab profiles (`KBL_LAB_PROFILE`)

| Profile | Kind config | Workers | Queue | Volcano demo |
|---------|-------------|---------|-------|--------------|
| `home` | `kind-config-home.yaml` | 2 (+ GPU label on w1) | 32 CPU / 48 GiB | 2-context wheel + parallel burst |
| `compact` | `kind-config-compact.yaml` | 1 | 8 CPU / 16 GiB | single-context wheel |
| `default` | `kind-config.yaml` (legacy) | 2 | 20 CPU / 64 GiB | same as home wheel |

Default for `lab/scripts/up.sh` is **`home`**.

### 2. Volcano demo enhancements

- **ComputeWheel** rotates `compute-a` → `compute-b` with `preProvisionNext: true` — two sequential VCJobs through queue `kbl-lab`.
- **Parallel burst** — `dominochain-volcano-burst.template.yaml` + `apply-volcano-burst.sh` submits two VCJobs pinned to different workers.
- **`verify-volcano.sh`** — queue state, VCJobs, PodGroups, pod `schedulerName: volcano`, node placement.

### 3. Home lab guide

`lab/HOME-LAB.md` — i9 setup, Docker sizing, remote kubectl from i7, GPU label semantics.

## Consequences

- Switching profiles requires cluster recreate (`down.sh` + `up.sh`).
- Burst demo requires two compute workers; skipped on `compact`.
- GPU label is documentation-only until GPU domino phases land.

## References

- [ADR 0029](0029-volcano-kind-lab.md), [ADR 0031](0031-computewheel-volcano-queue.md)
- [lab/HOME-LAB.md](../HOME-LAB.md)
- [lab/scripts/verify-volcano.sh](../scripts/verify-volcano.sh)

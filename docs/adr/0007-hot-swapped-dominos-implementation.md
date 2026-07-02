# ADR 0007: Hot-Swapped Dominos Implementation

## Status

Accepted

## Context

ADR 0004 proposed OpenKruise-based hot-swapped container dominos. Phase 4 implements in-cluster execution while preserving the existing local engine for memoization and replay logs.

## Decision

Introduce **DominoChain** CRD and reconciler with two runtimes:

### kubernetes-init (portable)
- Pod with sequential **initContainers**, one per domino step
- `emptyDir` handoff volume at `/kbl/handoff`
- Snapshot ConfigMap mounted at `/kbl/input`
- Works on any Kubernetes cluster without OpenKruise

### openkruise (hot-swap)
- Pod with **placeholder** containers (`pause:3.9`) named `slot-N`
- Controller issues OpenKruise **ContainerRecreateRequest** per step
- Swaps placeholder → `domino-runner` image with step command
- Only one slot active at a time; next slot pre-swappable (player-piano)

### domino-runner
- Small Go binary (`cmd/domino-runner`) used as container entrypoint
- Executes `builtin:*` commands against handoff files
- Same deterministic builtins as the local engine

### Workflow integration
- `spec.execution.runtime`: `local` | `kubernetes-init` | `openkruise`
- Non-local runtimes delegate to an owned DominoChain CR
- On chain completion, engine finalizes SQLite memo cache and replay log

## Consequences

- Container execution proves data handoff; engine still authoritative for hashes/replay
- OpenKruise mode requires `apps.kruise.io` CRD in cluster
- `kubernetes-init` validates full chain without OpenKruise dependencies
- Domino images can override `runnerImage` per chain or per step

## References

- ADR 0004 — Hot-Swapped Dominos (original proposal)
- *Hot swapping containers in daisy chain* (Apr 6, 2025)

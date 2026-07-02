# ADR 0006: Compute Wheel Time-Slice Rotation

## Status

Accepted

## Context

The blog series describes the Compute Wheel (Ferris Wheel) as a rotating set of compute contexts processing time slices continuously. Each seat on the wheel is a ComputeContext with its own node-local store. Workflows from ADR 0005 execute domino chains for a single snapshot — the wheel orchestrates *which* context runs *which* time slice *when*.

Player-piano scheduling pre-provisions the next slot while the current one executes.

## Decision

Introduce a **ComputeWheel reconciler** that:

1. Owns an ordered list of **ComputeContext** names
2. Maintains `status.currentTimeSlice` and `status.activeContextIndex`
3. Materializes a child **Workflow** per slot from `spec.workflowTemplate`
4. On Workflow completion:
   - Advances to the next context on the wheel
   - When all contexts complete a slice, advances the time slice by `spec.timeSliceInterval`
5. Records completed slots in `status.processedSlots`
6. Optionally pre-provisions the next Workflow when `spec.preProvisionNext: true`

Rotation logic lives in `pkg/wheel` (pure, testable). Store paths are derived from ComputeContext `spec.storePath` for data locality.

`spec.maxRotations` limits total slice rotations for demos and tests (0 = unlimited).

## Consequences

- ComputeWheel is the scheduling layer; Workflow remains the execution layer
- Two controllers cooperate: Wheel creates Workflows, Workflow executes domino chains
- CRD updated with `workflowTemplate`, `schedule`, `maxRotations`, `preProvisionNext`
- Standalone Snapshot/Domino CRDs still deferred; wheel uses inline workflow templates

## References

- *Higher Level Abstractions — Self Similarity pattern* (Jan 24, 2021)
- *Notes on the Uber language of compute* (Nov 10, 2021) — Ferris wheel, player-piano scheduling
- ADR 0005 — Workflow reconciler

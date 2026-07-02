# ADR 0013: Workflow References to Standalone Snapshot and Domino CRs

## Status

Accepted

## Context

ADR 0010 introduced standalone **Snapshot** and **Domino** reconcilers alongside the composed **Workflow** CRD, which embedded inline snapshot and domino specs. Teams adopting fine-grained CRs still had to duplicate specs when composing full chains in Workflow manifests.

Shared snapshots and dominos should be referenceable by name without copying YAML.

## Decision

1. **WorkflowSpec extensions** — optional `spec.snapshotRef` and `spec.dominoRefs[]` alongside existing inline `snapshot` and `dominos`
2. **Resolver** — `convert.ResolveEngineWorkflow(ctx, client, wf)` loads referenced CRs, validates seal/readiness, and builds the engine domain model
3. **Snapshot ID** — when resolving a Snapshot CR, use `status.snapshotID` from the reconciler (engine prefers status over recomputing hash)
4. **Validation** — domino `spec.snapshotRef` must match workflow `spec.snapshotRef` when both are set; inline path unchanged when refs are absent
5. **Pending requeue** — workflow reconciler sets phase `Pending` and requeues when the referenced snapshot is not yet sealed
6. **RBAC** — workflow reconciler gains `get/list/watch` on `snapshots` and `dominos`

Inline and reference-based specs are mutually composable at the API level; at least one snapshot source and one domino source must be present at reconcile time.

## Consequences

- Existing inline Workflow manifests require no changes
- Reference-based workflows depend on Snapshot/Domino CR lifecycle ordering
- Container/hot-swap execution (`DominoChain`) still uses inline conversion via `dominochain.FromWorkflow` until extended in a follow-up
- CRD `required` fields relax to `[execution]` only; validation occurs in the reconciler/resolver

## References

- ADR 0010 — Standalone Snapshot and Domino Reconcilers
- `controller/pkg/convert/resolve.go`
- Example: `examples/workflow-snapshot-refs/`

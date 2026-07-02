# ADR 0016: ComputeWheel Workflow Template CR References

## Status

Accepted

## Context

ADR 0006 introduced ComputeWheel with inline `workflowTemplate.snapshot` and `workflowTemplate.dominos`. ADRs 0010 and 0013 added standalone Snapshot/Domino CRs and Workflow reference fields, but the wheel still required duplicating specs in every wheel manifest.

Teams operating shared snapshot and domino libraries need wheels to stamp child Workflows that reference existing CRs.

## Decision

1. **WorkflowTemplateSpec extensions** — optional `snapshotRef` and `dominoRefs[]`, matching WorkflowSpec
2. **`wheel.BuildWorkflow`** — when refs are set, materializes child Workflows with `spec.snapshotRef` / `spec.dominoRefs` instead of inline specs; inline path unchanged (still stamps `timeSlice` per slot)
3. **CRD** — relax `workflowTemplate` required fields to `[execution]` only
4. **Execution unchanged** — Workflow reconciler resolves refs via existing `convert.ResolveEngineWorkflow`

When using CR refs, operators supply sealed Snapshot CRs appropriate for the wheel slot; the wheel does not mutate referenced Snapshot time slices.

## Consequences

- Inline wheel templates remain the default for per-slice data stamping
- Reference-based wheels depend on Snapshot/Domino CR lifecycle (same as Workflow refs)
- Pre-provisioned workflows inherit refs from the template

## References

- ADR 0006 — Compute Wheel Rotation
- ADR 0013 — Workflow CR References
- `controller/pkg/wheel/workflow_builder.go`
- Example: `examples/compute-wheel/wheel-refs.yaml`

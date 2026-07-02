# ADR 0014: DominoChain Resolution for Workflow CR References

## Status

Accepted

## Context

ADR 0013 added `spec.snapshotRef` and `spec.dominoRefs` to Workflow, with `convert.ResolveEngineWorkflow` powering the local execution path. Container and hot-swap execution still built DominoChain specs from inline `spec.snapshot` and `spec.dominos` via `dominochain.FromWorkflow`, leaving reference-based workflows unable to use `kubernetes-init` or `openkruise` runtimes.

## Decision

1. **`dominochain.FromWorkflow(ctx, client, wf, storePath)`** — resolves CR refs via `convert.ResolveEngineWorkflow`, then builds the DominoChain spec from the engine model
2. **`dominochain.FromEngineWorkflow`** — shared builder for resolved engine workflows (snapshot inline data copied into chain spec for ConfigMap mount)
3. **`NeedsContainerRuntime(ctx, client, wf)`** — also inspects Domino CR refs for non-empty `spec.image`
4. **Workflow reconciler** — `executeContainer` uses the resolver; requeues `Pending` when the referenced snapshot is not yet sealed (same as local path)

Inline Workflow manifests continue to work unchanged through the resolver's inline fallback.

## Consequences

- Reference-based workflows can run in-cluster via DominoChain
- DominoChain still embeds resolved inline snapshot data in its spec (not a snapshotRef field on DominoChain itself)
- Path/URI snapshot sources on referenced Snapshot CRs still hash metadata only until node-local ingestion lands

## References

- ADR 0013 — Workflow CR References
- `controller/pkg/dominochain/convert.go`
- Example: `examples/workflow-snapshot-refs/workflow-container.yaml`

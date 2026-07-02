# ADR 0010: Standalone Snapshot and Domino Reconcilers

## Status

Accepted

## Context

ADR 0005 introduced the **Workflow** CRD as the primary reconciliation unit, embedding inline snapshots and domino chains. CRDs for standalone **Snapshot** and **Domino** resources already existed but had no Go types or reconcilers.

Fine-grained resources enable:

- Shared snapshots referenced by multiple dominos or workflows
- Independent domino scheduling with explicit dependency edges
- Incremental adoption without rewriting existing Workflow manifests

## Decision

1. **Snapshot CR types + reconciler** — when `spec.sealed=true`, validate source, compute deterministic `status.snapshotID`, persist to node-local store via `store.OpenForSnapshot`
2. **Domino CR types + reconciler** — execute one domino via `engine.RunSingle` against a sealed Snapshot; load dependency outputs from store; memo cache yields `Cached` phase
3. **Engine extension** — `RunSingle(snapshotID, snap, domino, priorOutputs)` for standalone execution
4. **Store helpers** — `OpenForSnapshot`, `OpenForDomino` keyed by snapshot ref namespace/name
5. **Workflow unchanged** — composed Workflow path remains default; standalone CRDs are additive

Domino `spec.storePath` overrides the default `{store-root}/{namespace}/{snapshotRef}.db` path.

## Consequences

- Two execution paths coexist: composed Workflow chains and fine-grained Snapshot→Domino graphs
- Domino reconciler requeues every 5s while waiting for snapshot seal or dependency completion
- Path/URI snapshot sources hash metadata (not file contents) until node-local ingestion is added
- Container/hot-swap dominos remain Workflow/DominoChain scope for now

## References

- ADR 0005 — Kubernetes Controller Reconciler
- CRDs: `crds/snapshot.yaml`, `crds/domino.yaml`
- Example: `examples/standalone-snapshot-domino/`

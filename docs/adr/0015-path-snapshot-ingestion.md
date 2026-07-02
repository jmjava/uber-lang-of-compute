# ADR 0015: Node-Local Path Snapshot Ingestion

## Status

Accepted

## Context

ADR 0010 noted that `source.path` and non-file `source.uri` values contributed only metadata to snapshot ID hashing — not file contents. That prevented workflows from referencing large datasets mounted on compute nodes without duplicating them into CR manifests.

## Decision

1. **`snapshot.ResolveContent`** — loads inline data as-is; reads and parses node-local files for `source.path` and `file://` URIs
2. **JSON files** — parsed and hashed as structured data (same semantics as inline)
3. **Non-JSON files** — wrapped as `{ "path": "...", "raw": "..." }` for deterministic hashing
4. **Remote URIs** — unchanged metadata-only behavior until remote fetch is implemented
5. **Reconcilers** — Snapshot, Workflow, and DominoChain paths use resolved content for ID computation, store persistence, and domino inputs
6. **Pending requeue** — Snapshot reconciler stays `Pending` when the path file is not yet available

## Consequences

- Snapshot IDs change for existing path-based specs (now content-addressed, not path-metadata-only)
- Operators must ensure files exist on nodes before sealing path-based snapshots
- File reads occur at reconcile/execute time on the controller pod's filesystem (node-local paths assume hostPath or co-located data)

## References

- ADR 0010 — Standalone Snapshot and Domino Reconcilers
- `controller/pkg/snapshot/content.go`
- Example: `examples/path-snapshot/`

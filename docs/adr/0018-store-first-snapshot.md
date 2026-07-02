# ADR 0018: Store-First Snapshot Content on Hot Path

## Status

Accepted

## Context

Phases 12 and 15 ingest snapshot data from node-local paths and HTTP URIs at seal time. The engine previously re-resolved sources (including HTTP fetches) on every workflow run, which is slow and contradicts the framework goal: **bring compute to the data**.

ADR 0017 noted HTTP is bootstrap-only; production hot paths need local, cached access.

## Decision

1. **`snapshot.LoadContentPreferStore`** — when `snapshotID` is known, read persisted JSON from `store.Backend` first
2. **`snapshot.ResolveEngineContentPreferStore`** — store hit skips inline/path/HTTP resolution entirely
3. **Engine** — `Run`, `resolveInputs`, and `RunSingle` use store-first resolution
4. **Seal path unchanged** — Snapshot reconciler still ingests from source once; subsequent runs hit the store

Domino execution therefore reads snapshot bytes already on the node-local store/TSDB, not over REST on every chain run.

## Consequences

- First seal still pays ingestion cost (HTTP/path); hot reruns and memoized chains are fast
- Store must contain snapshot data before execute (Snapshot reconciler or prior run)
- Future work: zero-copy mmap, sidecar staging, TSDB streaming reads (README Phase 17)

## References

- ADR 0015 — Path Snapshot Ingestion
- ADR 0017 — HTTP Snapshot Ingestion
- `controller/pkg/snapshot/store_content.go`

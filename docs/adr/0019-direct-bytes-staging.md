# ADR 0019: Direct-Bytes Snapshot Staging

## Status

Accepted

## Context

Phases 12 and 15 read path/HTTP sources at seal time via `ResolveContent` → `json.Marshal`, which parsed JSON and re-serialized it — an extra copy and CPU cycle on the seal path. Phase 16 moved hot execution to store-first reads; seal-time staging should also minimize work.

## Decision

1. **`snapshot.SealPayload`** — single-pass seal: compute ID and store bytes together
2. **JSON path/HTTP sources** — hash from parsed content but **persist original bytes** (no parse→remarshal)
3. **Snapshot reconciler** — one `SealPayload` call instead of separate `ComputeID` + `MarshalData` (which read sources twice)
4. **Non-JSON sources** — wrapped as `{path|uri, raw}` with one marshal for storage

Combined with Phase 16 store-first reads, the pipeline is: **ingest once (direct bytes) → persist → hot path reads store only**.

## Consequences

- Snapshot store bytes may preserve source formatting (whitespace, key order) while IDs remain content-addressed via parsed JSON
- mmap and TSDB streaming deferred to Phase 18

## References

- ADR 0018 — Store-First Snapshot
- `controller/pkg/snapshot/staging.go`

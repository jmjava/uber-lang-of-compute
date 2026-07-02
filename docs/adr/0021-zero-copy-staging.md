# ADR 0021: Zero-Copy Snapshot Staging

## Status

Accepted

## Context

Phase 18 added mmap reads (≥1 MiB) and TSDB data sidecars, but large path seals still copied mapped bytes into the Go heap before persist, TSDB envelopes duplicated full payload in `{id}.json`, and the `/data` endpoint buffered sidecars in memory before writing.

ADR 0020 noted true zero-copy as deferred work.

## Decision

1. **`readPathBytesForSeal`** — mmap-backed views with explicit `release()`; seal path writes mapped bytes directly to TSDB sidecars
2. **`SealToBackend`** — reconciler seals and persists in one step via `SaveSnapshotPayload` without `string(raw)` conversion
3. **TSDB metadata-only envelopes** — `{id}.json` holds metadata; payload lives only in `{id}.data.json`
4. **`OpenSnapshotData`** — `/data` handler streams sidecar files with `io.Copy`
5. **`SnapshotPayloadSaver`** — optional store interface for byte-native persistence

## Consequences

- Large path seals avoid heap copy on TSDB backends (SQLite still converts once for TEXT storage)
- Backward compatible: `GetSnapshot` loads sidecar when envelope `data` is empty; legacy envelopes with embedded data still work
- Execute path still loads full JSON strings for domino parsing (deferred)

## References

- ADR 0020 — mmap + TSDB Streaming
- `controller/pkg/snapshot/staging.go`
- `controller/pkg/store/tsdb_engine.go`

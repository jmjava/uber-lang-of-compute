# ADR 0020: mmap Path Reads and TSDB Snapshot Data Streaming

## Status

Accepted

## Context

Phases 16–17 optimized execute and seal paths, but large path files still allocated full heap copies via `os.ReadFile`, and TSDB clients unmarshaled full snapshot envelope records to access payload bytes.

## Decision

1. **mmap path reads (Unix, ≥1 MiB)** — `ReadPathBytes` maps large node-local files before copying once for seal hashing
2. **TSDB data sidecars** — `SaveSnapshot` writes `{id}.data.json` alongside envelope metadata
3. **`GET /v1/snapshots/{id}/data`** — streams raw payload bytes without JSON envelope parsing
4. **`SnapshotDataGetter`** — optional store interface; TSDB client and SQLite implement `GetSnapshotData`
5. **`LoadDataPreferStore`** — hot path fetches bytes first, parses JSON once when content is needed

## Consequences

- mmap still copies once into Go heap for seal (true zero-copy deferred)
- Sidecar files add disk usage proportional to snapshot payload size
- Non-Unix builds fall back to `os.ReadFile`

## References

- ADR 0019 — Direct-Bytes Staging
- `controller/pkg/snapshot/mmap_unix.go`
- `controller/pkg/store/tsdb_engine.go`

# ADR 0017: HTTP/HTTPS Snapshot URI Ingestion

## Status

Accepted

## Context

ADR 0015 added node-local path and `file://` ingestion. Snapshot `source.uri` values using `http://` or `https://` still contributed only metadata to snapshot IDs, blocking workflows that reference remotely hosted immutable datasets.

## Decision

1. **`snapshot.loadHTTP`** — fetch `http://` and `https://` URIs with a 30s client timeout and 32 MiB body limit
2. **JSON parsing** — same semantics as path ingestion (structured JSON or `{uri, raw}` wrapper)
3. **Transient failures** — `IsURINotReady` detects 502/503/504/429, timeouts, and connection errors; reconcilers requeue `Pending`
4. **Other schemes** — `s3://`, `gs://`, etc. remain metadata-only until dedicated adapters exist
5. **`IsSourceNotReady`** — unified helper covering path and HTTP transient errors

## Consequences

- Snapshot IDs for HTTP-based specs become content-addressed
- Operators must ensure URIs are reachable from the controller at seal time
- No authentication headers yet — public or in-cluster URLs only for now
- **HTTP fetch is not the performance target** — REST ingestion is a bootstrap/convenience path; Phase 16 will prioritize node-local staging and zero-copy access for production hot paths (see README roadmap)

## References

- ADR 0015 — Path Snapshot Ingestion
- `controller/pkg/snapshot/content.go`
- Example: `examples/http-snapshot/`

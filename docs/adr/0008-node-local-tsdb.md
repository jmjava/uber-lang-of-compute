# ADR 0008: Node-Local TSDB DaemonSet

## Status

Accepted

## Context

ADR 0003 specified SQLite for MVP and TSDB DaemonSet as the production node-local data layer. Phases 1–4 proved compute physics with SQLite files. Phase 5 adds the TSDB path described in *Minimize Entropy while maximizing Caching*.

## Decision

1. **Store interface** (`store.Backend`) abstracts SQLite and TSDB backends
2. **kbl-tsdb** — lightweight HTTP TSDB server (`cmd/kbl-tsdb`) with file-backed:
   - `snapshots/` — immutable snapshot records
   - `memo/` — `(snapshot_id, domino_id, input_hash)` → result
   - `replay/` — append-only replay entries
3. **DaemonSet** (`deploy/node-local-tsdb/`) runs kbl-tsdb on each node with `hostNetwork: true` and `hostPath` data directory
4. **ComputeContext** gains `storeEndpoint`, `dataPath`; status exposes live endpoint and cache stats
5. **ComputeContext reconciler** pings TSDB health and publishes `status.storeEndpoint`
6. **Controllers** resolve store via `store.OpenForWorkflow` / `OpenForDominoChain`, honoring `ComputeContext.spec.storeType`

Workflow `provisioning.storePath` may be an `http://` URL to select TSDB directly for development.

## Consequences

- Engine and memoization logic unchanged; only persistence layer varies
- TSDB server is intentionally simple (not Prometheus/VictoriaMetrics) — swappable later
- Cross-node replication deferred to Debezium/Kafka (future phase)
- CLI `--store` flag still opens SQLite by default; use workflow YAML with HTTP storePath for TSDB

## References

- ADR 0003 — Node-Local Data
- *Minimize Entropy while maximizing Caching* (Apr 11, 2025)

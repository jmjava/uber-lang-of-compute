# ADR 0011: Read-Replica Materialization

## Status

Accepted

## Context

ADR 0009 introduced Multiverse routing and recorded routed snapshot events in `Multiverse.status.routedEvents`. Cross-universe **read replicas** — read-only copies of snapshot data and domino results in a target Pluggable Universe — were deferred.

Compute remains node-local (ADR 0003); read replicas enable downstream universes to query materialized results without re-running the source workflow.

## Decision

1. **ReadReplica CRD** — created by the Multiverse reconciler when a snapshot event is routed
2. **ReadReplica reconciler** — copies snapshot + domino chain results from source workflow store to target universe store
3. **`pkg/replica.Materialize`** — copies `GetSnapshot` + `GetLatestResult` per domino in source workflow chain
4. **Store extension** — `GetLatestResult(snapshotID, dominoID)` on all backends for full result replication
5. **Target store path** — `{store-root}/{namespace}/replicas/{targetUniverse}.db` unless `targetComputeContextRef` resolves a ComputeContext store

Multiverse handler flow: route event → append `status.routedEvents` → create ReadReplica CR (owner reference to Multiverse).

## Consequences

- Read replicas are eventually consistent; materialization runs asynchronously after routing
- Source workflow must exist and have completed (store populated) before ReadReplica can succeed
- TSDB cross-store replication uses HTTP `GetLatestResult` endpoint
- Full Debezium CDC can replace direct store copy in production (ADR 0009)

## References

- ADR 0009 — Multiverse Routing
- ADR 0003 — Node-Local Data
- `crds/readreplica.yaml`

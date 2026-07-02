# ADR 0012: Debezium CDC Cross-Universe Sync

## Status

Accepted

## Context

ADR 0011 materializes read replicas via direct store copy (`replica.Materialize`). Production multiverse deployments expect **Debezium-compatible CDC** over Kafka — snapshot and domino result rows replicated as change events, not pulled from source stores at reconcile time.

ADR 0009 deferred CDC in favor of routing audit trails. Phase 8 proved direct copy; Phase 9 adds the CDC path.

## Decision

1. **`pkg/cdc`** — Debezium-style envelopes for `snapshots` and `domino_results` tables
2. **Workflow CDC publish** — when Multiverse `spec.sync.enabled`, export completed workflow results as CDC events to `kbl.cdc.snapshots` (Kafka or in-memory bus)
3. **ReadReplica `replicationMode`** — `direct` (default) or `cdc`
4. **CDC consumer** — ReadReplica reconciler applies Kafka/memory CDC events to target store via `cdc.SyncFromConsumer`
5. **Multiverse integration** — when sync enabled, ReadReplica CRs default to `replicationMode: cdc`

In-memory CDC bus (`cdc.DefaultMemory()`) supports single-controller dev without Kafka.

## Consequences

- Two replication paths coexist: direct (Phase 8) and CDC (Phase 9)
- CDC mode requeues until all domino chain events arrive
- Real Debezium connectors can replace workflow-exported events in production
- Kafka CDC consumer uses short poll window; incomplete sync requeues after 5s

## References

- ADR 0011 — Read-Replica Materialization
- ADR 0009 — Multiverse Routing
- `pkg/cdc/`, `examples/multiverse-finance/multiverse.yaml`

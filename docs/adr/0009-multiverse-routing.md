# ADR 0009: Multiverse Routing via Kafka

## Status

Accepted

## Context

The 2020–2025 Uber Language of Compute design describes a **Multiverse** — routed, time-sliced replicas across **Pluggable Universes**. ADR 0001 defines Routing as a first-class DSL. Prior phases proved single-universe execution; Phase 6 adds cross-universe routing and event sync.

## Decision

1. **Multiverse CRD** — routing table: default universe, partition rules, time-slice overrides, optional Kafka sync
2. **PluggableUniverse CRD** — Go types + reconciler (CRD yaml existed; now fully wired)
3. **Event bus** (`pkg/events`) — `MemoryBus` (default) and `KafkaBus` (segmentio/kafka-go)
4. **Router** (`pkg/routing`) — resolves snapshot events to universe + compute context
5. **Workflow integration** — publishes `kbl.snapshot.completed` on chain completion
6. **Multiverse reconciler** — subscribes to bus, records `status.routedEvents`

Kafka topic default: `kbl.snapshot.events` (Debezium-compatible JSON payloads).

Workflow partition labels: `kbl.io/partition-<key>: <value>`.

Controller flags: `--kafka-brokers`, `--kafka-topic`.

## Consequences

- Cross-universe fan-out is event-driven; compute remains node-local (ADR 0003)
- Kafka optional — single-controller dev uses MemoryBus
- Read-replica materialization deferred; routing audit trail in Multiverse status
- Debezium CDC from TSDB → Kafka can replace direct publish in production

## References

- ADR 0001 — Four-DSL Model
- *Visualization: A Multiverse of Computation* (Apr 11, 2025)
- specs/routing-dsl.yaml

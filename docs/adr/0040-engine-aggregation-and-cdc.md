# ADR 0040: Compute-engine aggregation and CDC remaining work

## Status

Accepted — **Phases 37–40 verified**. M1–M32 stay frozen. Courseforge (Phase 32) is skipped.

## Context

Debezium was never proven as Debezium: `pkg/cdc` is a Debezium-*shaped* envelope over **MemoryBus**. Kafka clients existed; Kind does not run Strimzi or a connector. Phase 39 runs Redpanda via `lab/compose/research.yaml` and round-trips those envelopes.

Windowed Mandelbrot was a shape theorem (`Unfold` / `Coarsen`, F1). `WorkflowFromUnfold` ran as a live chain, but every node was `builtin:identity` — parents did not aggregate children.

## Decision

Engineering phases only (not theory milestones):

| Phase | Work | Proof |
|-------|------|-------|
| **37** | `builtin:coarsen` sums child `value`/`v`. Unfold parents coarsen; leaves identity the snapshot. | `TestCoarsenSumsChildValues`, `TestWorkflowFromUnfoldAggregatesToLeafCount` |
| **38** | Same chain as a sealed Workflow against catalog/lab data (local engine on Kind). | Compact `unfold-lab` Workflow; replay root `v=2`; Unfold still does not spawn wheel contexts |
| **39** | **Keep Kafka.** It is the bus Debezium already uses (retention, consumer groups, replay). Engine CDC envelopes round-trip on a real broker. This is not a substitute for Debezium capture. | `TestKafkaCDCRoundTrip` against Redpanda (`make research-kafka-test`) |
| **40** | Real Debezium connector **on that Kafka bus**, only when the store has a WAL Debezium can tail (not SQLite/kbl-tsdb HTTP). Compact Kind is not required. Strimzi is Kafka-on-K8s; Redpanda is the bus we already keep. | `TestDebeziumCapturesSealedSnapshot` (`make research-debezium-test`) |

Non-goals: \(z\mapsto z^2+c\) as scheduler, fractal pod spawn, Unfold spawning ComputeWheel contexts, joules, Everett.

## Consequences

- Continual aggregation is an engine fold-up: root `v` = leaf count × leaf `v` for a unit snapshot.
- Self-similarity of the live chain: `agg(d,k) = agg(d-1,k) * k` (LeafCount identity), distinct from pattern F1 on Unfold node values.
- CDC MemoryBus remains the default in-process path. Phase 39 proved engine envelopes on Kafka. Phase 40 proved Debezium's Postgres connector capturing sealed snapshot rows onto that same bus.

## References

- ADR 0012 — Debezium CDC envelopes
- ADR 0037 — Physics correspondence (F1 stays a pattern theorem)
- ADR 0038 — no new `kbl.io` kinds

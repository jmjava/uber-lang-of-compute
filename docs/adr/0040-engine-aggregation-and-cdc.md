# ADR 0040: Compute-engine aggregation and CDC remaining work

## Status

Accepted — **Phase 37 verified** (engine tests). Phases 38–40 scheduled. M1–M32 stay frozen. Courseforge (Phase 32) is skipped.

## Context

Debezium was never proven as Debezium: `pkg/cdc` is a Debezium-*shaped* envelope over **MemoryBus**. Kafka clients exist; Kind does not run Strimzi or a connector.

Windowed Mandelbrot was a shape theorem (`Unfold` / `Coarsen`, F1). `WorkflowFromUnfold` ran as a live chain, but every node was `builtin:identity` — parents did not aggregate children.

## Decision

Engineering phases only (not theory milestones):

| Phase | Work | Proof |
|-------|------|-------|
| **37** | `builtin:coarsen` sums child `value`/`v`. Unfold parents coarsen; leaves identity the snapshot. | `TestCoarsenSumsChildValues`, `TestWorkflowFromUnfoldAggregatesToLeafCount` |
| **38** | Same chain as a sealed Workflow against catalog/lab data (local engine on Kind). | Compact `finance-lab`-style Workflow; no new CRDs; Unfold still does not spawn wheel contexts |
| **39** | CDC Kafka round-trip in the engine (`KafkaBus` + `pkg/cdc` envelopes) without a Debezium connector. | Tests against a real broker or documented skip if none |
| **40** | Real Debezium connector + Strimzi **only** when Phase 39 is green and the cluster can hold Kafka. Compact Kind is not required to take this. | Connector capture of sealed snapshot rows |

Non-goals: \(z\mapsto z^2+c\) as scheduler, fractal pod spawn, Unfold spawning ComputeWheel contexts, joules, Everett.

## Consequences

- Continual aggregation is an engine fold-up: root `v` = leaf count × leaf `v` for a unit snapshot.
- Self-similarity of the live chain: `agg(d,k) = agg(d-1,k) * k` (LeafCount identity), distinct from pattern F1 on Unfold node values.
- CDC remains MemoryBus until Phase 39/40.

## References

- ADR 0012 — Debezium CDC envelopes
- ADR 0037 — Physics correspondence (F1 stays a pattern theorem)
- ADR 0038 — no new `kbl.io` kinds

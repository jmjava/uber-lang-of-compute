# ADR 0037: Physics Correspondence Layer

## Status

Accepted

## Context

The Uber Language of Compute blog series (2020–2025) named the fabric with physics vocabulary: low-entropy snapshots, Newtonian determinism, a routed multiverse, a Ferris-wheel scheduler, a Windowed Mandelbrot pattern, and a Kubernetes Based Lifeform. The implementation proved the *engineering* counterparts (seal gates, hash memoization, wheel rotation, Kafka routing) but left the physics words as identity claims.

A journal-style reading flags that as a category error: entropy was undefined, Mandelbrot was not an iterated map, multiverse invited Everett without amplitudes, `deterministic: true` was a flag the engine ignored, replica/CDC paths would copy unsealed snapshots, and snapshot IDs were 64-bit truncations of SHA-256.

Abandoning the analogies would discard the project's intellectual spine. Treating them as laws of nature would not survive review. The remaining option is the **neural-net move**: copy a structure nature already found useful, discard the wetware, and say which is which ([docs/theory/nature-inspired.md](../theory/nature-inspired.md)).

## Decision

1. **Nature-inspired, not identity.** Physics and biology terms in KBL are usable abstractions copied from nature, documented in [docs/theory/correspondences.md](../theory/correspondences.md) and [nature-inspired.md](../theory/nature-inspired.md). The remainder stays in nature (glia, joules, amplitudes, DNA).
2. **Formal model.** Discrete state space, hypotheses H1–H6, theorems, and a *graded* remainder live in [docs/theory/formal-model.md](../theory/formal-model.md).
3. **Executable counterparts.** `controller/pkg/theory` implements ensemble entropy, causal past, windowed aggregation, regularity classes, logical work, classical histories, and homeostasis.
4. **Patch the invariants the mimic requires.**
   - Snapshot IDs are 128-bit (`hash.SnapshotIDHexLen = 32`).
   - Cross-universe materialization and CDC export/apply refuse unsealed snapshots (Cauchy signaling).
   - `deterministic: true` is a Picard-style regularity contract; builtins discharge it as a theorem, Julia as a pinned discretization.
5. **Residual analogies are graded, not declined.** Landauer is bookkeeping of *steps*; Everett is non-interfering classical histories; \(z\mapsto z^2+c\) is an explorer iteration window; lifeform is spec/status homeostasis.
6. **Related work** includes nature-inspired computing (McCulloch–Pitts, Rosenblatt) as well as systems literature.

## Consequences

- Vision, vocabulary, and architecture text must point at the correspondence layer so "entropy" cannot be read as \(k\log W\) by accident.
- Windowed Mandelbrot is now a proven *pattern* (`Unfold` / `Coarsen`); wiring it into the cluster scheduler remains future work and is labeled as such.
- A nature-inspired computing paper can be extracted from the mimics plus theorems D1–R1. A physics venue should not receive this work as a claim about nature.

## References

- ADR 0001 Four-DSL Model
- ADR 0002 Snapshot Isolation
- ADR 0003 Node-Local Data
- ADR 0006 Compute Wheel
- ADR 0009 Multiverse Routing
- [docs/theory/peer-review.md](../theory/peer-review.md)
- Lamport (1978); Shannon (1948); Bennett (1973); Mandelbrot (1982) — see related-work.md

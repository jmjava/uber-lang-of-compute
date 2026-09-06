# ADR 0037: Physics Correspondence Layer

## Status

Accepted

## Context

The Uber Language of Compute blog series (2020–2025) named the fabric with physics vocabulary: low-entropy snapshots, Newtonian determinism, a routed multiverse, a Ferris-wheel scheduler, a Windowed Mandelbrot pattern, and a Kubernetes Based Lifeform. The implementation proved the *engineering* counterparts (seal gates, hash memoization, wheel rotation, Kafka routing) but left the physics words as identity claims.

A journal-style reading flags that as a category error: entropy was undefined, Mandelbrot was not an iterated map, multiverse invited Everett without amplitudes, `deterministic: true` was a flag the engine ignored, replica/CDC paths would copy unsealed snapshots, and snapshot IDs were 64-bit truncations of SHA-256.

Abandoning the analogies would discard the project's intellectual spine. Treating them as laws of nature would not survive review. The remaining option is to **prove the loose linkage**: each name becomes a correspondence with transferred properties, explicit limits, and tests.

## Decision

1. **Correspondence, not identity.** Physics terms in KBL are names of structure-preserving maps, documented in [docs/theory/correspondences.md](../theory/correspondences.md). What does not transfer is stated in the same document.
2. **Formal model.** Discrete state space, hypotheses H1–H6, and theorems D1, D2, M1, E1, W1, F1, R1, C1 live in [docs/theory/formal-model.md](../theory/formal-model.md).
3. **Executable counterparts.** `controller/pkg/theory` implements ensemble entropy, causal past, and windowed aggregation. Engine, wheel, routing, replica, and CDC tests are named after the theorems.
4. **Patch the invariants the analogy requires.**
   - Snapshot IDs are 128-bit (`hash.SnapshotIDHexLen = 32`).
   - Cross-universe materialization and CDC export/apply refuse unsealed snapshots (Cauchy signaling).
   - `deterministic: true` is documented as a purity *contract* the engine cannot discharge for arbitrary commands; the seal gate remains mandatory regardless.
5. **Decline three correspondences as theorems:** thermodynamic Landauer heat, Everett many-worlds, and the Mandelbrot polynomial \(z\mapsto z^2+c\). Keep memoization, product-of-systems routing, and coarsenable windowed aggregation instead.
6. **Related work** is scholarly, not only the blog ([docs/theory/related-work.md](../theory/related-work.md)).

## Consequences

- Vision, vocabulary, and architecture text must point at the correspondence layer so "entropy" cannot be read as \(k\log W\) by accident.
- Windowed Mandelbrot is now a proven *pattern* (`Unfold` / `Coarsen`); wiring it into the cluster scheduler remains future work and is labeled as such.
- A systems paper can be extracted from the fabric plus theorems D1–R1 without any physics venue.
- A physics venue should not receive this work as a claim about nature.

## References

- ADR 0001 Four-DSL Model
- ADR 0002 Snapshot Isolation
- ADR 0003 Node-Local Data
- ADR 0006 Compute Wheel
- ADR 0009 Multiverse Routing
- [docs/theory/peer-review.md](../theory/peer-review.md)
- Lamport (1978); Shannon (1948); Bennett (1973); Mandelbrot (1982) — see related-work.md

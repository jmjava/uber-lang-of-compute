# ADR 0002: Snapshot Isolation

## Status

Accepted

## Context

Deterministic, reproducible computation requires that inputs do not change mid-pipeline. The blog series uses the metaphor of "low-entropy snapshots" — frozen data views that make computation repeatable. Finance use cases (curve/risk calculations) demand that the same snapshot always produces the same curve.

Without snapshot isolation, memoization caches become unreliable and replay logs lose meaning.

## Decision

Every domino execution is bound to exactly one `Snapshot` resource:

- A Snapshot has a unique ID, a time slice identifier, and an immutable data payload (or reference to node-local storage)
- Dominos declare `snapshotRef` and may only read data from that snapshot plus outputs of prior dominos in the same chain
- Cross-snapshot reads are forbidden during execution
- Snapshot data is write-once: once sealed, the payload cannot be modified

The controller verifies snapshot seal status before executing any domino in the chain.

## Consequences

- Replay is trivial: re-run the chain against the same snapshot ID and compare replay logs
- Memoization keys include snapshot ID, ensuring cache entries are scoped correctly
- Storage must support point-in-time or immutable views (SQLite file copy, TSDB retention policy, or object-store versioning)

## References

- *Newtonian Physics, entropy, computational repeatability, and determinism* (Apr 10, 2025)
- *Applicability to Finance* (Apr 15, 2025)

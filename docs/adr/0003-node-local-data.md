# ADR 0003: Node-Local Data

## Status

Accepted

## Context

The KBL architecture principle is "bring compute to the data." Moving large datasets across the network for every computation violates data locality and increases latency and cost. The blog series describes node-local MemSQL, TSDB DaemonSets, and Lustre/FSx staging layers as progressively refined data layers.

## Decision

Each Compute Context owns node-local storage for:

1. Snapshot data (inputs)
2. Domino intermediate and final outputs
3. Memoization cache (input hash → output hash + payload)

**MVP:** SQLite database on a local filesystem path (`--store-path` flag or `emptyDir` volume).

**Target:** Node-local TSDB DaemonSet with hash-based memoization tables, optionally synced via Debezium/Kafka for cross-context read replicas (not for compute-time reads).

Scheduling: dominos for a given snapshot should run on the node where the snapshot data resides. The controller sets `nodeName` or uses node affinity derived from the Snapshot's `computeContextRef`.

## Consequences

- Domino chains for one snapshot are node-bound; cross-node fan-out requires explicit routing DSL configuration
- Node failure loses local cache (acceptable for MVP; target adds replication)
- SQLite is sufficient for MVP proof; production moves to TSDB with retention policies

## References

- *Optimizations in Time based Windowed APIs* (Jan 12, 2020)
- *Where is the mandelbrot?* (Jan 25, 2021)
- *HPC in Kubernetes* (Jul 17, 2021)
- *Looking under the hood at AWS* (Jan 3, 2022)
- *Minimize Entropy while maximizing Caching* (Apr 11, 2025)

# Node-Local TSDB (Target Architecture)

## MVP Status

The MVP uses **SQLite** on a node-local filesystem path as the store backend. This proves data locality and memoization without requiring a Kubernetes cluster.

## Target Architecture

From *Minimize Entropy while maximizing Caching* (Apr 11, 2025):

```
┌─────────────────────────────────────────┐
│           Kubernetes Node              │
│  ┌─────────────────────────────────┐   │
│  │  TSDB DaemonSet Pod             │   │
│  │  (node-local time-series DB)    │   │
│  │  - snapshot tables              │   │
│  │  - memo cache (hash → result)     │   │
│  │  - replay log                   │   │
│  └──────────────┬──────────────────┘   │
│                 │ hostPath / local PV  │
│  ┌──────────────▼──────────────────┐   │
│  │  ComputeContext Pod             │   │
│  │  - domino chain execution       │   │
│  │  - reads/writes via localhost   │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## Migration Path

1. **MVP (current):** SQLite at `--store-path`, single-node CLI
2. **Phase 2:** SQLite in `emptyDir` volume, domino runs as K8s Job
3. **Phase 3:** Node-local TSDB DaemonSet; controller connects via localhost
4. **Phase 4:** Debezium/Kafka sync for cross-node read replicas (not compute-time reads)

## Store Interface

The `pkg/store` package abstracts persistence. Swapping SQLite for TSDB requires implementing the same interface:

- `SaveSnapshot(snapshotID, timeSlice, data, sealed)`
- `LookupMemo(snapshotID, dominoID, inputHash)`
- `SaveResult(snapshotID, dominoID, inputHash, outputHash, output, reused)`
- `GetDominoOutput(snapshotID, dominoID)`

See [ADR 0003](../../docs/adr/0003-node-local-data.md) for the node-local data decision.

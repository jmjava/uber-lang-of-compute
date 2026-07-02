# KBL Compute Controller

The controller executes domino chains against immutable snapshots with hash-based memoization and deterministic replay logging.

## MVP Scope

This prototype proves the core physics of the KBL Compute Engine:

1. Load a **Snapshot** (immutable, sealed data view)
2. Execute a chain of **Dominos** in declared order
3. Store inputs/outputs in node-local SQLite
4. Hash inputs; skip domino if cached result exists
5. Emit replay log: snapshot ID, domino ID, input hash, output hash, reused vs recomputed

## Build

```bash
cd controller
go mod tidy
go build -o kbl-compute .
```

## Run

```bash
# Finance curve example (first run — all dominos computed)
./kbl-compute --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-finance/store.db \
  --replay-log /tmp/kbl-finance/replay-1.json

# Second run — memoization kicks in, all dominos reused
./kbl-compute --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-finance/store.db \
  --replay-log /tmp/kbl-finance/replay-2.json
```

## Simple domino chain

```bash
./kbl-compute --workflow ../examples/simple-domino-chain/workflow.yaml \
  --store /tmp/kbl-simple/store.db
```

## Architecture

```
main.go
  └── pkg/engine     Chain execution, input resolution, memoization
        ├── pkg/store    SQLite: snapshots, domino_results, replay_log
        ├── pkg/hash     SHA-256 input/output hashing
        ├── pkg/builtin  builtin:identity, interpolate, risk-dv01
        └── pkg/types    Snapshot, Domino, Workflow, ReplayLogEntry
```

## Post-MVP

- Kubernetes controller-runtime reconciler for CRDs
- OpenKruise hot-swapped container dominos
- Node-local TSDB DaemonSet backend
- Debezium/Kafka routing across pluggable universes

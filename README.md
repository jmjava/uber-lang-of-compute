# KBL Compute Engine

**A time-sliced, data-local, Kubernetes-native compute fabric**

The KBL Compute Engine processes immutable time-sliced data snapshots through modular, deterministic compute dominos placed near local data stores. It uses DSLs/CRDs to describe execution, data, provisioning, and routing — minimizing entropy through snapshot isolation and maximizing reuse through memoized intermediate results.

Derived from the [Uber Language of Compute](https://github.com/jmjava/uber-lang-of-compute) blog series (2020–2025).

## Core Concepts

| Term | Role |
|------|------|
| **Snapshot** | Immutable data view for a time slice |
| **Domino** | Single deterministic, referentially transparent compute step |
| **Compute Context** | Node-associated unit of compute + data locality |
| **Compute Wheel** | Rotating set of contexts processing time slices |
| **Pluggable Universe** | Swappable compute environment with its own execution/data/provisioning laws |

See [docs/vocabulary.md](docs/vocabulary.md) for the full glossary.

## Repository Structure

```
docs/           Vision, architecture, vocabulary, ADRs
specs/          Four DSL schemas + workflow example
crds/           Kubernetes CRD definitions (Snapshot, Domino, Workflow, …)
controller/     Go runtime — CLI + Kubernetes controller
examples/       Finance curve, simple domino chain, node-local TSDB target
tests/          Snapshot replay, memoization, scheduling (planned)
```

## MVP: Quick Start

### CLI (local)

```bash
make build

./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/replay.json

# Run again — all dominos reused from memo cache
./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/replay-2.json
```

### Kubernetes Controller

```bash
kubectl apply -f crds/
kubectl apply -f examples/finance-curve-snapshot/workflow-crd.yaml
./controller/bin/kbl-controller --store-root /tmp/kbl-store
kubectl get workflows -o wide
kubectl get configmap finance-curve-replay -o yaml
```

### Compute Wheel (time-slice rotation)

```bash
kubectl apply -f examples/compute-wheel/computecontexts.yaml
kubectl apply -f examples/compute-wheel/wheel.yaml
kubectl get computewheels -w
kubectl get workflows -l kbl.io/computewheel=finance-wheel
```

Run tests:

```bash
make test
```

### Multiverse routing (cross-universe)

```bash
kubectl apply -f examples/multiverse-finance/multiverse.yaml
kubectl apply -f examples/multiverse-finance/workflow-rates.yaml
./controller/bin/kbl-controller --store-root /tmp/kbl-store --kafka-brokers kafka:9092
kubectl get multiverses -o yaml   # status.routedEvents
```

See [examples/multiverse-finance/README.md](examples/multiverse-finance/README.md) and [ADR 0009](docs/adr/0009-multiverse-routing.md).

### Standalone Snapshot + Domino

```bash
kubectl apply -f examples/standalone-snapshot-domino/snapshot.yaml
kubectl apply -f examples/standalone-snapshot-domino/dominos.yaml
./controller/bin/kbl-controller --store-root /tmp/kbl-store
kubectl get snapshots,dominos -o wide
```

See [examples/standalone-snapshot-domino/README.md](examples/standalone-snapshot-domino/README.md) and [ADR 0010](docs/adr/0010-standalone-snapshot-domino.md).

### Read-replica materialization

```bash
kubectl apply -f examples/multiverse-finance/multiverse.yaml
# After workflows complete and Multiverse routes events:
kubectl get readreplicas -o wide
```

See [ADR 0011](docs/adr/0011-read-replica-materialization.md).

### Debezium CDC sync (Phase 9)

When Multiverse `spec.sync.enabled: true`, workflows publish CDC events and ReadReplicas use `replicationMode: cdc`:

```bash
kubectl get readreplicas -o jsonpath='{.items[*].spec.replicationMode}'
```

See [ADR 0012](docs/adr/0012-debezium-cdc-sync.md).

### Workflow CR references (Phase 10)

Reference standalone Snapshot and Domino CRs instead of inline specs:

```bash
kubectl apply -f examples/standalone-snapshot-domino/snapshot.yaml
kubectl apply -f examples/standalone-snapshot-domino/dominos.yaml
kubectl apply -f examples/workflow-snapshot-refs/workflow.yaml
kubectl get workflows -o wide
```

See [examples/workflow-snapshot-refs/README.md](examples/workflow-snapshot-refs/README.md) and [ADR 0013](docs/adr/0013-workflow-cr-references.md).

## What the MVP Proves

1. **Snapshot isolation** — sealed snapshots gate execution
2. **Deterministic dominos** — same inputs → same outputs, always
3. **Node-local storage** — SQLite store at configurable path
4. **Memoization** — input hash lookup skips recomputation
5. **Replay log** — audit trail with snapshot ID, domino ID, hashes, reused/recomputed

## Documentation

- [Vision](docs/vision.md)
- [Architecture](docs/architecture.md)
- [Vocabulary](docs/vocabulary.md)
- [ADR 0001: Four-DSL Model](docs/adr/0001-four-dsl-model.md)
- [ADR 0002: Snapshot Isolation](docs/adr/0002-snapshot-isolation.md)
- [ADR 0003: Node-Local Data](docs/adr/0003-node-local-data.md)
- [ADR 0004: Hot-Swapped Dominos](docs/adr/0004-hot-swapped-dominos.md)

- [ADR 0005: Kubernetes Controller](docs/adr/0005-kubernetes-controller.md)
- [ADR 0006: Compute Wheel Rotation](docs/adr/0006-compute-wheel-rotation.md)
- [ADR 0007: Hot-Swapped Dominos](docs/adr/0007-hot-swapped-dominos-implementation.md)
- [ADR 0008: Node-Local TSDB](docs/adr/0008-node-local-tsdb.md)
- [ADR 0009: Multiverse Routing](docs/adr/0009-multiverse-routing.md)
- [ADR 0010: Standalone Snapshot/Domino](docs/adr/0010-standalone-snapshot-domino.md)
- [ADR 0011: Read-Replica Materialization](docs/adr/0011-read-replica-materialization.md)
- [ADR 0012: Debezium CDC Sync](docs/adr/0012-debezium-cdc-sync.md)
- [ADR 0013: Workflow CR References](docs/adr/0013-workflow-cr-references.md)

## Roadmap

| Phase | Focus |
|-------|-------|
| **MVP** | CLI runtime, SQLite store, builtin dominos, replay log |
| **Phase 2** | Workflow CRD + `kbl-controller` reconciler |
| **Phase 3** | ComputeWheel time-slice rotation + player-piano pre-provision |
| **Phase 4** | Hot-swapped dominos — DominoChain CRD, init chain + OpenKruise |
| **Phase 5** | Node-local TSDB DaemonSet + store.Backend abstraction |
| **Phase 6** | Multiverse routing via Kafka + PluggableUniverse |
| **Phase 7** | Standalone Snapshot + Domino CRD reconcilers |
| **Phase 8** | Read-replica materialization from Multiverse routing |
| **Phase 9** | Debezium CDC sync for cross-universe read replicas |
| **Phase 10 (current)** | Workflow references to standalone Snapshot/Domino CRs |

## License

See repository license file.

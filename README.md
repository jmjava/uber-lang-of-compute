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

## Roadmap

| Phase | Focus |
|-------|-------|
| **MVP (current)** | CLI runtime, SQLite store, builtin dominos, replay log |
| **Phase 2 (current)** | Workflow CRD + `kbl-controller` reconciler |
| **Phase 3 (current)** | ComputeWheel time-slice rotation + player-piano pre-provision |
| Phase 4 | OpenKruise hot-swapped container dominos |
| Phase 5 | Node-local TSDB DaemonSet |
| Phase 6 | Multiverse routing via Debezium/Kafka |

## License

See repository license file.

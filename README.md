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

### DominoChain with CR references (Phase 11)

Container/hot-swap workflows can reference standalone CRs:

```bash
kubectl apply -f examples/standalone-snapshot-domino/snapshot.yaml
kubectl apply -f examples/standalone-snapshot-domino/dominos.yaml
kubectl apply -f examples/workflow-snapshot-refs/workflow-container.yaml
kubectl get dominochains -w
```

See [ADR 0014](docs/adr/0014-dominochain-cr-references.md).

### Path snapshot ingestion (Phase 12)

Load snapshot data from node-local files for content-addressed sealing:

```bash
sudo mkdir -p /var/kbl/data && sudo cp examples/path-snapshot/data/curve.json /var/kbl/data/
kubectl apply -f examples/path-snapshot/snapshot.yaml
kubectl get snapshots curve-file -o wide
```

See [examples/path-snapshot/README.md](examples/path-snapshot/README.md) and [ADR 0015](docs/adr/0015-path-snapshot-ingestion.md).

### ComputeWheel CR references (Phase 13)

Wheels can stamp Workflows that reference standalone Snapshot/Domino CRs:

```bash
kubectl apply -f examples/standalone-snapshot-domino/snapshot.yaml
kubectl apply -f examples/standalone-snapshot-domino/dominos.yaml
kubectl apply -f examples/compute-wheel/wheel-refs.yaml
kubectl get computewheels,workflows -l kbl.io/computewheel=finance-wheel-refs
```

See [ADR 0016](docs/adr/0016-computewheel-cr-references.md).

### HTTP snapshot ingestion (Phase 15)

Fetch snapshot data from HTTP/HTTPS URIs:

```bash
cd examples/http-snapshot/data && python3 -m http.server 8080 &
kubectl apply -f examples/http-snapshot/snapshot.yaml
kubectl get snapshots curve-http -o wide
```

See [ADR 0017](docs/adr/0017-http-snapshot-ingestion.md). After seal, domino runs use **store-first** reads (ADR 0018) and do not re-fetch the URI.

### Store-first hot path (Phase 16)

Once a snapshot is sealed into the node-local store, workflow execution reads persisted JSON from the store — no repeat HTTP or path resolution on the hot path.

See [ADR 0018](docs/adr/0018-store-first-snapshot.md).

### Direct-bytes staging (Phase 17)

Seal path/HTTP JSON in one pass — original file bytes persisted without parse→remarshal.

See [ADR 0019](docs/adr/0019-direct-bytes-staging.md).

### mmap + TSDB streaming (Phase 18)

Large path files (≥1 MiB) use mmap on Unix at seal time. TSDB stores snapshot payload sidecars and serves `GET /v1/snapshots/{id}/data` for streaming reads without envelope parsing.

See [ADR 0020](docs/adr/0020-mmap-tsdb-streaming.md).

### Zero-copy snapshot staging (Phase 19)

Large path seals write mmap-backed bytes directly to TSDB sidecars without a heap copy. TSDB envelopes store metadata only; `GET /v1/snapshots/{id}/data` streams sidecar files with `io.Copy`.

See [ADR 0021](docs/adr/0021-zero-copy-staging.md).

### Julia pluggable execution (Phase 14)

Run dominos via Julia subprocess using `julia:<script>` commands:

```bash
julia --project=controller/julia -e 'using Pkg; Pkg.instantiate()'
./controller/bin/kbl-compute --workflow examples/julia-domino-chain/workflow.yaml
```

See [examples/julia-domino-chain/README.md](examples/julia-domino-chain/README.md) and [ADR 0022](docs/adr/0022-julia-pluggable-execution.md). For in-cluster deployment choices (multi-container vs single-container multi-process), see [ADR 0023](docs/adr/0023-julia-deployment-models.md). Build the Julia runner image with `make docker-domino-runner-julia` (ADR 0024). CI builds both runner images on every relevant PR (ADR 0025).

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
- [ADR 0014: DominoChain CR References](docs/adr/0014-dominochain-cr-references.md)
- [ADR 0015: Path Snapshot Ingestion](docs/adr/0015-path-snapshot-ingestion.md)
- [ADR 0016: ComputeWheel CR References](docs/adr/0016-computewheel-cr-references.md)
- [ADR 0017: HTTP Snapshot Ingestion](docs/adr/0017-http-snapshot-ingestion.md)
- [ADR 0018: Store-First Snapshot](docs/adr/0018-store-first-snapshot.md)
- [ADR 0019: Direct-Bytes Staging](docs/adr/0019-direct-bytes-staging.md)
- [ADR 0020: mmap + TSDB Streaming](docs/adr/0020-mmap-tsdb-streaming.md)
- [ADR 0021: Zero-Copy Staging](docs/adr/0021-zero-copy-staging.md)
- [ADR 0022: Julia Pluggable Execution](docs/adr/0022-julia-pluggable-execution.md)
- [ADR 0023: Julia Deployment Models](docs/adr/0023-julia-deployment-models.md)
- [ADR 0024: Julia In-Cluster Execution](docs/adr/0024-julia-in-cluster.md)
- [ADR 0025: Julia Production Wiring](docs/adr/0025-julia-production-wiring.md)

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
| **Phase 10** | Workflow references to standalone Snapshot/Domino CRs |
| **Phase 11** | DominoChain container path resolves Workflow CR refs |
| **Phase 12** | Node-local path snapshot ingestion |
| **Phase 13** | ComputeWheel workflow template CR references |
| **Phase 14** | Julia pluggable execution — `julia:` domino commands via subprocess + bundled scripts |
| **Phase 15** | HTTP/HTTPS snapshot URI ingestion |
| **Phase 16** | Store-first snapshot reads — hot path skips re-fetching HTTP/path sources |
| **Phase 17** | Direct-bytes snapshot staging — single-pass seal without parse→remarshal |
| **Phase 18** | mmap path reads (≥1 MiB) + TSDB snapshot data sidecars and streaming `/data` endpoint |
| **Phase 19** | Zero-copy path staging — metadata-only TSDB envelopes, mmap seal-to-sidecar, streaming `/data` reads |
| **Phase 20** | Julia in-cluster — Julia domino-runner image, DominoChain env wiring, kubernetes-init examples |
| **Phase 21 (current)** | Julia production wiring — Docker CI, multiverse Julia universe, OpenKruise env parity |

## Performance note

Phase 15 HTTP ingestion is intended for convenience and cross-node bootstrap, not the hot compute path. **Phase 16** loads persisted snapshot JSON from the node-local store on execute; **Phase 17** seals path/HTTP sources in one pass; **Phase 18** adds mmap for large path files and TSDB `/data` streaming sidecars; **Phase 19** eliminates heap copies on large path seals and streams TSDB sidecars without buffering. Production workloads should still prefer **node-local paths** (Phase 12) or **pre-sealed snapshots** on the TSDB/store — bring compute to the data.

## License

See repository license file.

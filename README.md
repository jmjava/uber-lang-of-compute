# KBL Compute Engine

**A time-sliced, data-local, Kubernetes-native compute fabric**

The KBL Compute Engine processes immutable time-sliced data snapshots through modular, deterministic compute dominos placed near local data stores. It uses DSLs/CRDs to describe execution, data, provisioning, and routing — minimizing entropy through snapshot isolation and maximizing reuse through memoized intermediate results.

Derived from the [Uber Language of Compute](https://github.com/jmjava/uber-lang-of-compute) blog series (2020–2025), originally published at **[jmenke.blogspot.com](https://jmenke.blogspot.com/)**.

## Documentation

**Start here:** [docs/README.md](docs/README.md) — central index, blog mapping, examples, ADRs by topic.

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | CLI → Kind lab → verify Volcano + OpenKruise |
| [**Architecture Diagrams**](docs/diagrams.md) | Mermaid — topology, runtimes, troubleshooting |
| [Architecture](docs/architecture.md) | System layers, data flow, runtimes |
| [Provisioning Runtimes](docs/provisioning-runtimes.md) | `kubernetes-init`, `openkruise`, `volcano-init` |
| [Vocabulary](docs/vocabulary.md) | Glossary |
| [Vision](docs/vision.md) | Design goals |
| [Kind Lab](lab/README.md) | Local multi-node cluster operations |
| [Blog — jmenke.blogspot.com](https://jmenke.blogspot.com/) | Original Uber Language of Compute series |

All ADRs: [docs/README.md#adrs-by-topic](docs/README.md#adrs-by-topic). Foundation: [ADR 0001](docs/adr/0001-four-dsl-model.md) (Four-DSL Model), [ADR 0002](docs/adr/0002-snapshot-isolation.md) (Snapshot Isolation), [ADR 0003](docs/adr/0003-node-local-data.md) (Node-Local Data), [ADR 0004](docs/adr/0004-hot-swapped-dominos.md) (Hot-Swapped Dominos).

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
| **Phase 21** | Julia production wiring — Docker CI, multiverse Julia universe, OpenKruise env parity |
| **Phase 22** | Kind local lab — controller + TSDB images, kustomize, up/down scripts; AWS CDK scaffold (VPC, ECR, EKS) |
| **Phase 23** | Julia FinanceModels curve dominos — `ZeroRateCurve` interpolation, bump-and-reprice DV01, locked Manifest |
| **Phase 24** | Julia Greeks — bond duration/convexity, rate-bucket DV01, Black–Scholes option greeks via `julia:greeks` |
| **Phase 25** | Volcano Kind lab — multi-node cluster, Volcano install, queue + TSDB Data Pond node pin |
| **Phase 26** | Controller Volcano emission — `runtime: volcano-init` on DominoChain, reconciler creates VCJob |
| **Phase 27** | ComputeWheel Volcano queue — wheel assigns queue/nodeSelector/runner per time slice → Workflow → VCJob |
| **Phase 28** | OpenKruise Kind lab — Helm install, Julia hot-swap DominoChain demo via ContainerRecreateRequest |
| **Phase 29** | Documentation hub — getting-started, provisioning-runtimes guide, architecture refresh |
| **Phase 30 (current)** | Architecture diagrams — Mermaid visual reference (topology, sequences, troubleshooting) |

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
docs/           Vision, architecture, guides, vocabulary, ADRs (start at docs/README.md)
specs/          Four DSL schemas + workflow example
crds/           Kubernetes CRD definitions (Snapshot, Domino, Workflow, …)
controller/     Go runtime — CLI + Kubernetes controller
lab/            Kind local lab — kustomize, scripts, sample manifests
infra/          AWS CDK scaffold (EKS + ECR)
examples/       Finance curve, simple domino chain, node-local TSDB target
tests/          Snapshot replay, memoization, scheduling (planned)
```

## Quick Start

For step-by-step paths (CLI, controller, Kind lab, Volcano, OpenKruise), see **[docs/getting-started.md](docs/getting-started.md)**.

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

### Kind lab (recommended)

```bash
make lab-up          # Kind cluster + controller + TSDB + Volcano + OpenKruise demos
kubectl get workflows finance-lab -o wide
make lab-down        # delete Kind cluster
```

See [lab/README.md](lab/README.md) for flags (`KBL_LAB_VOLCANO=0`, `KBL_LAB_OPENKURISE=0`) and manifests.

### Kubernetes controller

```bash
kubectl apply -f crds/
kubectl apply -f examples/finance-curve-snapshot/workflow-crd.yaml
./controller/bin/kbl-controller --store-root /tmp/kbl-store
kubectl get workflows -o wide
```

Run tests: `make test`

### Examples by topic

| Topic | Example README |
|-------|----------------|
| Multiverse routing | [examples/multiverse-finance/README.md](examples/multiverse-finance/README.md) |
| Standalone Snapshot/Domino | [examples/standalone-snapshot-domino/README.md](examples/standalone-snapshot-domino/README.md) |
| Workflow CR references | [examples/workflow-snapshot-refs/README.md](examples/workflow-snapshot-refs/README.md) |
| Path snapshot ingestion | [examples/path-snapshot/README.md](examples/path-snapshot/README.md) |
| Julia domino chain | [examples/julia-domino-chain/README.md](examples/julia-domino-chain/README.md) |
| AWS CDK (EKS + ECR) | [infra/aws/cdk/README.md](infra/aws/cdk/README.md) |

## What the MVP Proves

1. **Snapshot isolation** — sealed snapshots gate execution
2. **Deterministic dominos** — same inputs → same outputs, always
3. **Node-local storage** — SQLite store at configurable path
4. **Memoization** — input hash lookup skips recomputation
5. **Replay log** — audit trail with snapshot ID, domino ID, hashes, reused/recomputed

## Performance note

Phase 15 HTTP ingestion is intended for convenience and cross-node bootstrap, not the hot compute path. **Phase 16** loads persisted snapshot JSON from the node-local store on execute; **Phase 17** seals path/HTTP sources in one pass; **Phase 18** adds mmap for large path files and TSDB `/data` streaming sidecars; **Phase 19** eliminates heap copies on large path seals and streams TSDB sidecars without buffering. Production workloads should still prefer **node-local paths** (Phase 12) or **pre-sealed snapshots** on the TSDB/store — bring compute to the data.

## License

See repository license file.

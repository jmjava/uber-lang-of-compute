# KBL Compute Engine — Documentation

Central index for the [Uber Language of Compute](https://jmenke.blogspot.com/) implementation in this repository.

## Start here

| Guide | Audience | Content |
|-------|----------|---------|
| [Getting Started](getting-started.md) | New operators | CLI → Kind lab → verify end-to-end |
| [**Architecture Diagrams**](diagrams.md) | Everyone | **Mermaid** — topology, sequences, troubleshooting |
| [Kind Lab Guide](../lab/README.md) | Local development | Multi-node Kind, Volcano, OpenKruise, demos |
| [Provisioning Runtimes](provisioning-runtimes.md) | Platform engineers | `kubernetes-init`, `openkruise`, `volcano-init` compared |
| [Architecture](architecture.md) | System design | Layers, data flow, [Multiverse communication](architecture.md#multiverse-communication) |
| [Vocabulary](vocabulary.md) | Everyone | Glossary aligned with the blog series |
| [Vision](vision.md) | Context | Original design goals |

## Blog alignment

The [jmenke.blogspot.com](https://jmenke.blogspot.com/) series describes four DSLs and a Kubernetes-native compute fabric. This repo maps blog concepts to code as follows:

| Blog concept | Repo artifact |
|--------------|---------------|
| Execution DSL | `Workflow`, `Domino`, `DominoChain` CRDs |
| Data DSL | `Snapshot`, node-local TSDB/SQLite, sealed snapshots |
| Provisioning DSL | DominoChain runtimes, Kind lab, Volcano queue, OpenKruise CRR |
| Routing DSL | `PluggableUniverse`, `Multiverse`, `ComputeContext` — multiple KBL fabrics coordinate **event-driven via Kafka**, not controller RPC ([architecture § Multiverse communication](architecture.md#multiverse-communication)) |
| Ferris Wheel / time slices | `ComputeWheel` reconciler ([ADR 0006](adr/0006-compute-wheel-rotation.md)) |
| Player-piano scheduling | `preProvisionNext` on ComputeWheel |
| Data Pond | Node-local TSDB + `kbl.io/tsdb-node` worker pin in Kind lab |
| Volcano SyncSet / batch | `runtime: volcano-init`, VCJob emission ([ADR 0030](adr/0030-controller-volcano-emission.md)) |
| Hot-swapped dominos | `runtime: openkruise`, ContainerRecreateRequest ([ADR 0007](adr/0007-hot-swapped-dominos-implementation.md)) |

## Examples by topic

| Topic | Path |
|-------|------|
| Finance curve (CLI) | [examples/finance-curve-snapshot](../examples/finance-curve-snapshot/) |
| Simple domino chain | [examples/simple-domino-chain](../examples/simple-domino-chain/) |
| Compute Wheel | [examples/compute-wheel](../examples/compute-wheel/) |
| Julia + FinanceModels | [examples/julia-domino-chain](../examples/julia-domino-chain/) |
| Hot-swap domino chain | [examples/hot-swap-domino-chain](../examples/hot-swap-domino-chain/) |
| Node-local TSDB | [examples/node-local-tsdb](../examples/node-local-tsdb/) |
| Multiverse routing | [examples/multiverse-finance](../examples/multiverse-finance/) |
| Workflow CR refs | [examples/workflow-snapshot-refs](../examples/workflow-snapshot-refs/) |

## Kind lab demos (Phases 25–28)

After `make lab-up`:

| Demo | Resource | Runtime |
|------|----------|---------|
| Finance workflow (local engine) | `Workflow/finance-lab` | `local` / builtin |
| Volcano time slice | `ComputeWheel/julia-finance-wheel` | `volcano-init` |
| OpenKruise hot-swap | `DominoChain/julia-finance-openkruise` | `openkruise` |

Skip components: `KBL_LAB_VOLCANO=0`, `KBL_LAB_OPENKURISE=0`. See [lab/README.md](../lab/README.md).

## ADRs by topic

### Foundation
- [0001 Four-DSL Model](adr/0001-four-dsl-model.md)
- [0002 Snapshot Isolation](adr/0002-snapshot-isolation.md)
- [0003 Node-Local Data](adr/0003-node-local-data.md)
- [0005 Kubernetes Controller](adr/0005-kubernetes-controller.md)

### Scheduling & rotation
- [0006 Compute Wheel Rotation](adr/0006-compute-wheel-rotation.md)
- [0016 ComputeWheel CR References](adr/0016-computewheel-cr-references.md)

### In-cluster execution
- [0007 Hot-Swapped Dominos Implementation](adr/0007-hot-swapped-dominos-implementation.md)
- [0014 DominoChain CR References](adr/0014-dominochain-cr-references.md)

### Data path & TSDB
- [0008 Node-Local TSDB](adr/0008-node-local-tsdb.md)
- [0015–0021 Path/HTTP/staging/streaming ADRs](adr/0015-path-snapshot-ingestion.md)

### Julia
- [0022 Julia Pluggable Execution](adr/0022-julia-pluggable-execution.md)
- [0023 Julia Deployment Models](adr/0023-julia-deployment-models.md)
- [0024 Julia In-Cluster](adr/0024-julia-in-cluster.md)
- [0027 FinanceModels Curves](adr/0027-julia-financemodels-curves.md)
- [0028 Julia Greeks](adr/0028-julia-greeks.md)

### Lab & cloud
- [0026 Kind Lab + AWS CDK](adr/0026-kind-lab-aws-cdk.md)
- [0029 Volcano Kind Lab](adr/0029-volcano-kind-lab.md)
- [0030 Controller Volcano Emission](adr/0030-controller-volcano-emission.md)
- [0031 ComputeWheel Volcano Queue](adr/0031-computewheel-volcano-queue.md)
- [0032 OpenKruise Kind Lab](adr/0032-openkruise-kind-lab.md)

### Meta
- [0033 Documentation Phase](adr/0033-documentation-phase.md)
- [0034 Architecture Diagrams](adr/0034-documentation-diagrams.md)

## Phase roadmap

Full phase table lives in the [root README](../README.md#roadmap). Recent phases:

| Phase | Deliverable |
|-------|-------------|
| 22 | Kind lab + AWS CDK scaffold |
| 23–24 | Julia FinanceModels curves + Greeks |
| 25–27 | Volcano install, controller VCJob emission, ComputeWheel queue |
| 28 | OpenKruise lab demo |
| 29 | Documentation hub (this index) |
| 30 | Architecture diagrams — Mermaid visual reference |

## Visual reference

**[diagrams.md](diagrams.md)** — 13 diagrams: DSL layers, Compute Wheel, Kind lab topology, init/OpenKruise/Volcano sequences, memoization, troubleshooting, AWS target.

## Component READMEs

- [controller/README.md](../controller/README.md) — Go runtime layout
- [lab/README.md](../lab/README.md) — Kind lab operations
- [infra/aws/cdk/README.md](../infra/aws/cdk/README.md) — AWS deployment scaffold

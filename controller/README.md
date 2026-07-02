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
make build
# or
cd controller
go build -o bin/kbl-compute ./cmd/kbl-compute
go build -o bin/kbl-controller ./cmd/kbl-controller
```

## Run CLI (local development)

```bash
# Finance curve example (first run — all dominos computed)
./bin/kbl-compute --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-finance/store.db \
  --replay-log /tmp/kbl-finance/replay-1.json

# Second run — memoization kicks in, all dominos reused
./bin/kbl-compute --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-finance/store.db \
  --replay-log /tmp/kbl-finance/replay-2.json
```

## Run Kubernetes Controller

Install CRDs (includes Workflow):

```bash
kubectl apply -f ../crds/
```

Deploy and apply a Workflow:

```bash
kubectl apply -f ../examples/finance-curve-snapshot/workflow-crd.yaml
./bin/kbl-controller --store-root /var/kbl/store
```

The controller reconciles Workflow resources: executes the domino chain, updates status, and writes replay JSON to a `{name}-replay` ConfigMap.

## Architecture

```
cmd/kbl-compute/       CLI for local/CI execution
cmd/kbl-controller/    Kubernetes controller-runtime reconciler
api/v1alpha1/          Workflow, ComputeWheel, Multiverse, PluggableUniverse types
internal/controller/   Workflow, ComputeWheel, DominoChain, Multiverse, Snapshot, Domino reconcilers
pkg/dominochain/       Init chain + OpenKruise pod builders, domino-runner handoff
cmd/domino-runner/     Container entrypoint for in-cluster domino steps
pkg/wheel/             Time-slice rotation logic and workflow builder
pkg/replica/            Cross-store read-replica materialization
pkg/cdc/                Debezium-compatible CDC export, publish, and apply
pkg/snapshot/          Snapshot sealing and deterministic ID computation
pkg/events/            Memory + Kafka event bus for snapshot completion events
pkg/routing/           Multiverse partition and time-slice routing
pkg/engine/            Chain execution, input resolution, memoization
pkg/store/             SQLite + TSDB backends, resolver, HTTP TSDB client
cmd/kbl-tsdb/          Node-local TSDB DaemonSet server
deploy/node-local-tsdb/ DaemonSet manifest
pkg/convert/           CRD → engine domain model
pkg/hash/              SHA-256 input/output hashing
pkg/builtin/           builtin:identity, interpolate, risk-dv01
```

## Compute Wheel

The ComputeWheel reconciler rotates contexts through time slices:

```bash
kubectl apply -f ../examples/compute-wheel/computecontexts.yaml
kubectl apply -f ../examples/compute-wheel/wheel.yaml
kubectl get computewheels -w
```

See [ADR 0006](../docs/adr/0006-compute-wheel-rotation.md).

## Multiverse routing

The Multiverse reconciler routes snapshot completion events across Pluggable Universes:

```bash
kubectl apply -f ../examples/multiverse-finance/multiverse.yaml
kubectl apply -f ../examples/multiverse-finance/workflow-rates.yaml
./bin/kbl-controller --store-root /var/kbl/store --kafka-brokers kafka:9092
kubectl get multiverses -o yaml
```

Controller flags: `--kafka-brokers` (comma-separated), `--kafka-topic` (default `kbl.snapshot.events`).

See [ADR 0009](../docs/adr/0009-multiverse-routing.md).

## Standalone Snapshot + Domino

Fine-grained reconcilers for sealed snapshots and individual domino steps:

```bash
kubectl apply -f ../examples/standalone-snapshot-domino/snapshot.yaml
kubectl apply -f ../examples/standalone-snapshot-domino/dominos.yaml
./bin/kbl-controller --store-root /var/kbl/store
kubectl get snapshots,dominos -o wide
```

See [ADR 0010](../docs/adr/0010-standalone-snapshot-domino.md).

## Read-replica materialization

When Multiverse routes snapshot events, **ReadReplica** CRs copy snapshot data and domino results to target universe stores:

```bash
kubectl get readreplicas -o wide
```

See [ADR 0011](../docs/adr/0011-read-replica-materialization.md).

## Debezium CDC sync

When Multiverse sync is enabled, workflows publish CDC events to `kbl.cdc.snapshots` and ReadReplicas replicate via `replicationMode: cdc` instead of direct store copy.

See [ADR 0012](../docs/adr/0012-debezium-cdc-sync.md).

## Post-MVP

- ~~Kubernetes controller-runtime reconciler for CRDs~~ (Workflow reconciler shipped in Phase 2)
- ~~Compute Wheel time-slice scheduling~~ (ComputeWheel reconciler shipped in Phase 3)
- ~~OpenKruise hot-swapped container dominos~~ (DominoChain reconciler shipped in Phase 4)
- ~~Node-local TSDB DaemonSet backend~~ (kbl-tsdb + store.Backend shipped in Phase 5)
- ~~Multiverse routing via Debezium/Kafka~~ (Multiverse + PluggableUniverse shipped in Phase 6)
- ~~Standalone Snapshot/Domino CRD reconcilers~~ (shipped in Phase 7)
- ~~Read-replica materialization from routed multiverse events~~ (shipped in Phase 8)
- ~~Debezium CDC replacing direct store copy for cross-universe sync~~ (shipped in Phase 9)
- Workflow references to standalone Snapshot/Domino CRs

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
api/v1alpha1/          Workflow, ComputeWheel, ComputeContext types
internal/controller/   Workflow + ComputeWheel + DominoChain reconcilers
pkg/dominochain/       Init chain + OpenKruise pod builders, domino-runner handoff
cmd/domino-runner/     Container entrypoint for in-cluster domino steps
pkg/wheel/             Time-slice rotation logic and workflow builder
pkg/engine/            Chain execution, input resolution, memoization
pkg/store/             SQLite: snapshots, domino_results, replay_log
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

## Post-MVP

- ~~Kubernetes controller-runtime reconciler for CRDs~~ (Workflow reconciler shipped in Phase 2)
- ~~Compute Wheel time-slice scheduling~~ (ComputeWheel reconciler shipped in Phase 3)
- ~~OpenKruise hot-swapped container dominos~~ (DominoChain reconciler shipped in Phase 4)
- Node-local TSDB DaemonSet backend
- Multiverse routing via Debezium/Kafka
- Standalone Snapshot/Domino CRD reconcilers

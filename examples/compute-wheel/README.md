# Compute Wheel Example

Demonstrates time-sliced rotation across multiple ComputeContexts — the Ferris Wheel pattern from the [Uber Language of Compute blog](https://jmenke.blogspot.com/).

## How It Works

```
Time slice T ──────────────────────────────────────────────►
  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
  │  node-a-ctx │───►│  node-b-ctx │───►│  node-a-ctx │───►  T+24h
  │  Workflow   │    │  Workflow   │    │  Workflow   │
  └─────────────┘    └─────────────┘    └─────────────┘
```

1. **ComputeWheel** defines contexts, time slice interval, and a workflow template
2. For each slot (context × time slice), the controller creates a **Workflow**
3. When the Workflow completes, the wheel rotates to the next context
4. After all contexts process a slice, the time slice advances
5. With `preProvisionNext: true`, the next slot's Workflow is created while the current one runs (player-piano scheduling)

## Manifests

| File | Purpose |
|------|---------|
| `wheel.yaml` | Inline snapshot/dominos, builtin dominos, continuous rotation |
| `wheel-refs.yaml` | CR refs (`snapshotRef`, `dominoRefs`), `maxRotations: 2` |
| `wheel-volcano.yaml` | Volcano batch runtime per time slice ([ADR 0031](../../docs/adr/0031-computewheel-volcano-queue.md)) |
| `computecontexts.yaml` | Sample node-local contexts |

## Deploy (manual cluster)

```bash
kubectl apply -f ../../crds/
kubectl apply -f computecontexts.yaml
kubectl apply -f wheel.yaml

# Start controller (if not in-cluster)
make -C ../.. build
./../../controller/bin/kbl-controller --store-root /var/kbl/store

kubectl get computewheels -w
kubectl get workflows -l kbl.io/computewheel=finance-wheel
```

## Kind lab (Volcano wheel)

The lab applies `ComputeWheel/julia-finance-wheel` automatically:

```bash
make lab-up
kubectl get wheel julia-finance-wheel -o wide
kubectl get wf -l kbl.io/computewheel=julia-finance-wheel
kubectl get vcjob -l kbl.io/volcano-demo=true
```

See [docs/getting-started.md](../../docs/getting-started.md) and [lab/README.md](../../lab/README.md).

## Volcano + time slices (Phase 27)

Set wheel-level queue and runtime on the workflow template:

```yaml
spec:
  volcanoQueue: kbl-lab
  nodeSelector:
    kbl.io/lab-role: compute
  workflowTemplate:
    execution:
      runtime: volcano-init
      chain: [load-curve-data, interpolate-curve, compute-greeks]
    provisioning:
      runnerImage: kbl-domino-runner-julia:lab
```

Pipeline: **ComputeWheel → Workflow → DominoChain → Volcano Job**.

```bash
kubectl apply -f wheel-volcano.yaml   # requires Volcano + queue kbl-lab
```

Full runtime comparison: [docs/provisioning-runtimes.md](../../docs/provisioning-runtimes.md).

## Status Fields

| Field | Meaning |
|-------|---------|
| `status.currentTimeSlice` | Active immutable data window |
| `status.activeContext` | ComputeContext currently processing |
| `status.rotationCount` | Completed slice rotations |
| `status.processedSlots` | Audit trail of completed context×slice pairs |

## Demo With Limited Rotations

Set `maxRotations: 1` to process one full slice across all contexts then stop (lab default for `julia-finance-wheel`).

## CR references (Phase 13)

Use shared Snapshot/Domino CRs in the wheel template:

```bash
kubectl apply -f ../standalone-snapshot-domino/snapshot.yaml
kubectl apply -f ../standalone-snapshot-domino/dominos.yaml
kubectl apply -f wheel-refs.yaml
```

See [ADR 0016](../../docs/adr/0016-computewheel-cr-references.md).

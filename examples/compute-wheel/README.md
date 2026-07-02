# Compute Wheel Example

Demonstrates time-sliced rotation across multiple ComputeContexts — the Ferris Wheel pattern from the Uber Language of Compute blog series.

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

## Deploy

```bash
kubectl apply -f ../../crds/
kubectl apply -f computecontexts.yaml
kubectl apply -f wheel.yaml

# Start controller
./controller/bin/kbl-controller --store-root /var/kbl/store

# Watch rotation
kubectl get computewheels -w
kubectl get workflows -l kbl.io/computewheel=finance-wheel
```

## Status Fields

| Field | Meaning |
|-------|---------|
| `status.currentTimeSlice` | Active immutable data window |
| `status.activeContext` | ComputeContext currently processing |
| `status.rotationCount` | Completed slice rotations |
| `status.processedSlots` | Audit trail of completed context×slice pairs |

## Demo With Limited Rotations

Set `maxRotations: 1` to process one full slice across all contexts then stop (useful for tests).

## CR references (Phase 13)

Use shared Snapshot/Domino CRs in the wheel template:

```bash
kubectl apply -f ../standalone-snapshot-domino/snapshot.yaml
kubectl apply -f ../standalone-snapshot-domino/dominos.yaml
kubectl apply -f wheel-refs.yaml
```

See [ADR 0016](../../docs/adr/0016-computewheel-cr-references.md).

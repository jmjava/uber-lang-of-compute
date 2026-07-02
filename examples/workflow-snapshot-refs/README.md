# Workflow with Snapshot/Domino CR References

Compose a full domino chain by referencing standalone **Snapshot** and **Domino** CRs instead of embedding inline specs.

## Concepts

| Field | Role |
|-------|------|
| `spec.snapshotRef` | Name of a sealed Snapshot CR in the same namespace |
| `spec.dominoRefs` | Ordered list of Domino CR names; `execution.chain` must match |

Inline `spec.snapshot` and `spec.dominos` remain supported for backward compatibility.

Container execution with CR refs:

```bash
kubectl apply -f workflow-container.yaml
kubectl get dominochains -w
```

## Deploy

```bash
kubectl apply -f ../../crds/
kubectl apply -f ../standalone-snapshot-domino/snapshot.yaml
kubectl apply -f ../standalone-snapshot-domino/dominos.yaml
kubectl apply -f workflow.yaml

./../../controller/bin/kbl-controller --store-root /tmp/kbl-store

kubectl get workflows -o wide
kubectl get configmap curve-from-refs-replay -o yaml
```

## Ordering

1. Snapshot reconciler seals `curve-snap` and sets `status.snapshotID`
2. Workflow reconciler resolves refs via `convert.ResolveEngineWorkflow`
3. If the snapshot is not yet sealed, the workflow stays `Pending` and requeues

See [ADR 0013](../../docs/adr/0013-workflow-cr-references.md) and [ADR 0014](../../docs/adr/0014-dominochain-cr-references.md).

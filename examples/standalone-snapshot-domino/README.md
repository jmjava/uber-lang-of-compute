# Standalone Snapshot + Domino

Fine-grained reconciliation of immutable snapshots and individual domino steps — complementing the composed **Workflow** CRD.

## Concepts

| Resource | Role |
|----------|------|
| **Snapshot** | Sealed immutable data view; computes deterministic `status.snapshotID` |
| **Domino** | Single compute step against a sealed Snapshot; respects `dependsOn` ordering |

Workflow remains the primary unit for full chains. Standalone CRDs enable incremental execution, shared snapshots across teams, and dependency-aware domino scheduling.

## Deploy

```bash
kubectl apply -f ../../crds/
kubectl apply -f snapshot.yaml
kubectl apply -f dominos.yaml

./../../controller/bin/kbl-controller --store-root /tmp/kbl-store

kubectl get snapshots,dominos -o wide
```

## Ordering

1. Snapshot reconciler seals when `spec.sealed: true` and persists to node-local store
2. Domino reconciler waits for sealed snapshot, then executes when dependencies complete
3. Second run of the same domino hits memo cache (`status.phase: Cached`)

See [ADR 0010](../../docs/adr/0010-standalone-snapshot-domino.md).

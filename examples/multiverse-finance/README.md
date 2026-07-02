# Multiverse Finance Routing

Routes snapshot completion events across **Pluggable Universes** via in-process bus and optional **Kafka/Debezium** sync.

## Concepts

| Resource | Role |
|----------|------|
| **PluggableUniverse** | Compute environment laws (engine, data layer, provisioning) |
| **Multiverse** | Partition + time-slice routing table across universes |
| **Snapshot event** | Published when a Workflow completes; routed by Multiverse |

## Routing rules (priority)

1. **Time slice override** — `spec.timeSliceRoutes`
2. **Partition match** — workflow labels `kbl.io/partition-<key>: <value>`
3. **Default universe** — `spec.defaultUniverse`

## Deploy

```bash
kubectl apply -f ../../crds/
kubectl apply -f multiverse.yaml
kubectl apply -f workflow-rates.yaml

./../../controller/bin/kbl-controller \
  --store-root /var/kbl/store \
  --kafka-brokers kafka:9092

kubectl get multiverses -o yaml
# status.routedEvents shows routed snapshot completions
```

## Kafka / Debezium sync

When `spec.sync.enabled: true`, completed snapshots are also published to `kbl.snapshot.events`. This mirrors Debezium CDC routing for cross-universe read replicas (not compute-time reads).

Local dev uses the in-memory bus (default when `--kafka-brokers` is omitted).

## Read replicas

Routed events automatically create **ReadReplica** CRs that materialize snapshot + domino results to the target universe store. Check progress with:

```bash
kubectl get readreplicas -o wide
```

See [read-replica.md](read-replica.md) and [ADR 0011](../../docs/adr/0011-read-replica-materialization.md).

## Workflow labels

```yaml
metadata:
  labels:
    kbl.io/partition-asset_class: rates
spec:
  routing:
    multiverseRef: finance-multiverse
```

See [ADR 0009](../../docs/adr/0009-multiverse-routing.md).

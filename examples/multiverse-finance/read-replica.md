## Read replicas

When a Multiverse routes a snapshot completion event, a **ReadReplica** CR is created and materialized to the target universe store:

```bash
kubectl get readreplicas -o wide
# status.phase: Ready, status.targetStorePath shows local replica DB
```

See [ADR 0011](../../docs/adr/0011-read-replica-materialization.md).

# Path-Based Snapshot Ingestion

Load snapshot data from node-local files instead of embedding inline YAML.

## Concepts

| Source | Behavior |
|--------|----------|
| `source.inline` | Embedded data (unchanged) |
| `source.path` | Read JSON (or raw text) from node-local path at seal time |
| `source.uri` (`file://`) | Same as path via file URI |

Remote URIs (`s3://`, `https://`) remain metadata-only until remote ingestion is added.

## Deploy

Mount data on the node and point the Snapshot CR at the file:

```bash
# Copy sample data to the node store path
sudo mkdir -p /var/kbl/data
sudo cp data/curve.json /var/kbl/data/curve.json

kubectl apply -f ../../crds/
kubectl apply -f snapshot.yaml

./../../controller/bin/kbl-controller --store-root /tmp/kbl-store
kubectl get snapshots curve-file -o wide
```

The reconciler reads file contents for deterministic `status.snapshotID` and persists resolved JSON to the node-local store. If the file is not yet present, the snapshot stays `Pending` and requeues.

See [ADR 0015](../../docs/adr/0015-path-snapshot-ingestion.md).

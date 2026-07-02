# HTTP Snapshot Ingestion

Fetch snapshot data from `http://` or `https://` URIs at seal time.

## Concepts

| Source | Behavior |
|--------|----------|
| `source.uri` (`http://`, `https://`) | Fetched with 30s timeout; JSON parsed when valid |
| `source.uri` (`file://`) | Node-local file (Phase 12) |
| `source.uri` (`s3://`, etc.) | Metadata-only until remote adapters land |

Maximum response body: 32 MiB.

## Local demo

```bash
# Terminal 1 — serve sample data
cd data && python3 -m http.server 8080

# Terminal 2
kubectl apply -f ../../crds/
kubectl apply -f snapshot.yaml
./../../controller/bin/kbl-controller --store-root /tmp/kbl-store
kubectl get snapshots curve-http -o wide
```

Transient HTTP errors (502, 503, 504, 429, timeouts) leave the snapshot `Pending` and requeue.

See [ADR 0017](../../docs/adr/0017-http-snapshot-ingestion.md).

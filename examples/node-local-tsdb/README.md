# Node-Local TSDB (Phase 5)

Node-local time-series store via **kbl-tsdb** DaemonSet — replaces direct SQLite access for production compute contexts.

## Architecture

```
Node
├── kbl-tsdb DaemonSet (:9090, hostPath /var/kbl/tsdb)
│   ├── snapshots/   immutable snapshot series
│   ├── memo/        hash → domino result cache
│   └── replay/      append-only replay log
└── kbl-controller / domino pods
    └── HTTP client → http://127.0.0.1:9090
```

## Deploy TSDB DaemonSet

```bash
kubectl apply -f ../../deploy/node-local-tsdb/daemonset.yaml
kubectl -n kbl-system get pods -l app.kubernetes.io/name=kbl-tsdb
curl http://127.0.0.1:9090/healthz   # on any node
```

## Local development (no Kubernetes)

```bash
# Terminal 1
./controller/bin/kbl-tsdb --addr=:9090 --data-dir=/tmp/kbl-tsdb

# Terminal 2 — point workflow at TSDB endpoint
./controller/bin/kbl-compute \
  --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --store http://127.0.0.1:9090
```

Note: CLI `--store` accepts HTTP URLs for TSDB when using updated store resolver (use workflow with `storePath: http://127.0.0.1:9090`).

## ComputeContext with TSDB

```bash
kubectl apply -f workflow-tsdb.yaml
```

Workflows with `routing.computeContextRef` pointing to a `storeType: tsdb` context use the node-local TSDB endpoint automatically.

## Store interface

All backends implement `store.Backend`:

| Backend | Config |
|---------|--------|
| SQLite | `{type: sqlite, path: "/var/kbl/store.db"}` |
| TSDB | `{type: tsdb, endpoint: "http://127.0.0.1:9090"}` |

See [ADR 0008](../../docs/adr/0008-node-local-tsdb.md).

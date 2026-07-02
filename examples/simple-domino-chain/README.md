# Simple Domino Chain

Minimal two-step domino chain proving snapshot isolation and chained execution.

## Run

```bash
cd controller
go build -o kbl-compute .
./kbl-compute --workflow ../examples/simple-domino-chain/workflow.yaml
```

Both dominos use `builtin:identity` — the chain validates ordering, hashing, and replay logging without domain-specific logic.

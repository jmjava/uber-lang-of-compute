# Snapshot Replay Tests

Integration tests verifying deterministic replay across multiple runs.

## Run

```bash
cd controller
go test ./pkg/engine/... -run TestSnapshotReplay -v
```

## What Is Verified

- Same workflow run twice produces identical snapshot IDs
- Input and output hashes match across runs
- Final output is byte-identical
- First run computes; structure of replay log is consistent

See `controller/pkg/engine/engine_test.go` for implementation.

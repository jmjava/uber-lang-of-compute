# Memoization Tests

Tests verifying hash-based memoization skips recomputation on cache hit.

## Run

```bash
cd controller
go test ./pkg/engine/... -run TestMemoization -v
```

## What Is Verified

- First run: all dominos show `reused: false`
- Second run (same store): all dominos show `reused: true`
- Final output unchanged between runs
- Memo cache keyed by `(snapshot_id, domino_id, input_hash)`

See `controller/pkg/engine/engine_test.go` for implementation.

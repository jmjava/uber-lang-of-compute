# Finance Curve Snapshot Example

Demonstrates deterministic curve interpolation and DV01 risk calculation from an immutable snapshot.

## Domain

From the blog post *Applicability to Finance* (Apr 15, 2025):

- Snapshot contains rate instruments (US2Y, US5Y, US10Y)
- Domino chain: load → interpolate → compute risk
- Node-local store holds curve points and memoized sub-curves
- Replay log enables historical audit of any calculation

## Run

```bash
cd controller
go build -o kbl-compute .

# First run — all dominos computed
./kbl-compute \
  --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/kbl-finance/replay-1.json

# Second run — all dominos reused from memo cache
./kbl-compute \
  --workflow ../examples/finance-curve-snapshot/workflow.yaml \
  --replay-log /tmp/kbl-finance/replay-2.json
```

Compare replay logs: first run shows `"reused": false` for all entries; second run shows `"reused": true`.

## Expected Output

The final domino produces risk metrics with DV01 values per tenor, derived deterministically from the sealed snapshot.

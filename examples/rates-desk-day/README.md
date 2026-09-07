# Rates desk day

A realistic overnight book: an 8-point UST par curve, two desks sharing memoized work, a T+1 steepener, then replica/CDC to a second store.

This is **engine use**, not a new theorem. The 3-instrument [finance-curve-snapshot](../finance-curve-snapshot/) example stays the workshop receipt.

## What happens

| Step | Who | What |
|------|-----|------|
| T 04-15 NY risk | `workflow-risk.yaml` | load → interpolate → KR01 DV01 on 1s–30s |
| T 04-15 LN mid | `workflow-mid.yaml` | load → interpolate only; **same domino names** so the sub-curve is reused |
| T+1 04-16 | `workflow-risk-next.yaml` | bear steepener (2s −8bp, 10s +6bp, 30s +10bp); new snapshot, full re-eval |
| London replica | tests | `replica.Materialize` copies the T spine; second store replays with memo |
| Wheel seats | `wheel.yaml` | NY then LN on the same slice (`windowDepth: 1`, `windowArity: 2`) |

Unfold still does **not** spawn live ComputeWheel contexts. The wheel YAML only names two seats.

## Run (CLI)

```bash
make build

# NY risk book — 3 evaluations
./controller/bin/kbl-compute \
  --workflow examples/rates-desk-day/workflow-risk.yaml \
  --store /tmp/kbl-rates-desk/ny.db \
  --replay-log /tmp/kbl-rates-desk/ny-risk.json

# London mid — load + interpolate reused from the NY store
./controller/bin/kbl-compute \
  --workflow examples/rates-desk-day/workflow-mid.yaml \
  --store /tmp/kbl-rates-desk/ny.db \
  --replay-log /tmp/kbl-rates-desk/ln-mid.json

# Next session — new Cauchy slice, no memo
./controller/bin/kbl-compute \
  --workflow examples/rates-desk-day/workflow-risk-next.yaml \
  --store /tmp/kbl-rates-desk/ny.db \
  --replay-log /tmp/kbl-rates-desk/ny-tplus1.json
```

On-the-run 3Y/7Y/10Y land on the input grid, so interpolate must print those par rates exactly (4.62 / 4.35 / 4.25 on T).

## Tests

```bash
cd controller && go test ./pkg/engine/ ./pkg/builtin/ ./internal/controller/ -run 'DeskDay|RatesDesk|Treasury' -count=1
```

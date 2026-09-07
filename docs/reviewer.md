# Reviewer walkthrough

A stranger can repeat the fabric without Kind, Kafka, or a cluster. This is the workshop path: **use the copy**, not invent a 33rd invariant.

## Command

From the repository root (Go 1.22+):

```bash
make review
```

That builds `kbl-compute` and `kbl-review`, runs the proof tests for the walkthrough, then executes:

```bash
./controller/bin/kbl-review --workflow examples/finance-curve-snapshot/workflow.yaml
```

Exit code 0 means every step below held. The JSON report on stdout is the receipt.

To spin the same assets for future research (toolchain + live `kbl-tsdb` on `:9090`):

```bash
make research-up
make research-test
```

See [research-env.md](research-env.md).

## What it proves, in order

| Step | What a reviewer should see |
|------|----------------------------|
| Sealed snapshot | `snapshot_id` is a content address of the finance curve YAML |
| Builtin chain | load → interpolate → risk evaluates (`first_evaluations: 3`) |
| Memo replay | second run reuses all three (`second_reuses: 3`), same `head_link` |
| Wheel lookahead | `lookahead_name` is a function of wheel state (player-piano next seat) |
| HeadLink fan-out | `fanout_universes` copies the same sealed record to the other universes |
| Idempotent bus | the same `event_id` is not delivered twice on retry |

## What it is not

- Not a Kind lab, Volcano job, or OpenKruise hot-swap. Those remain [getting-started.md](getting-started.md) Path 2.
- Not Unfold spawning live ComputeWheel contexts. M32 still only checks that optional `windowDepth`/`windowArity` seats equal \(k^d\). Wiring Unfold → Contexts is later engineering.
- Not joules, Hilbert space, or a physics claim. Physics-shaped names are [nature-inspired](theory/nature-inspired.md) copies.

## Manual CLI (same chain, no fan-out)

```bash
make build
./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-review/store.db \
  --replay-log /tmp/kbl-review/replay-1.json
./controller/bin/kbl-compute \
  --workflow examples/finance-curve-snapshot/workflow.yaml \
  --store /tmp/kbl-review/store.db \
  --replay-log /tmp/kbl-review/replay-2.json
```

The second replay log must show `"reused": true` on every entry.

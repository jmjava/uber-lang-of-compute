# Scheduling Tests

Scheduling is implemented via the **ComputeWheel reconciler** (Phase 3).

## Run

```bash
cd controller
go test ./internal/controller/... -run TestComputeWheel -v
go test ./pkg/wheel/... -v
```

## What Is Verified

- ComputeWheel creates Workflow resources per context×time-slice slot
- Context rotation within a time slice (ctx-a → ctx-b → …)
- Time slice advance after all contexts complete
- `maxRotations` stops the wheel after N slice rotations
- `preProvisionNext` creates the next slot's Workflow while current runs (player-piano)

## Future Tests

- Clock-driven requeue when interval elapses with no in-flight work
- Node affinity enforcement via ComputeContext → nodeName
- OpenKruise pre-warmed container slots

See [ADR 0006](../../docs/adr/0006-compute-wheel-rotation.md).

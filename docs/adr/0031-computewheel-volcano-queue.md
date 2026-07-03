# ADR 0031: ComputeWheel Volcano Queue Assignment

## Status

Accepted

## Context

Phase 26 added `runtime: volcano-init` on DominoChain so the controller emits Volcano Jobs. Phase 25 installed Volcano in the Kind lab with queue `kbl-lab`. ADR 0030 explicitly deferred **ComputeWheel → Volcano queue assignment** to Phase 27.

ComputeWheel (ADR 0006) rotates contexts through time slices and materializes a child **Workflow** per slot from `workflowTemplate`. Workflow → DominoChain → VCJob is already wired for container runtimes. The missing link was propagating **Volcano queue**, **node affinity**, and **runner image** from the wheel template into child Workflows and DominoChains.

## Decision

### 1. Wheel-level defaults

Extend `ComputeWheel.spec`:

| Field | Purpose |
|-------|---------|
| `volcanoQueue` | Default Volcano queue for all slots (overridable in `workflowTemplate.execution.volcanoQueue`) |
| `nodeSelector` | Default node pin for volcano-init chains (overridable in `workflowTemplate.provisioning.nodeSelector`) |

### 2. Workflow template extensions

Extend `ExecutionSpec` and `ProvisioningSpec` (Workflow + wheel template):

| Field | Location | Flows to |
|-------|----------|----------|
| `runtime` | `execution` | DominoChain runtime (CRD enum includes `volcano-init`) |
| `volcanoQueue` | `execution` | DominoChain `volcanoQueue` → VCJob `spec.queue` |
| `runnerImage` | `provisioning` | DominoChain `runnerImage` |
| `nodeSelector` | `provisioning` | DominoChain `nodeSelector` → task pod template |

### 3. Builder changes

- `pkg/wheel/workflow_builder.go` — merge wheel-level `volcanoQueue` / `nodeSelector` into child Workflow; label `kbl.io/volcano-queue`
- `pkg/dominochain/convert.go` — copy `volcanoQueue`, `runnerImage`, `nodeSelector` from Workflow to DominoChain

Existing reconcilers unchanged: ComputeWheel → Workflow → DominoChain (`volcano-init`) → VCJob.

### 4. Lab demo

Replace standalone `DominoChain` demo with `ComputeWheel julia-finance-wheel`:

- `maxRotations: 1` for deterministic lab completion
- `volcanoQueue: kbl-lab`, `runtime: volcano-init`, Julia dominos
- `up.sh` waits for wheel phase `Idle`

## Consequences

- Time-slice rotation and Volcano batch scheduling compose without hand-written VCJobs per slice
- Wheel template can override wheel-level queue/nodeSelector per workflow if needed
- `preProvisionNext` pre-creates the next slot's Workflow (and thus next VCJob chain) while current runs
- Per-context queue maps (context name → queue) remain future work

## References

- ADR 0006 — Compute Wheel Rotation
- ADR 0016 — ComputeWheel Workflow Template CR References
- ADR 0030 — Controller Volcano Emission
- `examples/compute-wheel/wheel-volcano.yaml`
- `lab/manifests/volcano/computewheel-julia-finance.yaml`

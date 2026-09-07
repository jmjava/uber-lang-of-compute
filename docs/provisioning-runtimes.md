# Provisioning Runtimes

How KBL executes domino chains in-cluster. Each runtime is selected via `DominoChain.spec.runtime` or `Workflow.spec.execution.runtime`.

**Full diagram set:** [diagrams.md](diagrams.md) (§5–8: runtime comparison, init pod, OpenKruise, Volcano pipeline).

## Overview

```mermaid
flowchart TB
  subgraph triggers [Entry points]
    WF[Workflow]
    DC[DominoChain]
    CW[ComputeWheel]
  end

  subgraph runtimes [Provisioning runtimes]
    LOCAL[local — engine in controller]
    INIT[kubernetes-init — Pod init chain]
    OK[openkruise — runner slots]
    VOL[volcano-init — VCJob batch]
  end

  CW --> WF
  WF -->|NeedsContainerRuntime| DC
  DC --> INIT
  DC --> OK
  DC --> VOL
  WF --> LOCAL
```

## Runtime comparison

| | `local` | `kubernetes-init` | `openkruise` | `volcano-init` |
|---|---------|-------------------|--------------|----------------|
| **K8s objects** | None (in-process) | Pod + init containers | Pod with runner slots | Volcano Job + init containers |
| **Scheduler** | N/A | default-scheduler | default-scheduler | Volcano |
| **Handoff** | In-memory / store | emptyDir `/kbl/handoff` | emptyDir `/kbl/handoff` | emptyDir `/kbl/handoff` |
| **Step execution** | Sequential in engine | Sequential init containers | Parallel start, file-wait serialize | Sequential init containers in VCJob task |
| **Dependencies** | None | Standard K8s | OpenKruise CRD | Volcano CRD |
| **Blog mapping** | CLI / dev path | Standard domino chain | Player-piano hot-swap | SyncSet / batch provisioning |
| **ADR** | MVP | [0007](adr/0007-hot-swapped-dominos-implementation.md) | [0007](adr/0007-hot-swapped-dominos-implementation.md) | [0030](adr/0030-controller-volcano-emission.md) |

## `local`

Default when `spec.execution.runtime` is empty or `local`. The workflow reconciler runs `engine.Run()` in-process. No Pods created.

**Use for:** CLI, unit tests, `finance-lab` workflow in Kind.

```yaml
spec:
  execution:
    runtime: local
    chain: [load, transform]
```

## `kubernetes-init`

DominoChain reconciler creates a Pod whose **init containers** run steps sequentially. A pause container keeps the Pod alive after inits complete. Snapshot data comes from a ConfigMap volume.

**Use for:** Any Kubernetes cluster without extra schedulers.

```yaml
apiVersion: kbl.io/v1alpha1
kind: DominoChain
spec:
  runtime: kubernetes-init
  runnerImage: kbl-domino-runner-julia:lab
  steps:
    - { name: load, command: julia:identity }
    - { name: interp, command: julia:interpolate }
```

Example: [examples/julia-domino-chain/dominochain-init.yaml](../examples/julia-domino-chain/dominochain-init.yaml). Pod layout: [diagrams.md §6](diagrams.md#6-kubernetes-init-pod-anatomy).

### Environment (all container runtimes)

| Variable | Purpose |
|----------|---------|
| `KBL_COMMAND` | Domino command (e.g. `julia:greeks`) |
| `KBL_INPUT` | Input JSON path |
| `KBL_OUTPUT` | Output JSON path |
| `KBL_JULIA_PROJECT` | Set automatically for `julia:*` commands |

## `openkruise`

`kruise-daemon` **ImagePullJob** prefetches the runner image onto `spec.nodeSelector` nodes. After the job succeeds, one Pod runs a **runner container per step**. Containers start together and serialize through a shared `emptyDir` handoff (`output-N.json`). OpenKruise 1.6 CRR cannot change image or env, so the lab does not use pause placeholders plus hot-swap.

**Use for:** Same-pod player-piano slots on an OpenKruise-enabled cluster, with an actual OpenKruise CR (ImagePullJob).

**Requires:** OpenKruise installed (`lab/scripts/install-openkruise.sh`). Missing ImagePullJob CRD fails the chain.

```yaml
spec:
  runtime: openkruise
  nodeSelector:
    kbl.io/lab-role: compute
  steps:
    - { name: load, command: julia:identity }
```

Lab demo: `Workflow/julia-finance-openkruise` → `DominoChain/julia-finance-openkruise-dchain` from catalog Snapshot + Domino CRs (Phase 35). Phase 36 emits `ImagePullJob` before the runner-slot Pod.

## `volcano-init`

DominoChain reconciler creates a **Volcano Job** (`batch.volcano.sh/v1alpha1`) with a single task whose pod template mirrors the init-chain layout. Volcano handles queueing and gang scheduling.

**Use for:** Batch workloads, queue fairness, blog SyncSet semantics.

**Requires:** Volcano installed (`lab/scripts/install-volcano.sh`).

```yaml
spec:
  runtime: volcano-init
  volcanoQueue: kbl-lab
  nodeSelector:
    kbl.io/lab-role: compute
  steps:
    - { name: load, command: julia:identity }
```

Optional fields:

| Field | Default | Purpose |
|-------|---------|---------|
| `volcanoQueue` | `default` | Volcano queue name |
| `nodeSelector` | none | Pin task pods to nodes |
| `runnerImage` | domino-runner default | Julia: `kbl-domino-runner-julia:lab` |

### ComputeWheel + Volcano (Phase 27)

ComputeWheel stamps child Workflows with `execution.runtime: volcano-init` and `volcanoQueue` from wheel spec or template:

```yaml
apiVersion: kbl.io/v1alpha1
kind: ComputeWheel
spec:
  volcanoQueue: kbl-lab
  workflowTemplate:
    execution:
      runtime: volcano-init
      chain: [load, interp, greeks]
```

Pipeline: **ComputeWheel → Workflow → DominoChain → VCJob**.

Example: [examples/compute-wheel/wheel-volcano.yaml](../examples/compute-wheel/wheel-volcano.yaml). Pipeline: [diagrams.md §8](diagrams.md#8-volcano-batch-path-lab-demo).

## Workflow → DominoChain bridge

When `Workflow.spec.execution.runtime` is not `local`, the workflow reconciler creates a child `DominoChain` named `{workflow}-dchain` and polls until complete.

`NeedsContainerRuntime` is true when:

- `runtime` is set and not `local`, or
- any domino specifies a container `image`

Provisioning fields propagate to DominoChain:

| Workflow field | DominoChain field |
|----------------|-------------------|
| `execution.runtime` | `spec.runtime` |
| `execution.volcanoQueue` | `spec.volcanoQueue` |
| `provisioning.runnerImage` | `spec.runnerImage` |
| `provisioning.nodeSelector` | `spec.nodeSelector` |

## Kind lab matrix

| Demo | Runtime | Install flag |
|------|---------|--------------|
| `finance-lab` | local | always |
| `julia-finance-wheel` | volcano-init | `KBL_LAB_VOLCANO=1` (default) |
| `julia-finance-openkruise` | openkruise | `KBL_LAB_OPENKURISE=1` (default) |

See [getting-started.md](getting-started.md) and [lab/README.md](../lab/README.md).

## References

- [ADR 0007 — Hot-Swapped Dominos Implementation](adr/0007-hot-swapped-dominos-implementation.md)
- [ADR 0029 — Volcano Kind Lab](adr/0029-volcano-kind-lab.md)
- [ADR 0030 — Controller Volcano Emission](adr/0030-controller-volcano-emission.md)
- [ADR 0031 — ComputeWheel Volcano Queue](adr/0031-computewheel-volcano-queue.md)
- [ADR 0032 — OpenKruise Kind Lab](adr/0032-openkruise-kind-lab.md)
- [ADR 0039 — OpenKruise ImagePullJob](adr/0039-openkruise-imagepulljob.md)

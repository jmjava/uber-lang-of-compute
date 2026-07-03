# Architecture Diagrams

Visual reference for the KBL Compute Engine. All diagrams use [Mermaid](https://mermaid.js.org/); they render on GitHub and in most Markdown viewers.

See also: [architecture.md](architecture.md) (§ [Multiverse communication](architecture.md#multiverse-communication)), [provisioning-runtimes.md](provisioning-runtimes.md), [getting-started.md](getting-started.md).

---

## 0. README overview — multiple KBL fabrics

Compact banner used at the top of the root README. Multiple **Pluggable Universes** coordinate via **Multiverse routing + Kafka**; controllers never call each other directly.

```mermaid
flowchart LR
  DSL["Describe<br/>4 DSLs"] --> K8S["Orchestrate<br/>CRDs + controller"]
  K8S --> FAB["Execute locally<br/>wheel · memo · store"]
  FAB --> UA["Universe A"]
  UA -->|sealed events · CDC| KFK[(Kafka)]
  KFK --> UB["Universe B …N"]
  MV["Multiverse<br/>routing rules"] -.-> KFK
```

---

## 1. Four DSLs → Kubernetes

Blog meta-model mapped to this repository:

```mermaid
flowchart TB
  subgraph dsl [Four DSLs specs/]
    EX[Execution DSL]
    DA[Data DSL]
    PR[Provisioning DSL]
    RO[Routing DSL]
  end

  subgraph crd [Kubernetes CRDs]
    WF[Workflow]
    SN[Snapshot]
    DM[Domino]
    DC[DominoChain]
    CC[ComputeContext]
    CW[ComputeWheel]
    PU[PluggableUniverse]
    MV[Multiverse]
  end

  subgraph ctrl [kbl-controller]
    WFR[Workflow reconciler]
    DCR[DominoChain reconciler]
    CWR[ComputeWheel reconciler]
    ENG[engine.Run + memo store]
  end

  EX --> WF
  EX --> DM
  EX --> DC
  DA --> SN
  PR --> DC
  RO --> PU
  RO --> MV
  RO --> CC

  CW --> WFR
  WF --> WFR
  WF --> DCR
  DC --> DCR
  WFR --> ENG
  DCR --> ENG
  CWR --> WF
```

---

## 2. End-to-end domino execution

```mermaid
sequenceDiagram
  participant Op as Operator
  participant WF as Workflow CR
  participant Ctrl as kbl-controller
  participant Store as Node-local store
  participant DC as DominoChain optional

  Op->>WF: apply sealed snapshot + chain
  Ctrl->>WF: reconcile

  alt runtime local
    Ctrl->>Store: open store path
    loop each domino
      Ctrl->>Store: hash inputs → memo lookup
      alt cache hit
        Ctrl->>Store: reuse output
      else cache miss
        Ctrl->>Ctrl: execute command
        Ctrl->>Store: persist output
      end
    end
    Ctrl->>WF: status Completed + replay log
  else runtime kubernetes-init / openkruise / volcano-init
    Ctrl->>DC: create DominoChain
    Ctrl->>DC: reconcile Pod / CRR / VCJob
    DC-->>Ctrl: phase Completed
    Ctrl->>Store: engine finalization
    Ctrl->>WF: status Completed
  end
```

---

## 3. Compute Wheel (Ferris Wheel)

Time-slice rotation across contexts ([ADR 0006](adr/0006-compute-wheel-rotation.md)):

```mermaid
stateDiagram-v2
  [*] --> Rotating: wheel created
  Rotating --> Processing: Workflow created for slot
  Processing --> Processing: Workflow Running
  Processing --> Rotating: Workflow Completed, next context
  Rotating --> Rotating: all contexts done → advance time slice
  Rotating --> Idle: maxRotations reached
  Processing --> Error: Workflow Error
  Idle --> [*]

  note right of Processing
    Slot = context × timeSlice
    preProvisionNext creates next Workflow early
  end note
```

```mermaid
flowchart LR
  subgraph sliceT [Time slice T]
    A[node-a-ctx] --> B[node-b-ctx] --> A2[node-a-ctx]
  end
  sliceT --> sliceT1[Time slice T+interval]

  A -.- WF1[Workflow T-a]
  B -.- WF2[Workflow T-b]
  A2 -.- WF3[Workflow T-a prime]
```

With Volcano (Phase 27), each stamped Workflow uses `runtime: volcano-init`:

```mermaid
flowchart LR
  CW[ComputeWheel] --> WF[Workflow per slot]
  WF --> DC[DominoChain]
  DC --> VJ[Volcano Job]
  VJ --> Q[kbl-lab queue]
```

---

## 4. Kind lab topology

Three-node cluster after `make lab-up` ([ADR 0029](adr/0029-volcano-kind-lab.md)):

```mermaid
flowchart TB
  subgraph cp [control-plane]
    CP[kube-system / CRDs registered]
  end

  subgraph w1 [worker w1 — kbl.io/lab-role=compute]
    VOL_POD[Volcano VCJob pods]
    OK_POD[OpenKruise chain pod]
  end

  subgraph w2 [worker w2 — kbl.io/tsdb-node=true]
    TSDB[kbl-tsdb Deployment]
    POND[/var/kbl Data Pond mount/]
  end

  subgraph kbl [kbl-system]
    CTRL[kbl-controller]
  end

  subgraph addons [optional addons]
    VS[volcano-system]
    KS[kruise-system]
  end

  CTRL --> TSDB
  VOL_POD --> w1
  OK_POD --> w1
  TSDB --> POND
  CTRL --> VS
  CTRL --> KS
```

| Node | Labels | Workloads |
|------|--------|-----------|
| control-plane | `kbl.io/lab-role=control-plane` | Kubernetes system |
| worker w1 | `kbl.io/lab-role=compute` | VCJob task pods, OpenKruise domino pod |
| worker w2 | `kbl.io/tsdb-node=true` | TSDB, node-local `/var/kbl` |

---

## 5. Provisioning runtimes compared

```mermaid
flowchart TB
  DC[DominoChain spec.runtime]

  DC -->|local via Workflow| LOC[engine in controller]
  DC -->|kubernetes-init| POD[Pod init chain]
  DC -->|openkruise| OK[Pod + CRR per step]
  DC -->|volcano-init| VJ[Volcano Job task]

  POD --> IC1[slot-0 init] --> IC2[slot-1 init] --> IC3[slot-2 init]
  OK --> PH[placeholder slots] --> CRR1[CRR swap slot-0] --> CRR2[CRR swap slot-1]
  VJ --> VQ[queue kbl-lab] --> VIC[init chain in task pod]
```

---

## 6. kubernetes-init pod anatomy

```mermaid
flowchart LR
  subgraph pod [Pod chain-name]
    subgraph vol [Volumes]
      CM[(ConfigMap snapshot)]
      ED[(emptyDir handoff)]
    end

    IC0[slot-0 init<br/>domino-runner]
    IC1[slot-1 init<br/>domino-runner]
    IC2[slot-2 init<br/>domino-runner]
    PAUSE[chain-complete<br/>pause]

    CM -->|/kbl/input| IC0
    IC0 -->|output.json| ED
    ED --> IC1
    IC1 --> ED
    ED --> IC2
    ED --> PAUSE
  end
```

Environment per init container: `KBL_COMMAND`, `KBL_INPUT`, `KBL_OUTPUT`, optional `KBL_JULIA_*`.

---

## 7. OpenKruise hot-swap sequence

Player-piano pattern ([ADR 0007](adr/0007-hot-swapped-dominos-implementation.md), lab demo Phase 28):

```mermaid
sequenceDiagram
  participant Ctrl as DominoChain reconciler
  participant Pod as Placeholder Pod
  participant CRR as ContainerRecreateRequest
  participant OK as OpenKruise controller

  Ctrl->>Pod: create pause slots slot-0..N
  loop each step index
    Ctrl->>CRR: create CRR for slot-i
    OK->>Pod: hot-swap slot-i → domino-runner
    Note over Pod: run julia:command<br/>read/write /kbl/handoff
    OK->>CRR: phase Completed
    Ctrl->>Ctrl: activeStep++
  end
  Ctrl->>Ctrl: engine.Run → Completed
```

```mermaid
flowchart TB
  subgraph slots [Same Pod over time]
    S0["slot-0: pause → runner → done"]
    S1["slot-1: pause → runner → done"]
    S2["slot-2: pause → runner → done"]
  end
  S0 --> S1 --> S2
  H[/kbl/handoff emptyDir persists across swaps/]
  S0 --- H
  S1 --- H
  S2 --- H
```

---

## 8. Volcano batch path (lab demo)

Full pipeline for `julia-finance-wheel`:

```mermaid
sequenceDiagram
  participant CW as ComputeWheel
  participant WFC as Wheel reconciler
  participant WF as Workflow
  participant WFR as Workflow reconciler
  participant DC as DominoChain
  participant DCR as DominoChain reconciler
  participant Vol as Volcano scheduler
  participant Job as VCJob

  WFC->>WF: BuildWorkflow timeSlice + volcanoQueue
  WFR->>DC: create dchain runtime volcano-init
  DCR->>Job: BuildVolcanoJob queue kbl-lab
  Vol->>Job: schedule task pod on compute worker
  Job-->>DCR: phase Completed
  DCR-->>WFR: chain Completed
  WFR-->>WFC: Workflow Completed → rotate wheel
```

---

## 9. Julia finance chain (lab demos)

Same three dominos, three provisioning paths:

```mermaid
flowchart LR
  SNAP[Sealed snapshot<br/>instruments JSON]

  SNAP --> I[julia:identity]
  I --> P[julia:interpolate]
  P --> G[julia:greeks]

  subgraph demos [Kind lab demos]
    L[finance-lab Workflow<br/>local engine]
    V[julia-finance-wheel<br/>volcano-init]
    O[julia-finance-openkruise<br/>openkruise]
  end

  G --> OUT[curve + greeks JSON]
  L --> G
  V --> G
  O --> G
```

---

## 10. Memoization and replay

```mermaid
flowchart TD
  IN[Domino inputs] --> HASH[SHA input hash]
  HASH --> LOOKUP{Store has hash?}
  LOOKUP -->|yes| REUSE[Skip execute reused=true]
  LOOKUP -->|no| RUN[Execute domino]
  RUN --> SAVE[Store output + hash]
  REUSE --> LOG[Replay log entry]
  SAVE --> LOG
  LOG --> NEXT[Next domino in chain]
```

---

## 11. Multiverse routing — multiple KBL fabrics

Event-driven coordination across Pluggable Universes. Works in one cluster (MemoryBus or Kafka) or across clusters sharing a Kafka/MSK backbone. Controllers do **not** peer with each other — only events and replicated sealed results cross universes.

```mermaid
flowchart LR
  subgraph uniA [Pluggable Universe A]
    WFA[Workflow + controller]
    StoreA[(node-local store)]
    WFA --> StoreA
  end

  subgraph bus [Event bus]
    MV[Multiverse routing]
    KFK[(Kafka / MemoryBus / Debezium CDC)]
    MV -.-> KFK
  end

  subgraph uniB [Pluggable Universe B …N]
    RRB[ReadReplica]
    StoreB[(node-local store)]
    RRB --> StoreB
  end

  WFA -->|snapshot completed| KFK
  StoreA -.->|CDC optional| KFK
  KFK -->|route + replicate| RRB
```

See [ADR 0009](adr/0009-multiverse-routing.md), [ADR 0011](adr/0011-read-replica-materialization.md), [ADR 0012](adr/0012-debezium-cdc-sync.md), and [architecture.md § Multiverse communication](architecture.md#multiverse-communication).

---

## 12. Kind lab troubleshooting

```mermaid
flowchart TD
  START[make lab-up failed?] --> KIND{Kind cluster up?}
  KIND -->|no| FIX1[kind delete cluster<br/>./lab/scripts/up.sh]
  KIND -->|yes| NODES{3 nodes?}
  NODES -->|no| FIX2[recreate cluster<br/>old single-node config]
  NODES -->|yes| CTRL{kbl-controller Running?}
  CTRL -->|no| FIX3[kubectl -n kbl-system logs<br/>deployment/kbl-controller]
  CTRL -->|yes| VOL{Volcano demo stuck?}
  VOL -->|yes| FIX4[kubectl -n volcano-system get pods<br/>kubectl get vcjob,wf,wheel]
  VOL -->|no| OK{OpenKruise demo stuck?}
  OK -->|yes| FIX5[kubectl -n kruise-system get pods<br/>kubectl get crr,pods -l kbl.io/openkruise-demo]
  OK -->|no| DONE[Check dchain phase + controller logs]
```

Common checks:

```bash
kubectl get nodes -L kbl.io/lab-role,kbl.io/tsdb-node
kubectl -n kbl-system get pods -o wide
kubectl get dchain,wf,wheel -A
kubectl -n kbl-system logs deployment/kbl-controller --tail=100
```

---

## 13. AWS target (CDK scaffold)

```mermaid
flowchart TB
  subgraph aws [KblPlatformStack Phase 22]
    VPC[VPC 2 AZ]
    ECR[ECR repos]
    EKS[EKS cluster]
  end

  subgraph deploy [Future phases]
    HELM[Volcano / OpenKruise Helm]
    FSX[FSx / node-local TSDB]
    IRSA[IRSA for controller]
  end

  LAB[Kind lab manifests] -->|reference| EKS
  ECR --> EKS
  EKS --> HELM
```

See [infra/aws/cdk/README.md](../infra/aws/cdk/README.md).

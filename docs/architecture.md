# Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     Uber Language of Compute                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌────────┐│
│  │  Execution   │ │     Data     │ │ Provisioning │ │Routing ││
│  │     DSL      │ │     DSL      │ │     DSL      │ │  DSL   ││
│  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘ └───┬────┘│
└─────────┼────────────────┼────────────────┼─────────────┼──────┘
          │                │                │             │
          ▼                ▼                ▼             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Control Plane                      │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐ │
│  │  Snapshot   │  │   Domino    │  │  PluggableUniverse /     │ │
│  │    CRD      │  │    CRD      │  │  ComputeWheel / Context  │ │
│  └──────┬──────┘  └──────┬──────┘  └────────────┬─────────────┘ │
└─────────┼────────────────┼──────────────────────┼───────────────┘
          │                │                      │
          ▼                ▼                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                      KBL Compute Fabric                          │
│                                                                  │
│   ┌────────────── Compute Wheel (time slices) ──────────────┐   │
│   │  Context A          Context B          Context C         │   │
│   │  ┌─────────┐       ┌─────────┐       ┌─────────┐        │   │
│   │  │ Snapshot│       │ Snapshot│       │ Snapshot│        │   │
│   │  │ Domino  │──►    │ Domino  │──►    │ Domino  │        │   │
│   │  │ chain   │       │ chain   │       │ chain   │        │   │
│   │  └────┬────┘       └────┬────┘       └────┬────┘        │   │
│   │       ▼                 ▼                 ▼              │   │
│   │  Node-local         Node-local         Node-local        │   │
│   │  store (TSDB/       store              store             │   │
│   │  SQLite/volume)                                            │   │
│   └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│   Memoization layer: input hash → output hash (skip if exists)   │
│   Replay log: snapshot_id, domino_id, hashes, reused/recomputed  │
└─────────────────────────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Multiverse Routing Layer                      │
│         Debezium/Kafka · Pluggable Universes · Time slices       │
└─────────────────────────────────────────────────────────────────┘
```

## Component Layers

### 1. DSL Layer (`specs/`)

Four YAML-based DSLs describe the compute fabric declaratively:

| DSL | Responsibility |
|-----|----------------|
| Execution | What compute steps run, in what order, with what dependencies |
| Data | Where snapshots live, schema, immutability guarantees |
| Provisioning | Node-local storage, container images, resource requests |
| Routing | Which universe handles which time slice / data partition |

### 2. CRD Layer (`crds/`)

Kubernetes Custom Resource Definitions translate DSL intent into cluster objects:

| CRD | Role |
|-----|------|
| `Snapshot` | Immutable data view for a time slice |
| `Domino` | Single deterministic compute step |
| `ComputeContext` | Node-associated unit of compute + data locality |
| `ComputeWheel` | Rotating set of contexts processing time slices |
| `PluggableUniverse` | Compute environment with its own execution/data/provisioning laws |

### 3. Controller Layer (`controller/`)

The controller reconciles CRDs and executes domino chains:

- Loads snapshot data from node-local store
- Resolves domino chain order (topological sort on `dependsOn`)
- Computes input hash; checks memoization store
- Executes or reuses result
- Writes replay log entry
- Updates CRD status

### 4. Node-Local Data Layer

Each Compute Context owns local storage:

- **MVP:** SQLite database on a local volume (`emptyDir` or host path)
- **Target:** Node-local TSDB DaemonSet, MemSQL, or Lustre/FSx staging layer

Data never crosses node boundaries during compute; routing happens at the scheduling layer.

## MVP Data Flow

```
1. Snapshot CRD created  →  data loaded into node-local store
2. Domino chain defined  →  controller resolves execution order
3. For each domino:
   a. Gather inputs (snapshot fields + prior domino outputs)
   b. Hash inputs → lookup memoization store
   c. HIT  → emit replay log (reused=true), pass output to next domino
   d. MISS → execute domino command, store result, emit replay log (reused=false)
4. Final output available in store + replay log
```

## Hot-Swapped Dominos (Future)

The 2025 design adds OpenKruise-based daisy-chain pods:

- 20 domino steps, only two active containers at once
- Placeholder containers + `emptyDir` for handoff
- Controller advances the chain in-place
- Enables granular updates without full pod restart

See [ADR 0004](./adr/0004-hot-swapped-dominos.md).

## Determinism Guarantees

| Guarantee | Mechanism |
|-----------|-----------|
| Same inputs → same outputs | Referentially transparent dominos, immutable snapshots |
| Reproducible replay | Replay log + snapshot ID + input/output hashes |
| Isolation | Snapshot immutability; no cross-snapshot reads during compute |
| Auditability | Every domino execution logged with hash chain |

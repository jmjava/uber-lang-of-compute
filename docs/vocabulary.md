# Vocabulary

Core terms from the Uber Language of Compute blog series, as used in the KBL Compute Engine.

| Term | Meaning |
|------|---------|
| **Uber Language of Compute** | Meta-language composed of four DSLs: Execution, Data, Provisioning, Routing. Describes the full compute fabric declaratively. |
| **Pluggable Universe** | A compute environment defined by its own laws: execution engine, data layer, and provisioning model. Can be swapped without changing the multiverse routing layer. |
| **Multiverse** | A routed collection of Pluggable Universes, often time-sliced replicas processing parallel or sequential data views. |
| **KBL / Kubernetes Based Lifeform** | A locality-aware Kubernetes compute organism: compute + local data + orchestration context, capable of spawning child KBLs. |
| **Compute Context** | Node-associated unit of compute and data locality. Binds a snapshot, domino chain, and node-local store together. |
| **Compute Wheel / Ferris Wheel** | A rotating set of compute contexts processing time slices continuously — like a Ferris wheel where each seat is a context. |
| **Windowed Mandelbrot** | Self-similar, hierarchical data/compute structure where only part of the graph exists at any given time. Enables fractal-style aggregation. |
| **Domino** | One deterministic, referentially transparent compute step tied to one immutable snapshot. Output depends only on declared inputs. |
| **Hot-Swapped Container** | A modular compute step swapped into a pod via OpenKruise ContainerRecreateRequest, without restarting the entire pipeline. |
| **Volcano Job / SyncSet** | Batch-scheduled domino chain executed by the Volcano scheduler; maps to `runtime: volcano-init` and `volcanoQueue` on DominoChain. |
| **Low-Entropy Snapshot** | An immutable data view that makes computation reproducible. Entropy is minimized by freezing the input state. |
| **Data Locality** | The principle that compute moves to data, not data to compute. Scheduling decisions prioritize node proximity to data. |
| **Player-Piano Scheduler** | A scheduler that pre-provisions resources ahead of need, like notes pre-positioned on a piano roll before they are played. |
| **Compute Fabric** | The overall system: time-sliced, data-local, Kubernetes-native infrastructure for deterministic chained computation. |
| **Replay Log** | An append-only record of domino executions: snapshot ID, domino ID, input hash, output hash, and whether the result was reused or recomputed. |

## Relationships

```
Multiverse
  └── PluggableUniverse (1..n)
        └── ComputeWheel
              └── ComputeContext (1..n, one per node/time-slice)
                    ├── Snapshot (immutable data view)
                    ├── Domino chain (ordered compute steps)
                    └── Node-local store (inputs, outputs, memo cache)
```

## Four DSLs

| DSL | Question It Answers |
|-----|---------------------|
| Execution | *What* runs, in what order, with what dependencies? |
| Data | *Where* does data live, what is its schema, how is immutability enforced? |
| Provisioning | *How* are resources (storage, containers, nodes) allocated? |
| Routing | *Which* universe/context handles which partition or time slice? |

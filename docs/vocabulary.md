# Vocabulary

Core terms from the Uber Language of Compute blog series, as used in the KBL Compute Engine.

Physics-shaped names are **nature-inspired abstractions** ([docs/theory/nature-inspired.md](./theory/nature-inspired.md)), like "neuron" in a neural net: a structure copied from nature, not a claim to be nature. Each name has a source-domain meaning, a computational copy, a grade (theorem / discretization / contract / bookkeeping / explorer / branching / control), and a remainder left in nature.

| Term | Meaning |
|------|---------|
| **Uber Language of Compute** | Meta-language composed of four DSLs: Execution, Data, Provisioning, Routing. Describes the full compute fabric declaratively. |
| **Pluggable Universe** | A compute environment defined by its own **laws** in the four-DSL sense: execution engine, data layer, and provisioning model — i.e. its own discrete evolution map \(\Phi_u\), not a Hamiltonian. Can be swapped without changing the multiverse routing layer. Each universe is a separate KBL fabric with its own node-local stores. |
| **Multiverse** | A **product of independent dynamical systems** plus routing: a routed collection of Pluggable Universes. Fan-out copies *sealed* records into other universes (branching mimic of many-worlds bookkeeping, not a wavefunction). Controllers do not call each other; the bus carries events. |
| **KBL / Kubernetes Based Lifeform** | A locality-aware Kubernetes compute unit. "Lifeform" copies homeostasis (spec vs status), the way "neural net" copies the neuron — not metabolism. |
| **Compute Context** | Node-associated unit of compute and data locality. Binds a snapshot, domino chain, and node-local store together. |
| **Compute Wheel / Ferris Wheel** | A rotating set of compute contexts processing time slices — discrete flow on a cylinder \(\mathbb{Z}_{n}\times T\): seats wrap, time only advances (theorem W1). Not rotational mechanics. |
| **Windowed Mandelbrot** | **Explorer mimic.** A finite unfolding of an infinite \(k\)-ary aggregation tree (`theory.Unfold`); only depth \(\le D\) is materialized. Coarsening recovers the shallower window (theorem F1). Escape-time is the iteration budget of a Mandelbrot-style viewport, not the scheduler's law of motion. |
| **Domino** | One deterministic, referentially transparent compute step tied to one immutable snapshot. Output depends only on declared inputs. |
| **Hot-Swapped Container** | A modular compute step swapped into a pod via OpenKruise ContainerRecreateRequest, without restarting the entire pipeline. |
| **Volcano Job / SyncSet** | Batch-scheduled domino chain executed by the Volcano scheduler; maps to `runtime: volcano-init` and `volcanoQueue` on DominoChain. |
| **Low-Entropy Snapshot** | An immutable data view that makes computation reproducible. **Entropy here is ensemble Shannon entropy** \(H_{\mathrm{ens}}\) of candidate inputs: sealing collapses the prior to a Dirac mass (\(H_{\mathrm{ens}}=0\)). It is not thermodynamic \(S\), and sealing does not compress the payload. |
| **Data Locality** | The principle that compute moves to data, not data to compute. Scheduling decisions prioritize node proximity to data. |
| **Player-Piano Scheduler** | A scheduler that pre-provisions resources ahead of need, like notes pre-positioned on a piano roll before they are played. `wheel.Lookahead` is that next note as a function of current state (M12). |
| **Compute Fabric** | The overall system: time-sliced, data-local, Kubernetes-native infrastructure for deterministic chained computation. |
| **Replay Log** | An append-only record of domino executions: snapshot ID, domino ID, input hash, output hash, and whether the result was reused or recomputed. |

## Relationships

```
Multiverse
  └── PluggableUniverse (1..n)     ← each = separate KBL fabric
        └── ComputeWheel
              └── ComputeContext (1..n, one per node/time-slice)
                    ├── Snapshot (immutable data view)
                    ├── Domino chain (ordered compute steps)
                    └── Node-local store (inputs, outputs, memo cache)

Cross-universe (not during domino compute):
  Universe A ── snapshot-completed / CDC events ──► Kafka / MemoryBus
       ► Multiverse routing rules ► ReadReplica ► Universe B local store
```

See [architecture.md § Multiverse communication](architecture.md#multiverse-communication).

Physics correspondences, theorems, and the simulated referee report: [theory/README.md](./theory/README.md).

## Four DSLs

| DSL | Question It Answers |
|-----|---------------------|
| Execution | *What* runs, in what order, with what dependencies? |
| Data | *Where* does data live, what is its schema, how is immutability enforced? |
| Provisioning | *How* are resources (storage, containers, nodes) allocated? |
| Routing | *Which* universe/context handles which partition or time slice? |

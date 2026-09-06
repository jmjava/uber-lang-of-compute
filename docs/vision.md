# KBL Compute Engine — Vision

**Subtitle:** A time-sliced, data-local, Kubernetes-native compute fabric

## What This Is

The KBL Compute Engine is not merely a batch runner or a streaming framework. It is a **compute fabric** — a locality-aware organism that processes immutable time-sliced data snapshots through modular, deterministic compute steps placed near local data stores.

The core thesis, drawn from the Uber Language of Compute blog series (2020–2025):

> Compute rearranges around data. Bring compute to the data, not data to compute.

Physics-shaped names are **nature-inspired abstractions**, in the same sense that a neural net is brain-inspired: we copy a structure nature already found useful (frozen initial data, unique trajectories, local interaction, recorded intermediates, independent worlds, homeostatic feedback, a windowed explorer of an infinite hierarchy) and leave the wetware in nature. Method, grades, and tests: [docs/theory](./theory/README.md), especially [nature-inspired.md](./theory/nature-inspired.md).

## Core Principles

1. **Snapshot isolation** — Every computation runs against an immutable, **ensemble-low-entropy** data view: sealing collapses the prior over candidate inputs to a single payload (Shannon \(H_{\mathrm{ens}}=0\)), not a drop in thermodynamic \(S\). Same snapshot + same inputs = same outputs, always.

2. **Data locality** — Each Compute Context binds compute to node-local storage. Work is scheduled where the data lives.

3. **Deterministic dominos** — A Domino is one referentially transparent compute step tied to one snapshot. Chains of dominos form reproducible pipelines.

4. **Entropy reduction through caching** — Hash inputs, memoize intermediate results, skip recomputation when a prior result exists.

5. **Time-sliced continuity** — The Compute Wheel rotates contexts through time slices, processing the multiverse of pluggable compute universes continuously.

6. **Four-language model** — Execution, Data, Provisioning, and Routing DSLs describe what runs, where data lives, how resources are provisioned, and how work is routed.

## What the MVP Proves

The first prototype does not build the full routed product of universes. It proves the computational correspondents of one slice (theorems D1, D2, M1 in [docs/theory](./theory/README.md)):

- Define a **Snapshot** resource (immutable data view)
- Define a **Domino** resource (deterministic compute step)
- Run a small chain of dominos against one snapshot
- Store inputs/outputs in node-local storage
- Hash inputs and skip a domino if the same result already exists
- Emit a **replay log**: snapshot ID, domino ID, input hash, output hash, reused vs recomputed

If replay + caching + locality work for one chain, the fabric scales.

## Long-Term Direction

- CRD/operator layer for Kubernetes-native lifecycle management
- Hot-swapped container dominos via OpenKruise daisy chains
- Node-local TSDB DaemonSet as live-data and cached-result store
- Debezium/Kafka routing across pluggable universes
- Self-similar hierarchical aggregation (Windowed Mandelbrot *pattern* — theorem F1; wheel seats = explorer leaves, milestone M3)
- Player-piano scheduling: pre-provision resources ahead of need

## Related Reading

See the [vocabulary](./vocabulary.md) for term definitions, [architecture](./architecture.md) for system design, and [theory](./theory/README.md) for the physics correspondences, formal model, and simulated peer review.

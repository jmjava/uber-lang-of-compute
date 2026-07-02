# KBL Compute Engine — Vision

**Subtitle:** A time-sliced, data-local, Kubernetes-native compute fabric

## What This Is

The KBL Compute Engine is not merely a batch runner or a streaming framework. It is a **compute fabric** — a locality-aware organism that processes immutable time-sliced data snapshots through modular, deterministic compute steps placed near local data stores.

The core thesis, drawn from the Uber Language of Compute blog series (2020–2025):

> Compute rearranges around data. Bring compute to the data, not data to compute.

## Core Principles

1. **Snapshot isolation** — Every computation runs against an immutable, low-entropy data view. Same snapshot + same inputs = same outputs, always.

2. **Data locality** — Each Compute Context binds compute to node-local storage. Work is scheduled where the data lives.

3. **Deterministic dominos** — A Domino is one referentially transparent compute step tied to one snapshot. Chains of dominos form reproducible pipelines.

4. **Entropy reduction through caching** — Hash inputs, memoize intermediate results, skip recomputation when a prior result exists.

5. **Time-sliced continuity** — The Compute Wheel rotates contexts through time slices, processing the multiverse of pluggable compute universes continuously.

6. **Four-language model** — Execution, Data, Provisioning, and Routing DSLs describe what runs, where data lives, how resources are provisioned, and how work is routed.

## What the MVP Proves

The first prototype does not build the full multiverse. It proves one slice of the physics:

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
- Self-similar hierarchical aggregation (Windowed Mandelbrot pattern)
- Player-piano scheduling: pre-provision resources ahead of need

## Related Reading

See the [vocabulary](./vocabulary.md) for term definitions and [architecture](./architecture.md) for system design.

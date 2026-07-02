# ADR 0004: Hot-Swapped Dominos

## Status

Proposed (post-MVP)

## Context

The 2025 blog posts describe a daisy-chain pod design where up to 20 domino steps share a pod, but only two containers are active at once. OpenKruise enables in-place container hot-swaps using placeholder containers and `emptyDir` handoff volumes. This allows granular updates to individual dominos without restarting the entire pipeline.

## Decision

Post-MVP, domino execution will migrate from CLI-invoked commands to Kubernetes pod-based hot-swapped containers:

1. A `DominoChain` pod template defines N placeholder slots
2. The controller activates domino N by hot-swapping container N+1 while N runs
3. Handoff data flows through `emptyDir` volumes shared between adjacent slots
4. Only two containers are live at any time (current + next pre-warmed)
5. In-place updates via OpenKruise `ContainerRecreateRequest` or equivalent

The MVP uses direct command execution to prove determinism and memoization without requiring OpenKruise in the cluster.

## Consequences

- MVP domino `spec.command` runs locally; future domino `spec.image` runs in-cluster
- Controller gains significant complexity (pod lifecycle, OpenKruise CRDs)
- Enables modular, granular updates — change one domino image without touching the chain
- Testing can isolate individual dominos in placeholder slots before chain integration

## References

- *Uber Language of Compute v2.0: AI powered design* (Apr 6, 2025)
- *Hot swapping containers in daisy chain* (Apr 6, 2025)

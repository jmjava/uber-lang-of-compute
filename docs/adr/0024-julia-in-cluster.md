# ADR 0024: Julia In-Cluster Execution (Multi-Container)

## Status

Accepted

## Context

ADR 0023 evaluated Julia deployment models and recommended **Model B (multi-container)**: one domino step per container via `domino-runner` inside DominoChain init containers or OpenKruise slots. Phase 14 shipped local subprocess execution only.

## Decision

1. **Julia runner image** — `controller/docker/domino-runner-julia/Dockerfile` bundles:
   - `domino-runner` Go binary
   - Julia 1.10 + pre-instantiated `controller/julia` project (`JSON.jl`)
   - Default tag: `ghcr.io/jmjava/kbl-domino-runner-julia:latest`

2. **DominoChain builder** — init containers running `julia:*` commands receive:
   - `KBL_JULIA_PROJECT=/opt/kbl/julia`
   - `KBL_JULIA_BIN=julia`

3. **Examples** — `examples/julia-domino-chain/dominochain-init.yaml` and `workflow-container.yaml` for kubernetes-init E2E

4. **Constants** — `DefaultJuliaRunnerImage`, `JuliaProjectContainerPath`, `IsJuliaCommand()`

Pooled single-container supervisor (Model C) remains deferred.

## Consequences

- Cluster operators must build/push the Julia runner image before in-cluster Julia chains run
- Builtin-only chains continue using `DefaultRunnerImage` (Go-only, smaller)
- OpenKruise hot-swap uses the same `runnerImage` annotation — Julia image works without builder changes beyond env injection

## References

- ADR 0023 — Julia Deployment Models
- ADR 0007 — Hot-Swapped Dominos Implementation
- `controller/docker/domino-runner-julia/Dockerfile`

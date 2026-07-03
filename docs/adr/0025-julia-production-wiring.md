# ADR 0025: Julia Production Wiring (CI + Multiverse)

## Status

Accepted

## Context

Phase 20 shipped the Julia domino-runner image and DominoChain init-container wiring (ADR 0024). Operators still needed CI validation of Docker builds, OpenKruise env parity, and a multiverse example tying `PluggableUniverse` routing to Julia workflows.

## Decision

1. **GitHub Actions** — `.github/workflows/docker-domino-runner.yml` builds standard and Julia runner images on relevant path changes
2. **Standard runner Dockerfile** — `controller/docker/domino-runner/Dockerfile` (distroless, builtin-only, smaller)
3. **OpenKruise CRR env** — `ContainerRecreateRequest` uses shared `stepEnv()` so `julia:*` steps get `KBL_JULIA_*` on hot-swap slots
4. **Multiverse example** — `julia-finance-universe` + `workflow-julia-rates.yaml` with partition label `kbl.io/partition-engine: julia`
5. **OpenKruise example** — `examples/julia-domino-chain/dominochain-openkruise.yaml`

GHCR publish credentials remain operator-specific; CI verifies builds only.

## Consequences

- Julia OpenKruise chains now receive the same env as init-container chains
- Multiverse routing demonstrates engine-based partition selection alongside asset-class routing
- Pooled Julia supervisor (Model C) and automated GHCR push remain deferred

## References

- ADR 0023 — Julia Deployment Models
- ADR 0024 — Julia In-Cluster Execution
- `.github/workflows/docker-domino-runner.yml`

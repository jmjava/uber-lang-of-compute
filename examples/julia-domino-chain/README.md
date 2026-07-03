# Julia domino chain example

Runs the finance curve workflow using **Julia subprocess dominos** instead of Go builtins.

## Prerequisites

1. [Julia](https://julialang.org/downloads/) 1.9+
2. Install domino script dependencies once:

```bash
julia --project=controller/julia -e 'using Pkg; Pkg.instantiate()'
```

## Run locally

```bash
make build
./controller/bin/kbl-compute --workflow examples/julia-domino-chain/workflow.yaml
```

Optional environment:

| Variable | Default | Purpose |
|----------|---------|---------|
| `KBL_JULIA_BIN` | `julia` | Julia executable |
| `KBL_JULIA_PROJECT` | `controller/julia` | Project with `JSON.jl` |

## Commands

Domino commands use the `julia:<script>` prefix:

| Command | Script |
|---------|--------|
| `julia:identity` | `scripts/identity.jl` |
| `julia:interpolate` | `scripts/interpolate.jl` |
| `julia:risk_dv01` | `scripts/risk_dv01.jl` |

## In-cluster deployment

Phase 14 runs Julia as a **local subprocess** (dev/CI). For Kubernetes, see [ADR 0023](../../docs/adr/0023-julia-deployment-models.md):

- **Recommended:** multi-container — one domino step per container via `domino-runner` + DominoChain (extends ADR 0007)
- **Optional spike:** single-container multi-process — shared Julia supervisor for lower step latency

### Build Julia runner image

From repo root:

```bash
make docker-domino-runner-julia
# or: docker build -f controller/docker/domino-runner-julia/Dockerfile \
#      -t ghcr.io/jmjava/kbl-domino-runner-julia:latest .
```

### Deploy init chain (kubernetes-init)

```bash
kubectl apply -f ../../crds/
kubectl apply -f dominochain-init.yaml
./../../controller/bin/kbl-controller --store-root /var/kbl/store
kubectl get dominochains julia-finance-init-chain -w
```

Or via Workflow with container runtime:

```bash
kubectl apply -f workflow-container.yaml
kubectl get dominochains -l kbl.io/dominochain -w
```

Each init container runs `domino-runner` with `KBL_JULIA_PROJECT=/opt/kbl/julia` injected automatically for `julia:*` commands.

## PluggableUniverse

The bundled `PluggableUniverse` sets `executionEngine.type: julia`. Workflow dominos declare explicit `julia:` commands; the universe CR documents the intended runtime for multiverse routing.

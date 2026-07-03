# Julia domino chain example

Runs the finance curve workflow using **Julia subprocess dominos** backed by [FinanceModels.jl](https://github.com/JuliaActuary/FinanceModels.jl) (ADR 0027).

## Prerequisites

1. [Julia](https://julialang.org/downloads/) 1.10+
2. Install domino script dependencies once (includes FinanceModels + locked `Manifest.toml`):

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
| `KBL_JULIA_PROJECT` | `controller/julia` | Project with `JSON.jl` + `FinanceModels.jl` |

## Commands

Domino commands use the `julia:<script>` prefix:

| Command | Script | FinanceModels usage |
|---------|--------|-------------------|
| `julia:identity` | `scripts/identity.jl` | Pass-through |
| `julia:interpolate` | `scripts/interpolate.jl` | `ZeroRateCurve(..., Spline.Linear())`, sample 3Y/7Y |
| `julia:risk_dv01` | `scripts/risk_dv01.jl` | Parallel +1bp bump, `present_value` on zero-coupon flows |

Rates in snapshot JSON are **percent quotes** (e.g. `4.25`); scripts convert to decimals internally.

## Sample output

After `interpolate`, `method` is `FinanceModels.ZeroRateCurve/Spline.Linear`. After `risk_dv01`, `method` is `FinanceModels.present_value/parallel_1bp`. Numeric results differ from Go `builtin:*` dominos (by design).

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

### Kind lab

After `make lab-up`:

```bash
kubectl apply -f examples/julia-domino-chain/dominochain-init.yaml
# Ensure runnerImage is kbl-domino-runner-julia:lab
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

OpenKruise hot-swap (requires OpenKruise installed):

```bash
kubectl apply -f dominochain-openkruise.yaml
kubectl get dominochains julia-finance-openkruise-chain -w
```

Each init container runs `domino-runner` with `KBL_JULIA_PROJECT=/opt/kbl/julia` injected automatically for `julia:*` commands.

## PluggableUniverse

The bundled `PluggableUniverse` sets `executionEngine.type: julia`. Workflow dominos declare explicit `julia:` commands; the universe CR documents the intended runtime for multiverse routing.

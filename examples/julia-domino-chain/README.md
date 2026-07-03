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
| `julia:greeks` | `scripts/greeks.jl` | Bond DV01/duration/convexity, zero-bucket DV01, Black–Scholes delta/gamma/vega/theta/rho |
| `julia:risk_dv01` | `scripts/risk_dv01.jl` | Legacy: parallel +1bp DV01 only |

Rates in snapshot JSON are **percent quotes** (e.g. `4.25`); scripts convert to decimals internally.

## Sample output

After `interpolate`, `method` is `FinanceModels.ZeroRateCurve/Spline.Linear`. After `greeks`, output includes `bond_greeks`, `rate_greeks`, and `option_greeks` (ADR 0028).

## In-cluster deployment

Phase 14 runs Julia as a **local subprocess** (dev/CI). For Kubernetes, see [ADR 0023](../../docs/adr/0023-julia-deployment-models.md) and [docs/provisioning-runtimes.md](../../docs/provisioning-runtimes.md).

| Runtime | Manifest | Lab demo (after `make lab-up`) |
|---------|----------|------------------------------|
| `kubernetes-init` | `dominochain-init.yaml` | Apply manually |
| `openkruise` | `dominochain-openkruise.yaml` | `DominoChain/julia-finance-openkruise` |
| `volcano-init` | `dominochain-volcano-init.yaml` | via `ComputeWheel/julia-finance-wheel` |

### Build Julia runner image

From repo root:

```bash
make docker-domino-runner-julia
# or: docker build -f controller/docker/domino-runner-julia/Dockerfile \
#      -t ghcr.io/jmjava/kbl-domino-runner-julia:latest .
```

### Kind lab

```bash
make lab-up

# Volcano wheel (automatic)
kubectl get wheel julia-finance-wheel -o wide
kubectl get vcjob -l kbl.io/volcano-demo=true

# OpenKruise hot-swap (automatic)
kubectl get dchain julia-finance-openkruise -o wide
kubectl logs -l kbl.io/openkruise-demo=true -c slot-2-compute-greeks

# Manual init chain
kubectl apply -f dominochain-init.yaml
# Set runnerImage: kbl-domino-runner-julia:lab
```

See [docs/getting-started.md](../../docs/getting-started.md).

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

### OpenKruise hot-swap

Requires OpenKruise (`lab/scripts/install-openkruise.sh` or `make lab-up`):

```bash
kubectl apply -f dominochain-openkruise.yaml
kubectl get dominochains julia-finance-openkruise-chain -w
kubectl get containerrecreaterequests.apps.kruise.io
```

### Volcano batch

Requires Volcano and queue `kbl-lab`:

```bash
kubectl apply -f dominochain-volcano-init.yaml
kubectl get vcjob -l kbl.io/dominochain -w
```

Each container runs `domino-runner` with `KBL_JULIA_PROJECT=/opt/kbl/julia` for `julia:*` commands.

## PluggableUniverse

The bundled `PluggableUniverse` sets `executionEngine.type: julia`. Workflow dominos declare explicit `julia:` commands; the universe CR documents the intended runtime for multiverse routing.

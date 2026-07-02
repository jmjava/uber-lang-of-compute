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

These mirror the Go `builtin:*` finance chain for deterministic comparison.

## PluggableUniverse

The bundled `PluggableUniverse` sets `executionEngine.type: julia`. Workflow dominos declare explicit `julia:` commands; the universe CR documents the intended runtime for multiverse routing.

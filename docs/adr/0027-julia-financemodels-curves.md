# ADR 0027: Julia FinanceModels Curve Dominos

## Status

Accepted

## Context

Phase 14 introduced `julia:*` domino commands with hand-rolled scripts mirroring Go `builtin:*` finance examples (linear interpolation, simplified DV01). Phase 22 added runnable in-cluster images. Users asked for **real curve math** using Julia finance libraries while preserving KBL snapshot isolation, memoization, and replay.

## Decision

### 1. FinanceModels.jl dependency

Add [FinanceModels.jl](https://github.com/JuliaActuary/FinanceModels.jl) to `controller/julia/Project.toml` with a committed **`Manifest.toml`** for reproducible Docker/CI builds.

### 2. Updated domino scripts

| Script | Implementation |
|--------|----------------|
| `interpolate.jl` | `ZeroRateCurve(rates, tenors, Spline.Linear())`; sample `zero(curve, t)` at 3Y and 7Y |
| `risk_dv01.jl` | Parallel +1bp pillar bump; `present_value` on `Cashflow(notional, tenor)` before/after |

Snapshot JSON rates remain **percent quotes** (e.g. `4.25`); scripts convert to decimals for FinanceModels and emit percent in `interpolated` for readability.

`interpolate` output includes `curve.pillars` metadata so `risk_dv01` can rebuild the same curve for bump-and-reprice.

### 3. Output schema

Backward-compatible with the finance workflow chain:

- `curve_points`, `interpolated`, `method` (now FinanceModels-specific strings)
- New: `curve.pillars` for downstream risk

Julia output **no longer matches** Go builtin numeric results (different interpolation and DV01 model). Tests assert **determinism** and FinanceModels method markers instead of builtin parity.

### 4. Docker / CI

`domino-runner-julia` precompiles `JSON` + `FinanceModels` at build time.

## Consequences

- Heavier Julia image and longer first CI build (FinanceModels dependency tree)
- Deterministic replay requires pinned Manifest versions
- Future dominos can add Nelson–Siegel fit, swap PV, etc. as new `julia:*` scripts
- Kind lab can apply `examples/julia-domino-chain/` with `kbl-domino-runner-julia:lab`

## References

- ADR 0022 — Julia pluggable execution
- ADR 0023 — Julia deployment models
- `examples/julia-domino-chain/`
- FinanceModels `ZeroRateCurve` docs

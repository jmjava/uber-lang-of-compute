# ADR 0028: Julia Greeks Domino

## Status

Accepted

## Context

Phase 23 added FinanceModels curve interpolation and simplified DV01 via `julia:risk_dv01`. Users asked for **Greeks** — sensitivities of prices to market inputs — for both rates and options.

## Decision

### 1. New domino `julia:greeks`

Single script `controller/julia/scripts/greeks.jl` computes:

| Bucket | Greeks | Method |
|--------|--------|--------|
| **Fixed-rate bond** | DV01, modified duration, convexity | Parallel ± bump on curve pillars; `Bond.Fixed` + `present_value` |
| **Zero-coupon buckets** (3Y, 7Y) | DV01 | +1bp parallel bump on pillar curve |
| **European option** | delta, gamma, vega, theta, rho | Closed-form Black–Scholes (`SpecialFunctions.erf`) |

### 2. Snapshot / payload contract

Optional fields on snapshot inline JSON (passed through `identity` → `interpolate` → `greeks`):

```yaml
bond:
  coupon_pct: 4.5
  frequency: 2
  maturity_years: 5
option:
  type: call
  spot: 100
  strike: 100
  maturity_years: 1
  rate_pct: 5
  volatility: 0.2
```

Defaults apply when omitted. `interpolate.jl` forwards `bond` and `option` keys unchanged.

### 3. Workflow change

`examples/julia-domino-chain/workflow.yaml` replaces final step `julia:risk_dv01` with `julia:greeks`. Legacy `risk_dv01.jl` remains for narrow DV01-only use.

### 4. Dependencies

Add `SpecialFunctions` to `controller/julia/Project.toml` for normal CDF in Black–Scholes formulas.

## Consequences

- Final workflow output is richer (bond + rate + option sensitivities)
- Option greeks use analytic BS, not FinanceModels `Option.EuroCall` (spot not wired in FM API without fit quotes)
- Rate greeks remain deterministic under pinned Manifest + fixed bump sizes
- Vega/rho reported per **1 vol point** / **1 rate point** (0.01); theta per calendar day

## References

- ADR 0027 — FinanceModels curves
- `examples/julia-domino-chain/`

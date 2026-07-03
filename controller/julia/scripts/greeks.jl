#!/usr/bin/env julia
# Rate/bond and Black-Scholes option Greeks via FinanceModels + parallel bump.
using JSON
using FinanceModels
using SpecialFunctions: erf

const NOTIONAL = 1_000_000.0
const BP_SHIFT = 0.0001
const BUMP_H = 1e-4  # symmetric finite-difference step (decimal rate / vol)

function tenor_years(label::AbstractString)
    endswith(label, "Y") || error("greeks: unsupported tenor label $(label)")
    return parse(Float64, label[1:end-1])
end

function build_curve(pillars)
    tenors = [Float64(p["tenor_years"]) for p in pillars]
    rates = [Float64(p["rate_pct"]) / 100.0 for p in pillars]
    return tenors, rates
end

function bond_greeks(rates, tenors, coupon_pct, frequency, maturity_years, notional)
    bond = Bond.Fixed(coupon_pct / 100.0, Periodic(frequency), maturity_years)
    curve = ZeroRateCurve(rates, tenors, Spline.Linear())
    curve_up = ZeroRateCurve(rates .+ BUMP_H, tenors, Spline.Linear())
    curve_dn = ZeroRateCurve(rates .- BUMP_H, tenors, Spline.Linear())

    pv0 = present_value(curve, bond) * notional
    pv_up = present_value(curve_up, bond) * notional
    pv_dn = present_value(curve_dn, bond) * notional

    dv01 = round(abs(pv0 - pv_up) / BUMP_H * BP_SHIFT; digits=2)
    mod_dur = round(-(pv_up - pv_dn) / (2 * BUMP_H * pv0); digits=6)
    conv = round((pv_up + pv_dn - 2 * pv0) / (pv0 * BUMP_H^2); digits=6)

    return Dict(
        "coupon_pct" => coupon_pct,
        "frequency" => frequency,
        "maturity_years" => maturity_years,
        "pv" => round(pv0; digits=2),
        "dv01" => dv01,
        "modified_duration" => mod_dur,
        "convexity" => conv,
    )
end

function zero_bucket_greeks(rates, tenors, interpolated, notional)
    curve = ZeroRateCurve(rates, tenors, Spline.Linear())
    bumped = ZeroRateCurve(rates .+ BP_SHIFT, tenors, Spline.Linear())
    metrics = []
    for (tenor, rate_pct) in sort(collect(interpolated))
        years = tenor_years(tenor)
        cf = Cashflow(notional, years)
        pv0 = present_value(curve, cf)
        pv1 = present_value(bumped, cf)
        push!(metrics, Dict(
            "tenor" => tenor,
            "rate" => Float64(rate_pct),
            "dv01" => round(abs(pv0 - pv1); digits=2),
        ))
    end
    metrics
end

# Closed-form Black-Scholes Greeks (European, no dividends).
function norm_cdf(x)
    0.5 * (1 + erf(x / sqrt(2.0)))
end

function norm_pdf(x)
    exp(-0.5 * x^2) / sqrt(2 * pi)
end

function black_scholes_greeks(spec)
    S = Float64(spec["spot"])
    K = Float64(spec["strike"])
    T = Float64(get(spec, "maturity_years", 1.0))
    r = Float64(get(spec, "rate_pct", 5.0)) / 100.0
    sigma = Float64(get(spec, "volatility", 0.2))
    opt_type = lowercase(string(get(spec, "type", "call")))

    sqrtT = sqrt(T)
    d1 = (log(S / K) + (r + 0.5 * sigma^2) * T) / (sigma * sqrtT)
    d2 = d1 - sigma * sqrtT
    nd1 = norm_cdf(d1)
    nd2 = norm_cdf(d2)
    npd1 = norm_pdf(d1)

    if opt_type == "call"
        delta = nd1
        theta = -(S * npd1 * sigma / (2 * sqrtT)) - r * K * exp(-r * T) * nd2
        rho = K * T * exp(-r * T) * nd2
    elseif opt_type == "put"
        delta = nd1 - 1.0
        theta = -(S * npd1 * sigma / (2 * sqrtT)) + r * K * exp(-r * T) * norm_cdf(-d2)
        rho = -K * T * exp(-r * T) * norm_cdf(-d2)
    else
        error("greeks: option.type must be call or put")
    end

    gamma = npd1 / (S * sigma * sqrtT)
    vega = S * npd1 * sqrtT

    Dict(
        "type" => opt_type,
        "spot" => S,
        "strike" => K,
        "maturity_years" => T,
        "rate_pct" => r * 100.0,
        "volatility" => sigma,
        "delta" => round(delta; digits=6),
        "gamma" => round(gamma; digits=6),
        "vega" => round(vega / 100.0; digits=6),  # per 1 vol point (0.01)
        "theta" => round(theta / 365.0; digits=6), # per calendar day
        "rho" => round(rho / 100.0; digits=6),     # per 1 rate point (0.01)
    )
end

input_path, output_path = ARGS[1], ARGS[2]
payload = JSON.parse(read(input_path, String))

curve_meta = get(payload, "curve", nothing)
curve_meta === nothing && error("greeks: missing curve metadata")
pillars = get(curve_meta, "pillars", nothing)
pillars === nothing && error("greeks: missing curve.pillars")

interpolated = get(payload, "interpolated", Dict{String, Any}())
tenors, rates = build_curve(pillars)

bond_spec = get(payload, "bond", Dict(
    "coupon_pct" => 4.5,
    "frequency" => 2,
    "maturity_years" => 5.0,
))
option_spec = get(payload, "option", Dict(
    "type" => "call",
    "spot" => 100.0,
    "strike" => 100.0,
    "maturity_years" => 1.0,
    "rate_pct" => 5.0,
    "volatility" => 0.2,
))

result = Dict(
    "method" => "FinanceModels.greeks/parallel_bump+BlackScholes",
    "notional" => NOTIONAL,
    "bond_greeks" => bond_greeks(
        rates,
        tenors,
        Float64(bond_spec["coupon_pct"]),
        Int(bond_spec["frequency"]),
        Float64(bond_spec["maturity_years"]),
        NOTIONAL,
    ),
    "rate_greeks" => zero_bucket_greeks(rates, tenors, interpolated, NOTIONAL),
    "option_greeks" => black_scholes_greeks(option_spec),
)

write(output_path, JSON.json(result))

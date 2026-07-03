#!/usr/bin/env julia
# Parallel +1bp pillar bump DV01 on zero-coupon notionals via FinanceModels.jl.
using JSON
using FinanceModels

const NOTIONAL = 1_000_000.0
const BP_SHIFT = 0.0001

function tenor_years(label::AbstractString)
    endswith(label, "Y") || error("risk_dv01: unsupported tenor label $(label)")
    return parse(Float64, label[1:end-1])
end

input_path, output_path = ARGS[1], ARGS[2]
payload = JSON.parse(read(input_path, String))

interpolated = get(payload, "interpolated", Dict{String, Any}())
interpolated == Dict{String, Any}() && error("risk_dv01: missing interpolated")

curve_meta = get(payload, "curve", nothing)
curve_meta === nothing && error("risk_dv01: missing curve metadata from interpolate step")

pillars = get(curve_meta, "pillars", nothing)
pillars === nothing && error("risk_dv01: missing curve.pillars")
tenors = [Float64(p["tenor_years"]) for p in pillars]
rates = [Float64(p["rate_pct"]) / 100.0 for p in pillars]

curve = ZeroRateCurve(rates, tenors, Spline.Linear())
bumped = ZeroRateCurve(rates .+ BP_SHIFT, tenors, Spline.Linear())

risk_metrics = []
for (tenor, rate_pct) in sort(collect(interpolated))
    years = tenor_years(tenor)
    cf = Cashflow(NOTIONAL, years)
    pv0 = present_value(curve, cf)
    pv1 = present_value(bumped, cf)
    push!(risk_metrics, Dict(
        "tenor" => tenor,
        "rate" => Float64(rate_pct),
        "dv01" => round(abs(pv0 - pv1); digits=2),
    ))
end
sort!(risk_metrics, by = r -> r["tenor"])

result = Dict(
    "risk_metrics" => risk_metrics,
    "notional" => NOTIONAL,
    "method" => "FinanceModels.present_value/parallel_1bp",
)

write(output_path, JSON.json(result))

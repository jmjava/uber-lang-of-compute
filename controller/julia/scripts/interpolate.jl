#!/usr/bin/env julia
# Build a zero-rate curve with FinanceModels.jl and sample 3Y / 7Y tenors.
using JSON
using FinanceModels

const REFERENCE_YEAR = 2025

function parse_instruments(payload)
    instruments = get(payload, "instruments", nothing)
    instruments === nothing && error("interpolate: missing instruments")
    return instruments
end

function maturity_to_years(maturity::AbstractString)
    length(maturity) >= 4 || return 0.0
    year = parse(Int, maturity[1:4])
    return float(year - REFERENCE_YEAR)
end

function rate_decimal(inst)
    r = Float64(inst["rate"])
    # Snapshot rates are quoted in percent (e.g. 4.25); FinanceModels expects decimals.
    return r / 100.0
end

function rate_percent(continuous_rate)
    round(100.0 * continuous_rate.continuous_value; digits=6)
end

input_path, output_path = ARGS[1], ARGS[2]
payload = JSON.parse(read(input_path, String))
instruments = parse_instruments(payload)

points = map(instruments) do inst
    Dict(
        "maturity" => string(inst["maturity"]),
        "rate" => Float64(inst["rate"]),
        "tenor_years" => maturity_to_years(string(inst["maturity"])),
    )
end
sort!(points, by = p -> p["maturity"])

tenors = [p["tenor_years"] for p in points]
rates = [rate_decimal(inst) for inst in sort(instruments, by = i -> string(i["maturity"]))]

curve = ZeroRateCurve(rates, tenors, Spline.Linear())

interpolated = Dict{String, Float64}()
for (label, target) in (("3Y" => 3.0), ("7Y" => 7.0))
    interpolated[label] = rate_percent(zero(curve, target))
end

result = Dict(
    "curve_points" => points,
    "interpolated" => interpolated,
    "method" => "FinanceModels.ZeroRateCurve/Spline.Linear",
    "curve" => Dict(
        "pillars" => [Dict("tenor_years" => t, "rate_pct" => round(r * 100; digits=6)) for (t, r) in zip(tenors, rates)],
    ),
)
for key in ("bond", "option")
    if haskey(payload, key)
        result[key] = payload[key]
    end
end

write(output_path, JSON.json(result))

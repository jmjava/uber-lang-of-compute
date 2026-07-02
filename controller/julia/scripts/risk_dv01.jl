#!/usr/bin/env julia
using JSON

input_path, output_path = ARGS[1], ARGS[2]
payload = JSON.parse(read(input_path, String))

const NOTIONAL = 1_000_000.0
const BP_SHIFT = 0.0001

interpolated = get(payload, "interpolated", Dict{String, Any}())
risk_metrics = []
for (tenor, rate) in interpolated
    years = tenor == "7Y" ? 7.0 : 3.0
    dv01 = round(NOTIONAL * years * BP_SHIFT; digits=2)
    push!(risk_metrics, Dict("tenor" => tenor, "rate" => Float64(rate), "dv01" => dv01))
end
sort!(risk_metrics, by = r -> r["tenor"])

result = Dict(
    "risk_metrics" => risk_metrics,
    "notional" => NOTIONAL,
    "method" => "dv01_simplified",
)

write(output_path, JSON.json(result))

#!/usr/bin/env julia
using JSON

input_path, output_path = ARGS[1], ARGS[2]
payload = JSON.parse(read(input_path, String))

instruments = get(payload, "instruments", nothing)
if instruments === nothing
    error("interpolate: missing instruments")
end

function maturity_to_years(maturity::AbstractString)
    length(maturity) >= 4 || return 0.0
    year = parse(Int, maturity[1:4])
    return float(year - 2025)
end

function linear_interp(points, target_tenor)
    isempty(points) && return 0.0
    sorted = sort(points, by = p -> p["tenor_years"])
    if target_tenor <= sorted[1]["tenor_years"]
        return sorted[1]["rate"]
    end
    if target_tenor >= sorted[end]["tenor_years"]
        return sorted[end]["rate"]
    end
    for i in 1:length(sorted)-1
        a, b = sorted[i], sorted[i+1]
        if target_tenor >= a["tenor_years"] && target_tenor <= b["tenor_years"]
            denom = b["tenor_years"] - a["tenor_years"]
            denom == 0 && return a["rate"]
            t = (target_tenor - a["tenor_years"]) / denom
            return a["rate"] + t * (b["rate"] - a["rate"])
        end
    end
    return sorted[end]["rate"]
end

points = map(instruments) do inst
    Dict(
        "maturity" => inst["maturity"],
        "rate" => Float64(inst["rate"]),
        "tenor_years" => maturity_to_years(string(inst["maturity"])),
    )
end
sort!(points, by = p -> p["maturity"])

interpolated = Dict(
    "3Y" => linear_interp(points, 3.0),
    "7Y" => linear_interp(points, 7.0),
)

result = Dict(
    "curve_points" => points,
    "interpolated" => interpolated,
    "method" => "linear",
)

write(output_path, JSON.json(result))

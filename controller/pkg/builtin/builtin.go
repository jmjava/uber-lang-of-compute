package builtin

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
)

// Instrument represents a rate instrument in finance examples.
type Instrument struct {
	InstrumentID string  `json:"instrument_id"`
	Rate         float64 `json:"rate"`
	Maturity     string  `json:"maturity"`
}

type curvePoint struct {
	Maturity string  `json:"maturity"`
	Rate     float64 `json:"rate"`
	Tenor    float64 `json:"tenor_years"`
}

// Execute runs a builtin command against the given input JSON string.
func Execute(command, inputJSON string) (string, error) {
	switch command {
	case "builtin:identity":
		return identity(inputJSON)
	case "builtin:interpolate":
		return interpolate(inputJSON)
	case "builtin:risk-dv01":
		return riskDV01(inputJSON)
	default:
		return "", fmt.Errorf("unknown builtin command: %s", command)
	}
}

func identity(inputJSON string) (string, error) {
	if inputJSON == "" || inputJSON == "null" {
		return inputJSON, nil
	}
	var data interface{}
	if err := json.Unmarshal([]byte(inputJSON), &data); err != nil {
		return inputJSON, nil
	}
	out, err := json.Marshal(data)
	return string(out), err
}

func interpolate(inputJSON string) (string, error) {
	var instruments []Instrument
	if err := json.Unmarshal([]byte(inputJSON), &instruments); err != nil {
		var wrapped struct {
			Instruments []Instrument `json:"instruments"`
		}
		if err2 := json.Unmarshal([]byte(inputJSON), &wrapped); err2 != nil {
			return "", fmt.Errorf("interpolate: parse input: %w", err)
		}
		instruments = wrapped.Instruments
	}

	sort.Slice(instruments, func(i, j int) bool {
		return instruments[i].Maturity < instruments[j].Maturity
	})

	points := make([]curvePoint, len(instruments))
	for i, inst := range instruments {
		points[i] = curvePoint{
			Maturity: inst.Maturity,
			Rate:     inst.Rate,
			Tenor:    maturityToYears(inst.Maturity),
		}
	}

	interpolated := map[string]float64{
		"3Y": linearInterp(points, 3.0),
		"7Y": linearInterp(points, 7.0),
	}

	result := map[string]interface{}{
		"curve_points": points,
		"interpolated": interpolated,
		"method":       "linear",
	}
	out, err := json.Marshal(result)
	return string(out), err
}

func riskDV01(inputJSON string) (string, error) {
	var curveData struct {
		CurvePoints []curvePoint         `json:"curve_points"`
		Interpolated map[string]float64 `json:"interpolated"`
	}
	if err := json.Unmarshal([]byte(inputJSON), &curveData); err != nil {
		return "", fmt.Errorf("risk-dv01: parse input: %w", err)
	}

	const notional = 1_000_000.0
	const bpShift = 0.0001

	type riskEntry struct {
		Tenor string  `json:"tenor"`
		Rate  float64 `json:"rate"`
		DV01  float64 `json:"dv01"`
	}

	var risks []riskEntry
	for tenor, rate := range curveData.Interpolated {
		years := 3.0
		if tenor == "7Y" {
			years = 7.0
		}
		dv01 := notional * years * bpShift
		risks = append(risks, riskEntry{
			Tenor: tenor,
			Rate:  rate,
			DV01:  math.Round(dv01*100) / 100,
		})
	}
	sort.Slice(risks, func(i, j int) bool {
		return risks[i].Tenor < risks[j].Tenor
	})

	result := map[string]interface{}{
		"risk_metrics": risks,
		"notional":     notional,
		"method":       "dv01_simplified",
	}
	out, err := json.Marshal(result)
	return string(out), err
}

func linearInterp(points []curvePoint, targetTenor float64) float64 {
	if len(points) == 0 {
		return 0
	}
	sorted := make([]curvePoint, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Tenor < sorted[j].Tenor
	})

	if targetTenor <= sorted[0].Tenor {
		return sorted[0].Rate
	}
	if targetTenor >= sorted[len(sorted)-1].Tenor {
		return sorted[len(sorted)-1].Rate
	}

	for i := 0; i < len(sorted)-1; i++ {
		if targetTenor >= sorted[i].Tenor && targetTenor <= sorted[i+1].Tenor {
			denom := sorted[i+1].Tenor - sorted[i].Tenor
			if denom == 0 {
				return sorted[i].Rate
			}
			t := (targetTenor - sorted[i].Tenor) / denom
			return sorted[i].Rate + t*(sorted[i+1].Rate-sorted[i].Rate)
		}
	}
	return sorted[len(sorted)-1].Rate
}

func maturityToYears(maturity string) float64 {
	if len(maturity) >= 4 {
		var year int
		fmt.Sscanf(maturity[:4], "%d", &year)
		return float64(year - 2025)
	}
	return 0
}

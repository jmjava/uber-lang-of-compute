package builtin_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/builtin"
)

func TestTreasuryParCurveLandsOnGrid(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "examples", "rates-desk-day", "curve-2025-04-15.json"))
	if err != nil {
		t.Fatal(err)
	}
	out, err := builtin.Execute("builtin:interpolate", string(raw))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Method       string             `json:"method"`
		Interpolated map[string]float64 `json:"interpolated"`
		CurvePoints  []struct {
			Tenor float64 `json:"tenor_years"`
			Rate  float64 `json:"rate"`
		} `json:"curve_points"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Method != "linear" {
		t.Fatalf("method %q", parsed.Method)
	}
	if len(parsed.CurvePoints) != 8 {
		t.Fatalf("curve points %d want 8", len(parsed.CurvePoints))
	}
	onGrid := map[string]float64{
		"1Y": 4.70, "2Y": 4.80, "3Y": 4.62, "5Y": 4.45,
		"7Y": 4.35, "10Y": 4.25, "20Y": 4.40, "30Y": 4.48,
	}
	for tenor, want := range onGrid {
		got, ok := parsed.Interpolated[tenor]
		if !ok {
			t.Fatalf("missing %s", tenor)
		}
		if math.Abs(got-want) > 1e-9 {
			t.Fatalf("%s: got %v want %v (on-the-run must be exact)", tenor, got, want)
		}
	}

	riskJSON, err := builtin.Execute("builtin:risk-dv01", out)
	if err != nil {
		t.Fatal(err)
	}
	var risk struct {
		Notional    float64 `json:"notional"`
		RiskMetrics []struct {
			Tenor string  `json:"tenor"`
			DV01  float64 `json:"dv01"`
		} `json:"risk_metrics"`
	}
	if err := json.Unmarshal([]byte(riskJSON), &risk); err != nil {
		t.Fatal(err)
	}
	if risk.Notional != 1_000_000 || len(risk.RiskMetrics) != 8 {
		t.Fatalf("risk envelope notional=%v metrics=%d", risk.Notional, len(risk.RiskMetrics))
	}
	byTenor := map[string]float64{}
	for _, m := range risk.RiskMetrics {
		byTenor[m.Tenor] = m.DV01
	}
	if byTenor["2Y"] != 200 || byTenor["10Y"] != 1000 || byTenor["30Y"] != 3000 {
		t.Fatalf("KR01 ladder %+v", byTenor)
	}
}

func TestCoarsenSumsChildValues(t *testing.T) {
	out, err := builtin.Execute("builtin:coarsen", `[{"v":1},{"v":1}]`)
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		V      float64 `json:"v"`
		Value  float64 `json:"value"`
		Parts  int     `json:"parts"`
		Method string  `json:"method"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.V != 2 || parsed.Value != 2 || parsed.Parts != 2 || parsed.Method != "sum" {
		t.Fatalf("coarsen %+v", parsed)
	}

	nested, err := builtin.Execute("builtin:coarsen", `[{"value":2,"v":2},{"value":2,"v":2}]`)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(nested), &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.V != 4 || parsed.Parts != 2 {
		t.Fatalf("nested coarsen %+v", parsed)
	}
}

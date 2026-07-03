package engine_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor/julia"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func ensureJuliaDeps(t *testing.T, cfg julia.Config) {
	t.Helper()
	if !julia.Available(cfg) {
		t.Skip("julia not installed")
	}
	cmd := exec.Command(cfg.Bin, "--project="+cfg.Project, "-e", "using Pkg; Pkg.instantiate()")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("julia deps not ready: %v: %s", err, out)
	}
}

func TestJuliaFinanceModelsWorkflowRuns(t *testing.T) {
	cfg := julia.DefaultConfig()
	ensureJuliaDeps(t, cfg)

	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "julia.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "julia-domino-chain")
	wf.Spec.Provisioning.StorePath = filepath.Join(dir, "julia.db")

	eng := engine.New(s)
	result, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("julia run: %v", err)
	}

	var greeks struct {
		Method      string `json:"method"`
		BondGreeks  struct {
			DV01              float64 `json:"dv01"`
			ModifiedDuration  float64 `json:"modified_duration"`
			Convexity         float64 `json:"convexity"`
		} `json:"bond_greeks"`
		RateGreeks []struct {
			Tenor string  `json:"tenor"`
			DV01  float64 `json:"dv01"`
		} `json:"rate_greeks"`
		OptionGreeks struct {
			Delta float64 `json:"delta"`
			Gamma float64 `json:"gamma"`
			Vega  float64 `json:"vega"`
		} `json:"option_greeks"`
	}
	if err := json.Unmarshal([]byte(result.FinalOutput), &greeks); err != nil {
		t.Fatalf("parse final output: %v", err)
	}
	if !strings.Contains(greeks.Method, "FinanceModels") {
		t.Fatalf("expected FinanceModels greeks method, got %q", greeks.Method)
	}
	if greeks.BondGreeks.DV01 <= 0 || greeks.BondGreeks.ModifiedDuration <= 0 {
		t.Fatalf("unexpected bond greeks: %+v", greeks.BondGreeks)
	}
	if len(greeks.RateGreeks) != 2 {
		t.Fatalf("expected 2 rate greeks, got %d", len(greeks.RateGreeks))
	}
	if greeks.OptionGreeks.Delta <= 0 || greeks.OptionGreeks.Gamma <= 0 {
		t.Fatalf("unexpected option greeks: %+v", greeks.OptionGreeks)
	}
}

func TestJuliaFinanceModelsWorkflowDeterministic(t *testing.T) {
	cfg := julia.DefaultConfig()
	ensureJuliaDeps(t, cfg)

	runOnce := func() string {
		dir := t.TempDir()
		s, err := store.Open(filepath.Join(dir, "julia.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()

		wf := loadTestWorkflow(t, "julia-domino-chain")
		wf.Spec.Provisioning.StorePath = filepath.Join(dir, "julia.db")
		result, err := engine.New(s).Run(wf)
		if err != nil {
			t.Fatalf("julia run: %v", err)
		}
		return result.FinalOutput
	}

	out1 := runOnce()
	out2 := runOnce()
	if out1 != out2 {
		t.Fatalf("non-deterministic julia output:\nfirst:  %s\nsecond: %s", out1, out2)
	}
}

func TestJuliaWorkflowExampleFileExists(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "julia-domino-chain", "workflow.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

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

	var risk struct {
		Method      string `json:"method"`
		RiskMetrics []struct {
			Tenor string  `json:"tenor"`
			DV01  float64 `json:"dv01"`
		} `json:"risk_metrics"`
	}
	if err := json.Unmarshal([]byte(result.FinalOutput), &risk); err != nil {
		t.Fatalf("parse final output: %v", err)
	}
	if !strings.Contains(risk.Method, "FinanceModels") {
		t.Fatalf("expected FinanceModels risk method, got %q", risk.Method)
	}
	if len(risk.RiskMetrics) != 2 {
		t.Fatalf("expected 2 risk metrics, got %d", len(risk.RiskMetrics))
	}
	for _, m := range risk.RiskMetrics {
		if m.DV01 <= 0 {
			t.Fatalf("expected positive dv01 for %s, got %v", m.Tenor, m.DV01)
		}
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

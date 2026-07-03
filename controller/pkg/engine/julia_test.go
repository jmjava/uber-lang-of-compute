package engine_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor/julia"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestJuliaWorkflowMatchesBuiltinChain(t *testing.T) {
	cfg := julia.DefaultConfig()
	if !julia.Available(cfg) {
		t.Skip("julia not installed")
	}

	project := cfg.Project
	cmd := exec.Command(cfg.Bin, "--project="+project, "-e", "using Pkg; Pkg.instantiate()")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("julia deps not ready: %v: %s", err, out)
	}

	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "julia.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "finance-curve-snapshot")
	juliaWF := loadTestWorkflow(t, "julia-domino-chain")
	juliaWF.Spec.Provisioning.StorePath = filepath.Join(dir, "julia.db")

	builtinEng := engine.New(s)
	builtinResult, err := builtinEng.Run(wf)
	if err != nil {
		t.Fatalf("builtin run: %v", err)
	}

	juliaStore, err := store.Open(filepath.Join(dir, "julia2.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer juliaStore.Close()

	juliaEng := engine.New(juliaStore)
	juliaResult, err := juliaEng.Run(juliaWF)
	if err != nil {
		t.Fatalf("julia run: %v", err)
	}

	if juliaResult.FinalOutput != builtinResult.FinalOutput {
		t.Fatalf("final output mismatch:\nbuiltin: %s\njulia:   %s", builtinResult.FinalOutput, juliaResult.FinalOutput)
	}
}

func TestJuliaWorkflowExampleFileExists(t *testing.T) {
	path := filepath.Join("..", "..", "..", "examples", "julia-domino-chain", "workflow.yaml")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

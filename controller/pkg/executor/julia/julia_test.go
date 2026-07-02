package julia_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor/julia"
)

type stubRunner struct {
	lastBin  string
	lastArgs []string
}

func (s *stubRunner) Run(bin string, args []string) error {
	s.lastBin = bin
	s.lastArgs = append([]string(nil), args...)
	dir := filepath.Dir(args[len(args)-2])
	_ = dir
	outPath := args[len(args)-1]
	return os.WriteFile(outPath, []byte(`{"ok":true}`), 0o644)
}

func TestExecuteUsesProjectAndScript(t *testing.T) {
	project := t.TempDir()
	scripts := filepath.Join(project, "scripts")
	if err := os.MkdirAll(scripts, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scripts, "identity.jl"), []byte("# stub"), 0o644); err != nil {
		t.Fatal(err)
	}

	runner := &stubRunner{}
	cfg := julia.Config{Bin: "julia", Project: project, Runner: runner}
	out, err := julia.Execute(cfg, "julia:identity", `{"v":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"ok":true}` {
		t.Fatalf("unexpected output %q", out)
	}
	if runner.lastBin != "julia" {
		t.Fatalf("expected julia bin, got %q", runner.lastBin)
	}
	if len(runner.lastArgs) < 2 || runner.lastArgs[0] != "--project="+project {
		t.Fatalf("unexpected args %v", runner.lastArgs)
	}
}

func TestExecuteJuliaIdentityIntegration(t *testing.T) {
	cfg := julia.DefaultConfig()
	if !julia.Available(cfg) {
		t.Skip("julia not installed")
	}

	out, err := julia.Execute(cfg, "julia:identity", `{"value":42}`)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"value":42}` {
		t.Fatalf("unexpected output %q", out)
	}
}

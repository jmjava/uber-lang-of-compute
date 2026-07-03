package julia

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Config controls Julia subprocess execution.
type Config struct {
	Bin     string
	Project string
	Runner  Runner
}

// Runner executes Julia with a script and input/output file paths.
type Runner interface {
	Run(bin string, args []string) error
}

type execRunner struct{}

func (execRunner) Run(bin string, args []string) error {
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// DefaultConfig loads Julia settings from the environment.
func DefaultConfig() Config {
	project := os.Getenv("KBL_JULIA_PROJECT")
	if project == "" {
		project = DefaultProjectRoot()
	}
	bin := os.Getenv("KBL_JULIA_BIN")
	if bin == "" {
		bin = "julia"
	}
	return Config{Bin: bin, Project: project, Runner: execRunner{}}
}

// DefaultProjectRoot returns the bundled Julia project directory.
func DefaultProjectRoot() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "julia"
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "julia"))
}

// Execute runs julia:<script> commands using file-based JSON handoff.
func Execute(cfg Config, command, inputJSON string) (string, error) {
	script, err := resolveScript(cfg, command)
	if err != nil {
		return "", err
	}

	dir, err := os.MkdirTemp("", "kbl-julia-*")
	if err != nil {
		return "", fmt.Errorf("julia temp dir: %w", err)
	}
	defer os.RemoveAll(dir)

	inputPath := filepath.Join(dir, "input.json")
	outputPath := filepath.Join(dir, "output.json")
	if err := os.WriteFile(inputPath, []byte(inputJSON), 0o644); err != nil {
		return "", fmt.Errorf("julia write input: %w", err)
	}

	runner := cfg.Runner
	if runner == nil {
		runner = execRunner{}
	}

	args := []string{"--project=" + cfg.Project, script, inputPath, outputPath}
	if err := runner.Run(cfg.Bin, args); err != nil {
		return "", fmt.Errorf("julia execute %q: %w", command, err)
	}

	out, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("julia read output: %w", err)
	}
	return string(out), nil
}

func resolveScript(cfg Config, command string) (string, error) {
	name := strings.TrimPrefix(command, "julia:")
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("julia command requires script name")
	}

	if strings.HasSuffix(name, ".jl") {
		if filepath.IsAbs(name) {
			return name, nil
		}
		return filepath.Join(cfg.Project, "scripts", name), nil
	}

	script := filepath.Join(cfg.Project, "scripts", name+".jl")
	if _, err := os.Stat(script); err != nil {
		return "", fmt.Errorf("julia script %q: %w", script, err)
	}
	return script, nil
}

// Available reports whether the configured Julia binary is on PATH.
func Available(cfg Config) bool {
	bin := cfg.Bin
	if bin == "" {
		bin = "julia"
	}
	_, err := exec.LookPath(bin)
	return err == nil
}

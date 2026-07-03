package executor

import (
	"fmt"
	"strings"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/builtin"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor/julia"
)

// EngineType names supported pluggable execution engines.
const (
	EngineBuiltin = "builtin"
	EngineJulia   = "julia"
	EnginePython  = "python"
	EngineGo      = "go"
)

// Config selects runtime binaries and script roots for pluggable engines.
type Config struct {
	Julia julia.Config
}

// DefaultConfig reads executor settings from the environment.
func DefaultConfig() Config {
	return Config{Julia: julia.DefaultConfig()}
}

// Execute runs a domino command against JSON input.
// Commands use engine prefixes: builtin:*, julia:*, python:* (future).
func Execute(cfg Config, command, inputJSON string) (string, error) {
	switch {
	case strings.HasPrefix(command, "builtin:"):
		return builtin.Execute(command, inputJSON)
	case strings.HasPrefix(command, "julia:"):
		return julia.Execute(cfg.Julia, command, inputJSON)
	case strings.HasPrefix(command, "python:"):
		return "", fmt.Errorf("python execution is not implemented yet (command %q)", command)
	default:
		return "", fmt.Errorf("unsupported command %q (expected builtin:, julia:, or python: prefix)", command)
	}
}

// ExecuteDefault runs a command using environment-based configuration.
func ExecuteDefault(command, inputJSON string) (string, error) {
	return Execute(DefaultConfig(), command, inputJSON)
}

// PrefixForEngine maps PluggableUniverse executionEngine.type to the domino command prefix.
func PrefixForEngine(engineType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(engineType)) {
	case "", EngineBuiltin, EngineGo:
		return "builtin:", nil
	case EngineJulia:
		return "julia:", nil
	case EnginePython:
		return "python:", nil
	default:
		return "", fmt.Errorf("unsupported execution engine type %q", engineType)
	}
}

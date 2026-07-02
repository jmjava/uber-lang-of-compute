package executor_test

import (
	"strings"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor"
)

func TestExecuteBuiltinIdentity(t *testing.T) {
	out, err := executor.ExecuteDefault("builtin:identity", `{"v":1}`)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"v":1}` {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestExecuteUnsupportedCommand(t *testing.T) {
	_, err := executor.ExecuteDefault("lua:print", `{}`)
	if err == nil || !strings.Contains(err.Error(), "unsupported command") {
		t.Fatalf("expected unsupported command error, got %v", err)
	}
}

func TestExecutePythonNotImplemented(t *testing.T) {
	_, err := executor.ExecuteDefault("python:pandas", `{}`)
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected python not implemented error, got %v", err)
	}
}

func TestPrefixForEngine(t *testing.T) {
	tests := map[string]string{
		"builtin": "builtin:",
		"go":      "builtin:",
		"julia":   "julia:",
		"python":  "python:",
	}
	for engine, want := range tests {
		got, err := executor.PrefixForEngine(engine)
		if err != nil || got != want {
			t.Fatalf("PrefixForEngine(%q) = %q, %v; want %q", engine, got, err, want)
		}
	}
}

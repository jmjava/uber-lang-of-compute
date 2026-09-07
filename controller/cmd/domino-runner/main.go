package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor"
)

// domino-runner is the container entrypoint for hot-swapped domino steps.
// Environment:
//   KBL_COMMAND  - builtin:identity, julia:interpolate, etc.
//   KBL_INPUT    - path to input JSON file
//   KBL_OUTPUT   - path to write output JSON
//   KBL_JULIA_BIN, KBL_JULIA_PROJECT - optional Julia runtime settings
func main() {
	cmd := os.Getenv("KBL_COMMAND")
	inputPath := os.Getenv("KBL_INPUT")
	outputPath := os.Getenv("KBL_OUTPUT")

	if cmd == "" {
		fmt.Fprintln(os.Stderr, "KBL_COMMAND is required")
		os.Exit(1)
	}
	if inputPath == "" || outputPath == "" {
		fmt.Fprintln(os.Stderr, "KBL_INPUT and KBL_OUTPUT are required")
		os.Exit(1)
	}

	input, err := waitAndReadInput(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	out, err := executor.ExecuteDefault(cmd, string(input))
	if err != nil {
		fmt.Fprintf(os.Stderr, "execute: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, []byte(out), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}

func waitAndReadInput(path string) ([]byte, error) {
	deadline := time.Now().Add(15 * time.Minute)
	for {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for input %s", path)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

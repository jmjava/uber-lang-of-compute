package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

func main() {
	workflowPath := flag.String("workflow", "", "Path to workflow YAML file")
	storePath := flag.String("store", "/tmp/kbl-store/cache.db", "Node-local SQLite store path")
	replayLogPath := flag.String("replay-log", "", "Write replay log JSON to this path (default: stdout)")
	flag.Parse()

	if *workflowPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: kbl-compute --workflow <path> [--store <path>] [--replay-log <path>]")
		os.Exit(1)
	}

	wf, err := loadWorkflow(*workflowPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load workflow: %v\n", err)
		os.Exit(1)
	}

	if wf.Spec.Provisioning.StorePath != "" {
		*storePath = wf.Spec.Provisioning.StorePath
	}

	s, err := store.Open(*storePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	eng := engine.New(s)
	result, err := eng.Run(wf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run workflow: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal replay log: %v\n", err)
		os.Exit(1)
	}

	if *replayLogPath != "" {
		dir := filepath.Dir(*replayLogPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "create replay log dir: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*replayLogPath, out, 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "write replay log: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Replay log written to %s\n", *replayLogPath)
	} else {
		fmt.Println(string(out))
	}

	// Summary
	reused := 0
	for _, e := range result.Entries {
		if e.Reused {
			reused++
		}
	}
	fmt.Fprintf(os.Stderr, "\nSnapshot: %s | Dominos: %d | Reused: %d | Recomputed: %d\n",
		result.SnapshotID, len(result.Entries), reused, len(result.Entries)-reused)
}

func loadWorkflow(path string) (*types.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var wf types.Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, err
	}
	return &wf, nil
}

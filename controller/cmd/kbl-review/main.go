package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/review"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

func main() {
	workflowPath := flag.String("workflow", "examples/finance-curve-snapshot/workflow.yaml", "Path to a sealed builtin workflow YAML")
	storePath := flag.String("store", "", "SQLite store path (default: a temp file)")
	flag.Parse()

	wf, err := loadWorkflow(*workflowPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load workflow: %v\n", err)
		os.Exit(1)
	}

	path := *storePath
	if path == "" {
		dir, err := os.MkdirTemp("", "kbl-review-*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "temp store: %v\n", err)
			os.Exit(1)
		}
		path = filepath.Join(dir, "review.db")
	}

	backend, err := store.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open store: %v\n", err)
		os.Exit(1)
	}
	defer backend.Close()

	report, err := review.Walkthrough(backend, wf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "review walkthrough: %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal report: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
	fmt.Fprintf(os.Stderr, "\nsealed snapshot %s\nbuiltin chain evaluated %d then reused %d\nlookahead %s\nfan-out HeadLink %s → %v\n",
		report.SnapshotID, report.FirstEvaluations, report.SecondReuses, report.LookaheadName, report.HeadLink, report.FanoutUniverses)
}

func loadWorkflow(path string) (*types.Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var wf types.Workflow
		err := dec.Decode(&wf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if wf.Kind == "Workflow" || wf.Metadata.Name != "" {
			return &wf, nil
		}
	}
	return nil, fmt.Errorf("no Workflow document in %s", path)
}

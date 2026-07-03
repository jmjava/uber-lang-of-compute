package engine_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

func loadTestWorkflow(t *testing.T, name string) *types.Workflow {
	t.Helper()
	path := filepath.Join("..", "..", "..", "examples", name, "workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var wf types.Workflow
		err := dec.Decode(&wf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse workflow: %v", err)
		}
		if wf.Kind == "Workflow" {
			return &wf
		}
	}
	t.Fatalf("no Workflow document in %s", path)
	return nil
}

func TestSnapshotReplayDeterministic(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "replay.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	eng := engine.New(s)

	result1, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	result2, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if result1.SnapshotID != result2.SnapshotID {
		t.Errorf("snapshot ID mismatch: %s vs %s", result1.SnapshotID, result2.SnapshotID)
	}

	if len(result1.Entries) != len(result2.Entries) {
		t.Fatalf("entry count mismatch: %d vs %d", len(result1.Entries), len(result2.Entries))
	}

	for i, e1 := range result1.Entries {
		e2 := result2.Entries[i]
		if e1.InputHash != e2.InputHash {
			t.Errorf("entry %d input hash mismatch", i)
		}
		if e1.OutputHash != e2.OutputHash {
			t.Errorf("entry %d output hash mismatch", i)
		}
	}

	if result1.FinalOutput != result2.FinalOutput {
		t.Errorf("final output mismatch:\n  run1: %s\n  run2: %s", result1.FinalOutput, result2.FinalOutput)
	}
}

func TestMemoizationReusesResults(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "memo.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "finance-curve-snapshot")
	eng := engine.New(s)

	result1, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	for _, e := range result1.Entries {
		if e.Reused {
			t.Errorf("first run domino %q should not be reused", e.DominoID)
		}
	}

	result2, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	for _, e := range result2.Entries {
		if !e.Reused {
			t.Errorf("second run domino %q should be reused (memo hit)", e.DominoID)
		}
	}

	if result1.FinalOutput != result2.FinalOutput {
		t.Errorf("final output changed between runs")
	}
}

func TestUnsealedSnapshotRejected(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "sealed.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	wf.Spec.Snapshot.Spec.Sealed = false

	eng := engine.New(s)
	_, err = eng.Run(wf)
	if err == nil {
		t.Fatal("expected error for unsealed snapshot")
	}
}

package replica_test

import (
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/replica"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestMaterializeCopiesSnapshotAndDominos(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.db")
	targetPath := filepath.Join(dir, "target.db")

	source, err := store.OpenSQLite(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	target, err := store.OpenSQLite(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	const snapshotID = "snap-abc123"
	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"key":"value"}`, true); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "in1", "out1", `{"loaded":true}`, false); err != nil {
		t.Fatal(err)
	}

	result, err := replica.Materialize(replica.MaterializeConfig{
		SnapshotID:  snapshotID,
		DominoChain: []string{"load"},
		Source:      source,
		Target:      target,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.SnapshotCopied {
		t.Fatal("expected snapshot copied")
	}
	if result.DominoCount != 1 {
		t.Fatalf("expected 1 domino, got %d", result.DominoCount)
	}

	_, data, sealed, err := target.GetSnapshot(snapshotID)
	if err != nil {
		t.Fatalf("target snapshot: %v", err)
	}
	if data == "" || !sealed {
		t.Fatalf("unexpected target snapshot data=%q sealed=%v", data, sealed)
	}

	out, err := target.GetDominoOutput(snapshotID, "load")
	if err != nil {
		t.Fatalf("target domino output: %v", err)
	}
	if out == "" {
		t.Fatal("expected domino output on target")
	}
}

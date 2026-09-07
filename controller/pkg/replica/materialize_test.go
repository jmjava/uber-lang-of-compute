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
	if err := source.SaveResult(snapshotID, "load", "in1", "out1", `{"loaded":true}`, false, "", ""); err != nil {
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

func TestMaterializeRejectsUnsealedSnapshot(t *testing.T) {
	dir := t.TempDir()
	source, err := store.OpenSQLite(filepath.Join(dir, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	target, err := store.OpenSQLite(filepath.Join(dir, "target.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	const snapshotID = "snap-live"
	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"key":"value"}`, false); err != nil {
		t.Fatal(err)
	}

	_, err = replica.Materialize(replica.MaterializeConfig{
		SnapshotID: snapshotID,
		Source:     source,
		Target:     target,
	})
	if err == nil {
		t.Fatal("expected error materializing unsealed snapshot")
	}
}

func TestMaterializeFailsOnMissingDomino(t *testing.T) {
	dir := t.TempDir()
	source, err := store.OpenSQLite(filepath.Join(dir, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	target, err := store.OpenSQLite(filepath.Join(dir, "target.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	const snapshotID = "snap-abc123"
	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"key":"value"}`, true); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "in1", "out1", `{"loaded":true}`, false, "", ""); err != nil {
		t.Fatal(err)
	}
	_, err = replica.Materialize(replica.MaterializeConfig{
		SnapshotID:  snapshotID,
		DominoChain: []string{"load", "missing"},
		Source:      source,
		Target:      target,
	})
	if err == nil {
		t.Fatal("incomplete chain must fail closed")
	}
}

func TestMaterializeCopiesLastSpineRowNotLexicalMemo(t *testing.T) {
	dir := t.TempDir()
	source, err := store.OpenSQLite(filepath.Join(dir, "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	target, err := store.OpenSQLite(filepath.Join(dir, "target.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	const snapshotID = "snap-spine"
	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "zzzz", "old", `{"n":1}`, false, snapshotID, "link-1"); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "aaaa", "new", `{"n":2}`, false, "link-1", "link-2"); err != nil {
		t.Fatal(err)
	}

	if _, err := replica.Materialize(replica.MaterializeConfig{
		SnapshotID:  snapshotID,
		DominoChain: []string{"load"},
		Source:      source,
		Target:      target,
	}); err != nil {
		t.Fatal(err)
	}
	in, out, payload, err := target.GetLatestResult(snapshotID, "load")
	if err != nil {
		t.Fatal(err)
	}
	if in != "aaaa" || out != "new" || payload != `{"n":2}` {
		t.Fatalf("replica copied the wrong memo row: in=%s out=%s payload=%s", in, out, payload)
	}
	rows, err := target.ListReplay(snapshotID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].PrevLink != "link-1" || rows[0].Link != "link-2" {
		t.Fatalf("replica must copy spine links, got %+v", rows)
	}
}

package cdc_test

import (
	"context"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestApplyReplicatesSnapshotAndDomino(t *testing.T) {
	source, err := store.OpenSQLite(t.TempDir() + "/source.db")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()

	target, err := store.OpenSQLite(t.TempDir() + "/target.db")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()

	const snapshotID = "snap-cdc-1"
	if err := source.SaveSnapshot(snapshotID, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := source.SaveResult(snapshotID, "load", "in", "out", `{"ok":true}`, false); err != nil {
		t.Fatal(err)
	}

	envs, err := cdc.ExportFromStore(source, snapshotID, []string{"load"})
	if err != nil {
		t.Fatal(err)
	}
	if len(envs) != 2 {
		t.Fatalf("expected 2 envelopes, got %d", len(envs))
	}

	progress, err := cdc.ApplyAll(target, snapshotID, envs)
	if err != nil {
		t.Fatal(err)
	}
	if !progress.IsComplete(1) {
		t.Fatalf("expected complete progress, got %+v", progress)
	}

	_, data, sealed, err := target.GetSnapshot(snapshotID)
	if err != nil || data == "" || !sealed {
		t.Fatalf("target snapshot missing: err=%v data=%q sealed=%v", err, data, sealed)
	}
}

func TestMemoryBusPublishConsume(t *testing.T) {
	bus := cdc.NewMemoryBus()
	ctx := context.Background()

	env := cdc.Envelope{
		Op:    cdc.OpCreate,
		Table: cdc.TableSnapshots,
		After: cdc.SnapshotRow{SnapshotID: "abc", TimeSlice: "2025-04-15", Data: "{}", Sealed: true},
	}
	if err := bus.Publish(ctx, "abc", env); err != nil {
		t.Fatal(err)
	}

	got, err := bus.Consume(ctx, "abc")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
}

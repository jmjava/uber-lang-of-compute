package snapshot_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestLoadContentPreferStore(t *testing.T) {
	dir := t.TempDir()
	backend, err := store.Open(dir + "/snap.db")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	payload := `{"instruments":[{"instrument_id":"US10Y"}]}`
	if err := backend.SaveSnapshot("snap-1", "2025-04-15T00:00:00Z", payload, true); err != nil {
		t.Fatal(err)
	}

	content, data, ok, err := snapshot.LoadContentPreferStore(backend, "snap-1")
	if err != nil || !ok {
		t.Fatalf("expected store hit, ok=%v err=%v", ok, err)
	}
	if data != payload {
		t.Fatalf("unexpected data: %s", data)
	}
	m, ok := content.(map[string]interface{})
	if !ok || m["instruments"] == nil {
		t.Fatalf("expected parsed content, got %+v", content)
	}
}

func TestLoadDataPreferStoreMissFallsThrough(t *testing.T) {
	dir := t.TempDir()
	backend, err := store.Open(dir + "/snap.db")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	data, ok, err := snapshot.LoadDataPreferStore(backend, "missing-snap")
	if err != nil || ok || data != "" {
		t.Fatalf("expected store miss, ok=%v data=%q err=%v", ok, data, err)
	}
}

func TestResolveEngineContentPreferStoreMissUsesInline(t *testing.T) {
	dir := t.TempDir()
	backend, err := store.Open(dir + "/snap.db")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	snap := types.Snapshot{
		Spec: types.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Source: types.SnapshotSource{
				Inline: map[string]interface{}{"value": 7},
			},
			Sealed: true,
		},
		Status: &types.SnapshotStatus{SnapshotID: "computed-but-not-persisted"},
	}

	content, _, id, err := snapshot.ResolveEngineContentPreferStore(backend, snap, "computed-but-not-persisted")
	if err != nil {
		t.Fatal(err)
	}
	if id != "computed-but-not-persisted" {
		t.Fatalf("expected computed-but-not-persisted, got %s", id)
	}
	m, ok := content.(map[string]interface{})
	if !ok || m["value"] != 7 {
		t.Fatalf("expected inline content value 7, got %+v", content)
	}
}

func TestResolveEngineContentPreferStoreSkipsHTTP(t *testing.T) {
	dir := t.TempDir()
	backend, err := store.Open(dir + "/snap.db")
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()

	payload := `{"value":42}`
	if err := backend.SaveSnapshot("sealed-id", "2025-04-15T00:00:00Z", payload, true); err != nil {
		t.Fatal(err)
	}

	snap := types.Snapshot{
		Spec: types.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Source: types.SnapshotSource{
				URI: "http://127.0.0.1:1/unreachable",
			},
			Sealed: true,
		},
		Status: &types.SnapshotStatus{SnapshotID: "sealed-id"},
	}

	content, _, id, err := snapshot.ResolveEngineContentPreferStore(backend, snap, "sealed-id")
	if err != nil {
		t.Fatal(err)
	}
	if id != "sealed-id" {
		t.Fatalf("expected sealed-id, got %s", id)
	}
	m, ok := content.(map[string]interface{})
	if !ok || m["value"].(float64) != 42 {
		t.Fatalf("expected stored content, got %+v", content)
	}
}

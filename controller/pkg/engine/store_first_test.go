package engine_test

import (
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestRunUsesStoreSnapshotWithoutSourceFetch(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "hotpath.db")
	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	const snapshotID = "presealed-snap-id"
	const payload = `{"instruments":[{"instrument_id":"US10Y","rate":4.25}]}`
	if err := s.SaveSnapshot(snapshotID, "2025-04-15T00:00:00Z", payload, true); err != nil {
		t.Fatal(err)
	}

	wf := &types.Workflow{
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				Spec: types.SnapshotSpec{
					TimeSlice: "2025-04-15T00:00:00Z",
					Source: types.SnapshotSource{
						URI: "http://127.0.0.1:1/unreachable",
					},
					Sealed: true,
				},
				Status: &types.SnapshotStatus{SnapshotID: snapshotID},
			},
			Dominos: []types.Domino{{
				Metadata: types.ObjectMeta{Name: "load"},
				Spec: types.DominoSpec{Command: "builtin:identity"},
			}},
			Execution: types.ExecutionConfig{Chain: []string{"load"}},
		},
	}

	eng := engine.New(s)
	result, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if result.SnapshotID != snapshotID {
		t.Fatalf("expected snapshot ID %s, got %s", snapshotID, result.SnapshotID)
	}
}

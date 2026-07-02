package cdc

import (
	"encoding/json"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ExportFromStore builds CDC envelopes from a node-local store.
func ExportFromStore(source store.Backend, snapshotID string, dominoChain []string) ([]Envelope, error) {
	timeSlice, data, sealed, err := source.GetSnapshot(snapshotID)
	if err != nil {
		return nil, err
	}

	out := []Envelope{{
		Op:    OpCreate,
		Table: TableSnapshots,
		After: SnapshotRow{
			SnapshotID: snapshotID,
			TimeSlice:  timeSlice,
			Data:       data,
			Sealed:     sealed,
		},
	}}

	for _, dominoID := range dominoChain {
		inHash, outHash, output, err := source.GetLatestResult(snapshotID, dominoID)
		if err != nil {
			continue
		}
		out = append(out, Envelope{
			Op:    OpCreate,
			Table: TableDominoResults,
			After: DominoResultRow{
				SnapshotID: snapshotID,
				DominoID:   dominoID,
				InputHash:  inHash,
				OutputHash: outHash,
				Output:     output,
				Reused:     false,
			},
		})
	}
	return out, nil
}

// ExportFromWorkflow builds CDC envelopes from a completed workflow run.
func ExportFromWorkflow(wf *kblv1alpha1.Workflow, result *types.RunResult) []Envelope {
	if result == nil || result.SnapshotID == "" {
		return nil
	}

	data, _ := snapshotDataJSON(wf)
	out := []Envelope{{
		Op:    OpCreate,
		Table: TableSnapshots,
		After: SnapshotRow{
			SnapshotID: result.SnapshotID,
			TimeSlice:  wf.Spec.Snapshot.TimeSlice,
			Data:       data,
			Sealed:     wf.Spec.Snapshot.Sealed,
		},
	}}

	for _, entry := range result.Entries {
		out = append(out, Envelope{
			Op:    OpCreate,
			Table: TableDominoResults,
			After: DominoResultRow{
				SnapshotID: result.SnapshotID,
				DominoID:   entry.DominoID,
				InputHash:  entry.InputHash,
				OutputHash: entry.OutputHash,
				Output:     entry.Output,
				Reused:     entry.Reused,
			},
		})
	}
	return out
}

func snapshotDataJSON(wf *kblv1alpha1.Workflow) (string, error) {
	if wf.Spec.Snapshot.Source.Inline != nil {
		b, err := json.Marshal(wf.Spec.Snapshot.Source.Inline)
		return string(b), err
	}
	if wf.Spec.Snapshot.Source.Path != "" {
		b, err := json.Marshal(map[string]string{"path": wf.Spec.Snapshot.Source.Path})
		return string(b), err
	}
	if wf.Spec.Snapshot.Source.URI != "" {
		b, err := json.Marshal(map[string]string{"uri": wf.Spec.Snapshot.Source.URI})
		return string(b), err
	}
	return "{}", nil
}

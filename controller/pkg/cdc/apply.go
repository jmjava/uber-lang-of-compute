package cdc

import (
	"encoding/json"
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// Apply writes a Debezium envelope to the target store.
func Apply(target store.Backend, env Envelope) error {
	switch env.Table {
	case TableSnapshots:
		return applySnapshot(target, env)
	case TableDominoResults:
		return applyDominoResult(target, env)
	default:
		return fmt.Errorf("unsupported cdc table %q", env.Table)
	}
}

func applySnapshot(target store.Backend, env Envelope) error {
	if env.Op == OpDelete {
		return nil
	}
	row, err := decodeAfter[SnapshotRow](env.After)
	if err != nil {
		return err
	}
	return target.SaveSnapshot(row.SnapshotID, row.TimeSlice, row.Data, row.Sealed)
}

func applyDominoResult(target store.Backend, env Envelope) error {
	if env.Op == OpDelete {
		return nil
	}
	row, err := decodeAfter[DominoResultRow](env.After)
	if err != nil {
		return err
	}
	return target.SaveResult(row.SnapshotID, row.DominoID, row.InputHash, row.OutputHash, row.Output, row.Reused)
}

func decodeAfter[T any](after interface{}) (T, error) {
	var zero T
	if after == nil {
		return zero, fmt.Errorf("missing after payload")
	}
	b, err := json.Marshal(after)
	if err != nil {
		return zero, err
	}
	var row T
	if err := json.Unmarshal(b, &row); err != nil {
		return zero, err
	}
	return row, nil
}

// ApplyAll applies envelopes and returns sync progress.
func ApplyAll(target store.Backend, snapshotID string, envs []Envelope) (SyncProgress, error) {
	progress := SyncProgress{}
	seenDominos := make(map[string]bool)

	for _, env := range envs {
		if !matchesSnapshot(snapshotID, env) {
			continue
		}
		if err := Apply(target, env); err != nil {
			return progress, err
		}
		switch env.Table {
		case TableSnapshots:
			progress.SnapshotApplied = true
		case TableDominoResults:
			row, err := decodeAfter[DominoResultRow](env.After)
			if err != nil {
				return progress, err
			}
			if !seenDominos[row.DominoID] {
				seenDominos[row.DominoID] = true
				progress.DominoCount++
			}
		}
	}
	return progress, nil
}

func matchesSnapshot(snapshotID string, env Envelope) bool {
	switch env.Table {
	case TableSnapshots:
		row, err := decodeAfter[SnapshotRow](env.After)
		return err == nil && row.SnapshotID == snapshotID
	case TableDominoResults:
		row, err := decodeAfter[DominoResultRow](env.After)
		return err == nil && row.SnapshotID == snapshotID
	default:
		return false
	}
}

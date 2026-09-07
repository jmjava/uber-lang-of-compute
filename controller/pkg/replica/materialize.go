package replica

import (
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// MaterializeConfig describes a cross-store read-replica copy.
type MaterializeConfig struct {
	SnapshotID  string
	DominoChain []string
	Source      store.Backend
	Target      store.Backend
}

// MaterializeResult summarizes what was copied to the target store.
type MaterializeResult struct {
	SnapshotCopied bool
	DominoCount    int
}

// Materialize copies a sealed snapshot and domino chain results to a target store.
func Materialize(cfg MaterializeConfig) (*MaterializeResult, error) {
	if cfg.Source == nil || cfg.Target == nil {
		return nil, fmt.Errorf("source and target stores are required")
	}
	if cfg.SnapshotID == "" {
		return nil, fmt.Errorf("snapshot ID is required")
	}

	timeSlice, data, sealed, err := cfg.Source.GetSnapshot(cfg.SnapshotID)
	if err != nil {
		return nil, fmt.Errorf("read source snapshot: %w", err)
	}
	if !sealed {
		return nil, fmt.Errorf("refusing to materialize unsealed snapshot %s (cross-universe signaling requires a sealed Cauchy view)", cfg.SnapshotID)
	}
	if err := cfg.Target.SaveSnapshot(cfg.SnapshotID, timeSlice, data, true); err != nil {
		return nil, fmt.Errorf("write target snapshot: %w", err)
	}

	result := &MaterializeResult{SnapshotCopied: true}
	rows, err := cfg.Source.ListReplay(cfg.SnapshotID)
	if err != nil {
		return nil, fmt.Errorf("read source replay: %w", err)
	}
	for _, dominoID := range cfg.DominoChain {
		rec, ok := store.LastSpineResult(rows, dominoID)
		if !ok {
			inHash, outHash, output, latestErr := cfg.Source.GetLatestResult(cfg.SnapshotID, dominoID)
			if latestErr != nil {
				return nil, fmt.Errorf("missing domino %q on sealed snapshot %s: %w", dominoID, cfg.SnapshotID, latestErr)
			}
			rec = store.ReplayEntry{
				SnapshotID: cfg.SnapshotID,
				DominoID:   dominoID,
				InputHash:  inHash,
				OutputHash: outHash,
				Output:     output,
			}
		}
		if err := cfg.Target.SaveResult(cfg.SnapshotID, rec.DominoID, rec.InputHash, rec.OutputHash, rec.Output, false, rec.PrevLink, rec.Link); err != nil {
			return nil, fmt.Errorf("copy domino %q: %w", dominoID, err)
		}
		result.DominoCount++
	}

	return result, nil
}

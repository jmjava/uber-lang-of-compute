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
	if err := cfg.Target.SaveSnapshot(cfg.SnapshotID, timeSlice, data, sealed); err != nil {
		return nil, fmt.Errorf("write target snapshot: %w", err)
	}

	result := &MaterializeResult{SnapshotCopied: true}
	for _, dominoID := range cfg.DominoChain {
		inHash, outHash, output, err := cfg.Source.GetLatestResult(cfg.SnapshotID, dominoID)
		if err != nil {
			continue
		}
		if err := cfg.Target.SaveResult(cfg.SnapshotID, dominoID, inHash, outHash, output, false); err != nil {
			return nil, fmt.Errorf("copy domino %q: %w", dominoID, err)
		}
		result.DominoCount++
	}

	return result, nil
}

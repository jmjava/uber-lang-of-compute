package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/builtin"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// Engine executes domino chains against snapshots with memoization.
type Engine struct {
	store store.Backend
}

// New creates an Engine backed by the given store.
func New(s store.Backend) *Engine {
	return &Engine{store: s}
}

// Run executes a workflow's domino chain and returns a replay log.
func (e *Engine) Run(wf *types.Workflow) (*types.RunResult, error) {
	snap := wf.Spec.Snapshot
	if !snap.Spec.Sealed {
		return nil, fmt.Errorf("snapshot %q is not sealed; cannot execute deterministically", snap.Metadata.Name)
	}

	snapshotData, err := json.Marshal(snap.Spec.Source.Inline)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot data: %w", err)
	}

	snapshotID, err := hash.SnapshotID(snap.Spec.TimeSlice, snap.Spec.Source.Inline)
	if err != nil {
		return nil, fmt.Errorf("compute snapshot ID: %w", err)
	}

	if err := e.store.SaveSnapshot(snapshotID, snap.Spec.TimeSlice, string(snapshotData), true); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	dominoMap := make(map[string]*types.Domino, len(wf.Spec.Dominos))
	for i := range wf.Spec.Dominos {
		d := &wf.Spec.Dominos[i]
		dominoMap[d.Metadata.Name] = d
	}

	chain := wf.Spec.Execution.Chain
	if len(chain) == 0 {
		return nil, fmt.Errorf("execution chain is empty")
	}

	var entries []types.ReplayLogEntry
	outputs := make(map[string]string)

	for _, dominoName := range chain {
		d, ok := dominoMap[dominoName]
		if !ok {
			return nil, fmt.Errorf("domino %q not found in workflow", dominoName)
		}

		inputJSON, err := e.resolveInputs(d, snap, outputs)
		if err != nil {
			return nil, fmt.Errorf("domino %q resolve inputs: %w", dominoName, err)
		}

		inputHash, err := hash.Compute(inputJSON)
		if err != nil {
			return nil, fmt.Errorf("domino %q hash inputs: %w", dominoName, err)
		}

		entry := types.ReplayLogEntry{
			Timestamp:  time.Now().UTC(),
			SnapshotID: snapshotID,
			DominoID:   dominoName,
			InputHash:  inputHash,
		}

		if outHash, out, found, err := e.store.LookupMemo(snapshotID, dominoName, inputHash); err != nil {
			return nil, fmt.Errorf("domino %q memo lookup: %w", dominoName, err)
		} else if found {
			entry.OutputHash = outHash
			entry.Reused = true
			entry.Output = out
			outputs[dominoName] = out

			if err := e.store.SaveResult(snapshotID, dominoName, inputHash, outHash, out, true); err != nil {
				return nil, fmt.Errorf("domino %q save replay: %w", dominoName, err)
			}
		} else {
			out, err := e.executeDomino(d, inputJSON)
			if err != nil {
				return nil, fmt.Errorf("domino %q execute: %w", dominoName, err)
			}

			outputHash, err := hash.Compute(out)
			if err != nil {
				return nil, fmt.Errorf("domino %q hash output: %w", dominoName, err)
			}

			entry.OutputHash = outputHash
			entry.Reused = false
			entry.Output = out
			outputs[dominoName] = out

			if err := e.store.SaveResult(snapshotID, dominoName, inputHash, outputHash, out, false); err != nil {
				return nil, fmt.Errorf("domino %q save result: %w", dominoName, err)
			}
		}

		entries = append(entries, entry)
	}

	finalOutput := outputs[chain[len(chain)-1]]
	return &types.RunResult{
		SnapshotID:  snapshotID,
		Entries:     entries,
		FinalOutput: finalOutput,
	}, nil
}

func (e *Engine) resolveInputs(d *types.Domino, snap types.Snapshot, priorOutputs map[string]string) (string, error) {
	if len(d.Spec.Inputs) == 0 {
		data, err := json.Marshal(snapshotContent(snap))
		return string(data), err
	}

	var parts []interface{}
	for _, input := range d.Spec.Inputs {
		if input.FromDomino != "" {
			out, ok := priorOutputs[input.FromDomino]
			if !ok {
				return "", fmt.Errorf("output from domino %q not available", input.FromDomino)
			}
			var parsed interface{}
			if err := json.Unmarshal([]byte(out), &parsed); err != nil {
				parts = append(parts, out)
			} else {
				parts = append(parts, parsed)
			}
		}
		if input.FromSnapshot != "" {
			data, err := json.Marshal(snapshotContent(snap))
			if err != nil {
				return "", err
			}
			var parsed interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				return "", err
			}
			parts = append(parts, parsed)
		}
	}

	if len(parts) == 1 {
		out, err := json.Marshal(parts[0])
		return string(out), err
	}

	combined, err := json.Marshal(parts)
	return string(combined), err
}

func (e *Engine) executeDomino(d *types.Domino, inputJSON string) (string, error) {
	cmd := d.Spec.Command
	if strings.HasPrefix(cmd, "builtin:") {
		return builtin.Execute(cmd, inputJSON)
	}
	return "", fmt.Errorf("unsupported command %q (only builtin: commands supported in MVP)", cmd)
}

// RunSingle executes one domino against a sealed snapshot with optional dependency outputs.
func (e *Engine) RunSingle(snapshotID string, snap types.Snapshot, domino types.Domino, priorOutputs map[string]string) (*types.ReplayLogEntry, error) {
	if !snap.Spec.Sealed {
		return nil, fmt.Errorf("snapshot %q is not sealed", snap.Metadata.Name)
	}
	if snapshotID == "" {
		return nil, fmt.Errorf("snapshot ID is required")
	}

	d := &domino
	inputJSON, err := e.resolveInputs(d, snap, priorOutputs)
	if err != nil {
		return nil, fmt.Errorf("resolve inputs: %w", err)
	}

	inputHash, err := hash.Compute(inputJSON)
	if err != nil {
		return nil, fmt.Errorf("hash inputs: %w", err)
	}

	entry := types.ReplayLogEntry{
		Timestamp:  time.Now().UTC(),
		SnapshotID: snapshotID,
		DominoID:   domino.Metadata.Name,
		InputHash:  inputHash,
	}

	if outHash, out, found, err := e.store.LookupMemo(snapshotID, domino.Metadata.Name, inputHash); err != nil {
		return nil, fmt.Errorf("memo lookup: %w", err)
	} else if found {
		entry.OutputHash = outHash
		entry.Reused = true
		entry.Output = out
		if err := e.store.SaveResult(snapshotID, domino.Metadata.Name, inputHash, outHash, out, true); err != nil {
			return nil, fmt.Errorf("save replay: %w", err)
		}
		return &entry, nil
	}

	out, err := e.executeDomino(d, inputJSON)
	if err != nil {
		return nil, err
	}

	outputHash, err := hash.Compute(out)
	if err != nil {
		return nil, fmt.Errorf("hash output: %w", err)
	}

	entry.OutputHash = outputHash
	entry.Reused = false
	entry.Output = out
	if err := e.store.SaveResult(snapshotID, domino.Metadata.Name, inputHash, outputHash, out, false); err != nil {
		return nil, fmt.Errorf("save result: %w", err)
	}
	return &entry, nil
}

func snapshotContent(snap types.Snapshot) interface{} {
	if snap.Spec.Source.Inline != nil {
		return snap.Spec.Source.Inline
	}
	if snap.Spec.Source.Path != "" {
		return map[string]string{"path": snap.Spec.Source.Path}
	}
	if snap.Spec.Source.URI != "" {
		return map[string]string{"uri": snap.Spec.Source.URI}
	}
	return map[string]interface{}{}
}

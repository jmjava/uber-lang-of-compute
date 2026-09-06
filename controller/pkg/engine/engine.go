package engine

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// Engine executes domino chains against snapshots with memoization.
type Engine struct {
	store    store.Backend
	executor executor.Config
	// DollarsPerEvaluation prices irreversible steps (M7). Zero disables charging.
	DollarsPerEvaluation float64
}

// New creates an Engine backed by the given store.
func New(s store.Backend) *Engine {
	return &Engine{store: s, executor: executor.DefaultConfig(), DollarsPerEvaluation: theory.DefaultDollarsPerEvaluation}
}

// NewWithExecutor creates an Engine with explicit pluggable runtime configuration.
func NewWithExecutor(s store.Backend, execCfg executor.Config) *Engine {
	return &Engine{store: s, executor: execCfg, DollarsPerEvaluation: theory.DefaultDollarsPerEvaluation}
}

// Run executes a workflow's domino chain and returns a replay log.
func (e *Engine) Run(wf *types.Workflow) (*types.RunResult, error) {
	snap := wf.Spec.Snapshot
	if !snap.Spec.Sealed {
		return nil, fmt.Errorf("snapshot %q is not sealed; cannot execute deterministically", snap.Metadata.Name)
	}

	content, snapshotData, snapshotID, err := snapshot.ResolveEngineContentPreferStore(e.store, snap, "")
	if err != nil {
		return nil, fmt.Errorf("resolve snapshot content: %w", err)
	}

	if snapshotID == "" {
		snapshotID, err = hash.SnapshotID(snap.Spec.TimeSlice, content)
		if err != nil {
			return nil, fmt.Errorf("compute snapshot ID: %w", err)
		}
	}

	if err := e.store.SaveSnapshot(snapshotID, snap.Spec.TimeSlice, snapshotData, true); err != nil {
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

		reg := theory.CommandRegularity(d.Spec.Command)
		if wf.Spec.Execution.Deterministic {
			sb := theory.SandboxPolicy{
				NetworkNone:  wf.Spec.Provisioning.SandboxNetworkNone,
				ReadOnlyRoot: wf.Spec.Provisioning.SandboxReadOnlyRoot,
			}
			if err := theory.AllowDeterministic(d.Spec.Command, sb); err != nil {
				return nil, fmt.Errorf("domino %q: %w", dominoName, err)
			}
		}

		inputJSON, err := e.resolveInputs(d, snap, snapshotID, outputs)
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
			Regularity: reg.String(),
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
	regs := make([]theory.Regularity, 0, len(entries))
	for _, name := range chain {
		regs = append(regs, theory.CommandRegularity(dominoMap[name].Spec.Command))
	}
	w := theory.WorkFromReplay(entries)
	return &types.RunResult{
		SnapshotID:      snapshotID,
		Entries:         entries,
		FinalOutput:     finalOutput,
		MinRegularity:   theory.MinRegularity(regs...).String(),
		WorkEvaluations: w.Evaluations,
		WorkReuses:      w.Reuses,
		WorkCostUSD:     theory.ChargeUSD(w, e.DollarsPerEvaluation),
	}, nil
}

func (e *Engine) resolveInputs(d *types.Domino, snap types.Snapshot, snapshotID string, priorOutputs map[string]string) (string, error) {
	content, _, _, err := snapshot.ResolveEngineContentPreferStore(e.store, snap, snapshotID)
	if err != nil {
		return "", fmt.Errorf("resolve snapshot content: %w", err)
	}

	if len(d.Spec.Inputs) == 0 {
		data, err := json.Marshal(content)
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
			data, err := json.Marshal(content)
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
	return executor.Execute(e.executor, d.Spec.Command, inputJSON)
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
	inputJSON, err := e.resolveInputs(d, snap, snapshotID, priorOutputs)
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
		Regularity: theory.CommandRegularity(domino.Spec.Command).String(),
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

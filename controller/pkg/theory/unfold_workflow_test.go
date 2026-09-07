package theory_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestWorkflowFromUnfoldRunsDeterministically(t *testing.T) {
	const depth, arity = 2, 2
	snap := types.Snapshot{
		Metadata: types.ObjectMeta{Name: "snap"},
		Spec: types.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Sealed:    true,
			Source:    types.SnapshotSource{Inline: map[string]interface{}{"v": 1}},
		},
	}
	wf := theory.WorkflowFromUnfold(depth, arity, snap)
	want := theory.NodeCount(depth, arity)
	if len(wf.Spec.Execution.Chain) != want {
		t.Fatalf("chain %d want node count %d", len(wf.Spec.Execution.Chain), want)
	}
	leaves := 0
	parents := 0
	for _, d := range wf.Spec.Dominos {
		switch d.Spec.Command {
		case "builtin:identity":
			leaves++
		case "builtin:coarsen":
			parents++
		default:
			t.Fatalf("unexpected command %s", d.Spec.Command)
		}
	}
	if leaves != theory.LeafCount(depth, arity) || parents != want-leaves {
		t.Fatalf("leaves=%d parents=%d", leaves, parents)
	}

	run := func() *types.RunResult {
		s, err := store.Open(filepath.Join(t.TempDir(), "unfold.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		res, err := engine.New(s).Run(wf)
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	a, b := run(), run()
	if a.SnapshotID != b.SnapshotID || a.FinalOutput != b.FinalOutput {
		t.Fatal("unfold workflow is not unique")
	}
	if len(a.Entries) != want || a.MinRegularity != "builtin" {
		t.Fatalf("entries=%d regularity=%s", len(a.Entries), a.MinRegularity)
	}
	if a.Entries[len(a.Entries)-1].DominoID != "n" {
		t.Fatalf("root should finish last, got %s", a.Entries[len(a.Entries)-1].DominoID)
	}
	if a.Entries[len(a.Entries)-1].Output != b.Entries[len(b.Entries)-1].Output {
		t.Fatal("root coarsen output is not unique")
	}
}

func TestWorkflowFromUnfoldAggregatesToLeafCount(t *testing.T) {
	const depth, arity = 2, 2
	snap := types.Snapshot{
		Metadata: types.ObjectMeta{Name: "snap"},
		Spec: types.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Sealed:    true,
			Source:    types.SnapshotSource{Inline: map[string]interface{}{"v": 1}},
		},
	}
	run := func(d, k int) float64 {
		wf := theory.WorkflowFromUnfold(d, k, snap)
		s, err := store.Open(filepath.Join(t.TempDir(), "agg.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		res, err := engine.New(s).Run(wf)
		if err != nil {
			t.Fatal(err)
		}
		var parsed struct {
			V float64 `json:"v"`
		}
		if err := json.Unmarshal([]byte(res.FinalOutput), &parsed); err != nil {
			t.Fatal(err)
		}
		return parsed.V
	}
	got := run(depth, arity)
	want := float64(theory.LeafCount(depth, arity))
	if got != want {
		t.Fatalf("root v=%v want leaf count %v", got, want)
	}
	shallower := run(depth-1, arity)
	if got != shallower*float64(arity) {
		t.Fatalf("live fold-up is not self-similar: depth %d → %v, depth %d → %v", depth, got, depth-1, shallower)
	}
}

func TestNodeCountPerfectTree(t *testing.T) {
	if theory.NodeCount(2, 2) != 7 {
		t.Fatalf("got %d want 7", theory.NodeCount(2, 2))
	}
	if theory.NodeCount(0, 3) != 1 {
		t.Fatal("depth 0 is a single node")
	}
}

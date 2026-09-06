package theory_test

import (
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
	// Post-order: root is last.
	if a.Entries[len(a.Entries)-1].DominoID != "n" {
		t.Fatalf("root should finish last, got %s", a.Entries[len(a.Entries)-1].DominoID)
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

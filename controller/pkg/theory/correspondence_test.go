package theory_test

import (
	"path/filepath"
	"testing"
	"time"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/replica"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

func identityWorkflow(sealed bool, payload map[string]interface{}) *types.Workflow {
	return &types.Workflow{
		Kind:     "Workflow",
		Metadata: types.ObjectMeta{Name: "correspondence"},
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				Metadata: types.ObjectMeta{Name: "snap"},
				Spec: types.SnapshotSpec{
					TimeSlice: "2025-04-15T00:00:00Z",
					Sealed:    sealed,
					Source:    types.SnapshotSource{Inline: payload},
				},
			},
			Dominos: []types.Domino{{
				Metadata: types.ObjectMeta{Name: "id"},
				Spec: types.DominoSpec{
					Command: "builtin:identity",
					Inputs:  []types.DominoInput{{FromSnapshot: "snap"}},
				},
			}},
			Execution: types.ExecutionConfig{Chain: []string{"id"}, Deterministic: true},
		},
	}
}

func TestTheoremD1UnsealedSnapshotHasNoTrajectory(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "d1.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	_, err = engine.New(s).Run(identityWorkflow(false, map[string]interface{}{"v": 1}))
	if err == nil {
		t.Fatal("Newtonian correspondence requires a Cauchy surface: unsealed snapshots must not evolve")
	}
}

func TestTheoremD2SealedSnapshotHasUniqueTrajectory(t *testing.T) {
	payload := map[string]interface{}{"rate": 4.25}
	run := func() *types.RunResult {
		s, err := store.Open(filepath.Join(t.TempDir(), "d2.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		res, err := engine.New(s).Run(identityWorkflow(true, payload))
		if err != nil {
			t.Fatal(err)
		}
		return res
	}
	a, b := run(), run()
	if a.SnapshotID != b.SnapshotID || a.FinalOutput != b.FinalOutput {
		t.Fatalf("unique trajectory failed: %+v vs %+v", a, b)
	}
	if a.Entries[0].InputHash != b.Entries[0].InputHash || a.Entries[0].OutputHash != b.Entries[0].OutputHash {
		t.Fatal("hash worldline is not unique")
	}
}

func TestTheoremM1MemoObservationallyEquivalent(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "m1.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	eng := engine.New(s)
	wf := identityWorkflow(true, map[string]interface{}{"v": 7})
	first, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	second, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	if first.Entries[0].Reused {
		t.Fatal("first run must compute")
	}
	if !second.Entries[0].Reused {
		t.Fatal("second run must hit memo")
	}
	if first.FinalOutput != second.FinalOutput || first.Entries[0].OutputHash != second.Entries[0].OutputHash {
		t.Fatal("memo hit is not observationally equivalent to recompute")
	}
}

func TestTheoremW1WheelHasUniqueSuccessor(t *testing.T) {
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{CurrentTimeSlice: start, ActiveContextIndex: 0}
	const n = 4
	interval := 24 * time.Hour

	seen := map[string]wheel.AdvanceResult{}
	key := func(s wheel.State) string {
		return s.CurrentTimeSlice.UTC().String() + "/" + string(rune('0'+s.ActiveContextIndex)) + "/" + string(rune('0'+s.RotationCount))
	}
	cur := state
	for i := 0; i < n*3; i++ {
		res := wheel.AdvanceAfterCompletion(cur, n, interval, 0)
		k := key(cur)
		if prev, ok := seen[k]; ok {
			if prev.State != res.State || prev.AdvancedSlice != res.AdvancedSlice {
				t.Fatalf("non-unique successor at %s", k)
			}
		}
		seen[k] = res
		cur = res.State
	}

	// n seat-advances wrap once: after n steps, slice += interval and seat = 0.
	cur = state
	for i := 0; i < n; i++ {
		cur = wheel.AdvanceAfterCompletion(cur, n, interval, 0).State
	}
	if cur.ActiveContextIndex != 0 || !cur.CurrentTimeSlice.Equal(start.Add(interval)) || cur.RotationCount != 1 {
		t.Fatalf("cylinder period failed: %+v", cur)
	}
}

func TestTheoremC1RoutingIsAFunction(t *testing.T) {
	spec := kblv1alpha1.MultiverseSpec{
		DefaultUniverse: "default",
		Universes: []kblv1alpha1.UniverseRouteSpec{
			{Name: "rates", PluggableUniverseRef: "rates-u", Partitions: []kblv1alpha1.PartitionRule{{Key: "asset_class", Values: []string{"rates"}}}},
			{Name: "default", PluggableUniverseRef: "def-u"},
		},
	}
	r := routing.NewRouter(spec)
	evt := events.SnapshotEvent{SnapshotID: "s", Partitions: map[string]string{"asset_class": "rates"}}
	a, err := r.Resolve(evt)
	if err != nil {
		t.Fatal(err)
	}
	b, err := r.Resolve(evt)
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatalf("routing is not a function: %+v vs %+v", a, b)
	}
}

func TestTheoremR1ReplicaRequiresSealedCauchyView(t *testing.T) {
	src, err := store.OpenSQLite(filepath.Join(t.TempDir(), "src.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer src.Close()
	dst, err := store.OpenSQLite(filepath.Join(t.TempDir(), "dst.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer dst.Close()
	if err := src.SaveSnapshot("live", "2025-04-15", `{}`, false); err != nil {
		t.Fatal(err)
	}
	if _, err := replica.Materialize(replica.MaterializeConfig{SnapshotID: "live", Source: src, Target: dst}); err == nil {
		t.Fatal("unsealed snapshot must not cross universes")
	}
}

func TestTheoremE1EngineSealCollapsesInputEnsemble(t *testing.T) {
	// Running two different sealed payloads yields two different snapshot IDs
	// (distinct Dirac masses). Same payload yields the same ID.
	id := func(v int) string {
		s, err := store.Open(filepath.Join(t.TempDir(), "e1.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		res, err := engine.New(s).Run(identityWorkflow(true, map[string]interface{}{"v": v}))
		if err != nil {
			t.Fatal(err)
		}
		return res.SnapshotID
	}
	if id(1) == id(2) {
		t.Fatal("distinct payloads must not share snapshot identity")
	}
	if id(1) != id(1) {
		t.Fatal("same payload must share snapshot identity")
	}
	if len(id(1)) != hash.SnapshotIDHexLen {
		t.Fatalf("snapshot ID must be %d hex chars (128 bits)", hash.SnapshotIDHexLen)
	}
}

func TestTheoremF1WindowedAggregation(t *testing.T) {
	deep := theory.Unfold(3, 2, "root", 8)
	got := theory.Coarsen(deep)
	want := theory.Unfold(2, 2, "root", 8)
	if !theory.ShapeEqual(got, want) || !theory.ValuesEqual(got, want) {
		t.Fatal("windowed aggregation is not self-similar under coarsening")
	}
}

package theory_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

func TestCommandRegularityLadder(t *testing.T) {
	if theory.CommandRegularity("builtin:identity") != theory.RegularityBuiltin {
		t.Fatal("builtin commands are the Lipschitz-grade (theorem) class")
	}
	if theory.CommandRegularity("julia:greeks") != theory.RegularityPinned {
		t.Fatal("julia commands are pinned discretizations")
	}
	if theory.CommandRegularity("image:custom") != theory.RegularityContract {
		t.Fatal("container commands remain a purity contract")
	}
	if theory.RegularityBuiltin.UniquenessGrade() != "theorem" {
		t.Fatal("expected theorem grade")
	}
	if theory.RegularityPinned.UniquenessGrade() != "empirical-discretization" {
		t.Fatal("expected empirical-discretization grade")
	}
}

func TestLogicalWorkReplaySaves(t *testing.T) {
	// Nature-inspired Landauer/Bennett: recording intermediates means the
	// second pass pays fewer irreversible evaluations. Not a joule claim.
	first := theory.LogicalWork{Evaluations: 3}
	second := theory.LogicalWork{Reuses: 3}
	if !theory.ReplaySaves(first, second) {
		t.Fatal("recorded replay should save irreversible steps")
	}
	if second.IrreversibleSteps() != 0 {
		t.Fatal("full reuse has zero irreversible steps")
	}
}

func TestHistoriesDoNotInterfereWhenSealed(t *testing.T) {
	parent := theory.History{Universe: "rates", SnapshotID: "abc", Worldline: "h1"}
	branches := theory.Branch(parent, []string{"credit", "equities"})
	if len(branches) != 2 {
		t.Fatalf("got %d branches", len(branches))
	}
	for _, b := range branches {
		if !theory.SameRecord(parent, b) {
			t.Fatal("fan-out must carry the sealed record")
		}
		if theory.Interfere(parent, b, false) {
			t.Fatal("sealed copies must not interfere")
		}
		if !theory.Interfere(parent, b, true) {
			t.Fatal("a live share would be the analogue of interference")
		}
	}
}

func TestHomeostasisMatchesSpec(t *testing.T) {
	if !theory.Homeostatic(theory.ReconcileError("Ready", "Ready")) {
		t.Fatal("matching spec/status is homeostatic")
	}
	if theory.Homeostatic(theory.ReconcileError("Ready", "Pending")) {
		t.Fatal("mismatch should be out of bound")
	}
}

func TestSimilarityDimensionAdditiveTreeIsOne(t *testing.T) {
	// Unfold partitions value by arity, so scale=arity and D=1.
	d := theory.SimilarityDimension(4, 4)
	if !theory.AlmostEqual(d, 1) {
		t.Fatalf("got %v want 1", d)
	}
}

func TestWheelWindowLeafCount(t *testing.T) {
	const depth, arity = 2, 3
	n := theory.Unfold(depth, arity, "root", 27)
	leaves := theory.Leaves(n)
	if len(leaves) != theory.LeafCount(depth, arity) {
		t.Fatalf("leaves %d want %d", len(leaves), theory.LeafCount(depth, arity))
	}
	// A wheel with that many seats can traverse the current window's leaves.
	if theory.LeafCount(depth, arity) != 9 {
		t.Fatal("expected 3^2 seats")
	}
}

func TestEscapeTimeIsAFiniteWindow(t *testing.T) {
	// Interior point c=0 never escapes; the window is the iteration budget.
	if got := theory.EscapeTime(0, 8, 2); got != 8 {
		t.Fatalf("interior window: got %d want 8", got)
	}
	// c=1 escapes quickly — deeper maxIter must not decrease the count.
	shallow := theory.EscapeTime(1, 4, 2)
	deep := theory.EscapeTime(1, 64, 2)
	if deep < shallow {
		t.Fatal("deeper explorer window must be monotonic")
	}
}

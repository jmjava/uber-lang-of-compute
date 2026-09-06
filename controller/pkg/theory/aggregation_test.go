package theory_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

func TestCoarsenRecoversShallowerWindow(t *testing.T) {
	// Theorem F1 (self-similarity of windowed aggregation):
	// Coarsen(Unfold(d, k)) is shape- and value-equal to Unfold(d-1, k)
	// under additive child partition.
	const arity = 3
	const seed = 27.0
	for d := 1; d <= 4; d++ {
		deep := theory.Unfold(d, arity, "root", seed)
		shallow := theory.Unfold(d-1, arity, "root", seed)
		got := theory.Coarsen(deep)
		if !theory.ShapeEqual(got, shallow) {
			t.Fatalf("depth %d: coarsened shape != shallower window", d)
		}
		if !theory.ValuesEqual(got, shallow) {
			t.Fatalf("depth %d: coarsened values != shallower window (conservation failed)", d)
		}
		if theory.Height(got) != theory.Height(shallow) {
			t.Fatalf("depth %d: height %d != %d", d, theory.Height(got), theory.Height(shallow))
		}
	}
}

func TestUnfoldHeight(t *testing.T) {
	n := theory.Unfold(3, 2, "r", 8)
	if h := theory.Height(n); h != 3 {
		t.Fatalf("height %d want 3", h)
	}
}

func TestWindowIsFinite(t *testing.T) {
	// Only depth ≤ D is materialized — the "windowed" half of the pattern.
	n := theory.Unfold(2, 4, "r", 16)
	if theory.Height(n) != 2 {
		t.Fatal("window leaked extra depth")
	}
	for _, c := range n.Children {
		for _, gc := range c.Children {
			if len(gc.Children) != 0 {
				t.Fatal("leaves must not hold children")
			}
		}
	}
}

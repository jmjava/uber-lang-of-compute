package wheel_test

import (
	"testing"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

func TestWindowSeatsMatchExplorerLeaves(t *testing.T) {
	const depth, arity = 2, 3
	if err := wheel.ValidateWindowSeats(9, depth, arity); err != nil {
		t.Fatal(err)
	}
	if err := wheel.ValidateWindowSeats(8, depth, arity); err == nil {
		t.Fatal("mismatched seat count should fail")
	}
	labels := wheel.SeatLabels(depth, arity)
	if len(labels) != wheel.SeatCount(depth, arity) {
		t.Fatalf("labels %d want %d", len(labels), wheel.SeatCount(depth, arity))
	}
	seen := map[string]bool{}
	for _, l := range labels {
		if seen[l] {
			t.Fatalf("duplicate seat label %s", l)
		}
		seen[l] = true
	}
}

func TestValidateExplorerWindow(t *testing.T) {
	if err := wheel.ValidateExplorerWindow(2, 0, 0); err != nil {
		t.Fatal("unset window must be allowed")
	}
	if err := wheel.ValidateExplorerWindow(9, 2, 3); err != nil {
		t.Fatal(err)
	}
	if err := wheel.ValidateExplorerWindow(4, 2, 3); err == nil {
		t.Fatal("seat count must match arity^depth")
	}
	if err := wheel.ValidateExplorerWindow(4, 2, 0); err == nil {
		t.Fatal("partial window spec must be rejected")
	}
}

func TestFullTurnVisitsEveryWindowLeaf(t *testing.T) {
	labels := wheel.SeatLabels(2, 2) // 4 leaves
	n := len(labels)
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{CurrentTimeSlice: start}
	for i := 0; i < n; i++ {
		if state.ActiveContextIndex != i {
			t.Fatalf("before step %d: seat %d", i, state.ActiveContextIndex)
		}
		res := wheel.AdvanceAfterCompletion(state, n, time.Hour, 0)
		if i < n-1 && res.AdvancedSlice {
			t.Fatal("slice must not advance before the last leaf")
		}
		if i == n-1 && !res.AdvancedSlice {
			t.Fatal("slice must advance after the last leaf")
		}
		state = res.State
	}
	if state.ActiveContextIndex != 0 {
		t.Fatalf("expected wrap to seat 0, got %d", state.ActiveContextIndex)
	}
}

func TestCoarsenReducesSeatsByArity(t *testing.T) {
	// F1: coarsen one level of a k-ary window divides leaf count by k.
	const arity = 3
	deep := wheel.SeatCount(3, arity)
	shallow := wheel.SeatCount(2, arity)
	if deep != shallow*arity {
		t.Fatalf("seat count %d -> %d want factor %d", deep, shallow, arity)
	}
}

package wheel_test

import (
	"testing"
	"time"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

func TestAdvanceAfterCompletionRotatesContext(t *testing.T) {
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{
		CurrentTimeSlice:   start,
		ActiveContextIndex: 0,
		RotationCount:      0,
	}

	result := wheel.AdvanceAfterCompletion(state, 3, 24*time.Hour, 0)

	if result.AdvancedSlice {
		t.Error("expected context rotation only, not slice advance")
	}
	if result.State.ActiveContextIndex != 1 {
		t.Errorf("expected context index 1, got %d", result.State.ActiveContextIndex)
	}
	if !result.State.CurrentTimeSlice.Equal(start) {
		t.Error("time slice should not change on context rotation")
	}
}

func TestAdvanceAfterCompletionAdvancesSlice(t *testing.T) {
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{
		CurrentTimeSlice:   start,
		ActiveContextIndex: 2,
		RotationCount:      0,
	}

	result := wheel.AdvanceAfterCompletion(state, 3, 24*time.Hour, 0)

	if !result.AdvancedSlice {
		t.Error("expected slice advance after last context")
	}
	if result.State.ActiveContextIndex != 0 {
		t.Errorf("expected context index reset to 0, got %d", result.State.ActiveContextIndex)
	}
	expected := start.Add(24 * time.Hour)
	if !result.State.CurrentTimeSlice.Equal(expected) {
		t.Errorf("expected time slice %v, got %v", expected, result.State.CurrentTimeSlice)
	}
	if result.State.RotationCount != 1 {
		t.Errorf("expected rotation count 1, got %d", result.State.RotationCount)
	}
}

func TestMaxRotationsStopsWheel(t *testing.T) {
	state := wheel.State{
		CurrentTimeSlice:   time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC),
		ActiveContextIndex: 1,
		RotationCount:      0,
	}

	result := wheel.AdvanceAfterCompletion(state, 2, time.Hour, 1)

	if !result.Done {
		t.Error("expected wheel to stop at max rotations")
	}
	if result.State.RotationCount != 1 {
		t.Errorf("expected rotation count 1, got %d", result.State.RotationCount)
	}
}

func TestCylinderPeriodAdvancesTimeOnce(t *testing.T) {
	// Correspondence V: n seats form a discrete circle; one full turn advances
	// the time coordinate by exactly one interval (helix on a cylinder).
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{CurrentTimeSlice: start}
	const n = 5
	interval := 12 * time.Hour
	for i := 0; i < n; i++ {
		state = wheel.AdvanceAfterCompletion(state, n, interval, 0).State
	}
	if state.ActiveContextIndex != 0 {
		t.Errorf("expected seat 0 after full turn, got %d", state.ActiveContextIndex)
	}
	if !state.CurrentTimeSlice.Equal(start.Add(interval)) {
		t.Errorf("expected slice %+v, got %+v", start.Add(interval), state.CurrentTimeSlice)
	}
}

func TestParseIntervalDays(t *testing.T) {
	d, err := wheel.ParseInterval("1d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 24*time.Hour {
		t.Errorf("expected 24h, got %v", d)
	}
}

func TestWorkflowNameWithinLimit(t *testing.T) {
	name := wheel.WorkflowName("finance-wheel", "node-a-context", "20250415t000000z")
	if len(name) > 63 {
		t.Errorf("workflow name exceeds 63 chars: %d", len(name))
	}
}

func TestLookaheadIsAFunctionOfState(t *testing.T) {
	contexts := []string{"compute-a", "compute-b"}
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{CurrentTimeSlice: start, ActiveContextIndex: 0}
	interval := 24 * time.Hour

	a, doneA, err := wheel.Lookahead("finance-wheel", contexts, state, interval, 0)
	if err != nil || doneA {
		t.Fatalf("lookahead: %v done=%v", err, doneA)
	}
	b, doneB, err := wheel.Lookahead("finance-wheel", contexts, state, interval, 0)
	if err != nil || doneB || a != b {
		t.Fatalf("lookahead must be a function: %v vs %v done=%v", a, b, doneB)
	}

	landed := wheel.AdvanceAfterCompletion(state, len(contexts), interval, 0)
	want := wheel.WorkflowName("finance-wheel", contexts[landed.State.ActiveContextIndex], wheel.FormatTimeSlice(landed.State.CurrentTimeSlice))
	if a.Name != want || a.Context != "compute-b" {
		t.Fatalf("pre-provisioned slot %q/%q want %q/compute-b", a.Name, a.Context, want)
	}
}

func TestLookaheadWrapsToNextSlice(t *testing.T) {
	contexts := []string{"compute-a", "compute-b"}
	start := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	state := wheel.State{CurrentTimeSlice: start, ActiveContextIndex: 1}
	slot, done, err := wheel.Lookahead("finance-wheel", contexts, state, 24*time.Hour, 0)
	if err != nil || done {
		t.Fatalf("lookahead: %v done=%v", err, done)
	}
	if slot.Context != "compute-a" {
		t.Fatalf("wrap should land on first context, got %s", slot.Context)
	}
	if slot.State.CurrentTimeSlice != start.Add(24*time.Hour) {
		t.Fatalf("wrap should advance the slice, got %v", slot.State.CurrentTimeSlice)
	}
}

func TestLookaheadStopsWhenDone(t *testing.T) {
	contexts := []string{"a", "b"}
	state := wheel.State{
		CurrentTimeSlice:   time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC),
		ActiveContextIndex: 1,
		RotationCount:      0,
	}
	slot, done, err := wheel.Lookahead("w", contexts, state, time.Hour, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !done {
		t.Fatal("expected no next note at max rotations")
	}
	if slot.Name != "" {
		t.Fatalf("done lookahead must not name a workflow, got %q", slot.Name)
	}
}

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

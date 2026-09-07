package review

import (
	"context"
	"fmt"
	"time"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

// Report is the operator-visible record of a reviewer walkthrough.
type Report struct {
	SnapshotID       string   `json:"snapshot_id"`
	HeadLink         string   `json:"head_link"`
	Worldline        string   `json:"worldline"`
	EventID          string   `json:"event_id"`
	FirstEvaluations int      `json:"first_evaluations"`
	SecondReuses     int      `json:"second_reuses"`
	LookaheadName    string   `json:"lookahead_name"`
	FanoutUniverses  []string `json:"fanout_universes"`
	FinalOutput      string   `json:"final_output"`
}

// Walkthrough runs the workshop path: sealed snapshot → builtin chain →
// memo replay → wheel lookahead → HeadLink fan-out.
func Walkthrough(backend store.Backend, wf *types.Workflow) (*Report, error) {
	if wf == nil {
		return nil, fmt.Errorf("workflow is required")
	}
	eng := engine.New(backend)

	first, err := eng.Run(wf)
	if err != nil {
		return nil, fmt.Errorf("first run: %w", err)
	}
	if err := theory.VerifySpine(first.SnapshotID, first.Entries); err != nil {
		return nil, fmt.Errorf("first spine: %w", err)
	}
	if first.WorkEvaluations == 0 || first.WorkReuses != 0 {
		return nil, fmt.Errorf("first run must evaluate, not replay: eval=%d reuse=%d", first.WorkEvaluations, first.WorkReuses)
	}

	second, err := eng.Run(wf)
	if err != nil {
		return nil, fmt.Errorf("memo run: %w", err)
	}
	if second.WorkEvaluations != 0 || second.WorkReuses != len(second.Entries) {
		return nil, fmt.Errorf("second run must be a full memo replay: eval=%d reuse=%d", second.WorkEvaluations, second.WorkReuses)
	}
	if first.SnapshotID != second.SnapshotID || first.HeadLink != second.HeadLink || first.FinalOutput != second.FinalOutput {
		return nil, fmt.Errorf("memo replay diverged from the first sealed run")
	}

	interval := 24 * time.Hour
	start := wf.Spec.Snapshot.Spec.TimeSlice
	if start == "" {
		start = "2025-04-15T00:00:00Z"
	}
	state, err := wheel.InitialState(start, interval, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("wheel state: %w", err)
	}
	slot, done, err := wheel.Lookahead("review-wheel", []string{"ctx-a", "ctx-b"}, state, interval, 0)
	if err != nil {
		return nil, fmt.Errorf("lookahead: %w", err)
	}
	if done || slot.Name == "" {
		return nil, fmt.Errorf("lookahead must name the next seat")
	}

	worldline := theory.Worldline(first.Entries)
	eventID, err := events.EventID(first.SnapshotID, worldline, "rates", wf.Metadata.Name)
	if err != nil {
		return nil, fmt.Errorf("event id: %w", err)
	}
	evt := events.SnapshotEvent{
		EventID:    eventID,
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: first.SnapshotID,
		TimeSlice:  wf.Spec.Snapshot.Spec.TimeSlice,
		Workflow:   wf.Metadata.Name,
		Universe:   "rates",
		Worldline:  worldline,
		HeadLink:   first.HeadLink,
	}
	router := routing.NewRouter(kblv1alpha1.MultiverseSpec{
		Universes: []kblv1alpha1.UniverseRouteSpec{
			{Name: "rates"},
			{Name: "credit"},
			{Name: "equities"},
		},
	})
	branches, err := router.Fanout(evt)
	if err != nil {
		return nil, fmt.Errorf("fanout: %w", err)
	}
	if len(branches) != 2 {
		return nil, fmt.Errorf("fanout want 2 branches, got %d", len(branches))
	}
	parent := routing.HistoryFromEvent(evt)
	names := make([]string, 0, len(branches))
	for _, b := range branches {
		if !theory.SameRecord(parent, b) || b.HeadLink != first.HeadLink {
			return nil, fmt.Errorf("fan-out must copy the sealed HeadLink")
		}
		names = append(names, b.Universe)
	}

	bus := events.NewMemoryBus()
	defer bus.Close()
	if err := bus.Publish(context.Background(), evt); err != nil {
		return nil, fmt.Errorf("publish: %w", err)
	}
	if err := bus.Publish(context.Background(), evt); err != nil {
		return nil, fmt.Errorf("retry publish: %w", err)
	}
	if len(bus.Published()) != 1 {
		return nil, fmt.Errorf("event bus retry must be idempotent, published %d", len(bus.Published()))
	}

	return &Report{
		SnapshotID:       first.SnapshotID,
		HeadLink:         first.HeadLink,
		Worldline:        worldline,
		EventID:          eventID,
		FirstEvaluations: first.WorkEvaluations,
		SecondReuses:     second.WorkReuses,
		LookaheadName:    slot.Name,
		FanoutUniverses:  names,
		FinalOutput:      first.FinalOutput,
	}, nil
}

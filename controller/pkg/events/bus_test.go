package events_test

import (
	"context"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
)

func TestMemoryBusPublishIsIdempotentOnEventID(t *testing.T) {
	bus := events.NewMemoryBus()
	defer bus.Close()

	n := 0
	if err := bus.Subscribe(context.Background(), func(context.Context, events.SnapshotEvent) error {
		n++
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	id, err := events.EventID("snap", "h1|h2", "rates", "wf")
	if err != nil {
		t.Fatal(err)
	}
	evt := events.SnapshotEvent{
		EventID:    id,
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: "snap",
		Worldline:  "h1|h2",
		Universe:   "rates",
		Workflow:   "wf",
		HeadLink:   "link-head",
	}
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("retry of the same EventID must not re-deliver, got %d deliveries", n)
	}
	if len(bus.Published()) != 1 {
		t.Fatalf("published log %d want 1", len(bus.Published()))
	}
}

func TestMemoryBusRejectsMissingEventID(t *testing.T) {
	bus := events.NewMemoryBus()
	defer bus.Close()

	err := bus.Publish(context.Background(), events.SnapshotEvent{
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: "snap",
		Worldline:  "h1|h2",
		Universe:   "rates",
		Workflow:   "wf",
	})
	if err == nil {
		t.Fatal("publish without EventID must be refused")
	}
	if len(bus.Published()) != 0 {
		t.Fatalf("refused event must not be recorded, got %d", len(bus.Published()))
	}
}

package events_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
)

func TestSnapshotEventIDIsFunctionOfPayload(t *testing.T) {
	a, err := events.EventID("snap", "h1|h2", "rates", "wf")
	if err != nil {
		t.Fatal(err)
	}
	b, err := events.EventID("snap", "h1|h2", "rates", "wf")
	if err != nil || a != b {
		t.Fatalf("EventID must be a function: %q vs %q err=%v", a, b, err)
	}
	c, err := events.EventID("snap", "h1|h2", "credit", "wf")
	if err != nil || a == c {
		t.Fatal("universe must be in the event identity")
	}
}

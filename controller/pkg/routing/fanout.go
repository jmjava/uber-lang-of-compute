package routing

import (
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

// HistoryFromEvent lifts a snapshot-completed event into a classical history.
func HistoryFromEvent(evt events.SnapshotEvent) theory.History {
	return theory.History{
		Universe:   evt.Universe,
		SnapshotID: evt.SnapshotID,
		Worldline:  evt.Worldline,
		HeadLink:   evt.HeadLink,
	}
}

// Fanout copies a sealed parent history into every other universe on the
// Multiverse spec. Branches share snapshot ID, worldline, and HeadLink; they
// do not share live stores (Interfere with liveShare=false is false).
func (r *Router) Fanout(evt events.SnapshotEvent) ([]theory.History, error) {
	if r == nil {
		return nil, fmt.Errorf("router is nil")
	}
	if evt.SnapshotID == "" || evt.Worldline == "" {
		return nil, fmt.Errorf("fanout requires a complete sealed record")
	}
	parent := HistoryFromEvent(evt)
	names := make([]string, 0, len(r.spec.Universes))
	for _, u := range r.spec.Universes {
		names = append(names, u.Name)
	}
	return theory.Branch(parent, names), nil
}

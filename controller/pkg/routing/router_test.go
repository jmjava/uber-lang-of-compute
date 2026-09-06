package routing_test

import (
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

func TestRouterPartitionMatch(t *testing.T) {
	spec := kblv1alpha1.MultiverseSpec{
		DefaultUniverse: "finance-local",
		Universes: []kblv1alpha1.UniverseRouteSpec{
			{
				Name:                 "rates-universe",
				PluggableUniverseRef: "rates-universe",
				Partitions:           []kblv1alpha1.PartitionRule{{Key: "asset_class", Values: []string{"rates"}}},
			},
			{
				Name:                 "finance-local",
				PluggableUniverseRef: "finance-universe",
			},
		},
	}
	r := routing.NewRouter(spec)

	target, err := r.Resolve(events.SnapshotEvent{
		SnapshotID: "abc",
		Partitions: map[string]string{"asset_class": "rates"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if target.Universe != "rates-universe" {
		t.Errorf("expected rates-universe, got %s", target.Universe)
	}
}

func TestRouterTimeSliceOverride(t *testing.T) {
	spec := kblv1alpha1.MultiverseSpec{
		DefaultUniverse: "finance-local",
		Universes: []kblv1alpha1.UniverseRouteSpec{
			{Name: "finance-local", PluggableUniverseRef: "finance-universe"},
			{Name: "credit-universe", PluggableUniverseRef: "credit-universe"},
		},
		TimeSliceRoutes: []kblv1alpha1.TimeSliceRoute{{
			TimeSlice: "2025-04-15",
			Universe:  "credit-universe",
		}},
	}
	r := routing.NewRouter(spec)

	target, err := r.Resolve(events.SnapshotEvent{
		TimeSlice:  "2025-04-15",
		Partitions: map[string]string{"asset_class": "rates"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if target.Universe != "credit-universe" {
		t.Errorf("expected credit-universe, got %s", target.Universe)
	}
}

func TestFanoutCarriesSealedHistoryWithoutInterference(t *testing.T) {
	spec := kblv1alpha1.MultiverseSpec{
		Universes: []kblv1alpha1.UniverseRouteSpec{
			{Name: "rates", PluggableUniverseRef: "rates-u"},
			{Name: "credit", PluggableUniverseRef: "credit-u"},
			{Name: "equities", PluggableUniverseRef: "eq-u"},
		},
	}
	r := routing.NewRouter(spec)
	evt := events.SnapshotEvent{
		SnapshotID: "snap-sealed",
		Universe:   "rates",
		Worldline:  "h1|h2",
	}
	branches := r.Fanout(evt)
	if len(branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches))
	}
	parent := routing.HistoryFromEvent(evt)
	for _, b := range branches {
		if b.Universe == "rates" {
			t.Fatal("fan-out must not include the parent universe")
		}
		if !theory.SameRecord(parent, b) {
			t.Fatal("sealed snapshot ID and worldline must be copied")
		}
		if theory.Interfere(parent, b, false) {
			t.Fatal("sealed fan-out must not interfere")
		}
	}
}

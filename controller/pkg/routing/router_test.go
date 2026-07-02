package routing_test

import (
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
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

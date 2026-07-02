package routing

import (
	"fmt"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
)

// Target is the resolved routing destination for a snapshot event.
type Target struct {
	Universe          string
	PluggableUniverse string
	ComputeContextRef string
}

// Router resolves snapshot events to universe targets using Multiverse rules.
type Router struct {
	spec kblv1alpha1.MultiverseSpec
	refs map[string]string // universe name -> pluggableUniverseRef
}

// NewRouter builds a router from Multiverse spec.
func NewRouter(spec kblv1alpha1.MultiverseSpec) *Router {
	refs := make(map[string]string, len(spec.Universes))
	for _, u := range spec.Universes {
		refs[u.Name] = u.PluggableUniverseRef
	}
	return &Router{spec: spec, refs: refs}
}

// Resolve selects the target universe for an event.
func (r *Router) Resolve(evt events.SnapshotEvent) (Target, error) {
	if r == nil {
		return Target{}, fmt.Errorf("router is nil")
	}

	// Time-slice override takes precedence.
	for _, ts := range r.spec.TimeSliceRoutes {
		if ts.TimeSlice != "" && ts.TimeSlice == evt.TimeSlice {
			return Target{
				Universe:          ts.Universe,
				PluggableUniverse: r.refs[ts.Universe],
				ComputeContextRef: ts.ComputeContextRef,
			}, nil
		}
	}

	// Partition-based routing.
	for _, u := range r.spec.Universes {
		if matchesPartitions(u.Partitions, evt.Partitions) {
			return Target{
				Universe:          u.Name,
				PluggableUniverse: u.PluggableUniverseRef,
				ComputeContextRef: u.ComputeContextRef,
			}, nil
		}
	}

	// Default universe fallback.
	if r.spec.DefaultUniverse != "" {
		for _, u := range r.spec.Universes {
			if u.Name == r.spec.DefaultUniverse {
				return Target{
					Universe:          u.Name,
					PluggableUniverse: u.PluggableUniverseRef,
					ComputeContextRef: u.ComputeContextRef,
				}, nil
			}
		}
		return Target{
			Universe:          r.spec.DefaultUniverse,
			PluggableUniverse: r.refs[r.spec.DefaultUniverse],
		}, nil
	}

	return Target{}, fmt.Errorf("no route for snapshot %s", evt.SnapshotID)
}

func matchesPartitions(rules []kblv1alpha1.PartitionRule, labels map[string]string) bool {
	if len(rules) == 0 {
		return false
	}
	for _, rule := range rules {
		val, ok := labels[rule.Key]
		if !ok {
			return false
		}
		if !contains(rule.Values, val) {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

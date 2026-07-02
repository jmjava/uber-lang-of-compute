package controller

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func (r *WorkflowReconciler) publishSnapshotEvent(ctx context.Context, wf *kblv1alpha1.Workflow, result *types.RunResult) {
	bus := r.EventBus
	if bus == nil {
		bus = events.NewNoopBus()
	}

	evt := events.SnapshotEvent{
		EventID:    uuid.NewString(),
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: result.SnapshotID,
		TimeSlice:  wf.Spec.Snapshot.TimeSlice,
		Workflow:   wf.Name,
		Namespace:  wf.Namespace,
		Universe:   wf.Spec.Routing.Universe,
		Multiverse: wf.Spec.Routing.MultiverseRef,
		Partitions: partitionLabels(wf),
		OccurredAt: time.Now().UTC(),
	}
	if len(result.Entries) > 0 {
		evt.FinalOutput = result.Entries[len(result.Entries)-1].OutputHash
	}

	if wf.Spec.Routing.MultiverseRef != "" {
		var mv kblv1alpha1.Multiverse
		if err := r.Get(ctx, client.ObjectKey{Namespace: wf.Namespace, Name: wf.Spec.Routing.MultiverseRef}, &mv); err == nil {
			_ = PublishSnapshotEvent(ctx, bus, &mv, evt)
			return
		}
	}
	_ = bus.Publish(ctx, evt)
}

func partitionLabels(wf *kblv1alpha1.Workflow) map[string]string {
	out := make(map[string]string)
	for k, v := range wf.Labels {
		if strings.HasPrefix(k, "kbl.io/partition-") {
			out[strings.TrimPrefix(k, "kbl.io/partition-")] = v
		}
	}
	return out
}

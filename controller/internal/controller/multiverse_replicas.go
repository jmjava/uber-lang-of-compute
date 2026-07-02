package controller

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
)

var readReplicaNameSanitizer = regexp.MustCompile(`[^a-z0-9-]+`)

func (r *MultiverseReconciler) ensureReadReplica(ctx context.Context, mv *kblv1alpha1.Multiverse, evt events.SnapshotEvent, target routing.Target) error {
	name := readReplicaName(evt, target)
	key := client.ObjectKey{Namespace: mv.Namespace, Name: name}

	var existing kblv1alpha1.ReadReplica
	err := r.Get(ctx, key, &existing)
	if err == nil {
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return err
	}

	rr := &kblv1alpha1.ReadReplica{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: mv.Namespace,
			Labels: map[string]string{
				"kbl.io/multiverse":      mv.Name,
				"kbl.io/target-universe": target.Universe,
				"kbl.io/source-workflow": evt.Workflow,
			},
		},
		Spec: kblv1alpha1.ReadReplicaSpec{
			MultiverseRef:           mv.Name,
			RoutedEventID:           evt.EventID,
			SourceSnapshotID:        evt.SnapshotID,
			SourceWorkflow:          evt.Workflow,
			SourceNamespace:         evt.Namespace,
			TimeSlice:               evt.TimeSlice,
			TargetUniverse:          target.Universe,
			TargetComputeContextRef: target.ComputeContextRef,
			PluggableUniverseRef:    target.PluggableUniverse,
			FinalOutputHash:         evt.FinalOutput,
			Partitions:              evt.Partitions,
		},
	}

	if err := controllerutil.SetControllerReference(mv, rr, r.Scheme); err != nil {
		return fmt.Errorf("set owner ref: %w", err)
	}
	return r.Create(ctx, rr)
}

func readReplicaName(evt events.SnapshotEvent, target routing.Target) string {
	parts := []string{
		"replica",
		evt.Workflow,
		truncateID(evt.SnapshotID, 8),
		target.Universe,
	}
	name := strings.ToLower(strings.Join(parts, "-"))
	name = readReplicaNameSanitizer.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 63 {
		name = name[:63]
		name = strings.TrimRight(name, "-")
	}
	return name
}

func truncateID(id string, n int) string {
	if len(id) <= n {
		return id
	}
	return id[:n]
}

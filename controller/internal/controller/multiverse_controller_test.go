package controller_test

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	kblcontroller "github.com/jmjava/uber-lang-of-compute/controller/internal/controller"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
)

func TestMultiverseRoutesSnapshotEvent(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	universe := &kblv1alpha1.PluggableUniverse{
		ObjectMeta: metav1.ObjectMeta{Name: "finance-universe"},
		Spec: kblv1alpha1.PluggableUniverseSpec{
			ExecutionEngine: kblv1alpha1.ExecutionEngineSpec{Type: "builtin"},
			DataLayer:       kblv1alpha1.DataLayerSpec{Type: "tsdb"},
		},
	}

	mv := &kblv1alpha1.Multiverse{
		ObjectMeta: metav1.ObjectMeta{Name: "finance-mv", Namespace: "default"},
		Spec: kblv1alpha1.MultiverseSpec{
			DefaultUniverse: "finance-local",
			Universes: []kblv1alpha1.UniverseRouteSpec{{
				Name:                 "finance-local",
				PluggableUniverseRef: "finance-universe",
				ComputeContextRef:    "node-a-ctx",
			}},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(mv).
		WithObjects(mv, universe).
		Build()

	bus := events.NewMemoryBus()
	r := &kblcontroller.MultiverseReconciler{
		Client: cl,
		Scheme: scheme,
		Bus:    bus,
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "finance-mv", Namespace: "default"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = bus.Publish(ctx, events.SnapshotEvent{
		EventID:    "evt-1",
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: "snap-123",
		TimeSlice:  "2025-04-15",
		Workflow:   "finance-curve",
		Namespace:  "default",
		Multiverse: "finance-mv",
	})

	time.Sleep(100 * time.Millisecond)

	var updated kblv1alpha1.Multiverse
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get multiverse: %v", err)
	}
	if len(updated.Status.RoutedEvents) == 0 {
		t.Fatal("expected routed event in status")
	}
	if updated.Status.RoutedEvents[0].TargetUniverse != "finance-local" {
		t.Errorf("expected finance-local, got %s", updated.Status.RoutedEvents[0].TargetUniverse)
	}
}

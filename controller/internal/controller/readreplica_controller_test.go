package controller_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	kblcontroller "github.com/jmjava/uber-lang-of-compute/controller/internal/controller"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestReadReplicaMaterializesFromWorkflow(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	storeDir := t.TempDir()
	sourcePath := filepath.Join(storeDir, "default", "rates-curve.db")
	targetPath := filepath.Join(storeDir, "default", "replicas", "finance-local.db")

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "rates-curve", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.WorkflowSpec{
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{
						"instruments": []interface{}{
							map[string]interface{}{"instrument_id": "US10Y", "rate": 4.25},
						},
					},
				},
				Sealed: true,
			},
			Dominos: []kblv1alpha1.DominoSpec{
				{Name: "load", SnapshotRef: "snap", Command: "builtin:identity"},
			},
			Execution:    kblv1alpha1.ExecutionSpec{Chain: []string{"load"}, Deterministic: true},
			Provisioning: kblv1alpha1.ProvisioningSpec{StorePath: sourcePath, NodeLocal: true},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&kblv1alpha1.Workflow{}, &kblv1alpha1.ReadReplica{}).
		WithObjects(wf).
		Build()

	wfRec := &kblcontroller.WorkflowReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}
	wfReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: "rates-curve", Namespace: "default"}}
	for i := 0; i < 2; i++ {
		if _, err := wfRec.Reconcile(context.Background(), wfReq); err != nil {
			t.Fatalf("workflow reconcile %d: %v", i, err)
		}
	}

	var completed kblv1alpha1.Workflow
	if err := cl.Get(context.Background(), wfReq.NamespacedName, &completed); err != nil {
		t.Fatalf("get workflow: %v", err)
	}

	rr := &kblv1alpha1.ReadReplica{
		ObjectMeta: metav1.ObjectMeta{Name: "replica-rates-curve", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.ReadReplicaSpec{
			MultiverseRef:    "finance-mv",
			RoutedEventID:    "evt-1",
			SourceSnapshotID: completed.Status.SnapshotID,
			SourceWorkflow:   "rates-curve",
			SourceNamespace:  "default",
			TimeSlice:        "2025-04-15T00:00:00Z",
			TargetUniverse:   "finance-local",
		},
	}
	if err := cl.Create(context.Background(), rr); err != nil {
		t.Fatalf("create read replica: %v", err)
	}

	rrRec := &kblcontroller.ReadReplicaReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}
	rrReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: "replica-rates-curve", Namespace: "default"}}
	if _, err := rrRec.Reconcile(context.Background(), rrReq); err != nil {
		t.Fatalf("read replica reconcile: %v", err)
	}

	var updated kblv1alpha1.ReadReplica
	if err := cl.Get(context.Background(), rrReq.NamespacedName, &updated); err != nil {
		t.Fatalf("get read replica: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.ReadReplicaPhaseReady {
		t.Fatalf("expected Ready, got %s: %s", updated.Status.Phase, updated.Status.Message)
	}

	target, err := store.OpenSQLite(targetPath)
	if err != nil {
		t.Fatalf("open target store: %v", err)
	}
	defer target.Close()

	_, data, sealed, err := target.GetSnapshot(completed.Status.SnapshotID)
	if err != nil || data == "" || !sealed {
		t.Fatalf("target snapshot missing: err=%v sealed=%v", err, sealed)
	}
}

func TestMultiverseCreatesReadReplica(t *testing.T) {
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
	r := &kblcontroller.MultiverseReconciler{Client: cl, Scheme: scheme, Bus: bus}

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
		EventID:    "evt-replica-1",
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: "snap-1234567890",
		TimeSlice:  "2025-04-15",
		Workflow:   "finance-curve",
		Namespace:  "default",
		Multiverse: "finance-mv",
	})

	time.Sleep(100 * time.Millisecond)

	var rrList kblv1alpha1.ReadReplicaList
	if err := cl.List(context.Background(), &rrList); err != nil {
		t.Fatalf("list read replicas: %v", err)
	}
	if len(rrList.Items) == 0 {
		t.Fatal("expected ReadReplica created by multiverse routing")
	}
	if rrList.Items[0].Spec.TargetUniverse != "finance-local" {
		t.Fatalf("expected finance-local, got %s", rrList.Items[0].Spec.TargetUniverse)
	}
}

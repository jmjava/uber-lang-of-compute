package controller_test

import (
	"context"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	kblcontroller "github.com/jmjava/uber-lang-of-compute/controller/internal/controller"
)

func TestSnapshotReconcilerSealsSnapshot(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	storeDir := t.TempDir()

	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "curve-snap",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.SnapshotSpec{
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
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(snap).WithObjects(snap).Build()
	r := &kblcontroller.SnapshotReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: storeDir,
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "curve-snap", Namespace: "default"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var updated kblv1alpha1.Snapshot
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.SnapshotPhaseSealed {
		t.Fatalf("expected Sealed, got %s", updated.Status.Phase)
	}
	if updated.Status.SnapshotID == "" {
		t.Fatal("expected snapshot ID")
	}
	if _, err := filepath.Glob(filepath.Join(storeDir, "default", "*.db")); err != nil {
		t.Fatalf("glob store: %v", err)
	}
}

func TestDominoReconcilerExecutesAgainstSnapshot(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	storeDir := t.TempDir()
	storePath := filepath.Join(storeDir, "default", "curve-snap.db")

	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "curve-snap", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.SnapshotSpec{
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
		Status: kblv1alpha1.SnapshotStatus{
			Phase:      kblv1alpha1.SnapshotPhaseSealed,
			SnapshotID: "abc123def4567890",
		},
	}

	load := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:identity",
			StorePath:   storePath,
		},
	}

	interp := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "interpolate", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:interpolate",
			DependsOn:   []string{"load"},
			Inputs:      []kblv1alpha1.DominoInput{{FromDomino: "load"}},
			StorePath:   storePath,
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(snap, load, interp).
		WithObjects(snap, load, interp).
		Build()

	snapRec := &kblcontroller.SnapshotReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}
	domRec := &kblcontroller.DominoReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}

	snapReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: "curve-snap", Namespace: "default"}}
	if _, err := snapRec.Reconcile(context.Background(), snapReq); err != nil {
		t.Fatalf("seal snapshot: %v", err)
	}

	var sealed kblv1alpha1.Snapshot
	if err := cl.Get(context.Background(), snapReq.NamespacedName, &sealed); err != nil {
		t.Fatalf("get sealed snapshot: %v", err)
	}

	loadReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: "load", Namespace: "default"}}
	if _, err := domRec.Reconcile(context.Background(), loadReq); err != nil {
		t.Fatalf("load domino: %v", err)
	}

	interpReq := reconcile.Request{NamespacedName: types.NamespacedName{Name: "interpolate", Namespace: "default"}}
	if _, err := domRec.Reconcile(context.Background(), interpReq); err != nil {
		t.Fatalf("interpolate domino: %v", err)
	}

	var updated kblv1alpha1.Domino
	if err := cl.Get(context.Background(), interpReq.NamespacedName, &updated); err != nil {
		t.Fatalf("get interpolate: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoPhaseCompleted && updated.Status.Phase != kblv1alpha1.DominoPhaseCached {
		t.Fatalf("expected completed domino, got %s: %s", updated.Status.Phase, updated.Status.Message)
	}
	if updated.Status.OutputHash == "" {
		t.Fatal("expected output hash")
	}
	_ = sealed
}

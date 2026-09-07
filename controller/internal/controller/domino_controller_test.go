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
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	enginetypes "github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestDominoReconcilerContinuesReplaySpine(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	storeDir := t.TempDir()
	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "curve-snap", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Source: kblv1alpha1.SnapshotSource{
				Inline: map[string]interface{}{"message": "hello-kbl", "value": 42},
			},
			Sealed: true,
		},
	}
	load := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:identity",
		},
	}
	step := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "step-two", Namespace: "default", Generation: 1},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:identity",
			DependsOn:   []string{"load"},
			Inputs:      []kblv1alpha1.DominoInput{{FromDomino: "load"}},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(snap, load, step).
		WithObjects(snap, load, step).
		Build()

	if _, err := (&kblcontroller.SnapshotReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}).Reconcile(
		context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: snap.Name, Namespace: snap.Namespace}},
	); err != nil {
		t.Fatalf("seal snapshot: %v", err)
	}

	domRec := &kblcontroller.DominoReconciler{Client: cl, Scheme: scheme, StoreRoot: storeDir}
	if _, err := domRec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: load.Name, Namespace: load.Namespace},
	}); err != nil {
		t.Fatalf("load: %v", err)
	}
	if _, err := domRec.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: step.Name, Namespace: step.Namespace},
	}); err != nil {
		t.Fatalf("step-two: %v", err)
	}

	var first, second kblv1alpha1.Domino
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "load", Namespace: "default"}, &first); err != nil {
		t.Fatal(err)
	}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "step-two", Namespace: "default"}, &second); err != nil {
		t.Fatal(err)
	}
	if first.Status.HeadLink == "" || second.Status.HeadLink == "" {
		t.Fatalf("dominos must persist HeadLink: load=%q step=%q", first.Status.HeadLink, second.Status.HeadLink)
	}
	if second.Status.PrevLink != first.Status.HeadLink {
		t.Fatalf("step prevLink %q want load head %q", second.Status.PrevLink, first.Status.HeadLink)
	}

	var sealed kblv1alpha1.Snapshot
	if err := cl.Get(context.Background(), types.NamespacedName{Name: snap.Name, Namespace: snap.Namespace}, &sealed); err != nil {
		t.Fatal(err)
	}
	backend, err := store.Open(filepath.Join(storeDir, snap.Namespace, snap.Name+".db"))
	if err != nil {
		t.Fatal(err)
	}
	defer backend.Close()
	rows, err := backend.ListReplay(sealed.Status.SnapshotID)
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]enginetypes.ReplayLogEntry, len(rows))
	for i, r := range rows {
		entries[i] = enginetypes.ReplayLogEntry{
			SnapshotID: r.SnapshotID,
			DominoID:   r.DominoID,
			InputHash:  r.InputHash,
			OutputHash: r.OutputHash,
			PrevLink:   r.PrevLink,
			Link:       r.Link,
		}
	}
	if err := theory.VerifySpine(sealed.Status.SnapshotID, entries); err != nil {
		t.Fatalf("stepwise CR spine: %v", err)
	}
}

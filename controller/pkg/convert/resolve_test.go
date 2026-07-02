package convert_test

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
)

func TestResolveEngineWorkflowFromRefs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "curve-snap", Namespace: "default"},
		Spec: kblv1alpha1.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Source: kblv1alpha1.SnapshotSource{
				Inline: map[string]interface{}{"value": 1},
			},
			Sealed: true,
		},
		Status: kblv1alpha1.SnapshotStatus{
			Phase:      kblv1alpha1.SnapshotPhaseSealed,
			SnapshotID: "snap-id-123",
		},
	}

	load := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:identity",
		},
	}

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-refs", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			SnapshotRef: "curve-snap",
			DominoRefs:  []string{"load"},
			Execution: kblv1alpha1.ExecutionSpec{
				Chain:         []string{"load"},
				Deterministic: true,
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(snap, load, wf).Build()

	engineWF, err := convert.ResolveEngineWorkflow(context.Background(), cl, wf)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if engineWF.Spec.Snapshot.Status == nil || engineWF.Spec.Snapshot.Status.SnapshotID != "snap-id-123" {
		t.Fatalf("expected snapshot ID from CR status, got %+v", engineWF.Spec.Snapshot.Status)
	}
	if len(engineWF.Spec.Dominos) != 1 || engineWF.Spec.Dominos[0].Metadata.Name != "load" {
		t.Fatalf("expected one domino named load, got %+v", engineWF.Spec.Dominos)
	}
}

func TestResolveEngineWorkflowSnapshotNotReady(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "pending-snap", Namespace: "default"},
		Spec: kblv1alpha1.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Sealed:    true,
		},
		Status: kblv1alpha1.SnapshotStatus{Phase: kblv1alpha1.SnapshotPhasePending},
	}

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf-pending", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			SnapshotRef: "pending-snap",
			DominoRefs:  []string{"load"},
			Execution:   kblv1alpha1.ExecutionSpec{Chain: []string{"load"}},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(snap, wf).Build()
	_, err := convert.ResolveEngineWorkflow(context.Background(), cl, wf)
	if err == nil {
		t.Fatal("expected error for unready snapshot")
	}
	if got := err.Error(); !strings.Contains(got, "is not ready") {
		t.Fatalf("expected not-ready error, got %v", err)
	}
}

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
)

func TestWorkflowReconcilerExecutesChain(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	storeDir := t.TempDir()
	storePath := filepath.Join(storeDir, "workflow.db")

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test-finance",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.WorkflowSpec{
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{
						"instruments": []interface{}{
							map[string]interface{}{
								"instrument_id": "US10Y",
								"rate":          4.25,
								"maturity":      "2035-02-15",
							},
							map[string]interface{}{
								"instrument_id": "US2Y",
								"rate":          4.80,
								"maturity":      "2027-02-15",
							},
						},
					},
				},
				Sealed: true,
			},
			Dominos: []kblv1alpha1.DominoSpec{
				{Name: "load", SnapshotRef: "snap-1", Command: "builtin:identity"},
				{
					Name: "interpolate", SnapshotRef: "snap-1", Command: "builtin:interpolate",
					DependsOn: []string{"load"},
					Inputs:    []kblv1alpha1.DominoInput{{FromDomino: "load"}},
				},
			},
			Execution: kblv1alpha1.ExecutionSpec{
				Chain:         []string{"load", "interpolate"},
				Deterministic: true,
			},
			Provisioning: kblv1alpha1.ProvisioningSpec{
				StorePath: storePath,
				NodeLocal: true,
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wf).WithObjects(wf).Build()

	r := &kblcontroller.WorkflowReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: storeDir,
	}

	// First reconcile adds finalizer
	_, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: wf.Name, Namespace: wf.Namespace},
	})
	if err != nil {
		t.Fatalf("reconcile (finalizer): %v", err)
	}

	// Second reconcile executes
	_, err = r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: wf.Name, Namespace: wf.Namespace},
	})
	if err != nil {
		t.Fatalf("reconcile (execute): %v", err)
	}

	var updated kblv1alpha1.Workflow
	if err := cl.Get(context.Background(), types.NamespacedName{Name: wf.Name, Namespace: wf.Namespace}, &updated); err != nil {
		t.Fatalf("get workflow: %v", err)
	}

	if updated.Status.Phase != kblv1alpha1.WorkflowPhaseCompleted {
		t.Errorf("expected phase Completed, got %s (message: %s)", updated.Status.Phase, updated.Status.Message)
	}
	if updated.Status.SnapshotID == "" {
		t.Error("expected snapshot ID in status")
	}
	if updated.Status.DominoCount != 2 {
		t.Errorf("expected 2 dominos, got %d", updated.Status.DominoCount)
	}
	if updated.Status.RecomputedCount != 2 {
		t.Errorf("first run: expected 2 recomputed, got %d", updated.Status.RecomputedCount)
	}
	if updated.Status.ReplayLogRef == "" {
		t.Error("expected replay log ref in status")
	}

	// Verify ConfigMap created
	cm := &corev1.ConfigMap{}
	if err := cl.Get(context.Background(), types.NamespacedName{
		Name: "test-finance-replay", Namespace: "default",
	}, cm); err != nil {
		t.Fatalf("get replay configmap: %v", err)
	}
	if cm.Data["replay.json"] == "" {
		t.Error("expected replay.json in configmap")
	}
}

func TestWorkflowReconcilerSkipsCompleted(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "done",
			Namespace:         "default",
			Generation:        2,
			Finalizers:        []string{"kbl.io/workflow-finalizer"},
			ResourceVersion:   "1",
		},
		Spec: kblv1alpha1.WorkflowSpec{
			Snapshot: kblv1alpha1.SnapshotSpec{TimeSlice: "2025-01-01", Sealed: true},
			Execution: kblv1alpha1.ExecutionSpec{Chain: []string{"a"}},
		},
		Status: kblv1alpha1.WorkflowStatus{
			ObservedGeneration: 2,
			Phase:              kblv1alpha1.WorkflowPhaseCompleted,
			SnapshotID:         "abc123",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(wf).WithObjects(wf).Build()
	r := &kblcontroller.WorkflowReconciler{Client: cl, Scheme: scheme}

	start := time.Now()
	_, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "done", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if time.Since(start) > time.Second {
		t.Error("expected fast no-op for completed workflow")
	}
}

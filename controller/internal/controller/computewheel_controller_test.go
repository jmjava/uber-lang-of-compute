package controller_test

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	kblcontroller "github.com/jmjava/uber-lang-of-compute/controller/internal/controller"
)

func financeWheelTemplate() kblv1alpha1.WorkflowTemplateSpec {
	return kblv1alpha1.WorkflowTemplateSpec{
		Snapshot: kblv1alpha1.SnapshotSpec{
			Source: kblv1alpha1.SnapshotSource{
				Inline: map[string]interface{}{
					"instruments": []interface{}{
						map[string]interface{}{"instrument_id": "US2Y", "rate": 4.8, "maturity": "2027-02-15"},
						map[string]interface{}{"instrument_id": "US10Y", "rate": 4.25, "maturity": "2035-02-15"},
					},
				},
			},
			Sealed: true,
		},
		Dominos: []kblv1alpha1.DominoSpec{
			{Name: "load", Command: "builtin:identity"},
			{Name: "interpolate", Command: "builtin:interpolate", DependsOn: []string{"load"},
				Inputs: []kblv1alpha1.DominoInput{{FromDomino: "load"}}},
		},
		Execution: kblv1alpha1.ExecutionSpec{
			Chain:         []string{"load", "interpolate"},
			Deterministic: true,
		},
	}
}

func TestComputeWheelReconcilerCreatesWorkflow(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	fixedNow := time.Date(2025, 4, 15, 12, 0, 0, 0, time.UTC)

	wheel := &kblv1alpha1.ComputeWheel{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "finance-wheel",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.ComputeWheelSpec{
			Contexts:            []string{"ctx-a", "ctx-b"},
			TimeSliceInterval:   "24h",
			MaxRotations:        1,
			WorkflowTemplate:    financeWheelTemplate(),
			Schedule:            &kblv1alpha1.WheelScheduleSpec{StartTimeSlice: "2025-04-15T00:00:00Z"},
		},
	}

	ctxA := &kblv1alpha1.ComputeContext{
		ObjectMeta: metav1.ObjectMeta{Name: "ctx-a"},
		Spec:       kblv1alpha1.ComputeContextSpec{NodeName: "node-a", StorePath: "/var/kbl/ctx-a"},
	}
	ctxB := &kblv1alpha1.ComputeContext{
		ObjectMeta: metav1.ObjectMeta{Name: "ctx-b"},
		Spec:       kblv1alpha1.ComputeContextSpec{NodeName: "node-b", StorePath: "/var/kbl/ctx-b"},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(wheel, &kblv1alpha1.Workflow{}).
		WithObjects(wheel, ctxA, ctxB).
		Build()

	r := &kblcontroller.ComputeWheelReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
		Clock:     func() time.Time { return fixedNow },
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "finance-wheel", Namespace: "default"}}

	// finalizer
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile finalizer: %v", err)
	}

	// initialize + create first workflow
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile init: %v", err)
	}

	var wfList kblv1alpha1.WorkflowList
	if err := cl.List(context.Background(), &wfList); err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	if len(wfList.Items) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wfList.Items))
	}
	if wfList.Items[0].Labels["kbl.io/compute-context"] != "ctx-a" {
		t.Errorf("expected first workflow on ctx-a, got %s", wfList.Items[0].Labels["kbl.io/compute-context"])
	}

	// simulate workflow completion
	wfName := wfList.Items[0].Name
	var wf kblv1alpha1.Workflow
	if err := cl.Get(context.Background(), types.NamespacedName{Name: wfName, Namespace: "default"}, &wf); err != nil {
		t.Fatalf("get workflow: %v", err)
	}
	wf.Status.Phase = kblv1alpha1.WorkflowPhaseCompleted
	wf.Status.SnapshotID = "abc123"
	if err := cl.Status().Update(context.Background(), &wf); err != nil {
		t.Fatalf("update workflow status: %v", err)
	}

	// advance to ctx-b and create its workflow
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile advance context: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile create ctx-b workflow: %v", err)
	}

	if err := cl.List(context.Background(), &wfList); err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	if len(wfList.Items) != 2 {
		t.Fatalf("expected 2 workflows after context rotation, got %d", len(wfList.Items))
	}

	// complete second workflow
	if err := cl.List(context.Background(), &wfList); err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	for _, item := range wfList.Items {
		if item.Labels["kbl.io/compute-context"] == "ctx-b" {
			var wf2 kblv1alpha1.Workflow
			if err := cl.Get(context.Background(), types.NamespacedName{Name: item.Name, Namespace: "default"}, &wf2); err != nil {
				t.Fatalf("get wf2: %v", err)
			}
			wf2.Status.Phase = kblv1alpha1.WorkflowPhaseCompleted
			wf2.Status.SnapshotID = "def456"
			if err := cl.Status().Update(context.Background(), &wf2); err != nil {
				t.Fatalf("update wf2 status: %v", err)
			}
		}
	}

	// advance slice -> should hit maxRotations=1 and go Idle
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile advance slice: %v", err)
	}

	var updated kblv1alpha1.ComputeWheel
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get wheel: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.ComputeWheelPhaseIdle {
		t.Errorf("expected Idle after max rotations, got %s", updated.Status.Phase)
	}
	if len(updated.Status.ProcessedSlots) != 2 {
		t.Errorf("expected 2 processed slots, got %d", len(updated.Status.ProcessedSlots))
	}
}

func TestComputeWheelPreProvisionCreatesNextWorkflow(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	wheel := &kblv1alpha1.ComputeWheel{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "preprov-wheel",
			Namespace:  "default",
			Finalizers: []string{"kbl.io/computewheel-finalizer"},
		},
		Spec: kblv1alpha1.ComputeWheelSpec{
			Contexts:          []string{"ctx-a", "ctx-b"},
			TimeSliceInterval: "24h",
			PreProvisionNext:  true,
			WorkflowTemplate:  financeWheelTemplate(),
			Schedule:          &kblv1alpha1.WheelScheduleSpec{StartTimeSlice: "2025-04-15T00:00:00Z"},
		},
		Status: kblv1alpha1.ComputeWheelStatus{
			CurrentTimeSlice:   "2025-04-15T00:00:00Z",
			ActiveContextIndex: 0,
			ActiveContext:      "ctx-a",
			Phase:              kblv1alpha1.ComputeWheelPhaseProcessing,
		},
	}

	activeWF := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "preprov-wheel-ctx-a-20250415t000000z",
			Namespace: "default",
			Labels:    map[string]string{"kbl.io/compute-context": "ctx-a"},
		},
		Spec: kblv1alpha1.WorkflowSpec{
			Snapshot:  financeWheelTemplate().Snapshot,
			Dominos:   financeWheelTemplate().Dominos,
			Execution: financeWheelTemplate().Execution,
		},
		Status: kblv1alpha1.WorkflowStatus{Phase: kblv1alpha1.WorkflowPhaseRunning},
	}
	activeWF.Status.Phase = kblv1alpha1.WorkflowPhaseRunning

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(wheel, activeWF).
		WithObjects(wheel, activeWF).
		Build()

	r := &kblcontroller.ComputeWheelReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: "preprov-wheel", Namespace: "default"},
	}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	var wfList kblv1alpha1.WorkflowList
	if err := cl.List(context.Background(), &wfList); err != nil {
		t.Fatalf("list workflows: %v", err)
	}
	if len(wfList.Items) < 2 {
		t.Fatalf("expected pre-provisioned next workflow, got %d workflows", len(wfList.Items))
	}

	foundNext := false
	for _, wf := range wfList.Items {
		if wf.Labels["kbl.io/compute-context"] == "ctx-b" {
			foundNext = true
		}
	}
	if !foundNext {
		t.Error("expected pre-provisioned workflow for ctx-b")
	}
}

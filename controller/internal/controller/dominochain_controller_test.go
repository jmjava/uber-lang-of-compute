package controller_test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	kblcontroller "github.com/jmjava/uber-lang-of-compute/controller/internal/controller"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
)

func TestDominoChainInitChainCompletes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "demo-chain",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeKubernetesInit,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{"value": 42},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "step-one", Command: "builtin:identity"},
				{Name: "step-two", Command: "builtin:identity"},
			},
			StorePath: t.TempDir() + "/chain.db",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "demo-chain", Namespace: "default"}}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create resources reconcile: %v", err)
	}

	var pod corev1.Pod
	if err := cl.Get(context.Background(), types.NamespacedName{
		Name: "demo-chain-chain", Namespace: "default",
	}, &pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	pod.Status.Phase = corev1.PodRunning
	pod.Status.InitContainerStatuses = []corev1.ContainerStatus{
		{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
		{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
	}
	if err := cl.Status().Update(context.Background(), &pod); err != nil {
		t.Fatalf("update pod status: %v", err)
	}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("complete reconcile: %v", err)
	}

	var updated kblv1alpha1.DominoChain
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get chain: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoChainPhaseCompleted {
		t.Errorf("expected Completed, got %s (%s)", updated.Status.Phase, updated.Status.Message)
	}
	if updated.Status.SnapshotID == "" {
		t.Error("expected snapshot ID after completion")
	}
}

func TestDominoChainJuliaInitChainPodSpec(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "julia-chain",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeKubernetesInit,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{
						"instruments": []interface{}{
							map[string]interface{}{"instrument_id": "US10Y", "rate": 4.25, "maturity": "2035-02-15"},
						},
					},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "julia:identity"},
				{Name: "interpolate", Command: "julia:interpolate"},
			},
			StorePath:   t.TempDir() + "/julia-chain.db",
			RunnerImage: "ghcr.io/jmjava/kbl-domino-runner-julia:latest",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "julia-chain", Namespace: "default"}}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create resources reconcile: %v", err)
	}

	var pod corev1.Pod
	if err := cl.Get(context.Background(), types.NamespacedName{
		Name: "julia-chain-chain", Namespace: "default",
	}, &pod); err != nil {
		t.Fatalf("get pod: %v", err)
	}

	if pod.Spec.InitContainers[0].Image != "ghcr.io/jmjava/kbl-domino-runner-julia:latest" {
		t.Fatalf("expected julia runner image, got %s", pod.Spec.InitContainers[0].Image)
	}
	juliaProjectSet := false
	for _, e := range pod.Spec.InitContainers[0].Env {
		if e.Name == "KBL_JULIA_PROJECT" && e.Value == "/opt/kbl/julia" {
			juliaProjectSet = true
		}
	}
	if !juliaProjectSet {
		t.Fatal("expected KBL_JULIA_PROJECT on julia init container")
	}
}

func TestDominoChainVolcanoCompletes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "volcano-chain",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeVolcanoInit,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{"value": 42},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "step-one", Command: "builtin:identity"},
			},
			StorePath: t.TempDir() + "/volcano-chain.db",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "volcano-chain", Namespace: "default"}}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create resources reconcile: %v", err)
	}

	var job unstructured.Unstructured
	job.SetGroupVersionKind(dominochain.VolcanoJobGVK)
	if err := cl.Get(context.Background(), types.NamespacedName{
		Name: "volcano-chain-chain", Namespace: "default",
	}, &job); err != nil {
		t.Fatalf("get volcano job: %v", err)
	}

	_ = unstructured.SetNestedField(job.Object, map[string]interface{}{
		"phase": "Completed",
	}, "status", "state")
	if err := cl.Update(context.Background(), &job); err != nil {
		t.Fatalf("update volcano job status: %v", err)
	}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("complete reconcile: %v", err)
	}

	var updated kblv1alpha1.DominoChain
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get chain: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoChainPhaseCompleted {
		t.Errorf("expected Completed, got %s (%s)", updated.Status.Phase, updated.Status.Message)
	}
}

func TestDominoChainVolcanoCompletesWhenInitContainersFinish(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "volcano-init-done",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeVolcanoInit,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{"value": 42},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "step-one", Command: "builtin:identity"},
			},
			StorePath: t.TempDir() + "/volcano-init-done.db",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "volcano-init-done", Namespace: "default"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create resources reconcile: %v", err)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "volcano-init-done-chain-domino-chain-0",
			Namespace: "default",
			Labels: map[string]string{
				dominochain.LabelDominoChain: "volcano-init-done",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			InitContainerStatuses: []corev1.ContainerStatus{
				{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
			},
		},
	}
	if err := cl.Create(context.Background(), pod); err != nil {
		t.Fatalf("create volcano task pod: %v", err)
	}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("complete reconcile: %v", err)
	}

	var updated kblv1alpha1.DominoChain
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get chain: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoChainPhaseCompleted {
		t.Errorf("expected Completed, got %s (%s)", updated.Status.Phase, updated.Status.Message)
	}
}

func TestDominoChainJuliaVolcanoCompletesWithoutOperatorReplay(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "julia-volcano-chain",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeVolcanoInit,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{"value": 42},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "julia:identity"},
			},
			RunnerImage: "kbl-domino-runner-julia:lab",
			StorePath:   t.TempDir() + "/julia-volcano.db",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "julia-volcano-chain", Namespace: "default"}}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create resources reconcile: %v", err)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "julia-volcano-chain-chain-domino-chain-0",
			Namespace: "default",
			Labels: map[string]string{
				dominochain.LabelDominoChain: "julia-volcano-chain",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			InitContainerStatuses: []corev1.ContainerStatus{
				{State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
			},
		},
	}
	if err := cl.Create(context.Background(), pod); err != nil {
		t.Fatalf("create volcano task pod: %v", err)
	}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("complete reconcile: %v", err)
	}

	var updated kblv1alpha1.DominoChain
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get chain: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoChainPhaseCompleted {
		t.Errorf("expected Completed, got %s (%s)", updated.Status.Phase, updated.Status.Message)
	}
	if updated.Status.SnapshotID == "" {
		t.Error("expected snapshot ID after in-cluster julia completion")
	}
}

func TestDominoChainOpenKruiseCompletes(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kblv1alpha1.AddToScheme(scheme)

	chain := &kblv1alpha1.DominoChain{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "openkruise-chain",
			Namespace:  "default",
			Generation: 1,
		},
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeOpenKruise,
			Snapshot: kblv1alpha1.SnapshotSpec{
				TimeSlice: "2025-04-15T00:00:00Z",
				Source: kblv1alpha1.SnapshotSource{
					Inline: map[string]interface{}{"value": 42},
				},
				Sealed: true,
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "step-one", Command: "builtin:identity"},
				{Name: "step-two", Command: "builtin:identity"},
			},
			StorePath: t.TempDir() + "/openkruise-chain.db",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(chain, &corev1.Pod{}, &corev1.ConfigMap{}).
		WithObjects(chain).
		Build()

	r := &kblcontroller.DominoChainReconciler{
		Client:    cl,
		Scheme:    scheme,
		StoreRoot: t.TempDir(),
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: "openkruise-chain", Namespace: "default"}}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("finalizer reconcile: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create ImagePullJob reconcile: %v", err)
	}

	pull := &unstructured.Unstructured{}
	pull.SetGroupVersionKind(dominochain.ImagePullJobGVK)
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "openkruise-chain-pull", Namespace: "default"}, pull); err != nil {
		t.Fatalf("get ImagePullJob: %v", err)
	}
	var tooEarly corev1.Pod
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "openkruise-chain-chain", Namespace: "default"}, &tooEarly); err == nil {
		t.Fatal("runner-slot pod must wait for ImagePullJob")
	}
	if err := unstructured.SetNestedField(pull.Object, int64(1), "status", "succeeded"); err != nil {
		t.Fatalf("set succeeded: %v", err)
	}
	if err := unstructured.SetNestedField(pull.Object, int64(1), "status", "desired"); err != nil {
		t.Fatalf("set desired: %v", err)
	}
	if err := unstructured.SetNestedField(pull.Object, int64(0), "status", "failed"); err != nil {
		t.Fatalf("set failed: %v", err)
	}
	if err := cl.Update(context.Background(), pull); err != nil {
		t.Fatalf("update ImagePullJob: %v", err)
	}

	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("create pod reconcile: %v", err)
	}

	var pod corev1.Pod
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "openkruise-chain-chain", Namespace: "default"}, &pod); err != nil {
		t.Fatalf("get openkruise pod: %v", err)
	}
	pod.Status.Phase = corev1.PodSucceeded
	pod.Status.ContainerStatuses = []corev1.ContainerStatus{
		{Name: "slot-0-step-one", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
		{Name: "slot-1-step-two", State: corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{ExitCode: 0}}},
	}
	if err := cl.Status().Update(context.Background(), &pod); err != nil {
		t.Fatalf("update pod status: %v", err)
	}
	if _, err := r.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("complete reconcile: %v", err)
	}

	var updated kblv1alpha1.DominoChain
	if err := cl.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("get chain: %v", err)
	}
	if updated.Status.Phase != kblv1alpha1.DominoChainPhaseCompleted {
		t.Errorf("expected Completed, got %s (%s)", updated.Status.Phase, updated.Status.Message)
	}
}

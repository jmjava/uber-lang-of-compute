package controller_test

import (
	"context"
	"testing"

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

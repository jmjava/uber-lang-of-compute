package dominochain_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
)

func TestBuildInitChainPod(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeKubernetesInit,
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "builtin:identity"},
				{Name: "transform", Command: "builtin:identity"},
			},
		},
	}
	chain.Name = "test-chain"
	chain.Namespace = "default"

	b := &dominochain.Builder{}
	pod := b.BuildInitChainPod(chain)

	if len(pod.Spec.InitContainers) != 2 {
		t.Fatalf("expected 2 init containers, got %d", len(pod.Spec.InitContainers))
	}
	if pod.Spec.InitContainers[0].Env[0].Value != "builtin:identity" {
		t.Errorf("expected KBL_COMMAND env on first init container")
	}
	if len(pod.Spec.Volumes) != 2 {
		t.Fatalf("expected handoff + snapshot volumes")
	}
}

func TestBuildOpenKruisePodPlaceholders(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeOpenKruise,
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "a", Command: "builtin:identity"},
				{Name: "b", Command: "builtin:identity"},
			},
		},
	}
	chain.Name = "swap-chain"
	chain.Namespace = "default"

	pod := (&dominochain.Builder{}).BuildOpenKruisePod(chain)
	if len(pod.Spec.Containers) != 2 {
		t.Fatalf("expected 2 placeholder containers, got %d", len(pod.Spec.Containers))
	}
	for _, c := range pod.Spec.Containers {
		if c.Image != dominochain.PlaceholderImage {
			t.Errorf("expected placeholder image, got %s", c.Image)
		}
	}
}

func TestNeedsContainerRuntime(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)
	ctx := context.Background()

	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wf", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			Execution: kblv1alpha1.ExecutionSpec{Runtime: "openkruise"},
		},
	}
	if !dominochain.NeedsContainerRuntime(ctx, nil, wf) {
		t.Error("expected container runtime for openkruise")
	}

	dom := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "custom", Namespace: "default"},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "snap",
			Command:     "builtin:identity",
			Image:       "ghcr.io/example/runner:v1",
		},
	}
	wfRefs := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "refs", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			SnapshotRef: "snap",
			DominoRefs:  []string{"custom"},
			Execution:   kblv1alpha1.ExecutionSpec{Chain: []string{"custom"}},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(dom).Build()
	if !dominochain.NeedsContainerRuntime(ctx, cl, wfRefs) {
		t.Error("expected container runtime when domino ref has image")
	}
}

func TestFromWorkflowWithCRRefs(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)
	ctx := context.Background()

	snap := &kblv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{Name: "curve-snap", Namespace: "default"},
		Spec: kblv1alpha1.SnapshotSpec{
			TimeSlice: "2025-04-15T00:00:00Z",
			Source: kblv1alpha1.SnapshotSource{
				Inline: map[string]interface{}{"value": 42},
			},
			Sealed: true,
		},
		Status: kblv1alpha1.SnapshotStatus{
			Phase:      kblv1alpha1.SnapshotPhaseSealed,
			SnapshotID: "snap-abc",
		},
	}
	load := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "load", Namespace: "default"},
		Spec: kblv1alpha1.DominoResourceSpec{
			SnapshotRef: "curve-snap",
			Command:     "builtin:identity",
			Image:       "ghcr.io/example/runner:v1",
		},
	}
	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "container-refs", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			SnapshotRef: "curve-snap",
			DominoRefs:  []string{"load"},
			Execution: kblv1alpha1.ExecutionSpec{
				Chain:   []string{"load"},
				Runtime: "kubernetes-init",
			},
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(snap, load, wf).Build()
	chain, err := dominochain.FromWorkflow(ctx, cl, wf, "/var/kbl/store/wf.db")
	if err != nil {
		t.Fatalf("FromWorkflow: %v", err)
	}
	if !chain.Spec.Snapshot.Sealed {
		t.Fatal("expected sealed snapshot on chain")
	}
	if chain.Spec.Snapshot.Source.Inline["value"].(float64) != 42 {
		t.Fatalf("expected inline snapshot data from CR, got %+v", chain.Spec.Snapshot.Source.Inline)
	}
	if len(chain.Spec.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(chain.Spec.Steps))
	}
	if chain.Spec.Steps[0].Name != "load" || chain.Spec.Steps[0].Image != "ghcr.io/example/runner:v1" {
		t.Fatalf("unexpected step: %+v", chain.Spec.Steps[0])
	}
	if chain.Spec.Runtime != kblv1alpha1.DominoChainRuntimeKubernetesInit {
		t.Fatalf("expected kubernetes-init runtime, got %s", chain.Spec.Runtime)
	}
}

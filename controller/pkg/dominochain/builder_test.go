package dominochain_test

import (
	"testing"

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
	wf := &kblv1alpha1.Workflow{
		Spec: kblv1alpha1.WorkflowSpec{
			Execution: kblv1alpha1.ExecutionSpec{Runtime: "openkruise"},
		},
	}
	if !dominochain.NeedsContainerRuntime(wf) {
		t.Error("expected container runtime for openkruise")
	}
}

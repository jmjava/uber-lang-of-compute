package dominochain_test

import (
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestContainerRecreateRequestJuliaEnv(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeOpenKruise,
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "interp", Command: "julia:interpolate"},
			},
			RunnerImage: dominochain.DefaultJuliaRunnerImage,
		},
	}
	chain.Name = "julia-openkruise"
	chain.Namespace = "default"

	crr := dominochain.ContainerRecreateRequest(chain, &dominochain.Builder{}, 0)
	containers, found, err := unstructured.NestedSlice(crr.Object, "spec", "containers")
	if err != nil || !found || len(containers) == 0 {
		t.Fatalf("expected CRR containers: found=%v err=%v", found, err)
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected container map")
	}
	if container["name"] == "" {
		t.Fatal("expected container name")
	}
	podName, _, _ := unstructured.NestedString(crr.Object, "spec", "podName")
	if podName != "julia-openkruise-chain" {
		t.Fatalf("expected podName julia-openkruise-chain, got %q", podName)
	}
}

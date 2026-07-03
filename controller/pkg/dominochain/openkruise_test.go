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
	containers, found, err := unstructured.NestedSlice(crr.Object, "spec", "template", "containers")
	if err != nil || !found || len(containers) == 0 {
		t.Fatalf("expected CRR container template: found=%v err=%v", found, err)
	}

	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected container map")
	}
	envSlice, ok := container["env"].([]interface{})
	if !ok {
		t.Fatal("expected env slice")
	}

	names := map[string]string{}
	for _, raw := range envSlice {
		m, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		names[m["name"].(string)] = m["value"].(string)
	}
	if names["KBL_COMMAND"] != "julia:interpolate" {
		t.Fatalf("unexpected command: %v", names)
	}
	if names["KBL_JULIA_PROJECT"] != dominochain.JuliaProjectContainerPath {
		t.Fatalf("expected julia project env, got %v", names)
	}
}

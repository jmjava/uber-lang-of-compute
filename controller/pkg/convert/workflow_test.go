package convert_test

import (
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
)

func TestConvertPassesSandboxIntoEngineGate(t *testing.T) {
	wf := &kblv1alpha1.Workflow{
		Spec: kblv1alpha1.WorkflowSpec{
			Dominos: []kblv1alpha1.DominoSpec{{Name: "a", Command: "sandbox:identity"}},
			Execution: kblv1alpha1.ExecutionSpec{
				Chain:         []string{"a"},
				Deterministic: true,
			},
			Provisioning: kblv1alpha1.ProvisioningSpec{
				SandboxNetworkNone:  true,
				SandboxReadOnlyRoot: true,
			},
		},
	}
	eng := convert.ToEngineWorkflow(wf)
	if !eng.Spec.Provisioning.SandboxNetworkNone || !eng.Spec.Provisioning.SandboxReadOnlyRoot {
		t.Fatalf("sandbox flags did not reach engine provisioning: %+v", eng.Spec.Provisioning)
	}
}

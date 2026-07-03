package dominochain_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestFromEngineWorkflowVolcanoProvisioning(t *testing.T) {
	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "wheel-wf", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			Execution: kblv1alpha1.ExecutionSpec{
				Runtime:      string(kblv1alpha1.DominoChainRuntimeVolcanoInit),
				VolcanoQueue: "kbl-lab",
				Chain:        []string{"load"},
			},
			Provisioning: kblv1alpha1.ProvisioningSpec{
				RunnerImage:  "kbl-domino-runner-julia:lab",
				NodeSelector: map[string]string{"kbl.io/lab-role": "compute"},
			},
		},
	}
	engineWF := &types.Workflow{
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				Spec: types.SnapshotSpec{
					Source: types.SnapshotSource{Inline: map[string]interface{}{"value": 1}},
					Sealed: true,
				},
			},
			Dominos: []types.Domino{{
				Metadata: types.ObjectMeta{Name: "load"},
				Spec:     types.DominoSpec{Command: "julia:identity"},
			}},
		},
	}

	chain := dominochain.FromEngineWorkflow(engineWF, wf, "/var/kbl/store/w.db")
	if chain.Spec.Runtime != kblv1alpha1.DominoChainRuntimeVolcanoInit {
		t.Fatalf("expected volcano-init runtime, got %s", chain.Spec.Runtime)
	}
	if chain.Spec.VolcanoQueue != "kbl-lab" {
		t.Fatalf("expected volcano queue kbl-lab, got %q", chain.Spec.VolcanoQueue)
	}
	if chain.Spec.RunnerImage != "kbl-domino-runner-julia:lab" {
		t.Fatalf("expected runner image, got %q", chain.Spec.RunnerImage)
	}
	if chain.Spec.NodeSelector["kbl.io/lab-role"] != "compute" {
		t.Fatalf("expected node selector, got %+v", chain.Spec.NodeSelector)
	}
}

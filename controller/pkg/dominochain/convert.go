package dominochain

import (
	"encoding/json"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

// FromWorkflow builds a DominoChain spec from a Workflow with container runtime.
func FromWorkflow(wf *kblv1alpha1.Workflow, storePath string) *kblv1alpha1.DominoChain {
	runtime := kblv1alpha1.DominoChainRuntime(wf.Spec.Execution.Runtime)
	if runtime == "" {
		runtime = kblv1alpha1.DominoChainRuntimeKubernetesInit
	}

	steps := make([]kblv1alpha1.DominoStepSpec, len(wf.Spec.Dominos))
	for i, d := range wf.Spec.Dominos {
		steps[i] = kblv1alpha1.DominoStepSpec{
			Name:    d.Name,
			Image:   d.Image,
			Command: d.Command,
		}
	}

	return &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Snapshot:    wf.Spec.Snapshot,
			Steps:       steps,
			Runtime:     runtime,
			StorePath:   storePath,
			RunnerImage: "",
		},
	}
}

// SnapshotJSON marshals inline snapshot data for the chain ConfigMap.
func SnapshotJSON(snap kblv1alpha1.SnapshotSpec) (string, error) {
	data, err := json.Marshal(snap.Source.Inline)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NeedsContainerRuntime reports whether a workflow should run via DominoChain.
func NeedsContainerRuntime(wf *kblv1alpha1.Workflow) bool {
	if wf.Spec.Execution.Runtime != "" && wf.Spec.Execution.Runtime != string(kblv1alpha1.DominoChainRuntimeLocal) {
		return true
	}
	for _, d := range wf.Spec.Dominos {
		if d.Image != "" {
			return true
		}
	}
	return false
}

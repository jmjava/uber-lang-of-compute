package dominochain

import (
	"context"
	"encoding/json"

	"sigs.k8s.io/controller-runtime/pkg/client"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// FromWorkflow builds a DominoChain from a Workflow, resolving Snapshot/Domino CR references.
func FromWorkflow(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow, storePath string) (*kblv1alpha1.DominoChain, error) {
	engineWF, err := convert.ResolveEngineWorkflow(ctx, c, wf)
	if err != nil {
		return nil, err
	}
	return FromEngineWorkflow(engineWF, wf, storePath), nil
}

// FromEngineWorkflow builds a DominoChain spec from a resolved engine workflow.
func FromEngineWorkflow(engineWF *types.Workflow, wf *kblv1alpha1.Workflow, storePath string) *kblv1alpha1.DominoChain {
	runtime := kblv1alpha1.DominoChainRuntime(wf.Spec.Execution.Runtime)
	if runtime == "" {
		runtime = kblv1alpha1.DominoChainRuntimeKubernetesInit
	}

	steps := make([]kblv1alpha1.DominoStepSpec, len(engineWF.Spec.Dominos))
	for i, d := range engineWF.Spec.Dominos {
		steps[i] = kblv1alpha1.DominoStepSpec{
			Name:    d.Metadata.Name,
			Image:   d.Spec.Image,
			Command: d.Spec.Command,
		}
	}

	return &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Snapshot:     convert.ToCRSnapshotSpec(engineWF.Spec.Snapshot),
			Steps:        steps,
			Runtime:      runtime,
			StorePath:    storePath,
			RunnerImage:  wf.Spec.Provisioning.RunnerImage,
			NodeSelector: wf.Spec.Provisioning.NodeSelector,
			VolcanoQueue: wf.Spec.Execution.VolcanoQueue,
		},
	}
}

// SnapshotJSON marshals resolved snapshot content for the chain ConfigMap.
func SnapshotJSON(snap kblv1alpha1.SnapshotSpec) (string, error) {
	content, err := snapshot.ResolveContent(snap)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(content)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// NeedsContainerRuntime reports whether a workflow should run via DominoChain.
func NeedsContainerRuntime(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) bool {
	if wf.Spec.Execution.Runtime != "" && wf.Spec.Execution.Runtime != string(kblv1alpha1.DominoChainRuntimeLocal) {
		return true
	}
	for _, d := range wf.Spec.Dominos {
		if d.Image != "" {
			return true
		}
	}
	if c != nil && len(wf.Spec.DominoRefs) > 0 {
		for _, name := range wf.Spec.DominoRefs {
			var dom kblv1alpha1.Domino
			if err := c.Get(ctx, client.ObjectKey{Namespace: wf.Namespace, Name: name}, &dom); err == nil && dom.Spec.Image != "" {
				return true
			}
		}
	}
	return false
}

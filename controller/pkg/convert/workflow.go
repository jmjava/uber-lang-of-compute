package convert

import (
	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ToEngineWorkflow converts a Kubernetes Workflow CR to the engine domain model.
func ToEngineWorkflow(wf *kblv1alpha1.Workflow) *types.Workflow {
	snapshotName := wf.Name + "-snapshot"
	if len(wf.Spec.Dominos) > 0 && wf.Spec.Dominos[0].SnapshotRef != "" {
		snapshotName = wf.Spec.Dominos[0].SnapshotRef
	}

	dominos := make([]types.Domino, len(wf.Spec.Dominos))
	for i, d := range wf.Spec.Dominos {
		name := d.Name
		if name == "" && i < len(wf.Spec.Execution.Chain) {
			name = wf.Spec.Execution.Chain[i]
		}
		dominos[i] = types.Domino{
			APIVersion: "kbl.io/v1alpha1",
			Kind:       "Domino",
			Metadata:   types.ObjectMeta{Name: name},
			Spec: types.DominoSpec{
				SnapshotRef: d.SnapshotRef,
				Command:     d.Command,
				DependsOn:   d.DependsOn,
				Inputs:      toDominoInputs(d.Inputs),
				Image:       d.Image,
			},
		}
	}

	return &types.Workflow{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "Workflow",
		Metadata:   types.ObjectMeta{Name: wf.Name, Labels: wf.Labels},
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				APIVersion: "kbl.io/v1alpha1",
				Kind:       "Snapshot",
				Metadata:   types.ObjectMeta{Name: snapshotName},
				Spec: types.SnapshotSpec{
					TimeSlice:         wf.Spec.Snapshot.TimeSlice,
					Source:            toSnapshotSource(wf.Spec.Snapshot.Source),
					ComputeContextRef: wf.Spec.Snapshot.ComputeContextRef,
					Sealed:            wf.Spec.Snapshot.Sealed,
				},
			},
			Dominos: dominos,
			Execution: types.ExecutionConfig{
				Chain:         wf.Spec.Execution.Chain,
				Deterministic: wf.Spec.Execution.Deterministic,
			},
			Provisioning: types.ProvisioningConfig{
				StorePath: wf.Spec.Provisioning.StorePath,
				NodeLocal: wf.Spec.Provisioning.NodeLocal,
			},
			Routing: types.RoutingConfig{
				Universe:          wf.Spec.Routing.Universe,
				ComputeContextRef: wf.Spec.Routing.ComputeContextRef,
			},
		},
	}
}

func toSnapshotSource(s kblv1alpha1.SnapshotSource) types.SnapshotSource {
	return types.SnapshotSource{
		Inline: s.Inline,
		Path:   s.Path,
		URI:    s.URI,
	}
}

func toDominoInputs(inputs []kblv1alpha1.DominoInput) []types.DominoInput {
	out := make([]types.DominoInput, len(inputs))
	for i, in := range inputs {
		out[i] = types.DominoInput{
			FromDomino:   in.FromDomino,
			FromSnapshot: in.FromSnapshot,
		}
	}
	return out
}

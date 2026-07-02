package convert

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ResolveEngineWorkflow builds an engine workflow, resolving Snapshot and Domino CR references.
func ResolveEngineWorkflow(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) (*types.Workflow, error) {
	snap, err := resolveSnapshot(ctx, c, wf)
	if err != nil {
		return nil, err
	}

	dominos, chain, err := resolveDominos(ctx, c, wf)
	if err != nil {
		return nil, err
	}

	execChain := wf.Spec.Execution.Chain
	if len(execChain) == 0 {
		execChain = chain
	}
	if len(execChain) == 0 {
		return nil, fmt.Errorf("execution chain is empty")
	}

	return &types.Workflow{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "Workflow",
		Metadata:   types.ObjectMeta{Name: wf.Name, Labels: wf.Labels},
		Spec: types.WorkflowSpec{
			Snapshot: snap,
			Dominos:  dominos,
			Execution: types.ExecutionConfig{
				Chain:         execChain,
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
	}, nil
}

func resolveSnapshot(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) (types.Snapshot, error) {
	if wf.Spec.SnapshotRef != "" {
		var snapCR kblv1alpha1.Snapshot
		if err := c.Get(ctx, client.ObjectKey{Namespace: wf.Namespace, Name: wf.Spec.SnapshotRef}, &snapCR); err != nil {
			return types.Snapshot{}, fmt.Errorf("snapshot ref %q: %w", wf.Spec.SnapshotRef, err)
		}
		if !snapCR.Spec.Sealed {
			return types.Snapshot{}, fmt.Errorf("snapshot %q is not sealed", snapCR.Name)
		}
		if snapCR.Status.Phase != kblv1alpha1.SnapshotPhaseSealed || snapCR.Status.SnapshotID == "" {
			return types.Snapshot{}, fmt.Errorf("snapshot %q is not ready (phase=%s)", snapCR.Name, snapCR.Status.Phase)
		}
		engineSnap := ToEngineSnapshot(&snapCR)
		engineSnap.Status = &types.SnapshotStatus{
			Phase:      string(snapCR.Status.Phase),
			SnapshotID: snapCR.Status.SnapshotID,
		}
		if snapCR.Status.SealedAt != nil {
			engineSnap.Status.SealedAt = snapCR.Status.SealedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		return engineSnap, nil
	}

	if err := snapshot.Validate(wf.Spec.Snapshot); err != nil {
		return types.Snapshot{}, fmt.Errorf("inline snapshot: %w", err)
	}
	if !wf.Spec.Snapshot.Sealed {
		return types.Snapshot{}, fmt.Errorf("inline snapshot is not sealed")
	}

	snapshotName := wf.Name + "-snapshot"
	if len(wf.Spec.Dominos) > 0 && wf.Spec.Dominos[0].SnapshotRef != "" {
		snapshotName = wf.Spec.Dominos[0].SnapshotRef
	}

	snapshotID, err := snapshot.ComputeID(wf.Spec.Snapshot)
	if err != nil {
		return types.Snapshot{}, err
	}

	return types.Snapshot{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "Snapshot",
		Metadata:   types.ObjectMeta{Name: snapshotName},
		Spec: types.SnapshotSpec{
			TimeSlice:         wf.Spec.Snapshot.TimeSlice,
			Source:            toSnapshotSource(wf.Spec.Snapshot.Source),
			ComputeContextRef: wf.Spec.Snapshot.ComputeContextRef,
			Sealed:            wf.Spec.Snapshot.Sealed,
		},
		Status: &types.SnapshotStatus{SnapshotID: snapshotID},
	}, nil
}

func resolveDominos(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) ([]types.Domino, []string, error) {
	if len(wf.Spec.DominoRefs) > 0 {
		dominos := make([]types.Domino, 0, len(wf.Spec.DominoRefs))
		for _, name := range wf.Spec.DominoRefs {
			var domCR kblv1alpha1.Domino
			if err := c.Get(ctx, client.ObjectKey{Namespace: wf.Namespace, Name: name}, &domCR); err != nil {
				return nil, nil, fmt.Errorf("domino ref %q: %w", name, err)
			}
			if wf.Spec.SnapshotRef != "" && domCR.Spec.SnapshotRef != wf.Spec.SnapshotRef {
				return nil, nil, fmt.Errorf("domino %q references snapshot %q, workflow expects %q",
					name, domCR.Spec.SnapshotRef, wf.Spec.SnapshotRef)
			}
			dominos = append(dominos, ToEngineDomino(&domCR))
		}
		return dominos, append([]string(nil), wf.Spec.DominoRefs...), nil
	}

	if len(wf.Spec.Dominos) == 0 {
		return nil, nil, fmt.Errorf("dominos or dominoRefs is required")
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
	return dominos, wf.Spec.Execution.Chain, nil
}

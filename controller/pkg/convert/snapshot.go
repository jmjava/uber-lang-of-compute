package convert

import (
	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ToEngineSnapshot converts a Snapshot CR to the engine domain model.
func ToEngineSnapshot(snap *kblv1alpha1.Snapshot) types.Snapshot {
	return types.Snapshot{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "Snapshot",
		Metadata:   types.ObjectMeta{Name: snap.Name},
		Spec: types.SnapshotSpec{
			TimeSlice:         snap.Spec.TimeSlice,
			Source:            toSnapshotSource(snap.Spec.Source),
			ComputeContextRef: snap.Spec.ComputeContextRef,
			Sealed:            snap.Spec.Sealed,
		},
	}
}

// ToEngineDomino converts a Domino CR to the engine domain model.
func ToEngineDomino(d *kblv1alpha1.Domino) types.Domino {
	return types.Domino{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "Domino",
		Metadata:   types.ObjectMeta{Name: d.Name},
		Spec: types.DominoSpec{
			SnapshotRef: d.Spec.SnapshotRef,
			Command:     d.Spec.Command,
			DependsOn:   d.Spec.DependsOn,
			Inputs:      toDominoInputs(d.Spec.Inputs),
			Image:       d.Spec.Image,
		},
	}
}

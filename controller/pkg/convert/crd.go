package convert

import (
	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ToCRSnapshotSpec converts a resolved engine snapshot to a Kubernetes SnapshotSpec.
func ToCRSnapshotSpec(s types.Snapshot) kblv1alpha1.SnapshotSpec {
	return kblv1alpha1.SnapshotSpec{
		TimeSlice: s.Spec.TimeSlice,
		Source: kblv1alpha1.SnapshotSource{
			Inline: s.Spec.Source.Inline,
			Path:   s.Spec.Source.Path,
			URI:    s.Spec.Source.URI,
		},
		ComputeContextRef: s.Spec.ComputeContextRef,
		Sealed:            s.Spec.Sealed,
	}
}

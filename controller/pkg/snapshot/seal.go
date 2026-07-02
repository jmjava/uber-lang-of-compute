package snapshot

import (
	"fmt"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

// Validate checks that a snapshot spec has enough source material to seal.
func Validate(spec kblv1alpha1.SnapshotSpec) error {
	if spec.TimeSlice == "" {
		return fmt.Errorf("timeSlice is required")
	}
	if spec.Source.Inline == nil && spec.Source.Path == "" && spec.Source.URI == "" {
		return fmt.Errorf("snapshot source requires inline, path, or uri")
	}
	return nil
}

// ContentData returns the data payload used for snapshot ID hashing.
func ContentData(spec kblv1alpha1.SnapshotSpec) (interface{}, error) {
	return ResolveContent(spec)
}

// ComputeID returns a deterministic snapshot ID for the spec.
func ComputeID(spec kblv1alpha1.SnapshotSpec) (string, error) {
	id, _, err := SealPayload(spec)
	return id, err
}

// MarshalData serializes resolved snapshot content for store persistence.
func MarshalData(spec kblv1alpha1.SnapshotSpec) (string, error) {
	_, data, err := SealPayload(spec)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

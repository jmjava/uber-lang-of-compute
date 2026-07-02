package snapshot

import (
	"encoding/json"
	"fmt"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
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
func ContentData(spec kblv1alpha1.SnapshotSpec) interface{} {
	if spec.Source.Inline != nil {
		return spec.Source.Inline
	}
	if spec.Source.Path != "" {
		return map[string]string{"path": spec.Source.Path}
	}
	return map[string]string{"uri": spec.Source.URI}
}

// ComputeID returns a deterministic snapshot ID for the spec.
func ComputeID(spec kblv1alpha1.SnapshotSpec) (string, error) {
	if err := Validate(spec); err != nil {
		return "", err
	}
	return hash.SnapshotID(spec.TimeSlice, ContentData(spec))
}

// MarshalData serializes snapshot source for store persistence.
func MarshalData(spec kblv1alpha1.SnapshotSpec) (string, error) {
	data, err := json.Marshal(ContentData(spec))
	if err != nil {
		return "", fmt.Errorf("marshal snapshot data: %w", err)
	}
	return string(data), nil
}

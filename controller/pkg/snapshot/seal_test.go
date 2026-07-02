package snapshot_test

import (
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
)

func TestComputeIDDeterministic(t *testing.T) {
	spec := kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-04-15T00:00:00Z",
		Source: kblv1alpha1.SnapshotSource{
			Inline: map[string]interface{}{"key": "value"},
		},
		Sealed: true,
	}

	id1, err := snapshot.ComputeID(spec)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := snapshot.ComputeID(spec)
	if err != nil {
		t.Fatal(err)
	}
	if id1 != id2 {
		t.Fatalf("expected deterministic ID, got %s vs %s", id1, id2)
	}
	if len(id1) != 16 {
		t.Fatalf("expected 16-char snapshot ID, got %d", len(id1))
	}
}

func TestValidateRequiresSource(t *testing.T) {
	err := snapshot.Validate(kblv1alpha1.SnapshotSpec{TimeSlice: "2025-04-15"})
	if err == nil {
		t.Fatal("expected validation error for missing source")
	}
}

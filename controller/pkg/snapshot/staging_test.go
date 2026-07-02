package snapshot_test

import (
	"os"
	"path/filepath"
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
)

func TestSealPayloadPreservesPathBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "curve.json")
	original := `{"instruments":[{"instrument_id":"US10Y","rate":4.25}]}`
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	spec := kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-04-15T00:00:00Z",
		Source:    kblv1alpha1.SnapshotSource{Path: path},
		Sealed:    true,
	}

	id, data, err := snapshot.SealPayload(spec)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("expected snapshot ID")
	}
	if data != original {
		t.Fatalf("expected original bytes preserved, got %q", data)
	}

	// MarshalData via old path would reformat JSON; SealPayload must not.
	legacy, err := snapshot.MarshalData(spec)
	if err != nil {
		t.Fatal(err)
	}
	if legacy != original {
		t.Fatalf("MarshalData should delegate to SealPayload, got %q", legacy)
	}
}

package hash_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
)

func TestComputeDeterministic(t *testing.T) {
	data := map[string]interface{}{
		"instruments": []map[string]interface{}{
			{"instrument_id": "US10Y", "rate": 4.25},
		},
	}

	h1, err := hash.Compute(data)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	h2, err := hash.Compute(data)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if h1 != h2 {
		t.Errorf("same data produced different hashes: %s vs %s", h1, h2)
	}
}

func TestSnapshotIDStable(t *testing.T) {
	data := map[string]interface{}{"value": 42}
	id1, err := hash.SnapshotID("2025-07-02T00:00:00Z", data)
	if err != nil {
		t.Fatalf("snapshot ID: %v", err)
	}
	id2, err := hash.SnapshotID("2025-07-02T00:00:00Z", data)
	if err != nil {
		t.Fatalf("snapshot ID: %v", err)
	}
	if id1 != id2 {
		t.Errorf("snapshot ID not stable: %s vs %s", id1, id2)
	}
	if len(id1) != hash.SnapshotIDHexLen {
		t.Errorf("expected %d-char snapshot ID, got %d", hash.SnapshotIDHexLen, len(id1))
	}
}

func TestDifferentInputsDifferentHashes(t *testing.T) {
	h1, _ := hash.Compute(map[string]int{"a": 1})
	h2, _ := hash.Compute(map[string]int{"a": 2})
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestLinkIsDeterministicAndOrderSensitive(t *testing.T) {
	a, err := hash.Link("prev", "in", "out")
	if err != nil {
		t.Fatal(err)
	}
	b, err := hash.Link("prev", "in", "out")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Fatal("link must be a function of its parts")
	}
	c, err := hash.Link("prev", "out", "in")
	if err != nil {
		t.Fatal(err)
	}
	if a == c {
		t.Fatal("reordered parts must not collide")
	}
	if _, err := hash.Link(); err == nil {
		t.Fatal("empty link must be rejected")
	}
	if got := hash.ChainKey("s", "d", "i"); got != "s:d:i" {
		t.Fatalf("ChainKey = %q", got)
	}
}

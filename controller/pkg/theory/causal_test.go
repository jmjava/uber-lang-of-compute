package theory_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

func TestCausalPastIsPrefix(t *testing.T) {
	chain := []string{"load", "interpolate", "risk"}
	past, err := theory.CausalPast(chain, "interpolate")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"snapshot", "load", "interpolate"}
	if len(past) != len(want) {
		t.Fatalf("got %v want %v", past, want)
	}
	for i := range want {
		if past[i] != want[i] {
			t.Fatalf("got %v want %v", past, want)
		}
	}
}

func TestFutureReadRejected(t *testing.T) {
	chain := []string{"load", "interpolate", "risk"}
	err := theory.AllowedReads(chain, "load", []string{"snapshot", "risk"})
	if err == nil {
		t.Fatal("expected spacelike/future read to be rejected")
	}
}

func TestSnapshotReadAlwaysAllowed(t *testing.T) {
	chain := []string{"load"}
	if err := theory.AllowedReads(chain, "load", []string{"snapshot"}); err != nil {
		t.Fatal(err)
	}
}

func TestUnknownDominoRejected(t *testing.T) {
	if _, err := theory.CausalPast([]string{"a"}, "b"); err == nil {
		t.Fatal("expected error")
	}
}

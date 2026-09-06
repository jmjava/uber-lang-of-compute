package theory_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestVerifySpineEmptyLog(t *testing.T) {
	if err := theory.VerifySpine("snap", nil); err != nil {
		t.Fatal(err)
	}
}

func TestVerifySpineAcceptsLinkedChain(t *testing.T) {
	const snap = "snap-root"
	e1 := linkedEntry(t, snap, "in1", "out1")
	e2 := linkedEntry(t, e1.Link, "in2", "out2")
	if err := theory.VerifySpine(snap, []types.ReplayLogEntry{e1, e2}); err != nil {
		t.Fatal(err)
	}
}

func TestVerifySpineRejectsTamperedOutput(t *testing.T) {
	const snap = "snap-root"
	e1 := linkedEntry(t, snap, "in1", "out1")
	e2 := linkedEntry(t, e1.Link, "in2", "out2")
	e2.OutputHash = "tampered"
	if err := theory.VerifySpine(snap, []types.ReplayLogEntry{e1, e2}); err == nil {
		t.Fatal("tampered output must fail the spine")
	}
}

func TestVerifySpineRejectsBrokenPrev(t *testing.T) {
	const snap = "snap-root"
	e1 := linkedEntry(t, snap, "in1", "out1")
	e1.PrevLink = "not-the-snapshot"
	if err := theory.VerifySpine(snap, []types.ReplayLogEntry{e1}); err == nil {
		t.Fatal("broken prev-link must fail the spine")
	}
}

func linkedEntry(t *testing.T, prev, in, out string) types.ReplayLogEntry {
	t.Helper()
	link, err := hash.Link(prev, in, out)
	if err != nil {
		t.Fatal(err)
	}
	return types.ReplayLogEntry{
		PrevLink:   prev,
		InputHash:  in,
		OutputHash: out,
		Link:       link,
	}
}

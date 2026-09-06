package theory_test

import (
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

func TestShannonBitsKnownDistribution(t *testing.T) {
	// Two symbols, equal count → 1 bit.
	data := []byte{0, 1, 0, 1}
	if !theory.AlmostEqual(theory.ShannonBits(data), 1) {
		t.Fatalf("got %v want 1", theory.ShannonBits(data))
	}
	if theory.ShannonBits(nil) != 0 {
		t.Fatal("empty payload should have zero shannon entropy")
	}
}

func TestEnsembleEntropyCollapsesOnSeal(t *testing.T) {
	// Theorem E1: uniform ensemble of N distinct payloads has entropy log2(N);
	// after sealing one member, posterior entropy is 0.
	prior := []string{"curve-a", "curve-b", "curve-c", "curve-d"}
	hPrior := theory.EnsembleEntropyBits(prior)
	if !theory.AlmostEqual(hPrior, 2) {
		t.Fatalf("uniform 4-ensemble: got %v want 2", hPrior)
	}

	posterior := theory.SealEnsemble(prior, "curve-b")
	hPost := theory.EnsembleEntropyBits(posterior)
	if !theory.AlmostEqual(hPost, 0) {
		t.Fatalf("sealed ensemble: got %v want 0", hPost)
	}
	if len(posterior) != 1 || posterior[0] != "curve-b" {
		t.Fatalf("posterior should be the Dirac mass on the sealed payload, got %v", posterior)
	}
	// Sealing is conditioning, not compression: the payload bytes are unchanged,
	// so payload Shannon entropy is the same object before and after.
	payload := []byte("curve-b")
	if theory.ShannonBits(payload) != theory.ShannonBits([]byte(posterior[0])) {
		t.Fatal("payload entropy must be invariant under the act of sealing")
	}
}

func TestSealEnsembleRejectsUnknownPayload(t *testing.T) {
	if got := theory.SealEnsemble([]string{"a", "b"}, "c"); got != nil {
		t.Fatalf("expected empty posterior, got %v", got)
	}
}

func TestDuplicatePayloadsReduceEnsembleEntropy(t *testing.T) {
	h := theory.EnsembleEntropyBits([]string{"a", "a", "b", "b"})
	if !theory.AlmostEqual(h, 1) {
		t.Fatalf("got %v want 1", h)
	}
}

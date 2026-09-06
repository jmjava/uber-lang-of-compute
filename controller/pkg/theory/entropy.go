// Package theory implements the falsifiable computational side of the
// physics correspondences documented in docs/theory/.
//
// Nothing here claims that KBL is a physical theory. Each function is a
// discrete, testable counterpart of a named structure (Shannon entropy,
// unique Newtonian trajectories, causal pasts, coarse-graining).
package theory

import "math"

// ShannonBits is the Shannon entropy of a byte string treated as i.i.d. symbols
// over a 256-ary alphabet, in bits per symbol.
//
// This is payload compressibility, not the "low-entropy snapshot" claim.
// Sealing does not change ShannonBits of the payload; it changes ensemble entropy.
func ShannonBits(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	var freq [256]int
	for _, b := range data {
		freq[b]++
	}
	n := float64(len(data))
	var h float64
	for _, c := range freq {
		if c == 0 {
			continue
		}
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}

// EnsembleEntropyBits is the Shannon entropy of a discrete distribution over
// candidate payloads, in bits. Uniform on n distinct strings is log2(n).
//
// Correspondence: this is the epistemic entropy of "which input will compute see?"
// Sealing selects one payload and collapses the ensemble to a Dirac mass (entropy 0).
func EnsembleEntropyBits(payloads []string) float64 {
	if len(payloads) == 0 {
		return 0
	}
	counts := make(map[string]int, len(payloads))
	for _, p := range payloads {
		counts[p]++
	}
	n := float64(len(payloads))
	var h float64
	for _, c := range counts {
		p := float64(c) / n
		h -= p * math.Log2(p)
	}
	return h
}

// SealEnsemble returns the posterior ensemble after conditioning on a sealed
// payload. If sealed is not in the prior, the result is empty.
func SealEnsemble(prior []string, sealed string) []string {
	for _, p := range prior {
		if p == sealed {
			return []string{sealed}
		}
	}
	return nil
}

const entropyTol = 1e-12

// AlmostEqual reports whether two floats agree within a fixed absolute tolerance.
func AlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= entropyTol
}

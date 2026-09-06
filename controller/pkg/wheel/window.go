package wheel

import (
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
)

// SeatCount is the number of Ferris-wheel seats in a windowed unfolding of
// depth d and arity k: k^d leaves. Nature-inspired reading: the current
// Mandelbrot *explorer* viewport is the circuit the wheel plays.
func SeatCount(depth, arity int) int {
	return theory.LeafCount(depth, arity)
}

// SeatLabels names each leaf of Unfold(depth, arity) in left-to-right order.
func SeatLabels(depth, arity int) []string {
	root := theory.Unfold(depth, arity, "seat", 1)
	leaves := theory.Leaves(root)
	out := make([]string, len(leaves))
	for i, leaf := range leaves {
		out[i] = leaf.Label
	}
	return out
}

// ValidateWindowSeats reports whether a wheel's context count matches the
// explorer window (k^d leaves).
func ValidateWindowSeats(contextCount, depth, arity int) error {
	want := SeatCount(depth, arity)
	if contextCount != want {
		return fmt.Errorf("wheel seats %d != window leaves %d (arity^depth)", contextCount, want)
	}
	return nil
}

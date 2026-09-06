package theory

import "strings"

// Regularity is the Picard-style smoothness class of a domino command.
//
// Nature-inspired reading: uniqueness of Newtonian trajectories requires a
// regularity hypothesis (Lipschitz). Uniqueness of KBL worldlines requires a
// purity class. Neither claims the other; we copy the *conditional well-posedness*
// pattern, the way a neural net copies weighted summation without claiming wetware.
type Regularity int

const (
	// RegularityContract is an arbitrary container/image command. Uniqueness is
	// assumed by the deterministic:true contract, not discharged by the engine.
	RegularityContract Regularity = iota
	// RegularityPinned is a pinned-runtime command (Julia Manifest, fixed image).
	// Uniqueness is that of a *discretized* map, not of a real-valued function.
	RegularityPinned
	// RegularityBuiltin is a referentially transparent in-process function.
	RegularityBuiltin
)

// CommandRegularity classifies a domino command prefix.
func CommandRegularity(command string) Regularity {
	switch {
	case strings.HasPrefix(command, "builtin:"):
		return RegularityBuiltin
	case strings.HasPrefix(command, "julia:"):
		return RegularityPinned
	default:
		return RegularityContract
	}
}

// UniquenessGrade is how strongly uniqueness is warranted for this class.
func (r Regularity) UniquenessGrade() string {
	switch r {
	case RegularityBuiltin:
		return "theorem"
	case RegularityPinned:
		return "empirical-discretization"
	default:
		return "contract"
	}
}

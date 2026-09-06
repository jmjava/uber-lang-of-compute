package theory

import (
	"strings"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// History is one classical trajectory of one universe against one sealed snapshot.
//
// Nature-inspired reading: Everett's many-worlds is branching of a wavefunction.
// KBL copies the *branching bookkeeping* — independent evolutions that do not
// interfere — the way an ensemble of neural nets does not share activations.
// There is no amplitude, no inner product, no decoherence calculation.
// Sealed records are the only messages between histories (classical records).
type History struct {
	Universe   string
	SnapshotID string
	Worldline  string // concatenation of output hashes
}

// Worldline is the hash spine of a run: output hashes in chain order.
func Worldline(entries []types.ReplayLogEntry) string {
	if len(entries) == 0 {
		return ""
	}
	parts := make([]string, len(entries))
	for i, e := range entries {
		parts[i] = e.OutputHash
	}
	return strings.Join(parts, "|")
}

// HistoryFromRun builds the parent history of a completed sealed run.
func HistoryFromRun(universe, snapshotID string, entries []types.ReplayLogEntry) History {
	return History{
		Universe:   universe,
		SnapshotID: snapshotID,
		Worldline:  Worldline(entries),
	}
}

// Interfere reports whether two histories share a live (unsealed) channel.
// Under sealed-only copies — the fabric rule — the answer is always false:
// branches do not interfere. If liveShare were allowed, distinct universes
// holding the same snapshot identity would be a live coupling.
func Interfere(a, b History, liveShare bool) bool {
	if !liveShare {
		return false
	}
	return a.Universe != b.Universe && a.SnapshotID != "" && a.SnapshotID == b.SnapshotID
}

// Branch fans one completed history out to additional universes, carrying the
// same sealed snapshot identity. This is classical consistent-histories fan-out
// (routing), not unitary branching.
func Branch(parent History, universes []string) []History {
	out := make([]History, 0, len(universes))
	for _, u := range universes {
		if u == "" || u == parent.Universe {
			continue
		}
		out = append(out, History{
			Universe:   u,
			SnapshotID: parent.SnapshotID,
			Worldline:  parent.Worldline,
		})
	}
	return out
}

// SameRecord reports whether two histories are copies of the same sealed result.
func SameRecord(a, b History) bool {
	return a.SnapshotID != "" && a.SnapshotID == b.SnapshotID && a.Worldline == b.Worldline
}

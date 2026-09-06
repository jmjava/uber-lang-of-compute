package theory

import "github.com/jmjava/uber-lang-of-compute/controller/pkg/types"

// LogicalWork counts irreversible evaluations versus recorded reuses.
//
// Nature-inspired reading: Landauer/Bennett say *don't erase, record the result*.
// KBL copies that bookkeeping as memoization. This is not a joule meter — it is
// the same abstraction neural nets make when they copy "synaptic weight" without
// copying ATP. IrreversibleSteps is the computational analogue of "bits you paid
// to evaluate," not kT ln 2.
type LogicalWork struct {
	Evaluations int
	Reuses      int
}

// WorkFromReplay tallies a replay log.
func WorkFromReplay(entries []types.ReplayLogEntry) LogicalWork {
	var w LogicalWork
	for _, e := range entries {
		if e.Reused {
			w.Reuses++
		} else {
			w.Evaluations++
		}
	}
	return w
}

// IrreversibleSteps is the Landauer-shaped cost: evaluations that were not
// recovered from a recorded intermediate.
func (w LogicalWork) IrreversibleSteps() int {
	return w.Evaluations
}

// ReplaySaves reports whether a second pass paid strictly less irreversible
// work than the first, with at least one reuse.
func ReplaySaves(first, second LogicalWork) bool {
	return second.IrreversibleSteps() < first.IrreversibleSteps() && second.Reuses > 0
}

// DefaultDollarsPerEvaluation is a unit price for one irreversible step.
// Nature-inspired reading: this is a *cost accountant*, not Landauer heat.
const DefaultDollarsPerEvaluation = 0.01

// ChargeUSD is evaluations × unit price. Reuses are free.
func ChargeUSD(w LogicalWork, dollarsPerEval float64) float64 {
	if dollarsPerEval < 0 {
		dollarsPerEval = 0
	}
	return float64(w.Evaluations) * dollarsPerEval
}

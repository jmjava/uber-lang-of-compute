package theory

// ReconcileError is the deviation of an observed status from a desired spec.
// Zero means the essential variable is inside its bound (Ashby homeostasis).
//
// Nature-inspired reading: a lifeform keeps internals in range by feedback.
// Kubernetes reconcilers copy that loop. This is not metabolism or autopoiesis —
// it is the same move as an ANN copying "neuron fires" without copying glia.
func ReconcileError(desired, observed string) float64 {
	if desired == observed {
		return 0
	}
	return 1
}

// Homeostatic reports whether the essential variable is in bound.
func Homeostatic(err float64) bool {
	return err == 0
}

// DesiredWorkflowPhase is the essential variable the reconciler holds: a
// completed chain. Running/Pending/Error are out of bound.
const DesiredWorkflowPhase = "Completed"

// PhaseHomeostasis is the spec/status error for a Workflow phase.
func PhaseHomeostasis(observed string) float64 {
	return ReconcileError(DesiredWorkflowPhase, observed)
}

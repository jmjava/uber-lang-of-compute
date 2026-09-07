package controller

import (
	"strings"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/executor/julia"
)

func chainHasJuliaSteps(chain *kblv1alpha1.DominoChain) bool {
	for _, step := range chain.Spec.Steps {
		if strings.HasPrefix(step.Command, "julia:") {
			return true
		}
	}
	return false
}

func workflowHasJuliaDominos(wf *kblv1alpha1.Workflow) bool {
	for _, d := range wf.Spec.Dominos {
		if strings.HasPrefix(d.Command, "julia:") {
			return true
		}
	}
	return false
}

// skipJuliaOperatorReplay is true when Julia already ran in a runner pod but the
// operator image has no julia binary (distroless kbl-controller).
func skipJuliaOperatorReplay(chain *kblv1alpha1.DominoChain, wf *kblv1alpha1.Workflow) bool {
	hasJulia := false
	if chain != nil {
		hasJulia = chainHasJuliaSteps(chain)
	}
	if !hasJulia && wf != nil {
		hasJulia = workflowHasJuliaDominos(wf)
	}
	if !hasJulia {
		return false
	}
	return !julia.Available(julia.DefaultConfig())
}

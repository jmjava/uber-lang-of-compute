package wheel

import (
	"path/filepath"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

// BuildWorkflow materializes a Workflow CR for the given wheel slot.
func BuildWorkflow(
	wheel *kblv1alpha1.ComputeWheel,
	computeCtx *kblv1alpha1.ComputeContext,
	state State,
	storeRoot string,
) *kblv1alpha1.Workflow {
	timeSliceStr := state.CurrentTimeSlice.UTC().Format(time.RFC3339)
	timeSliceKey := FormatTimeSlice(state.CurrentTimeSlice)
	contextName := wheel.Spec.Contexts[state.ActiveContextIndex]

	tmpl := wheel.Spec.WorkflowTemplate
	snapshotRef := wheel.Name + "-" + timeSliceKey

	storePath := tmpl.Provisioning.StorePath
	if computeCtx != nil && computeCtx.Spec.StorePath != "" {
		storePath = filepath.Join(computeCtx.Spec.StorePath, wheel.Name, timeSliceKey+".db")
	} else if storePath == "" {
		storePath = filepath.Join(storeRoot, wheel.Namespace, wheel.Name, contextName, timeSliceKey+".db")
	}

	wfName := WorkflowName(wheel.Name, contextName, timeSliceKey)

	spec := kblv1alpha1.WorkflowSpec{
		Execution: tmpl.Execution,
		Provisioning: kblv1alpha1.ProvisioningSpec{
			StorePath: storePath,
			NodeLocal: true,
		},
		Routing: kblv1alpha1.RoutingSpec{
			Universe:          tmpl.Routing.Universe,
			ComputeContextRef: contextName,
		},
	}

	if tmpl.SnapshotRef != "" {
		spec.SnapshotRef = tmpl.SnapshotRef
	} else {
		snapshot := tmpl.Snapshot
		snapshot.TimeSlice = timeSliceStr
		snapshot.Sealed = true
		if snapshot.ComputeContextRef == "" {
			snapshot.ComputeContextRef = contextName
		}
		spec.Snapshot = snapshot
	}

	if len(tmpl.DominoRefs) > 0 {
		spec.DominoRefs = append([]string(nil), tmpl.DominoRefs...)
	} else {
		dominos := make([]kblv1alpha1.DominoSpec, len(tmpl.Dominos))
		for i, d := range tmpl.Dominos {
			dominos[i] = d
			if dominos[i].SnapshotRef == "" {
				dominos[i].SnapshotRef = snapshotRef
			}
		}
		spec.Dominos = dominos
	}

	return &kblv1alpha1.Workflow{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kbl.io/v1alpha1",
			Kind:       "Workflow",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      wfName,
			Namespace: wheel.Namespace,
			Labels: map[string]string{
				"kbl.io/computewheel":          wheel.Name,
				"kbl.io/compute-context":       contextName,
				"kbl.io/time-slice":            timeSliceKey,
				"app.kubernetes.io/managed-by": "kbl-controller",
			},
		},
		Spec: spec,
	}
}

// StateFromStatus reconstructs wheel.State from ComputeWheel status fields.
func StateFromStatus(status kblv1alpha1.ComputeWheelStatus) (State, error) {
	if status.CurrentTimeSlice == "" {
		return State{}, nil
	}
	t, err := time.Parse(time.RFC3339, status.CurrentTimeSlice)
	if err != nil {
		return State{}, err
	}
	return State{
		CurrentTimeSlice:   t.UTC(),
		ActiveContextIndex: status.ActiveContextIndex,
		RotationCount:      status.RotationCount,
	}, nil
}

// ApplyState writes wheel.State back into ComputeWheel status fields.
func ApplyState(status *kblv1alpha1.ComputeWheelStatus, state State, contextName string) {
	status.CurrentTimeSlice = state.CurrentTimeSlice.UTC().Format(time.RFC3339)
	status.ActiveContextIndex = state.ActiveContextIndex
	status.RotationCount = state.RotationCount
	status.ActiveContext = contextName
}

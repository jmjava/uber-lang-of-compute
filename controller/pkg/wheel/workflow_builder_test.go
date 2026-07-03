package wheel_test

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

func TestBuildWorkflowInlineTemplate(t *testing.T) {
	ts := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	w := &kblv1alpha1.ComputeWheel{
		ObjectMeta: metav1.ObjectMeta{Name: "finance-wheel", Namespace: "default"},
		Spec: kblv1alpha1.ComputeWheelSpec{
			Contexts: []string{"node-a"},
			WorkflowTemplate: kblv1alpha1.WorkflowTemplateSpec{
				Snapshot: kblv1alpha1.SnapshotSpec{
					Source: kblv1alpha1.SnapshotSource{
						Inline: map[string]interface{}{"value": 1},
					},
					Sealed: true,
				},
				Dominos: []kblv1alpha1.DominoSpec{
					{Name: "load", Command: "builtin:identity"},
				},
				Execution: kblv1alpha1.ExecutionSpec{
					Chain:         []string{"load"},
					Deterministic: true,
				},
			},
		},
	}

	wf := wheel.BuildWorkflow(w, nil, wheel.State{CurrentTimeSlice: ts, ActiveContextIndex: 0}, "/tmp/kbl")
	if wf.Spec.Snapshot.TimeSlice != ts.Format(time.RFC3339) {
		t.Fatalf("expected stamped time slice, got %s", wf.Spec.Snapshot.TimeSlice)
	}
	if wf.Spec.Dominos[0].SnapshotRef == "" {
		t.Fatal("expected generated snapshotRef on inline domino")
	}
}

func TestBuildWorkflowCRRefs(t *testing.T) {
	ts := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	w := &kblv1alpha1.ComputeWheel{
		ObjectMeta: metav1.ObjectMeta{Name: "refs-wheel", Namespace: "default"},
		Spec: kblv1alpha1.ComputeWheelSpec{
			Contexts: []string{"node-a"},
			WorkflowTemplate: kblv1alpha1.WorkflowTemplateSpec{
				SnapshotRef: "curve-snap",
				DominoRefs:  []string{"load-curve", "interpolate-curve"},
				Execution: kblv1alpha1.ExecutionSpec{
					Chain:         []string{"load-curve", "interpolate-curve"},
					Deterministic: true,
				},
			},
		},
	}

	wf := wheel.BuildWorkflow(w, nil, wheel.State{CurrentTimeSlice: ts, ActiveContextIndex: 0}, "/tmp/kbl")
	if wf.Spec.SnapshotRef != "curve-snap" {
		t.Fatalf("expected snapshotRef, got %+v", wf.Spec)
	}
	if len(wf.Spec.DominoRefs) != 2 {
		t.Fatalf("expected domino refs, got %+v", wf.Spec)
	}
	if wf.Spec.Snapshot.TimeSlice != "" {
		t.Fatal("inline snapshot should not be set when snapshotRef is used")
	}
}

func TestBuildWorkflowVolcanoQueue(t *testing.T) {
	ts := time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)
	w := &kblv1alpha1.ComputeWheel{
		ObjectMeta: metav1.ObjectMeta{Name: "volcano-wheel", Namespace: "default"},
		Spec: kblv1alpha1.ComputeWheelSpec{
			Contexts:     []string{"default-context"},
			VolcanoQueue: "kbl-lab",
			NodeSelector: map[string]string{"kbl.io/lab-role": "compute"},
			WorkflowTemplate: kblv1alpha1.WorkflowTemplateSpec{
				Snapshot: kblv1alpha1.SnapshotSpec{
					Source: kblv1alpha1.SnapshotSource{
						Inline: map[string]interface{}{"value": 1},
					},
					Sealed: true,
				},
				Dominos: []kblv1alpha1.DominoSpec{
					{Name: "load", Command: "julia:identity"},
				},
				Execution: kblv1alpha1.ExecutionSpec{
					Chain:         []string{"load"},
					Deterministic: true,
					Runtime:       string(kblv1alpha1.DominoChainRuntimeVolcanoInit),
				},
				Provisioning: kblv1alpha1.ProvisioningSpec{
					RunnerImage: "kbl-domino-runner-julia:lab",
				},
			},
		},
	}

	wf := wheel.BuildWorkflow(w, nil, wheel.State{CurrentTimeSlice: ts, ActiveContextIndex: 0}, "/tmp/kbl")
	if wf.Spec.Execution.VolcanoQueue != "kbl-lab" {
		t.Fatalf("expected wheel volcano queue on workflow, got %q", wf.Spec.Execution.VolcanoQueue)
	}
	if wf.Spec.Execution.Runtime != string(kblv1alpha1.DominoChainRuntimeVolcanoInit) {
		t.Fatalf("expected volcano-init runtime, got %q", wf.Spec.Execution.Runtime)
	}
	if wf.Spec.Provisioning.RunnerImage != "kbl-domino-runner-julia:lab" {
		t.Fatalf("expected runner image, got %q", wf.Spec.Provisioning.RunnerImage)
	}
	if wf.Spec.Provisioning.NodeSelector["kbl.io/lab-role"] != "compute" {
		t.Fatalf("expected wheel node selector, got %+v", wf.Spec.Provisioning.NodeSelector)
	}
	if wf.Labels["kbl.io/volcano-queue"] != "kbl-lab" {
		t.Fatalf("expected volcano queue label, got %q", wf.Labels["kbl.io/volcano-queue"])
	}
}

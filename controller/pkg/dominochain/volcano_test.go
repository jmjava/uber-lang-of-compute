package dominochain_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
)

func TestBuildVolcanoJob(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeVolcanoInit,
			VolcanoQueue: "kbl-lab",
			NodeSelector: map[string]string{"kbl.io/lab-role": "compute"},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "julia:identity"},
				{Name: "interp", Command: "julia:interpolate"},
			},
			RunnerImage: dominochain.DefaultJuliaRunnerImage,
		},
	}
	chain.Name = "julia-finance-volcano"
	chain.Namespace = "default"

	job := (&dominochain.Builder{}).BuildVolcanoJob(chain)
	if job.GetName() != "julia-finance-volcano-chain" {
		t.Fatalf("expected job name julia-finance-volcano-chain, got %s", job.GetName())
	}

	queue, _, _ := unstructuredNestedString(job.Object, "spec", "queue")
	if queue != "kbl-lab" {
		t.Fatalf("expected queue kbl-lab, got %q", queue)
	}

	scheduler, _, _ := unstructuredNestedString(job.Object, "spec", "schedulerName")
	if scheduler != dominochain.VolcanoSchedulerName {
		t.Fatalf("expected scheduler volcano, got %q", scheduler)
	}

	tasks, _, _ := unstructuredNestedSlice(job.Object, "spec", "tasks")
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	task, ok := tasks[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected task map")
	}
	template, ok := task["template"].(map[string]interface{})
	if !ok {
		t.Fatal("expected task template")
	}
	spec, ok := template["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("expected pod spec in template")
	}
	if spec["schedulerName"] != dominochain.VolcanoSchedulerName {
		t.Fatalf("expected task pod scheduler volcano, got %v", spec["schedulerName"])
	}

	inits, ok := spec["initContainers"].([]interface{})
	if !ok || len(inits) != 2 {
		t.Fatalf("expected 2 init containers, got %v", spec["initContainers"])
	}
}

func TestVolcanoJobPhaseHelpers(t *testing.T) {
	job := volcanoJobWithPhase("Completed", "")
	if !dominochain.IsVolcanoJobComplete(job) {
		t.Fatal("expected completed")
	}
	if dominochain.IsVolcanoJobFailed(job) {
		t.Fatal("completed should not be failed")
	}

	failed := volcanoJobWithPhase("Failed", "task pod failed")
	if !dominochain.IsVolcanoJobFailed(failed) {
		t.Fatal("expected failed")
	}
}

func volcanoJobWithPhase(phase, message string) *unstructured.Unstructured {
	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(dominochain.VolcanoJobGVK)
	_ = unstructured.SetNestedField(job.Object, map[string]interface{}{
		"phase":   phase,
		"message": message,
	}, "status", "state")
	return job
}

// Helpers mirror unstructured access without exporting test-only imports from production code.
func unstructuredNestedString(obj interface{}, fields ...string) (string, bool, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return "", false, nil
	}
	cur := m
	for i, f := range fields {
		if i == len(fields)-1 {
			v, ok := cur[f].(string)
			return v, ok, nil
		}
		next, ok := cur[f].(map[string]interface{})
		if !ok {
			return "", false, nil
		}
		cur = next
	}
	return "", false, nil
}

func unstructuredNestedSlice(obj interface{}, fields ...string) ([]interface{}, bool, error) {
	m, ok := obj.(map[string]interface{})
	if !ok {
		return nil, false, nil
	}
	cur := m
	for i, f := range fields {
		if i == len(fields)-1 {
			v, ok := cur[f].([]interface{})
			return v, ok, nil
		}
		next, ok := cur[f].(map[string]interface{})
		if !ok {
			return nil, false, nil
		}
		cur = next
	}
	return nil, false, nil
}

package dominochain

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

const (
	DefaultVolcanoQueue   = "default"
	VolcanoSchedulerName  = "volcano"
	volcanoInitTaskName   = "domino-chain"
)

// VolcanoJobGVK is the GroupVersionKind for Volcano batch Jobs.
var VolcanoJobGVK = schema.GroupVersionKind{
	Group:   "batch.volcano.sh",
	Version: "v1alpha1",
	Kind:    "Job",
}

// VolcanoJobName returns the Volcano Job name for a domino chain.
func VolcanoJobName(chain *kblv1alpha1.DominoChain) string {
	return chain.Name + "-chain"
}

// BuildVolcanoJob returns a Volcano Job whose single task runs the init-container domino chain.
func (b *Builder) BuildVolcanoJob(chain *kblv1alpha1.DominoChain) *unstructured.Unstructured {
	pod := b.BuildInitChainPod(chain)
	podSpec := pod.Spec.DeepCopy()
	podSpec.SchedulerName = VolcanoSchedulerName

	queue := chain.Spec.VolcanoQueue
	if queue == "" {
		queue = DefaultVolcanoQueue
	}

	taskTemplate := map[string]interface{}{
		"metadata": map[string]interface{}{
			"labels":      stringMapToInterface(pod.Labels),
			"annotations": stringMapToInterface(pod.Annotations),
		},
		"spec": podSpecToMap(*podSpec),
	}

	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(VolcanoJobGVK)
	job.SetName(VolcanoJobName(chain))
	job.SetNamespace(chain.Namespace)
	job.SetLabels(chainLabels(chain.Name))

	_ = unstructured.SetNestedField(job.Object, int64(1), "spec", "minAvailable")
	_ = unstructured.SetNestedField(job.Object, VolcanoSchedulerName, "spec", "schedulerName")
	_ = unstructured.SetNestedField(job.Object, queue, "spec", "queue")
	_ = unstructured.SetNestedSlice(job.Object, []interface{}{
		map[string]interface{}{
			"event":  "TaskCompleted",
			"action": "CompleteJob",
		},
	}, "spec", "policies")
	_ = unstructured.SetNestedSlice(job.Object, []interface{}{
		map[string]interface{}{
			"replicas": int64(1),
			"name":     volcanoInitTaskName,
			"policies": []interface{}{
				map[string]interface{}{
					"event":  "TaskCompleted",
					"action": "CompleteJob",
				},
			},
			"template": taskTemplate,
		},
	}, "spec", "tasks")

	job.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "DominoChain",
		Name:       chain.Name,
		UID:        chain.UID,
	}})

	return job
}

func podSpecToMap(spec corev1.PodSpec) map[string]interface{} {
	data, err := json.Marshal(spec)
	if err != nil {
		return map[string]interface{}{}
	}
	var out map[string]interface{}
	if err := json.Unmarshal(data, &out); err != nil {
		return map[string]interface{}{}
	}
	return out
}

func stringMapToInterface(in map[string]string) map[string]interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// VolcanoJobPhase extracts status.state.phase from a Volcano Job.
func VolcanoJobPhase(job *unstructured.Unstructured) string {
	phase, _, _ := unstructured.NestedString(job.Object, "status", "state", "phase")
	return phase
}

// IsVolcanoJobComplete reports whether Volcano marked the Job completed.
func IsVolcanoJobComplete(job *unstructured.Unstructured) bool {
	return VolcanoJobPhase(job) == "Completed"
}

// IsVolcanoJobFailed reports terminal failure phases on a Volcano Job.
func IsVolcanoJobFailed(job *unstructured.Unstructured) bool {
	switch VolcanoJobPhase(job) {
	case "Failed", "Aborted", "Terminated":
		return true
	default:
		return false
	}
}

// VolcanoJobStatusMessage returns a human-readable status message when present.
func VolcanoJobStatusMessage(job *unstructured.Unstructured) string {
	msg, _, _ := unstructured.NestedString(job.Object, "status", "state", "message")
	if msg != "" {
		return msg
	}
	return fmt.Sprintf("volcano job phase %s", VolcanoJobPhase(job))
}

package dominochain

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

// ContainerRecreateRequestGVK is the GroupVersionKind for OpenKruise CRRs.
func ContainerRecreateRequestGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apps.kruise.io",
		Version: "v1alpha1",
		Kind:    "ContainerRecreateRequest",
	}
}

// ContainerRecreateRequest builds an OpenKruise CRR to hot-swap a placeholder slot.
func ContainerRecreateRequest(chain *kblv1alpha1.DominoChain, b *Builder, stepIndex int) *unstructured.Unstructured {
	step := chain.Spec.Steps[stepIndex]
	containerName := StepContainerName(chain, stepIndex)
	image := b.stepImage(chain, step)

	crr := &unstructured.Unstructured{}
	crr.SetGroupVersionKind(ContainerRecreateRequestGVK())
	crr.SetName(fmt.Sprintf("%s-slot-%d", chain.Name, stepIndex))
	crr.SetNamespace(chain.Namespace)
	crr.SetLabels(map[string]string{
		LabelDominoChain: chain.Name,
		LabelManagedBy:   "kbl-controller",
		"kbl.io/step":      step.Name,
	})

	_ = unstructured.SetNestedField(crr.Object, chain.Name+"-chain", "spec", "podName")
	_ = unstructured.SetNestedSlice(crr.Object, []interface{}{containerName}, "spec", "containers")
	_ = unstructured.SetNestedMap(crr.Object, map[string]interface{}{
		"orderedRecreate": true,
	}, "spec", "strategy")
	_ = unstructured.SetNestedMap(crr.Object, map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"name":  containerName,
				"image": image,
				"env":   envVarsToCRR(b.stepEnv(step, inputPath(stepIndex))),
			},
		},
	}, "spec", "template")

	crr.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "DominoChain",
		Name:       chain.Name,
		UID:        chain.UID,
	}})

	return crr
}

func inputPath(stepIndex int) string {
	if stepIndex == 0 {
		return SnapshotMountPath + "/snapshot.json"
	}
	return HandoffMountPath + "/output.json"
}

func envVarsToCRR(env []corev1.EnvVar) []interface{} {
	out := make([]interface{}, len(env))
	for i, e := range env {
		out[i] = map[string]interface{}{"name": e.Name, "value": e.Value}
	}
	return out
}

// CRRPhase extracts phase from an unstructured CRR status.
func CRRPhase(crr *unstructured.Unstructured) string {
	phase, _, _ := unstructured.NestedString(crr.Object, "status", "phase")
	return phase
}

// IsCRRComplete returns true when OpenKruise marks the CRR complete.
func IsCRRComplete(crr *unstructured.Unstructured) bool {
	phase := CRRPhase(crr)
	return phase == "Completed" || phase == "Complete"
}

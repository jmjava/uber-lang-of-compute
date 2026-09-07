package dominochain

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

const imagePullJobNameSuffix = "-pull"

// ImagePullJobGVK is the GroupVersionKind for OpenKruise ImagePullJob.
var ImagePullJobGVK = schema.GroupVersionKind{
	Group:   "apps.kruise.io",
	Version: "v1alpha1",
	Kind:    "ImagePullJob",
}

// ContainerRecreateRequestGVK is the GroupVersionKind for OpenKruise CRRs.
// CRR 1.6 cannot change image or env; the reconciler does not emit CRRs.
func ContainerRecreateRequestGVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   "apps.kruise.io",
		Version: "v1alpha1",
		Kind:    "ContainerRecreateRequest",
	}
}

// ImagePullJobName is a DNS-1123 name ≤ 63 characters.
func ImagePullJobName(chain *kblv1alpha1.DominoChain, imageIndex int) string {
	suffix := imagePullJobNameSuffix
	if imageIndex > 0 {
		suffix = fmt.Sprintf("%s-%d", imagePullJobNameSuffix, imageIndex)
	}
	name := chain.Name + suffix
	if len(name) <= 63 {
		return name
	}
	max := 63 - len(suffix)
	if max < 1 {
		return strings.TrimRight(("kbl" + suffix)[:63], "-")
	}
	trimmed := strings.TrimRight(chain.Name[:max], "-")
	if trimmed == "" {
		trimmed = "kbl"
	}
	return trimmed + suffix
}

// OpenKruisePullImages is the unique runner images a chain will execute.
func (b *Builder) OpenKruisePullImages(chain *kblv1alpha1.DominoChain) []string {
	seen := map[string]struct{}{}
	var images []string
	if chain == nil {
		return images
	}
	for _, step := range chain.Spec.Steps {
		img := b.stepImage(chain, step)
		if img == "" {
			continue
		}
		if _, ok := seen[img]; ok {
			continue
		}
		seen[img] = struct{}{}
		images = append(images, img)
	}
	if len(images) == 0 {
		images = append(images, b.runnerImage(chain))
	}
	return images
}

// BuildImagePullJob asks kruise-daemon to prefetch the runner image onto
// nodes matching the chain's nodeSelector (all nodes if unset).
func BuildImagePullJob(chain *kblv1alpha1.DominoChain, image string, imageIndex int) *unstructured.Unstructured {
	job := &unstructured.Unstructured{}
	job.SetGroupVersionKind(ImagePullJobGVK)
	job.SetName(ImagePullJobName(chain, imageIndex))
	job.SetNamespace(chain.Namespace)
	labels := chainLabels(chain.Name)
	labels["kbl.io/image-pull"] = "true"
	for k, v := range chain.Labels {
		if strings.HasPrefix(k, "kbl.io/") {
			labels[k] = v
		}
	}
	job.SetLabels(labels)

	_ = unstructured.SetNestedField(job.Object, image, "spec", "image")
	_ = unstructured.SetNestedField(job.Object, "IfNotPresent", "spec", "imagePullPolicy")
	_ = unstructured.SetNestedField(job.Object, int64(1), "spec", "parallelism")
	_ = unstructured.SetNestedMap(job.Object, map[string]interface{}{
		"type":                    "Always",
		"activeDeadlineSeconds":   int64(900),
		"ttlSecondsAfterFinished": int64(600),
	}, "spec", "completionPolicy")
	_ = unstructured.SetNestedMap(job.Object, map[string]interface{}{
		"backoffLimit":   int64(3),
		"timeoutSeconds": int64(600),
	}, "spec", "pullPolicy")
	_ = unstructured.SetNestedMap(job.Object, map[string]interface{}{
		"labels": stringMapToInterface(labels),
	}, "spec", "sandboxConfig")

	if len(chain.Spec.NodeSelector) > 0 {
		_ = unstructured.SetNestedStringMap(job.Object, chain.Spec.NodeSelector, "spec", "selector", "matchLabels")
	}

	if chain.UID != "" {
		job.SetOwnerReferences([]metav1.OwnerReference{{
			APIVersion: "kbl.io/v1alpha1",
			Kind:       "DominoChain",
			Name:       chain.Name,
			UID:        chain.UID,
		}})
	}

	return job
}

func nestedStatusInt(job *unstructured.Unstructured, field string) int64 {
	if job == nil {
		return 0
	}
	val, found, err := unstructured.NestedFieldNoCopy(job.Object, "status", field)
	if !found || err != nil || val == nil {
		return 0
	}
	switch n := val.(type) {
	case int64:
		return n
	case int32:
		return int64(n)
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

// IsImagePullJobComplete is true when kruise-daemon finished at least the desired pulls.
func IsImagePullJobComplete(job *unstructured.Unstructured) bool {
	if job == nil {
		return false
	}
	succeeded := nestedStatusInt(job, "succeeded")
	desired := nestedStatusInt(job, "desired")
	if succeeded <= 0 {
		return false
	}
	if desired > 0 && succeeded < desired {
		return false
	}
	return true
}

// IsImagePullJobFailed is true when the job finished with no successful pulls.
func IsImagePullJobFailed(job *unstructured.Unstructured) bool {
	if job == nil {
		return false
	}
	succeeded := nestedStatusInt(job, "succeeded")
	desired := nestedStatusInt(job, "desired")
	failed := nestedStatusInt(job, "failed")
	completion, _, _ := unstructured.NestedString(job.Object, "status", "completionTime")
	if succeeded > 0 {
		return false
	}
	if failed > 0 && completion != "" {
		return true
	}
	if failed > 0 && desired > 0 && failed >= desired {
		return true
	}
	msg, _, _ := unstructured.NestedString(job.Object, "status", "message")
	if strings.Contains(strings.ToLower(msg), "deadline") && failed > 0 {
		return true
	}
	return false
}

// ImagePullJobStatusMessage is a short status string for chain.Status.Message.
func ImagePullJobStatusMessage(job *unstructured.Unstructured) string {
	if job == nil {
		return "image pull job missing"
	}
	msg, _, _ := unstructured.NestedString(job.Object, "status", "message")
	if msg != "" {
		return msg
	}
	return fmt.Sprintf("image pull succeeded=%d desired=%d failed=%d",
		nestedStatusInt(job, "succeeded"), nestedStatusInt(job, "desired"), nestedStatusInt(job, "failed"))
}

// ContainerRecreateRequest builds an OpenKruise CRR to hot-swap a placeholder slot.
// Unused by the reconciler: OpenKruise 1.6 CRR cannot change image or env.
func ContainerRecreateRequest(chain *kblv1alpha1.DominoChain, b *Builder, stepIndex int) *unstructured.Unstructured {
	step := chain.Spec.Steps[stepIndex]
	containerName := StepContainerName(chain, stepIndex)
	_ = b

	crr := &unstructured.Unstructured{}
	crr.SetGroupVersionKind(ContainerRecreateRequestGVK())
	crr.SetName(fmt.Sprintf("%s-slot-%d", chain.Name, stepIndex))
	crr.SetNamespace(chain.Namespace)
	crr.SetLabels(map[string]string{
		LabelDominoChain: chain.Name,
		LabelManagedBy:   "kbl-controller",
		"kbl.io/step":    step.Name,
	})

	_ = unstructured.SetNestedField(crr.Object, chain.Name+"-chain", "spec", "podName")
	_ = unstructured.SetNestedSlice(crr.Object, []interface{}{
		map[string]interface{}{"name": containerName},
	}, "spec", "containers")
	_ = unstructured.SetNestedMap(crr.Object, map[string]interface{}{
		"orderedRecreate": true,
	}, "spec", "strategy")

	crr.SetOwnerReferences([]metav1.OwnerReference{{
		APIVersion: "kbl.io/v1alpha1",
		Kind:       "DominoChain",
		Name:       chain.Name,
		UID:        chain.UID,
	}})

	return crr
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

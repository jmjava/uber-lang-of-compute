package dominochain_test

import (
	"strings"
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestContainerRecreateRequestJuliaEnv(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime: kblv1alpha1.DominoChainRuntimeOpenKruise,
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "interp", Command: "julia:interpolate"},
			},
			RunnerImage: dominochain.DefaultJuliaRunnerImage,
		},
	}
	chain.Name = "julia-openkruise"
	chain.Namespace = "default"

	crr := dominochain.ContainerRecreateRequest(chain, &dominochain.Builder{}, 0)
	containers, found, err := unstructured.NestedSlice(crr.Object, "spec", "containers")
	if err != nil || !found || len(containers) == 0 {
		t.Fatalf("expected CRR containers: found=%v err=%v", found, err)
	}
	container, ok := containers[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected container map")
	}
	if container["name"] == "" {
		t.Fatal("expected container name")
	}
	podName, _, _ := unstructured.NestedString(crr.Object, "spec", "podName")
	if podName != "julia-openkruise-chain" {
		t.Fatalf("expected podName julia-openkruise-chain, got %q", podName)
	}
}

func TestBuildImagePullJob(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime:     kblv1alpha1.DominoChainRuntimeOpenKruise,
			RunnerImage: "kbl-domino-runner-julia:lab",
			NodeSelector: map[string]string{
				"kbl.io/lab-role": "compute",
			},
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "julia:identity"},
				{Name: "greeks", Command: "julia:greeks"},
			},
		},
	}
	chain.Name = "julia-finance-openkruise-dchain"
	chain.Namespace = "default"

	b := &dominochain.Builder{}
	images := b.OpenKruisePullImages(chain)
	if len(images) != 1 || images[0] != "kbl-domino-runner-julia:lab" {
		t.Fatalf("expected one julia runner image, got %v", images)
	}

	job := dominochain.BuildImagePullJob(chain, images[0], 0)
	if job.GetName() != "julia-finance-openkruise-dchain-pull" {
		t.Fatalf("job name: %s", job.GetName())
	}
	if job.GroupVersionKind() != dominochain.ImagePullJobGVK {
		t.Fatalf("gvk: %s", job.GroupVersionKind())
	}
	image, _, _ := unstructured.NestedString(job.Object, "spec", "image")
	if image != "kbl-domino-runner-julia:lab" {
		t.Fatalf("image: %s", image)
	}
	policy, _, _ := unstructured.NestedString(job.Object, "spec", "imagePullPolicy")
	if policy != "IfNotPresent" {
		t.Fatalf("imagePullPolicy: %s", policy)
	}
	sel, _, _ := unstructured.NestedStringMap(job.Object, "spec", "selector", "matchLabels")
	if sel["kbl.io/lab-role"] != "compute" {
		t.Fatalf("selector: %v", sel)
	}
}

func TestBuildImagePullJobCopiesChainLabels(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{
		Spec: kblv1alpha1.DominoChainSpec{
			Runtime:     kblv1alpha1.DominoChainRuntimeOpenKruise,
			RunnerImage: "kbl-domino-runner-julia:lab",
			Steps: []kblv1alpha1.DominoStepSpec{
				{Name: "load", Command: "julia:identity"},
			},
		},
	}
	chain.Name = "demo"
	chain.Namespace = "default"
	chain.Labels = map[string]string{
		"kbl.io/openkruise-demo": "true",
		"other":                  "skip",
	}
	job := dominochain.BuildImagePullJob(chain, "kbl-domino-runner-julia:lab", 0)
	if job.GetLabels()["kbl.io/openkruise-demo"] != "true" {
		t.Fatalf("labels: %v", job.GetLabels())
	}
	if _, ok := job.GetLabels()["other"]; ok {
		t.Fatal("non-kbl labels must not copy")
	}
}

func TestImagePullJobNameTruncates(t *testing.T) {
	chain := &kblv1alpha1.DominoChain{}
	chain.Name = strings.Repeat("a", 60)
	name := dominochain.ImagePullJobName(chain, 0)
	if len(name) > 63 {
		t.Fatalf("name length %d: %s", len(name), name)
	}
	if !strings.HasSuffix(name, "-pull") {
		t.Fatalf("expected -pull suffix, got %s", name)
	}
}

func TestImagePullJobCompleteAndFailed(t *testing.T) {
	complete := &unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"succeeded": int64(1),
			"desired":   int64(1),
			"failed":    int64(0),
		},
	}}
	if !dominochain.IsImagePullJobComplete(complete) {
		t.Fatal("expected complete")
	}
	if dominochain.IsImagePullJobFailed(complete) {
		t.Fatal("complete job is not failed")
	}

	pending := &unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"succeeded": int64(0),
			"desired":   int64(1),
			"active":    int64(1),
		},
	}}
	if dominochain.IsImagePullJobComplete(pending) {
		t.Fatal("pending should not be complete")
	}

	failed := &unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"succeeded":      int64(0),
			"desired":        int64(1),
			"failed":         int64(1),
			"completionTime": "2025-04-15T00:00:00Z",
			"message":        "pull failed",
		},
	}}
	if !dominochain.IsImagePullJobFailed(failed) {
		t.Fatal("expected failed")
	}
	if dominochain.IsImagePullJobComplete(failed) {
		t.Fatal("failed job is not complete")
	}

	floatStatus := &unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{
			"succeeded": float64(1),
			"desired":   float64(1),
			"failed":    float64(0),
		},
	}}
	if !dominochain.IsImagePullJobComplete(floatStatus) {
		t.Fatal("expected complete from float64 status numbers")
	}
}

package universe_test

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/universe"
)

func TestApplyToWorkflowFillsRunnerImageFromUniverse(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)

	u := &kblv1alpha1.PluggableUniverse{
		ObjectMeta: metav1.ObjectMeta{Name: "julia-finance-universe"},
		Spec: kblv1alpha1.PluggableUniverseSpec{
			ExecutionEngine: kblv1alpha1.ExecutionEngineSpec{
				Type:         "julia",
				RuntimeImage: "kbl-domino-runner-julia:lab",
			},
			DataLayer: kblv1alpha1.DataLayerSpec{Type: "tsdb"},
		},
	}
	dom := &kblv1alpha1.Domino{
		ObjectMeta: metav1.ObjectMeta{Name: "julia-identity", Namespace: "default"},
		Spec:       kblv1alpha1.DominoResourceSpec{SnapshotRef: "julia-curve-2025-04-15", Command: "julia:identity"},
	}
	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Name: "julia-wf", Namespace: "default"},
		Spec: kblv1alpha1.WorkflowSpec{
			DominoRefs: []string{"julia-identity"},
			Execution:  kblv1alpha1.ExecutionSpec{Chain: []string{"julia-identity"}},
			Routing:    kblv1alpha1.RoutingSpec{Universe: "julia-finance-universe"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(u, dom).Build()

	if err := universe.ApplyToWorkflow(context.Background(), cl, wf); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if wf.Spec.Provisioning.RunnerImage != "kbl-domino-runner-julia:lab" {
		t.Fatalf("expected universe runner image, got %q", wf.Spec.Provisioning.RunnerImage)
	}
	if wf.Spec.Execution.Runtime != string(kblv1alpha1.DominoChainRuntimeKubernetesInit) {
		t.Fatalf("expected default julia runtime kubernetes-init, got %q", wf.Spec.Execution.Runtime)
	}
}

func TestApplyToWorkflowRejectsNonJuliaRunner(t *testing.T) {
	wf := &kblv1alpha1.Workflow{
		Spec: kblv1alpha1.WorkflowSpec{
			Dominos: []kblv1alpha1.DominoSpec{{Name: "load", Command: "julia:identity"}},
			Provisioning: kblv1alpha1.ProvisioningSpec{
				RunnerImage: "kbl-domino-runner:lab",
			},
		},
	}
	err := universe.ApplyToWorkflow(context.Background(), nil, wf)
	if err == nil || !strings.Contains(err.Error(), "julia runner image") {
		t.Fatalf("expected julia runner rejection, got %v", err)
	}
}

func TestApplyToWorkflowLeavesLocalBuiltinAlone(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = kblv1alpha1.AddToScheme(scheme)
	wf := &kblv1alpha1.Workflow{
		Spec: kblv1alpha1.WorkflowSpec{
			Dominos: []kblv1alpha1.DominoSpec{{Name: "load", Command: "builtin:identity"}},
			Routing: kblv1alpha1.RoutingSpec{Universe: "missing-universe"},
		},
	}
	cl := fake.NewClientBuilder().WithScheme(scheme).Build()
	if err := universe.ApplyToWorkflow(context.Background(), cl, wf); err != nil {
		t.Fatalf("local builtin should ignore missing universe: %v", err)
	}
	if wf.Spec.Provisioning.RunnerImage != "" {
		t.Fatalf("did not expect runner image, got %q", wf.Spec.Provisioning.RunnerImage)
	}
	if wf.Spec.Execution.Runtime != "" {
		t.Fatalf("did not expect runtime, got %q", wf.Spec.Execution.Runtime)
	}
}

func TestApplyToWorkflowDefaultJuliaImage(t *testing.T) {
	wf := &kblv1alpha1.Workflow{
		Spec: kblv1alpha1.WorkflowSpec{
			Dominos: []kblv1alpha1.DominoSpec{{Name: "load", Command: "julia:greeks"}},
		},
	}
	if err := universe.ApplyToWorkflow(context.Background(), nil, wf); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if wf.Spec.Provisioning.RunnerImage != dominochain.DefaultJuliaRunnerImage {
		t.Fatalf("expected default julia image, got %q", wf.Spec.Provisioning.RunnerImage)
	}
}

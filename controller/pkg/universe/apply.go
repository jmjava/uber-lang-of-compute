package universe

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
)

// ApplyToWorkflow fills runner image and, for Julia universes, a default
// container runtime when the workflow left those fields empty.
func ApplyToWorkflow(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) error {
	if wf == nil {
		return fmt.Errorf("workflow is nil")
	}

	hasJulia, err := workflowHasJulia(ctx, c, wf)
	if err != nil {
		return err
	}

	if name := strings.TrimSpace(wf.Spec.Routing.Universe); name != "" && c != nil {
		var u kblv1alpha1.PluggableUniverse
		if err := c.Get(ctx, client.ObjectKey{Name: name}, &u); err != nil {
			if !apierrors.IsNotFound(err) {
				return fmt.Errorf("pluggable universe %q: %w", name, err)
			}
			if isContainerRuntime(wf.Spec.Execution.Runtime) {
				return fmt.Errorf("pluggable universe %q not found", name)
			}
		} else {
			if wf.Spec.Provisioning.RunnerImage == "" && u.Spec.ExecutionEngine.RuntimeImage != "" {
				wf.Spec.Provisioning.RunnerImage = u.Spec.ExecutionEngine.RuntimeImage
			}
			if wf.Spec.Execution.Runtime == "" && strings.EqualFold(u.Spec.ExecutionEngine.Type, "julia") && hasJulia {
				wf.Spec.Execution.Runtime = string(kblv1alpha1.DominoChainRuntimeKubernetesInit)
			}
		}
	}

	if hasJulia && wf.Spec.Provisioning.RunnerImage == "" {
		wf.Spec.Provisioning.RunnerImage = dominochain.DefaultJuliaRunnerImage
	}
	if hasJulia && !dominochain.IsJuliaRunnerImage(wf.Spec.Provisioning.RunnerImage) {
		return fmt.Errorf("julia steps require a julia runner image, got %q", wf.Spec.Provisioning.RunnerImage)
	}
	return nil
}

func isContainerRuntime(runtime string) bool {
	if runtime == "" || runtime == string(kblv1alpha1.DominoChainRuntimeLocal) {
		return false
	}
	return true
}

func workflowHasJulia(ctx context.Context, c client.Client, wf *kblv1alpha1.Workflow) (bool, error) {
	for _, d := range wf.Spec.Dominos {
		if strings.HasPrefix(d.Command, "julia:") {
			return true, nil
		}
	}
	if c == nil || len(wf.Spec.DominoRefs) == 0 {
		return false, nil
	}
	for _, name := range wf.Spec.DominoRefs {
		var d kblv1alpha1.Domino
		if err := c.Get(ctx, client.ObjectKey{Namespace: wf.Namespace, Name: name}, &d); err != nil {
			return false, fmt.Errorf("domino ref %q: %w", name, err)
		}
		if strings.HasPrefix(d.Spec.Command, "julia:") {
			return true, nil
		}
	}
	return false, nil
}

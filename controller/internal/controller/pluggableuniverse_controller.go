package controller

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
)

// PluggableUniverseReconciler maintains pluggable universe status.
type PluggableUniverseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kbl.io,resources=pluggableuniverses,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=pluggableuniverses/status,verbs=get;update;patch

func (r *PluggableUniverseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var u kblv1alpha1.PluggableUniverse
	if err := r.Get(ctx, req.NamespacedName, &u); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	phase := kblv1alpha1.PluggableUniversePhaseActive
	message := "universe active"

	engineType := strings.ToLower(strings.TrimSpace(u.Spec.ExecutionEngine.Type))
	if u.Spec.ExecutionEngine.Type == "" {
		phase = kblv1alpha1.PluggableUniversePhaseDegraded
		message = "executionEngine.type is required"
	} else if engineType == "julia" {
		message = "julia subprocess runtime (julia: commands)"
		if u.Spec.ExecutionEngine.RuntimeImage != "" {
			message = fmt.Sprintf("julia runtime image %s", u.Spec.ExecutionEngine.RuntimeImage)
		}
	} else if engineType == "python" {
		message = "python execution planned; use builtin: or julia: commands today"
	}
	if u.Spec.DataLayer.Type == "" {
		phase = kblv1alpha1.PluggableUniversePhaseDegraded
		message = "dataLayer.type is required"
	}

	u.Status.Phase = phase
	u.Status.Message = message
	if err := r.Status().Update(ctx, &u); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("pluggable universe reconciled", "universe", u.Name, "phase", phase)
	return ctrl.Result{}, nil
}

func (r *PluggableUniverseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.PluggableUniverse{}).
		Complete(r)
}

// ValidateUniverseRef checks that a PluggableUniverse exists.
func ValidateUniverseRef(ctx context.Context, c client.Client, ref string) error {
	if ref == "" {
		return fmt.Errorf("empty pluggable universe ref")
	}
	var u kblv1alpha1.PluggableUniverse
	return c.Get(ctx, client.ObjectKey{Name: ref}, &u)
}

package controller

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

const (
	wheelFinalizerName = "kbl.io/computewheel-finalizer"
)

// ComputeWheelReconciler rotates compute contexts through time slices.
type ComputeWheelReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
	Clock     func() time.Time
}

func (r *ComputeWheelReconciler) now() time.Time {
	if r.Clock != nil {
		return r.Clock().UTC()
	}
	return time.Now().UTC()
}

// +kubebuilder:rbac:groups=kbl.io,resources=computewheels,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kbl.io,resources=computewheels/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=computewheels/finalizers,verbs=update
// +kubebuilder:rbac:groups=kbl.io,resources=computecontexts,verbs=get;list;watch
// +kubebuilder:rbac:groups=kbl.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete

func (r *ComputeWheelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var w kblv1alpha1.ComputeWheel
	if err := r.Get(ctx, req.NamespacedName, &w); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if w.DeletionTimestamp != nil {
		return r.finalizeWheel(ctx, &w)
	}

	if !controllerutil.ContainsFinalizer(&w, wheelFinalizerName) {
		controllerutil.AddFinalizer(&w, wheelFinalizerName)
		if err := r.Update(ctx, &w); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if len(w.Spec.Contexts) == 0 {
		return r.failWheel(ctx, &w, fmt.Errorf("spec.contexts must not be empty"))
	}

	interval, err := wheel.ParseInterval(w.Spec.TimeSliceInterval)
	if err != nil {
		return r.failWheel(ctx, &w, err)
	}

	state, err := wheel.StateFromStatus(w.Status)
	if err != nil {
		return r.failWheel(ctx, &w, err)
	}
	if w.Status.CurrentTimeSlice == "" {
		startSlice := ""
		if w.Spec.Schedule != nil {
			startSlice = w.Spec.Schedule.StartTimeSlice
		}
		state, err = wheel.InitialState(startSlice, interval, r.now())
		if err != nil {
			return r.failWheel(ctx, &w, err)
		}
		contextName := w.Spec.Contexts[state.ActiveContextIndex]
		wheel.ApplyState(&w.Status, state, contextName)
		w.Status.Phase = kblv1alpha1.ComputeWheelPhaseRotating
		if err := r.Status().Update(ctx, &w); err != nil {
			return ctrl.Result{}, err
		}
	}

	if w.Status.Phase == kblv1alpha1.ComputeWheelPhaseIdle && w.Spec.MaxRotations > 0 &&
		w.Status.RotationCount >= w.Spec.MaxRotations {
		return ctrl.Result{}, nil
	}

	contextName, err := wheel.ActiveContextName(w.Spec.Contexts, state.ActiveContextIndex)
	if err != nil {
		return r.failWheel(ctx, &w, err)
	}

	var computeCtx kblv1alpha1.ComputeContext
	computeCtxPtr := &computeCtx
	if err := r.Get(ctx, client.ObjectKey{Name: contextName}, &computeCtx); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		computeCtxPtr = nil
		logger.Info("ComputeContext not found; using default store path", "context", contextName)
	}

	wf := wheel.BuildWorkflow(&w, computeCtxPtr, state, r.StoreRoot)
	wfKey := client.ObjectKeyFromObject(wf)

	var existing kblv1alpha1.Workflow
	getErr := r.Get(ctx, wfKey, &existing)
	if apierrors.IsNotFound(getErr) {
		if err := controllerutil.SetControllerReference(&w, wf, r.Scheme); err != nil {
			return r.failWheel(ctx, &w, err)
		}
		if err := r.Create(ctx, wf); err != nil {
			return r.failWheel(ctx, &w, fmt.Errorf("create workflow: %w", err))
		}
		w.Status.Phase = kblv1alpha1.ComputeWheelPhaseProcessing
		w.Status.ActiveWorkflow = wf.Name
		w.Status.Message = fmt.Sprintf("processing %s on %s", contextName, w.Status.CurrentTimeSlice)
		if err := r.Status().Update(ctx, &w); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("created workflow for wheel slot", "wheel", w.Name, "workflow", wf.Name, "context", contextName)
		return ctrl.Result{RequeueAfter: wheel.RequeueDelay(interval)}, nil
	}
	if getErr != nil {
		return ctrl.Result{}, getErr
	}

	w.Status.ActiveWorkflow = existing.Name

	if w.Spec.PreProvisionNext {
		if err := r.preProvisionNext(ctx, &w, state, interval, computeCtxPtr); err != nil {
			logger.Error(err, "pre-provision next slot failed")
		}
	}

	switch existing.Status.Phase {
	case kblv1alpha1.WorkflowPhaseCompleted:
		return r.advance(ctx, &w, state, interval, contextName, &existing)
	case kblv1alpha1.WorkflowPhaseError:
		w.Status.Phase = kblv1alpha1.ComputeWheelPhaseError
		w.Status.Message = existing.Status.Message
		if err := r.Status().Update(ctx, &w); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, fmt.Errorf("workflow %s failed: %s", existing.Name, existing.Status.Message)
	default:
		w.Status.Phase = kblv1alpha1.ComputeWheelPhaseProcessing
		w.Status.Message = fmt.Sprintf("waiting for workflow %s", existing.Name)
		if err := r.Status().Update(ctx, &w); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: wheel.RequeueDelay(interval)}, nil
	}
}

func (r *ComputeWheelReconciler) advance(
	ctx context.Context,
	w *kblv1alpha1.ComputeWheel,
	state wheel.State,
	interval time.Duration,
	contextName string,
	wf *kblv1alpha1.Workflow,
) (ctrl.Result, error) {
	now := r.now()
	w.Status.ProcessedSlots = append(w.Status.ProcessedSlots, kblv1alpha1.ProcessedSlot{
		TimeSlice:   w.Status.CurrentTimeSlice,
		Context:     contextName,
		Workflow:    wf.Name,
		SnapshotID:  wf.Status.SnapshotID,
		CompletedAt: now.Format(time.RFC3339),
	})

	result := wheel.AdvanceAfterCompletion(state, len(w.Spec.Contexts), interval, w.Spec.MaxRotations)
	wheel.ApplyState(&w.Status, result.State, "")
	w.Status.LastRotation = &metav1.Time{Time: now}

	if result.Done {
		w.Status.Phase = kblv1alpha1.ComputeWheelPhaseIdle
		w.Status.Message = fmt.Sprintf("completed %d rotation(s)", w.Status.RotationCount)
		w.Status.ActiveWorkflow = ""
		if err := r.Status().Update(ctx, w); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	nextContext, err := wheel.ActiveContextName(w.Spec.Contexts, result.State.ActiveContextIndex)
	if err != nil {
		return r.failWheel(ctx, w, err)
	}
	wheel.ApplyState(&w.Status, result.State, nextContext)
	w.Status.Phase = kblv1alpha1.ComputeWheelPhaseRotating
	w.Status.ActiveWorkflow = ""
	if result.AdvancedSlice {
		w.Status.Message = fmt.Sprintf("advanced to time slice %s", w.Status.CurrentTimeSlice)
	} else {
		w.Status.Message = fmt.Sprintf("rotated to context %s", nextContext)
	}
	if err := r.Status().Update(ctx, w); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{Requeue: true}, nil
}

func (r *ComputeWheelReconciler) preProvisionNext(
	ctx context.Context,
	w *kblv1alpha1.ComputeWheel,
	state wheel.State,
	interval time.Duration,
	currentCtx *kblv1alpha1.ComputeContext,
) error {
	next := wheel.AdvanceAfterCompletion(state, len(w.Spec.Contexts), interval, 0)
	if next.Done {
		return nil
	}

	nextContextName, err := wheel.ActiveContextName(w.Spec.Contexts, next.State.ActiveContextIndex)
	if err != nil {
		return err
	}

	var nextCtx kblv1alpha1.ComputeContext
	nextCtxPtr := &nextCtx
	if err := r.Get(ctx, client.ObjectKey{Name: nextContextName}, &nextCtx); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		nextCtxPtr = currentCtx
	}

	nextWF := wheel.BuildWorkflow(w, nextCtxPtr, next.State, r.StoreRoot)
	var existing kblv1alpha1.Workflow
	if err := r.Get(ctx, client.ObjectKeyFromObject(nextWF), &existing); err == nil {
		return nil
	} else if !apierrors.IsNotFound(err) {
		return err
	}

	if err := controllerutil.SetControllerReference(w, nextWF, r.Scheme); err != nil {
		return err
	}
	return r.Create(ctx, nextWF)
}

func (r *ComputeWheelReconciler) failWheel(ctx context.Context, w *kblv1alpha1.ComputeWheel, execErr error) (ctrl.Result, error) {
	w.Status.Phase = kblv1alpha1.ComputeWheelPhaseError
	w.Status.Message = execErr.Error()
	if err := r.Status().Update(ctx, w); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, execErr
}

func (r *ComputeWheelReconciler) finalizeWheel(ctx context.Context, w *kblv1alpha1.ComputeWheel) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(w, wheelFinalizerName) {
		controllerutil.RemoveFinalizer(w, wheelFinalizerName)
		if err := r.Update(ctx, w); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler with the manager.
func (r *ComputeWheelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.ComputeWheel{}).
		Owns(&kblv1alpha1.Workflow{}).
		Complete(r)
}

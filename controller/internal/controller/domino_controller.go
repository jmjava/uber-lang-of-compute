package controller

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

const dominoDependencyRequeue = 5 * time.Second

// DominoReconciler executes standalone Domino resources against sealed Snapshots.
type DominoReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
}

// +kubebuilder:rbac:groups=kbl.io,resources=dominos,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=dominos/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=snapshots,verbs=get;list;watch

func (r *DominoReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var dom kblv1alpha1.Domino
	if err := r.Get(ctx, req.NamespacedName, &dom); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if dom.Status.ObservedGeneration == dom.Generation &&
		(dom.Status.Phase == kblv1alpha1.DominoPhaseCompleted || dom.Status.Phase == kblv1alpha1.DominoPhaseCached) {
		return ctrl.Result{}, nil
	}

	var snap kblv1alpha1.Snapshot
	if err := r.Get(ctx, client.ObjectKey{Namespace: dom.Namespace, Name: dom.Spec.SnapshotRef}, &snap); err != nil {
		return r.fail(ctx, &dom, fmt.Errorf("snapshot %q: %w", dom.Spec.SnapshotRef, err))
	}

	if snap.Status.Phase != kblv1alpha1.SnapshotPhaseSealed || snap.Status.SnapshotID == "" {
		dom.Status.Phase = kblv1alpha1.DominoPhasePending
		dom.Status.Message = fmt.Sprintf("waiting for snapshot %q to seal", dom.Spec.SnapshotRef)
		if err := r.Status().Update(ctx, &dom); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: dominoDependencyRequeue}, nil
	}

	priorOutputs, requeue, err := r.loadDependencyOutputs(ctx, &dom, snap.Status.SnapshotID, &snap)
	if err != nil {
		return r.fail(ctx, &dom, err)
	}
	if requeue {
		dom.Status.Phase = kblv1alpha1.DominoPhasePending
		dom.Status.Message = "waiting for dependency dominos"
		if err := r.Status().Update(ctx, &dom); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: dominoDependencyRequeue}, nil
	}

	dom.Status.Phase = kblv1alpha1.DominoPhaseRunning
	dom.Status.Message = "executing domino"
	if err := r.Status().Update(ctx, &dom); err != nil {
		return ctrl.Result{}, err
	}

	backend, err := store.OpenForDomino(ctx, r.Client, &dom, &snap, r.StoreRoot)
	if err != nil {
		return r.fail(ctx, &dom, fmt.Errorf("open store: %w", err))
	}
	defer backend.Close()

	eng := engine.New(backend)
	engineSnap := convert.ToEngineSnapshot(&snap)
	engineDom := convert.ToEngineDomino(&dom)

	entry, err := eng.RunSingle(snap.Status.SnapshotID, engineSnap, engineDom, priorOutputs)
	if err != nil {
		return r.fail(ctx, &dom, err)
	}

	now := metav1.Now()
	phase := kblv1alpha1.DominoPhaseCompleted
	if entry.Reused {
		phase = kblv1alpha1.DominoPhaseCached
	}

	dom.Status.ObservedGeneration = dom.Generation
	dom.Status.Phase = phase
	dom.Status.InputHash = entry.InputHash
	dom.Status.OutputHash = entry.OutputHash
	dom.Status.Reused = entry.Reused
	dom.Status.CompletedAt = &now
	dom.Status.Message = fmt.Sprintf("%s: input=%s output=%s", phase, entry.InputHash[:8], entry.OutputHash[:8])
	dom.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             string(phase),
		Message:            dom.Status.Message,
		LastTransitionTime: now,
	}}

	if err := r.Status().Update(ctx, &dom); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("domino completed", "domino", dom.Name, "phase", phase, "reused", entry.Reused)
	return ctrl.Result{}, nil
}

func (r *DominoReconciler) loadDependencyOutputs(ctx context.Context, dom *kblv1alpha1.Domino, snapshotID string, snap *kblv1alpha1.Snapshot) (map[string]string, bool, error) {
	if len(dom.Spec.DependsOn) == 0 {
		return nil, false, nil
	}

	backend, err := store.OpenForDomino(ctx, r.Client, dom, snap, r.StoreRoot)
	if err != nil {
		return nil, false, err
	}
	defer backend.Close()

	outputs := make(map[string]string, len(dom.Spec.DependsOn))
	for _, dep := range dom.Spec.DependsOn {
		var depDom kblv1alpha1.Domino
		if err := r.Get(ctx, client.ObjectKey{Namespace: dom.Namespace, Name: dep}, &depDom); err != nil {
			return nil, false, fmt.Errorf("dependency domino %q: %w", dep, err)
		}
		if depDom.Status.Phase != kblv1alpha1.DominoPhaseCompleted &&
			depDom.Status.Phase != kblv1alpha1.DominoPhaseCached {
			return nil, true, nil
		}
		out, err := backend.GetDominoOutput(snapshotID, dep)
		if err != nil {
			return nil, false, fmt.Errorf("dependency output %q: %w", dep, err)
		}
		outputs[dep] = out
	}
	return outputs, false, nil
}

func (r *DominoReconciler) fail(ctx context.Context, dom *kblv1alpha1.Domino, err error) (ctrl.Result, error) {
	dom.Status.Phase = kblv1alpha1.DominoPhaseError
	dom.Status.Message = err.Error()
	dom.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "Error",
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
	}}
	_ = r.Status().Update(ctx, dom)
	return ctrl.Result{}, err
}

func (r *DominoReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.Domino{}).
		Complete(r)
}

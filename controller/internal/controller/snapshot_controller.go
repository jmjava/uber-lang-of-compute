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
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// SnapshotReconciler seals Snapshot resources and persists them to node-local stores.
type SnapshotReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
}

// +kubebuilder:rbac:groups=kbl.io,resources=snapshots,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=snapshots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=computecontexts,verbs=get;list;watch

func (r *SnapshotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var snap kblv1alpha1.Snapshot
	if err := r.Get(ctx, req.NamespacedName, &snap); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !snap.Spec.Sealed {
		return r.updateStatus(ctx, &snap, kblv1alpha1.SnapshotPhasePending, "", "waiting for spec.sealed=true")
	}

	if snap.Status.ObservedGeneration == snap.Generation &&
		snap.Status.Phase == kblv1alpha1.SnapshotPhaseSealed &&
		snap.Status.SnapshotID != "" {
		return ctrl.Result{}, nil
	}

	if err := snapshot.Validate(snap.Spec); err != nil {
		return r.fail(ctx, &snap, err)
	}

	snapshotID, err := snapshot.ComputeID(snap.Spec)
	if err != nil {
		if snapshot.IsPathNotReady(err) {
			if _, uerr := r.updateStatus(ctx, &snap, kblv1alpha1.SnapshotPhasePending, "", err.Error()); uerr != nil {
				return ctrl.Result{}, uerr
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return r.fail(ctx, &snap, err)
	}

	data, err := snapshot.MarshalData(snap.Spec)
	if err != nil {
		if snapshot.IsPathNotReady(err) {
			if _, uerr := r.updateStatus(ctx, &snap, kblv1alpha1.SnapshotPhasePending, "", err.Error()); uerr != nil {
				return ctrl.Result{}, uerr
			}
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		return r.fail(ctx, &snap, err)
	}

	backend, err := store.OpenForSnapshot(ctx, r.Client, &snap, r.StoreRoot)
	if err != nil {
		return r.fail(ctx, &snap, fmt.Errorf("open store: %w", err))
	}
	defer backend.Close()

	if err := backend.SaveSnapshot(snapshotID, snap.Spec.TimeSlice, data, true); err != nil {
		return r.fail(ctx, &snap, fmt.Errorf("persist snapshot: %w", err))
	}

	now := metav1.Now()
	snap.Status.ObservedGeneration = snap.Generation
	snap.Status.Phase = kblv1alpha1.SnapshotPhaseSealed
	snap.Status.SnapshotID = snapshotID
	snap.Status.SealedAt = &now
	snap.Status.Message = fmt.Sprintf("sealed snapshot %s", snapshotID)
	snap.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "Sealed",
		Message:            snap.Status.Message,
		LastTransitionTime: now,
	}}

	if err := r.Status().Update(ctx, &snap); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("snapshot sealed", "snapshot", snap.Name, "snapshotID", snapshotID)
	return ctrl.Result{}, nil
}

func (r *SnapshotReconciler) updateStatus(ctx context.Context, snap *kblv1alpha1.Snapshot, phase kblv1alpha1.SnapshotPhase, snapshotID, message string) (ctrl.Result, error) {
	snap.Status.Phase = phase
	snap.Status.SnapshotID = snapshotID
	snap.Status.Message = message
	if err := r.Status().Update(ctx, snap); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SnapshotReconciler) fail(ctx context.Context, snap *kblv1alpha1.Snapshot, err error) (ctrl.Result, error) {
	snap.Status.Phase = kblv1alpha1.SnapshotPhaseError
	snap.Status.Message = err.Error()
	snap.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "Error",
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
	}}
	_ = r.Status().Update(ctx, snap)
	return ctrl.Result{}, err
}

func (r *SnapshotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.Snapshot{}).
		Complete(r)
}

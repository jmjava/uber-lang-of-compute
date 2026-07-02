package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/replica"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// ReadReplicaReconciler materializes read-only snapshot copies in target universes.
type ReadReplicaReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
}

// +kubebuilder:rbac:groups=kbl.io,resources=readreplicas,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=readreplicas/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=workflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=kbl.io,resources=computecontexts,verbs=get;list;watch

func (r *ReadReplicaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var rr kblv1alpha1.ReadReplica
	if err := r.Get(ctx, req.NamespacedName, &rr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if rr.Status.ObservedGeneration == rr.Generation && rr.Status.Phase == kblv1alpha1.ReadReplicaPhaseReady {
		return ctrl.Result{}, nil
	}

	var wf kblv1alpha1.Workflow
	wfKey := client.ObjectKey{Namespace: rr.Spec.SourceNamespace, Name: rr.Spec.SourceWorkflow}
	if err := r.Get(ctx, wfKey, &wf); err != nil {
		return r.fail(ctx, &rr, fmt.Errorf("source workflow: %w", err))
	}

	targetStore, targetPath, err := r.openTargetStore(ctx, &rr)
	if err != nil {
		return r.fail(ctx, &rr, fmt.Errorf("open target store: %w", err))
	}
	defer targetStore.Close()

	rr.Status.Phase = kblv1alpha1.ReadReplicaPhaseMaterializing
	mode := rr.Spec.ReplicationMode
	if mode == "" {
		mode = kblv1alpha1.ReplicationModeDirect
	}
	rr.Status.Message = fmt.Sprintf("materializing via %s", mode)
	if err := r.Status().Update(ctx, &rr); err != nil {
		return ctrl.Result{}, err
	}

	var dominoCount int
	switch mode {
	case kblv1alpha1.ReplicationModeCDC:
		consumer := cdcConsumerForReplica(&rr)
		defer consumer.Close()
		readCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		progress, err := cdc.SyncFromConsumer(readCtx, consumer, targetStore, rr.Spec.SourceSnapshotID, wf.Spec.Execution.Chain)
		if err != nil {
			rr.Status.Phase = kblv1alpha1.ReadReplicaPhasePending
			rr.Status.Message = fmt.Sprintf("waiting for cdc events: %v", err)
			_ = r.Status().Update(ctx, &rr)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
		}
		dominoCount = progress.DominoCount
	default:
		sourceStore, err := store.OpenForWorkflow(ctx, r.Client, &wf, r.StoreRoot)
		if err != nil {
			return r.fail(ctx, &rr, fmt.Errorf("open source store: %w", err))
		}
		defer sourceStore.Close()

		result, err := replica.Materialize(replica.MaterializeConfig{
			SnapshotID:  rr.Spec.SourceSnapshotID,
			DominoChain: wf.Spec.Execution.Chain,
			Source:      sourceStore,
			Target:      targetStore,
		})
		if err != nil {
			return r.fail(ctx, &rr, err)
		}
		dominoCount = result.DominoCount
	}

	now := metav1.Now()
	rr.Status.ObservedGeneration = rr.Generation
	rr.Status.Phase = kblv1alpha1.ReadReplicaPhaseReady
	rr.Status.DominoCount = dominoCount
	rr.Status.TargetStorePath = targetPath
	rr.Status.MaterializedAt = &now
	rr.Status.Message = fmt.Sprintf("%s: snapshot %s with %d dominos to %s",
		mode, rr.Spec.SourceSnapshotID, dominoCount, targetPath)
	rr.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "Materialized",
		Message:            rr.Status.Message,
		LastTransitionTime: now,
	}}

	if err := r.Status().Update(ctx, &rr); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("read replica materialized",
		"replica", rr.Name,
		"mode", mode,
		"snapshot", rr.Spec.SourceSnapshotID,
		"universe", rr.Spec.TargetUniverse,
		"dominos", dominoCount,
	)
	return ctrl.Result{}, nil
}

func (r *ReadReplicaReconciler) openTargetStore(ctx context.Context, rr *kblv1alpha1.ReadReplica) (store.Backend, string, error) {
	root := r.StoreRoot
	if root == "" {
		root = "/var/kbl/store"
	}

	cfg := store.ResolveConfig{
		ComputeContextRef: rr.Spec.TargetComputeContextRef,
		Namespace:         rr.Namespace,
		StoreRoot:         root,
	}

	path := filepath.Join(root, rr.Namespace, "replicas", rr.Spec.TargetUniverse+".db")
	if rr.Spec.TargetComputeContextRef == "" {
		cfg.StorePath = path
	}

	backend, err := store.OpenResolved(ctx, r.Client, cfg)
	if err != nil {
		return nil, "", err
	}

	if cfg.StoreType == store.TypeTSDB || strings.HasPrefix(cfg.StoreEndpoint, "http") {
		if cfg.StoreEndpoint != "" {
			path = cfg.StoreEndpoint
		}
	} else if cfg.StorePath != "" {
		path = cfg.StorePath
	}
	return backend, path, nil
}

func (r *ReadReplicaReconciler) fail(ctx context.Context, rr *kblv1alpha1.ReadReplica, err error) (ctrl.Result, error) {
	_ = r.failStatus(ctx, rr, err)
	return ctrl.Result{RequeueAfter: 30 * time.Second}, err
}

func (r *ReadReplicaReconciler) failStatus(ctx context.Context, rr *kblv1alpha1.ReadReplica, err error) error {
	rr.Status.Phase = kblv1alpha1.ReadReplicaPhaseError
	rr.Status.Message = err.Error()
	rr.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionFalse,
		Reason:             "Error",
		Message:            err.Error(),
		LastTransitionTime: metav1.Now(),
	}}
	return r.Status().Update(ctx, rr)
}

func (r *ReadReplicaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.ReadReplica{}).
		Complete(r)
}

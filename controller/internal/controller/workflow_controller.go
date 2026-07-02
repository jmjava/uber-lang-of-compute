package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

const (
	finalizerName    = "kbl.io/workflow-finalizer"
	replayConfigKey  = "replay.json"
	conditionReady   = "Ready"
	conditionFailed  = "Failed"
)

// WorkflowReconciler reconciles Workflow resources by executing domino chains.
type WorkflowReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
}

// +kubebuilder:rbac:groups=kbl.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kbl.io,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=workflows/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var wf kblv1alpha1.Workflow
	if err := r.Get(ctx, req.NamespacedName, &wf); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if wf.DeletionTimestamp != nil {
		return r.finalize(ctx, &wf)
	}

	if !controllerutil.ContainsFinalizer(&wf, finalizerName) {
		controllerutil.AddFinalizer(&wf, finalizerName)
		if err := r.Update(ctx, &wf); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if wf.Status.ObservedGeneration == wf.Generation && wf.Status.Phase == kblv1alpha1.WorkflowPhaseCompleted {
		return ctrl.Result{}, nil
	}

	return r.execute(ctx, &wf, logger)
}

func (r *WorkflowReconciler) execute(ctx context.Context, wf *kblv1alpha1.Workflow, logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(err error, msg string, keysAndValues ...interface{})
}) (ctrl.Result, error) {
	wf.Status.Phase = kblv1alpha1.WorkflowPhaseRunning
	wf.Status.Message = "executing domino chain"
	if err := r.Status().Update(ctx, wf); err != nil {
		return ctrl.Result{}, err
	}

	storePath := wf.Spec.Provisioning.StorePath
	if storePath == "" {
		root := r.StoreRoot
		if root == "" {
			root = "/var/kbl/store"
		}
		storePath = filepath.Join(root, wf.Namespace, wf.Name+".db")
	}

	s, err := store.Open(storePath)
	if err != nil {
		return r.fail(ctx, wf, fmt.Errorf("open store: %w", err))
	}
	defer s.Close()

	eng := engine.New(s)
	engineWF := convert.ToEngineWorkflow(wf)
	result, err := eng.Run(engineWF)
	if err != nil {
		return r.fail(ctx, wf, err)
	}

	reused := 0
	dominoResults := make([]kblv1alpha1.DominoResult, len(result.Entries))
	for i, e := range result.Entries {
		if e.Reused {
			reused++
		}
		dominoResults[i] = kblv1alpha1.DominoResult{
			DominoID:   e.DominoID,
			InputHash:  e.InputHash,
			OutputHash: e.OutputHash,
			Reused:     e.Reused,
		}
	}

	replayRef, err := r.writeReplayLog(ctx, wf, result)
	if err != nil {
		logger.Error(err, "failed to write replay log ConfigMap")
	}

	now := metav1.NewTime(time.Now().UTC())
	wf.Status.ObservedGeneration = wf.Generation
	wf.Status.Phase = kblv1alpha1.WorkflowPhaseCompleted
	wf.Status.SnapshotID = result.SnapshotID
	wf.Status.DominoCount = len(result.Entries)
	wf.Status.ReusedCount = reused
	wf.Status.RecomputedCount = len(result.Entries) - reused
	wf.Status.LastRunTime = &now
	wf.Status.ReplayLogRef = replayRef
	wf.Status.DominoResults = dominoResults
	wf.Status.Message = fmt.Sprintf("completed: %d dominos, %d reused", len(result.Entries), reused)
	wf.Status.Conditions = []metav1.Condition{
		{
			Type:               conditionReady,
			Status:             metav1.ConditionTrue,
			Reason:             "ChainCompleted",
			Message:            wf.Status.Message,
			LastTransitionTime: now,
			ObservedGeneration: wf.Generation,
		},
	}

	if err := r.Status().Update(ctx, wf); err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("workflow completed",
		"workflow", wf.Name,
		"snapshotID", result.SnapshotID,
		"dominos", len(result.Entries),
		"reused", reused,
	)

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) fail(ctx context.Context, wf *kblv1alpha1.Workflow, execErr error) (ctrl.Result, error) {
	now := metav1.NewTime(time.Now().UTC())
	wf.Status.Phase = kblv1alpha1.WorkflowPhaseError
	wf.Status.Message = execErr.Error()
	wf.Status.Conditions = []metav1.Condition{
		{
			Type:               conditionFailed,
			Status:             metav1.ConditionTrue,
			Reason:             "ExecutionFailed",
			Message:            execErr.Error(),
			LastTransitionTime: now,
			ObservedGeneration: wf.Generation,
		},
	}
	if err := r.Status().Update(ctx, wf); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, execErr
}

func (r *WorkflowReconciler) writeReplayLog(ctx context.Context, wf *kblv1alpha1.Workflow, result *types.RunResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	cmName := wf.Name + "-replay"
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: wf.Namespace,
			Labels: map[string]string{
				"kbl.io/workflow": wf.Name,
				"app.kubernetes.io/managed-by": "kbl-controller",
			},
		},
		Data: map[string]string{
			replayConfigKey: string(data),
		},
	}

	if err := controllerutil.SetControllerReference(wf, cm, r.Scheme); err != nil {
		return "", err
	}

	existing := &corev1.ConfigMap{}
	err = r.Get(ctx, client.ObjectKeyFromObject(cm), existing)
	if apierrors.IsNotFound(err) {
		if err := r.Create(ctx, cm); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else {
		existing.Data = cm.Data
		existing.Labels = cm.Labels
		if err := r.Update(ctx, existing); err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("configmap/%s/%s", wf.Namespace, cmName), nil
}

func (r *WorkflowReconciler) finalize(ctx context.Context, wf *kblv1alpha1.Workflow) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(wf, finalizerName) {
		controllerutil.RemoveFinalizer(wf, finalizerName)
		if err := r.Update(ctx, wf); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the reconciler with the manager.
func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.Workflow{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

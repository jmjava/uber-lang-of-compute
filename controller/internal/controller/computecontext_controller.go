package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// ComputeContextReconciler maintains node-local store endpoints for compute contexts.
type ComputeContextReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=kbl.io,resources=computecontexts,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=computecontexts/status,verbs=get;update;patch

func (r *ComputeContextReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var ctxCR kblv1alpha1.ComputeContext
	if err := r.Get(ctx, req.NamespacedName, &ctxCR); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	storeType := ctxCR.Spec.StoreType
	if storeType == "" {
		storeType = string(store.TypeSQLite)
	}

	endpoint := ctxCR.Spec.StoreEndpoint
	if storeType == string(store.TypeTSDB) && endpoint == "" {
		endpoint = store.DefaultTSDBEndpoint
	}

	phase := kblv1alpha1.ComputeContextPhaseReady
	snapshots, cacheEntries := 0, 0
	message := "ready"

	if storeType == string(store.TypeTSDB) {
		if err := pingTSDB(endpoint); err != nil {
			phase = kblv1alpha1.ComputeContextPhaseDegraded
			message = fmt.Sprintf("tsdb unreachable: %v", err)
			logger.Info("TSDB health check failed", "context", ctxCR.Name, "endpoint", endpoint, "error", err)
			if err := r.updateStatus(ctx, &ctxCR, phase, endpoint, snapshots, cacheEntries, message); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		if stats, err := fetchTSDBStats(endpoint); err == nil {
			snapshots = stats.Snapshots
			cacheEntries = stats.MemoEntries
		}
	}

	if err := r.updateStatus(ctx, &ctxCR, phase, endpoint, snapshots, cacheEntries, message); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil
}

func (r *ComputeContextReconciler) updateStatus(ctx context.Context, ctxCR *kblv1alpha1.ComputeContext, phase kblv1alpha1.ComputeContextPhase, endpoint string, snapshots, cache int, message string) error {
	ctxCR.Status.Phase = phase
	ctxCR.Status.StoreEndpoint = endpoint
	ctxCR.Status.SnapshotCount = snapshots
	ctxCR.Status.CacheEntries = cache
	status := metav1.ConditionFalse
	if phase == kblv1alpha1.ComputeContextPhaseReady {
		status = metav1.ConditionTrue
	}
	ctxCR.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             status,
		Reason:             string(phase),
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}}
	return r.Status().Update(ctx, ctxCR)
}

func pingTSDB(endpoint string) error {
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(endpoint + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}

type tsdbStats struct {
	Snapshots   int `json:"snapshots"`
	MemoEntries int `json:"memo_entries"`
}

func fetchTSDBStats(endpoint string) (*tsdbStats, error) {
	c := &http.Client{Timeout: 3 * time.Second}
	resp, err := c.Get(endpoint + "/v1/stats")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	var stats tsdbStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// SetupWithManager registers the reconciler.
func (r *ComputeContextReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.ComputeContext{}).
		Complete(r)
}

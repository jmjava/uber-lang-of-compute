package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/dominochain"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

const dominoChainFinalizer = "kbl.io/dominochain-finalizer"

// DominoChainReconciler reconciles in-cluster domino chains.
type DominoChainReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	StoreRoot string
	Builder   dominochain.Builder
}

// Reconcile implements the domino chain lifecycle.
func (r *DominoChainReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var chain kblv1alpha1.DominoChain
	if err := r.Get(ctx, req.NamespacedName, &chain); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if chain.DeletionTimestamp != nil {
		return r.finalizeChain(ctx, &chain)
	}

	if !controllerutil.ContainsFinalizer(&chain, dominoChainFinalizer) {
		controllerutil.AddFinalizer(&chain, dominoChainFinalizer)
		if err := r.Update(ctx, &chain); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if chain.Status.Phase == kblv1alpha1.DominoChainPhaseCompleted {
		return ctrl.Result{}, nil
	}

	if len(chain.Spec.Steps) == 0 {
		return r.failChain(ctx, &chain, fmt.Errorf("spec.steps must not be empty"))
	}

	runtime := chain.Spec.Runtime
	if runtime == "" {
		runtime = kblv1alpha1.DominoChainRuntimeKubernetesInit
	}

	snapshotJSON, err := dominochain.SnapshotJSON(chain.Spec.Snapshot)
	if err != nil {
		return r.failChain(ctx, &chain, err)
	}

	cm := r.builder().SnapshotConfigMap(&chain, snapshotJSON)
	if err := r.ensureConfigMap(ctx, &chain, cm); err != nil {
		return r.failChain(ctx, &chain, err)
	}

	switch runtime {
	case kblv1alpha1.DominoChainRuntimeOpenKruise:
		return r.reconcileOpenKruise(ctx, &chain, logger)
	default:
		return r.reconcileInitChain(ctx, &chain, logger)
	}
}

func (r *DominoChainReconciler) reconcileInitChain(ctx context.Context, chain *kblv1alpha1.DominoChain, logger interface {
	Info(msg string, keysAndValues ...interface{})
}) (ctrl.Result, error) {
	pod := r.builder().BuildInitChainPod(chain)
	if err := r.ensurePod(ctx, chain, pod); err != nil {
		return r.failChain(ctx, chain, err)
	}

	chain.Status.PodName = pod.Name
	chain.Status.Phase = kblv1alpha1.DominoChainPhaseRunning

	var live corev1.Pod
	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), &live); err != nil {
		return ctrl.Result{}, err
	}

	if done, err := r.initChainComplete(&live, len(chain.Spec.Steps)); err != nil {
		return r.failChain(ctx, chain, err)
	} else if done {
		return r.completeChain(ctx, chain, logger)
	}

	if live.Status.Phase == corev1.PodFailed {
		return r.failChain(ctx, chain, fmt.Errorf("pod %s failed", live.Name))
	}

	if err := r.Status().Update(ctx, chain); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *DominoChainReconciler) reconcileOpenKruise(ctx context.Context, chain *kblv1alpha1.DominoChain, logger interface {
	Info(msg string, keysAndValues ...interface{})
}) (ctrl.Result, error) {
	pod := r.builder().BuildOpenKruisePod(chain)
	if err := r.ensurePod(ctx, chain, pod); err != nil {
		return r.failChain(ctx, chain, err)
	}

	chain.Status.PodName = pod.Name
	chain.Status.Phase = kblv1alpha1.DominoChainPhaseRunning

	var live corev1.Pod
	if err := r.Get(ctx, client.ObjectKeyFromObject(pod), &live); err != nil {
		return ctrl.Result{}, err
	}

	step := chain.Status.ActiveStep
	if step >= len(chain.Spec.Steps) {
		return r.completeChain(ctx, chain, logger)
	}

	crrName := fmt.Sprintf("%s-slot-%d", chain.Name, step)
	var crr unstructured.Unstructured
	crr.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps.kruise.io",
		Version: "v1alpha1",
		Kind:    "ContainerRecreateRequest",
	})
	err := r.Get(ctx, client.ObjectKey{Namespace: chain.Namespace, Name: crrName}, &crr)
	if apierrors.IsNotFound(err) {
		newCRR := dominochain.ContainerRecreateRequest(chain, r.builderPtr(), step)
		if err := controllerutil.SetControllerReference(chain, newCRR, r.Scheme); err != nil {
			return r.failChain(ctx, chain, err)
		}
		if err := r.Create(ctx, newCRR); err != nil {
			if meta.IsNoMatchError(err) {
				return r.failChain(ctx, chain, fmt.Errorf("openkruise CRD not installed: %w", err))
			}
			return r.failChain(ctx, chain, fmt.Errorf("create CRR: %w", err))
		}
		logger.Info("created ContainerRecreateRequest", "chain", chain.Name, "step", step)
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	if dominochain.IsCRRComplete(&crr) || r.stepContainerSucceeded(&live, chain, step) {
		chain.Status.StepResults = appendStepResult(chain.Status.StepResults, chain, step, "Completed")
		chain.Status.ActiveStep = step + 1

		if chain.Status.ActiveStep >= len(chain.Spec.Steps) {
			return r.completeChain(ctx, chain, logger)
		}

		chain.Status.Message = fmt.Sprintf("advanced to step %d", chain.Status.ActiveStep)
		if err := r.Status().Update(ctx, chain); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.Status().Update(ctx, chain); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
}

func (r *DominoChainReconciler) completeChain(ctx context.Context, chain *kblv1alpha1.DominoChain, logger interface {
	Info(msg string, keysAndValues ...interface{})
}) (ctrl.Result, error) {
	storePath := chain.Spec.StorePath
	if storePath == "" {
		root := r.StoreRoot
		if root == "" {
			root = "/var/kbl/store"
		}
		storePath = filepath.Join(root, chain.Namespace, chain.Name+".db")
	}

	s, err := store.Open(storePath)
	if err != nil {
		return r.failChain(ctx, chain, fmt.Errorf("open store: %w", err))
	}
	defer s.Close()

	eng := engine.New(s)
	wf := dominoChainToWorkflow(chain)
	result, err := eng.Run(wf)
	if err != nil {
		return r.failChain(ctx, chain, err)
	}

	finalHash, _ := hash.Compute(result.FinalOutput)
	chain.Status.Phase = kblv1alpha1.DominoChainPhaseCompleted
	chain.Status.SnapshotID = result.SnapshotID
	chain.Status.FinalOutputHash = finalHash
	chain.Status.Message = fmt.Sprintf("completed %d steps", len(chain.Spec.Steps))
	chain.Status.StepResults = make([]kblv1alpha1.StepResult, len(result.Entries))
	for i, e := range result.Entries {
		chain.Status.StepResults[i] = kblv1alpha1.StepResult{
			Name:       e.DominoID,
			Index:      i,
			OutputHash: e.OutputHash,
			Phase:      stepPhase(e.Reused),
		}
	}
	chain.Status.Conditions = []metav1.Condition{{
		Type:               conditionReady,
		Status:             metav1.ConditionTrue,
		Reason:             "ChainCompleted",
		Message:            chain.Status.Message,
		LastTransitionTime: metav1.Now(),
	}}

	if err := r.Status().Update(ctx, chain); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("domino chain completed", "chain", chain.Name, "snapshotID", result.SnapshotID)
	return ctrl.Result{}, nil
}

func (r *DominoChainReconciler) initChainComplete(pod *corev1.Pod, steps int) (bool, error) {
	if pod.Status.Phase == corev1.PodSucceeded {
		return true, nil
	}
	if len(pod.Status.InitContainerStatuses) < steps {
		return false, nil
	}
	for i := 0; i < steps; i++ {
		st := pod.Status.InitContainerStatuses[i].State
		if st.Terminated == nil {
			return false, nil
		}
		if st.Terminated.ExitCode != 0 {
			return false, fmt.Errorf("init container %d exit code %d", i, st.Terminated.ExitCode)
		}
	}
	// All inits done; main container may still be pause — treat as complete.
	return pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded, nil
}

func (r *DominoChainReconciler) stepContainerSucceeded(pod *corev1.Pod, chain *kblv1alpha1.DominoChain, step int) bool {
	name := dominochain.StepContainerName(chain, step)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == name && cs.State.Terminated != nil && cs.State.Terminated.ExitCode == 0 {
			return true
		}
	}
	return false
}

func (r *DominoChainReconciler) ensureConfigMap(ctx context.Context, owner *kblv1alpha1.DominoChain, cm *corev1.ConfigMap) error {
	if err := controllerutil.SetControllerReference(owner, cm, r.Scheme); err != nil {
		return err
	}
	var existing corev1.ConfigMap
	err := r.Get(ctx, client.ObjectKeyFromObject(cm), &existing)
	if apierrors.IsNotFound(err) {
		return r.Create(ctx, cm)
	}
	if err != nil {
		return err
	}
	existing.Data = cm.Data
	return r.Update(ctx, &existing)
}

func (r *DominoChainReconciler) ensurePod(ctx context.Context, owner *kblv1alpha1.DominoChain, pod *corev1.Pod) error {
	if err := controllerutil.SetControllerReference(owner, pod, r.Scheme); err != nil {
		return err
	}
	var existing corev1.Pod
	err := r.Get(ctx, client.ObjectKeyFromObject(pod), &existing)
	if apierrors.IsNotFound(err) {
		return r.Create(ctx, pod)
	}
	return err
}

func (r *DominoChainReconciler) failChain(ctx context.Context, chain *kblv1alpha1.DominoChain, execErr error) (ctrl.Result, error) {
	chain.Status.Phase = kblv1alpha1.DominoChainPhaseError
	chain.Status.Message = execErr.Error()
	_ = r.Status().Update(ctx, chain)
	return ctrl.Result{}, execErr
}

func (r *DominoChainReconciler) finalizeChain(ctx context.Context, chain *kblv1alpha1.DominoChain) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(chain, dominoChainFinalizer) {
		controllerutil.RemoveFinalizer(chain, dominoChainFinalizer)
		if err := r.Update(ctx, chain); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *DominoChainReconciler) builder() *dominochain.Builder {
	b := r.Builder
	return &b
}

func (r *DominoChainReconciler) builderPtr() *dominochain.Builder {
	return r.builder()
}

// SetupWithManager registers the reconciler.
func (r *DominoChainReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.DominoChain{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

func dominoChainToWorkflow(chain *kblv1alpha1.DominoChain) *types.Workflow {
	dominos := make([]types.Domino, len(chain.Spec.Steps))
	snapshotRef := chain.Name + "-snapshot"
	for i, s := range chain.Spec.Steps {
		dominos[i] = types.Domino{
			Metadata: types.ObjectMeta{Name: s.Name},
			Spec: types.DominoSpec{
				SnapshotRef: snapshotRef,
				Command:     s.Command,
			},
		}
		if i > 0 {
			dominos[i].Spec.DependsOn = []string{chain.Spec.Steps[i-1].Name}
			dominos[i].Spec.Inputs = []types.DominoInput{{FromDomino: chain.Spec.Steps[i-1].Name}}
		}
	}
	chainNames := make([]string, len(chain.Spec.Steps))
	for i, s := range chain.Spec.Steps {
		chainNames[i] = s.Name
	}
	return &types.Workflow{
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				Metadata: types.ObjectMeta{Name: snapshotRef},
				Spec: types.SnapshotSpec{
					TimeSlice: chain.Spec.Snapshot.TimeSlice,
					Source:    types.SnapshotSource{Inline: chain.Spec.Snapshot.Source.Inline},
					Sealed:    chain.Spec.Snapshot.Sealed,
				},
			},
			Dominos: dominos,
			Execution: types.ExecutionConfig{
				Chain:         chainNames,
				Deterministic: true,
			},
			Provisioning: types.ProvisioningConfig{StorePath: chain.Spec.StorePath},
		},
	}
}

func appendStepResult(existing []kblv1alpha1.StepResult, chain *kblv1alpha1.DominoChain, step int, phase string) []kblv1alpha1.StepResult {
	name := chain.Spec.Steps[step].Name
	for _, r := range existing {
		if r.Index == step {
			return existing
		}
	}
	return append(existing, kblv1alpha1.StepResult{Name: name, Index: step, Phase: phase})
}

func stepPhase(reused bool) string {
	if reused {
		return "Cached"
	}
	return "Completed"
}

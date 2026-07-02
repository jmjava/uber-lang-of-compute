package controller

import (
	"context"
	"fmt"
	"sync"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
)

const multiverseFinalizer = "kbl.io/multiverse-finalizer"

// MultiverseReconciler routes snapshot events across pluggable universes.
type MultiverseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Bus    events.Bus

	mu          sync.Mutex
	subscribed  map[string]bool
}

// +kubebuilder:rbac:groups=kbl.io,resources=multiverses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kbl.io,resources=multiverses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=kbl.io,resources=multiverses/finalizers,verbs=update
// +kubebuilder:rbac:groups=kbl.io,resources=pluggableuniverses,verbs=get;list;watch
// +kubebuilder:rbac:groups=kbl.io,resources=readreplicas,verbs=get;list;watch;create;update;patch

func (r *MultiverseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var mv kblv1alpha1.Multiverse
	if err := r.Get(ctx, req.NamespacedName, &mv); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if mv.DeletionTimestamp != nil {
		return r.finalizeMultiverse(ctx, &mv)
	}

	if !controllerutil.ContainsFinalizer(&mv, multiverseFinalizer) {
		controllerutil.AddFinalizer(&mv, multiverseFinalizer)
		if err := r.Update(ctx, &mv); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	for _, u := range mv.Spec.Universes {
		if err := ValidateUniverseRef(ctx, r.Client, u.PluggableUniverseRef); err != nil {
			return r.failMultiverse(ctx, &mv, fmt.Errorf("universe %q ref %q: %w", u.Name, u.PluggableUniverseRef, err))
		}
	}

	kafkaOK := false
	if mv.Spec.Sync != nil && mv.Spec.Sync.Enabled && mv.Spec.Sync.Kafka != nil {
		if err := events.Ping(ctx, mv.Spec.Sync.Kafka.Brokers); err != nil {
			mv.Status.Phase = kblv1alpha1.MultiversePhaseDegraded
			mv.Status.KafkaConnected = false
			mv.Status.Message = fmt.Sprintf("kafka unreachable: %v", err)
		} else {
			kafkaOK = true
			mv.Status.KafkaConnected = true
		}
	}

	if mv.Status.Phase != kblv1alpha1.MultiversePhaseDegraded {
		mv.Status.Phase = kblv1alpha1.MultiversePhaseActive
		mv.Status.Message = fmt.Sprintf("routing %d universes", len(mv.Spec.Universes))
	}
	mv.Status.UniverseCount = len(mv.Spec.Universes)
	mv.Status.Conditions = []metav1.Condition{{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             string(mv.Status.Phase),
		Message:            mv.Status.Message,
		LastTransitionTime: metav1.Now(),
	}}

	if err := r.Status().Update(ctx, &mv); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ensureSubscription(ctx, &mv); err != nil {
		logger.Error(err, "multiverse subscription failed")
	}

	logger.Info("multiverse reconciled", "multiverse", mv.Name, "kafka", kafkaOK)
	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

func (r *MultiverseReconciler) ensureSubscription(ctx context.Context, mv *kblv1alpha1.Multiverse) error {
	bus := r.Bus
	if bus == nil {
		bus = events.NewNoopBus()
	}

	r.mu.Lock()
	key := mv.Namespace + "/" + mv.Name
	if r.subscribed == nil {
		r.subscribed = make(map[string]bool)
	}
	if r.subscribed[key] {
		r.mu.Unlock()
		return nil
	}
	r.subscribed[key] = true
	r.mu.Unlock()

	router := routing.NewRouter(mv.Spec)
	mvName := mv.Name
	mvNS := mv.Namespace

	return bus.Subscribe(ctx, func(hctx context.Context, evt events.SnapshotEvent) error {
		if evt.Multiverse != "" && evt.Multiverse != mvName {
			return nil
		}
		target, err := router.Resolve(evt)
		if err != nil {
			return nil
		}

		var latest kblv1alpha1.Multiverse
		if err := r.Get(hctx, client.ObjectKey{Namespace: mvNS, Name: mvName}, &latest); err != nil {
			return err
		}

		record := kblv1alpha1.RoutedEvent{
			EventID:           evt.EventID,
			SnapshotID:        evt.SnapshotID,
			TimeSlice:         evt.TimeSlice,
			SourceWorkflow:    evt.Namespace + "/" + evt.Workflow,
			TargetUniverse:    target.Universe,
			ComputeContextRef: target.ComputeContextRef,
			RoutedAt:          time.Now().UTC().Format(time.RFC3339),
		}
		latest.Status.RoutedEvents = appendRoutedEvent(latest.Status.RoutedEvents, record, 50)
		if err := r.Status().Update(hctx, &latest); err != nil {
			return err
		}
		return r.ensureReadReplica(hctx, &latest, evt, target)
	})
}

func appendRoutedEvent(existing []kblv1alpha1.RoutedEvent, evt kblv1alpha1.RoutedEvent, max int) []kblv1alpha1.RoutedEvent {
	for _, e := range existing {
		if e.EventID == evt.EventID {
			return existing
		}
	}
	out := append(existing, evt)
	if len(out) > max {
		out = out[len(out)-max:]
	}
	return out
}

func (r *MultiverseReconciler) failMultiverse(ctx context.Context, mv *kblv1alpha1.Multiverse, err error) (ctrl.Result, error) {
	mv.Status.Phase = kblv1alpha1.MultiversePhaseDegraded
	mv.Status.Message = err.Error()
	_ = r.Status().Update(ctx, mv)
	return ctrl.Result{}, err
}

func (r *MultiverseReconciler) finalizeMultiverse(ctx context.Context, mv *kblv1alpha1.Multiverse) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(mv, multiverseFinalizer) {
		controllerutil.RemoveFinalizer(mv, multiverseFinalizer)
		if err := r.Update(ctx, mv); err != nil {
			return ctrl.Result{}, err
		}
	}
	r.mu.Lock()
	delete(r.subscribed, mv.Namespace+"/"+mv.Name)
	r.mu.Unlock()
	return ctrl.Result{}, nil
}

func (r *MultiverseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kblv1alpha1.Multiverse{}).
		Complete(r)
}

// PublishSnapshotEvent emits a snapshot completion event to the bus and optional Kafka sync topic.
func PublishSnapshotEvent(ctx context.Context, bus events.Bus, mv *kblv1alpha1.Multiverse, evt events.SnapshotEvent) error {
	if bus == nil {
		return nil
	}
	if mv != nil && mv.Spec.Sync != nil && mv.Spec.Sync.Enabled && mv.Spec.Sync.Kafka != nil {
		kbus, err := events.NewKafkaBus(events.KafkaConfig{
			Brokers: mv.Spec.Sync.Kafka.Brokers,
			Topic:   mv.Spec.Sync.Kafka.Topic,
			GroupID: mv.Spec.Sync.Kafka.GroupID,
		})
		if err == nil {
			defer kbus.Close()
			_ = kbus.Publish(ctx, evt)
		}
	}
	return bus.Publish(ctx, evt)
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies the receiver into out.
func (in *Workflow) DeepCopyInto(out *Workflow) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a deep copy.
func (in *Workflow) DeepCopy() *Workflow {
	if in == nil {
		return nil
	}
	out := new(Workflow)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject implements runtime.Object.
func (in *Workflow) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *WorkflowList) DeepCopyInto(out *WorkflowList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Workflow, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *WorkflowList) DeepCopy() *WorkflowList {
	if in == nil {
		return nil
	}
	out := new(WorkflowList)
	in.DeepCopyInto(out)
	return out
}

func (in *WorkflowList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *WorkflowSpec) DeepCopyInto(out *WorkflowSpec) {
	*out = *in
	in.Snapshot.DeepCopyInto(&out.Snapshot)
	if in.Dominos != nil {
		out.Dominos = make([]DominoSpec, len(in.Dominos))
		for i := range in.Dominos {
			in.Dominos[i].DeepCopyInto(&out.Dominos[i])
		}
	}
	if in.DominoRefs != nil {
		out.DominoRefs = make([]string, len(in.DominoRefs))
		copy(out.DominoRefs, in.DominoRefs)
	}
	if in.Execution.Chain != nil {
		out.Execution.Chain = make([]string, len(in.Execution.Chain))
		copy(out.Execution.Chain, in.Execution.Chain)
	}
	out.Execution.Deterministic = in.Execution.Deterministic
	out.Provisioning = in.Provisioning
	out.Routing = in.Routing
}

func (in *SnapshotSpec) DeepCopyInto(out *SnapshotSpec) {
	*out = *in
	if in.Source.Inline != nil {
		out.Source.Inline = deepCopyMap(in.Source.Inline)
	}
	out.Source.Path = in.Source.Path
	out.Source.URI = in.Source.URI
}

func (in *DominoSpec) DeepCopyInto(out *DominoSpec) {
	*out = *in
	if in.DependsOn != nil {
		out.DependsOn = make([]string, len(in.DependsOn))
		copy(out.DependsOn, in.DependsOn)
	}
	if in.Inputs != nil {
		out.Inputs = make([]DominoInput, len(in.Inputs))
		copy(out.Inputs, in.Inputs)
	}
}

func (in *WorkflowStatus) DeepCopyInto(out *WorkflowStatus) {
	*out = *in
	if in.LastRunTime != nil {
		t := *in.LastRunTime
		out.LastRunTime = &t
	}
	if in.DominoResults != nil {
		out.DominoResults = make([]DominoResult, len(in.DominoResults))
		copy(out.DominoResults, in.DominoResults)
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func deepCopyMap(in map[string]interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = deepCopyValue(v)
	}
	return out
}

func deepCopyValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		return deepCopyMap(val)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, item := range val {
			out[i] = deepCopyValue(item)
		}
		return out
	default:
		return v
	}
}

func (in *ComputeContext) DeepCopyInto(out *ComputeContext) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ComputeContext) DeepCopy() *ComputeContext {
	if in == nil {
		return nil
	}
	out := new(ComputeContext)
	in.DeepCopyInto(out)
	return out
}

func (in *ComputeContext) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ComputeContextList) DeepCopyInto(out *ComputeContextList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ComputeContext, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ComputeContextList) DeepCopy() *ComputeContextList {
	if in == nil {
		return nil
	}
	out := new(ComputeContextList)
	in.DeepCopyInto(out)
	return out
}

func (in *ComputeContextList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ComputeContextStatus) DeepCopyInto(out *ComputeContextStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *ComputeWheel) DeepCopyInto(out *ComputeWheel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ComputeWheel) DeepCopy() *ComputeWheel {
	if in == nil {
		return nil
	}
	out := new(ComputeWheel)
	in.DeepCopyInto(out)
	return out
}

func (in *ComputeWheel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ComputeWheelList) DeepCopyInto(out *ComputeWheelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ComputeWheel, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ComputeWheelList) DeepCopy() *ComputeWheelList {
	if in == nil {
		return nil
	}
	out := new(ComputeWheelList)
	in.DeepCopyInto(out)
	return out
}

func (in *ComputeWheelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ComputeWheelSpec) DeepCopyInto(out *ComputeWheelSpec) {
	*out = *in
	if in.Contexts != nil {
		out.Contexts = make([]string, len(in.Contexts))
		copy(out.Contexts, in.Contexts)
	}
	in.WorkflowTemplate.DeepCopyInto(&out.WorkflowTemplate)
	if in.Schedule != nil {
		s := *in.Schedule
		out.Schedule = &s
	}
	out.MaxRotations = in.MaxRotations
	out.PreProvisionNext = in.PreProvisionNext
}

func (in *WorkflowTemplateSpec) DeepCopyInto(out *WorkflowTemplateSpec) {
	in.Snapshot.DeepCopyInto(&out.Snapshot)
	if in.Dominos != nil {
		out.Dominos = make([]DominoSpec, len(in.Dominos))
		for i := range in.Dominos {
			in.Dominos[i].DeepCopyInto(&out.Dominos[i])
		}
	}
	if in.DominoRefs != nil {
		out.DominoRefs = make([]string, len(in.DominoRefs))
		copy(out.DominoRefs, in.DominoRefs)
	}
	if in.Execution.Chain != nil {
		out.Execution.Chain = make([]string, len(in.Execution.Chain))
		copy(out.Execution.Chain, in.Execution.Chain)
	}
	out.Execution.Deterministic = in.Execution.Deterministic
	out.Provisioning = in.Provisioning
	out.Routing = in.Routing
}

func (in *ComputeWheelStatus) DeepCopyInto(out *ComputeWheelStatus) {
	*out = *in
	if in.LastRotation != nil {
		t := *in.LastRotation
		out.LastRotation = &t
	}
	if in.ProcessedSlots != nil {
		out.ProcessedSlots = make([]ProcessedSlot, len(in.ProcessedSlots))
		copy(out.ProcessedSlots, in.ProcessedSlots)
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *DominoChain) DeepCopyInto(out *DominoChain) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *DominoChain) DeepCopy() *DominoChain {
	if in == nil {
		return nil
	}
	out := new(DominoChain)
	in.DeepCopyInto(out)
	return out
}

func (in *DominoChain) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DominoChainList) DeepCopyInto(out *DominoChainList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]DominoChain, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *DominoChainList) DeepCopy() *DominoChainList {
	if in == nil {
		return nil
	}
	out := new(DominoChainList)
	in.DeepCopyInto(out)
	return out
}

func (in *DominoChainList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DominoChainSpec) DeepCopyInto(out *DominoChainSpec) {
	*out = *in
	in.Snapshot.DeepCopyInto(&out.Snapshot)
	if in.Steps != nil {
		out.Steps = make([]DominoStepSpec, len(in.Steps))
		copy(out.Steps, in.Steps)
	}
	if in.NodeSelector != nil {
		out.NodeSelector = make(map[string]string, len(in.NodeSelector))
		for k, v := range in.NodeSelector {
			out.NodeSelector[k] = v
		}
	}
}

func (in *DominoChainStatus) DeepCopyInto(out *DominoChainStatus) {
	*out = *in
	if in.StepResults != nil {
		out.StepResults = make([]StepResult, len(in.StepResults))
		copy(out.StepResults, in.StepResults)
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *PluggableUniverse) DeepCopyInto(out *PluggableUniverse) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

func (in *PluggableUniverse) DeepCopy() *PluggableUniverse {
	if in == nil {
		return nil
	}
	out := new(PluggableUniverse)
	in.DeepCopyInto(out)
	return out
}

func (in *PluggableUniverse) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *PluggableUniverseList) DeepCopyInto(out *PluggableUniverseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]PluggableUniverse, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *PluggableUniverseList) DeepCopy() *PluggableUniverseList {
	if in == nil {
		return nil
	}
	out := new(PluggableUniverseList)
	in.DeepCopyInto(out)
	return out
}

func (in *PluggableUniverseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *PluggableUniverseSpec) DeepCopyInto(out *PluggableUniverseSpec) {
	*out = *in
	if in.DataLayer.Config != nil {
		out.DataLayer.Config = deepCopyMap(in.DataLayer.Config)
	}
	if in.ProvisioningModel != nil {
		p := *in.ProvisioningModel
		out.ProvisioningModel = &p
	}
}

func (in *Multiverse) DeepCopyInto(out *Multiverse) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *Multiverse) DeepCopy() *Multiverse {
	if in == nil {
		return nil
	}
	out := new(Multiverse)
	in.DeepCopyInto(out)
	return out
}

func (in *Multiverse) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MultiverseList) DeepCopyInto(out *MultiverseList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Multiverse, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *MultiverseList) DeepCopy() *MultiverseList {
	if in == nil {
		return nil
	}
	out := new(MultiverseList)
	in.DeepCopyInto(out)
	return out
}

func (in *MultiverseList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *MultiverseSpec) DeepCopyInto(out *MultiverseSpec) {
	*out = *in
	if in.Universes != nil {
		out.Universes = make([]UniverseRouteSpec, len(in.Universes))
		copy(out.Universes, in.Universes)
	}
	if in.TimeSliceRoutes != nil {
		out.TimeSliceRoutes = make([]TimeSliceRoute, len(in.TimeSliceRoutes))
		copy(out.TimeSliceRoutes, in.TimeSliceRoutes)
	}
	if in.Sync != nil {
		s := *in.Sync
		out.Sync = &s
		if in.Sync.Kafka != nil {
			k := *in.Sync.Kafka
			out.Sync.Kafka = &k
			if in.Sync.Kafka.Brokers != nil {
				out.Sync.Kafka.Brokers = make([]string, len(in.Sync.Kafka.Brokers))
				copy(out.Sync.Kafka.Brokers, in.Sync.Kafka.Brokers)
			}
		}
	}
}

func (in *MultiverseStatus) DeepCopyInto(out *MultiverseStatus) {
	*out = *in
	if in.RoutedEvents != nil {
		out.RoutedEvents = make([]RoutedEvent, len(in.RoutedEvents))
		copy(out.RoutedEvents, in.RoutedEvents)
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *Snapshot) DeepCopyInto(out *Snapshot) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *Snapshot) DeepCopy() *Snapshot {
	if in == nil {
		return nil
	}
	out := new(Snapshot)
	in.DeepCopyInto(out)
	return out
}

func (in *Snapshot) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *SnapshotList) DeepCopyInto(out *SnapshotList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Snapshot, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *SnapshotList) DeepCopy() *SnapshotList {
	if in == nil {
		return nil
	}
	out := new(SnapshotList)
	in.DeepCopyInto(out)
	return out
}

func (in *SnapshotList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *SnapshotStatus) DeepCopyInto(out *SnapshotStatus) {
	*out = *in
	if in.SealedAt != nil {
		t := *in.SealedAt
		out.SealedAt = &t
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *Domino) DeepCopyInto(out *Domino) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *Domino) DeepCopy() *Domino {
	if in == nil {
		return nil
	}
	out := new(Domino)
	in.DeepCopyInto(out)
	return out
}

func (in *Domino) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DominoList) DeepCopyInto(out *DominoList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Domino, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *DominoList) DeepCopy() *DominoList {
	if in == nil {
		return nil
	}
	out := new(DominoList)
	in.DeepCopyInto(out)
	return out
}

func (in *DominoList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *DominoResourceSpec) DeepCopyInto(out *DominoResourceSpec) {
	*out = *in
	if in.DependsOn != nil {
		out.DependsOn = make([]string, len(in.DependsOn))
		copy(out.DependsOn, in.DependsOn)
	}
	if in.Inputs != nil {
		out.Inputs = make([]DominoInput, len(in.Inputs))
		copy(out.Inputs, in.Inputs)
	}
}

func (in *DominoResourceStatus) DeepCopyInto(out *DominoResourceStatus) {
	*out = *in
	if in.CompletedAt != nil {
		t := *in.CompletedAt
		out.CompletedAt = &t
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *ReadReplica) DeepCopyInto(out *ReadReplica) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ReadReplica) DeepCopy() *ReadReplica {
	if in == nil {
		return nil
	}
	out := new(ReadReplica)
	in.DeepCopyInto(out)
	return out
}

func (in *ReadReplica) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ReadReplicaList) DeepCopyInto(out *ReadReplicaList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ReadReplica, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ReadReplicaList) DeepCopy() *ReadReplicaList {
	if in == nil {
		return nil
	}
	out := new(ReadReplicaList)
	in.DeepCopyInto(out)
	return out
}

func (in *ReadReplicaList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ReadReplicaSpec) DeepCopyInto(out *ReadReplicaSpec) {
	*out = *in
	if in.Partitions != nil {
		out.Partitions = make(map[string]string, len(in.Partitions))
		for k, v := range in.Partitions {
			out.Partitions[k] = v
		}
	}
	if in.CDCSync != nil {
		s := *in.CDCSync
		out.CDCSync = &s
		if in.CDCSync.Brokers != nil {
			out.CDCSync.Brokers = make([]string, len(in.CDCSync.Brokers))
			copy(out.CDCSync.Brokers, in.CDCSync.Brokers)
		}
	}
}

func (in *ReadReplicaStatus) DeepCopyInto(out *ReadReplicaStatus) {
	*out = *in
	if in.MaterializedAt != nil {
		t := *in.MaterializedAt
		out.MaterializedAt = &t
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

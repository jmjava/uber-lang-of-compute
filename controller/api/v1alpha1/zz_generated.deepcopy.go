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

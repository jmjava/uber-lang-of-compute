package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComputeContextPhase represents the lifecycle phase of a ComputeContext.
type ComputeContextPhase string

const (
	ComputeContextPhaseReady    ComputeContextPhase = "Ready"
	ComputeContextPhaseDegraded ComputeContextPhase = "Degraded"
	ComputeContextPhaseOffline  ComputeContextPhase = "Offline"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=ctx
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="Store",type=string,JSONPath=`.spec.storeType`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// ComputeContext is a node-associated unit of compute and data locality.
type ComputeContext struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeContextSpec   `json:"spec,omitempty"`
	Status ComputeContextStatus `json:"status,omitempty"`
}

type ComputeContextSpec struct {
	NodeName      string `json:"nodeName"`
	StorePath     string `json:"storePath,omitempty"`
	StoreType     string `json:"storeType,omitempty"`
	StoreEndpoint string `json:"storeEndpoint,omitempty"`
	DataPath      string `json:"dataPath,omitempty"`
}

type ComputeContextStatus struct {
	Phase         ComputeContextPhase `json:"phase,omitempty"`
	StoreEndpoint string              `json:"storeEndpoint,omitempty"`
	SnapshotCount int                 `json:"snapshotCount,omitempty"`
	CacheEntries  int                 `json:"cacheEntries,omitempty"`
	Conditions    []metav1.Condition  `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ComputeContextList contains a list of ComputeContext resources.
type ComputeContextList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeContext `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeContext{}, &ComputeContextList{})
}

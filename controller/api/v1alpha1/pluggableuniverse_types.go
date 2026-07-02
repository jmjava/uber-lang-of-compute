package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PluggableUniversePhase represents universe lifecycle phase.
type PluggableUniversePhase string

const (
	PluggableUniversePhaseActive   PluggableUniversePhase = "Active"
	PluggableUniversePhaseDegraded PluggableUniversePhase = "Degraded"
	PluggableUniversePhaseOffline PluggableUniversePhase = "Offline"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=universe
// +kubebuilder:printcolumn:name="Engine",type=string,JSONPath=`.spec.executionEngine.type`
// +kubebuilder:printcolumn:name="Data",type=string,JSONPath=`.spec.dataLayer.type`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// PluggableUniverse defines a compute environment with its own execution/data/provisioning laws.
type PluggableUniverse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluggableUniverseSpec   `json:"spec,omitempty"`
	Status PluggableUniverseStatus `json:"status,omitempty"`
}

type PluggableUniverseSpec struct {
	DisplayName        string                    `json:"name,omitempty"`
	ExecutionEngine    ExecutionEngineSpec       `json:"executionEngine"`
	DataLayer          DataLayerSpec             `json:"dataLayer"`
	ProvisioningModel  *ProvisioningModelSpec    `json:"provisioningModel,omitempty"`
}

type ExecutionEngineSpec struct {
	Type         string `json:"type,omitempty"`
	RuntimeImage string `json:"runtimeImage,omitempty"`
}

type DataLayerSpec struct {
	Type   string                 `json:"type,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type ProvisioningModelSpec struct {
	Autoscaler bool `json:"autoscaler,omitempty"`
	MinNodes   int  `json:"minNodes,omitempty"`
	MaxNodes   int  `json:"maxNodes,omitempty"`
}

type PluggableUniverseStatus struct {
	Phase         PluggableUniversePhase `json:"phase,omitempty"`
	ContextCount  int                    `json:"contextCount,omitempty"`
	Message       string                 `json:"message,omitempty"`
}

// +kubebuilder:object:root=true

// PluggableUniverseList contains a list of PluggableUniverse resources.
type PluggableUniverseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PluggableUniverse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PluggableUniverse{}, &PluggableUniverseList{})
}

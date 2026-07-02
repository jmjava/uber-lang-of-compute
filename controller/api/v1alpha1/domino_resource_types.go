package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DominoPhase represents standalone domino lifecycle phase.
type DominoPhase string

const (
	DominoPhasePending   DominoPhase = "Pending"
	DominoPhaseRunning   DominoPhase = "Running"
	DominoPhaseCompleted DominoPhase = "Completed"
	DominoPhaseCached    DominoPhase = "Cached"
	DominoPhaseError     DominoPhase = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.snapshotRef`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Reused",type=boolean,JSONPath=`.status.reused`

// Domino is a single deterministic compute step against a sealed Snapshot.
type Domino struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DominoResourceSpec   `json:"spec,omitempty"`
	Status DominoResourceStatus `json:"status,omitempty"`
}

// DominoResourceSpec defines one standalone domino (name comes from metadata.name).
type DominoResourceSpec struct {
	SnapshotRef string        `json:"snapshotRef"`
	Command     string        `json:"command"`
	DependsOn   []string      `json:"dependsOn,omitempty"`
	Inputs      []DominoInput `json:"inputs,omitempty"`
	Image       string        `json:"image,omitempty"`
	StorePath   string        `json:"storePath,omitempty"`
}

// DominoResourceStatus defines observed state of a Domino.
type DominoResourceStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Phase              DominoPhase        `json:"phase,omitempty"`
	InputHash          string             `json:"inputHash,omitempty"`
	OutputHash         string             `json:"outputHash,omitempty"`
	Reused             bool               `json:"reused,omitempty"`
	CompletedAt        *metav1.Time       `json:"completedAt,omitempty"`
	Message            string             `json:"message,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// DominoList contains a list of Domino resources.
type DominoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Domino `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Domino{}, &DominoList{})
}

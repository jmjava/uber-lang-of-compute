package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SnapshotPhase represents snapshot lifecycle phase.
type SnapshotPhase string

const (
	SnapshotPhasePending SnapshotPhase = "Pending"
	SnapshotPhaseReady   SnapshotPhase = "Ready"
	SnapshotPhaseSealed  SnapshotPhase = "Sealed"
	SnapshotPhaseError   SnapshotPhase = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=snap
// +kubebuilder:printcolumn:name="TimeSlice",type=string,JSONPath=`.spec.timeSlice`
// +kubebuilder:printcolumn:name="Sealed",type=boolean,JSONPath=`.spec.sealed`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// Snapshot is an immutable data view for a time slice.
type Snapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SnapshotSpec   `json:"spec,omitempty"`
	Status SnapshotStatus `json:"status,omitempty"`
}

// SnapshotStatus defines observed state of a Snapshot.
type SnapshotStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Phase              SnapshotPhase      `json:"phase,omitempty"`
	SnapshotID         string             `json:"snapshotID,omitempty"`
	SealedAt           *metav1.Time       `json:"sealedAt,omitempty"`
	Message            string             `json:"message,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// SnapshotList contains a list of Snapshot resources.
type SnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Snapshot{}, &SnapshotList{})
}

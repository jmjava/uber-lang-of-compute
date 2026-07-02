package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReadReplicaPhase represents read-replica materialization lifecycle.
type ReadReplicaPhase string

const (
	ReadReplicaPhasePending       ReadReplicaPhase = "Pending"
	ReadReplicaPhaseMaterializing ReadReplicaPhase = "Materializing"
	ReadReplicaPhaseReady         ReadReplicaPhase = "Ready"
	ReadReplicaPhaseError         ReadReplicaPhase = "Error"
)

const (
	ReplicationModeDirect = "direct"
	ReplicationModeCDC    = "cdc"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rr
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.spec.sourceSnapshotID`
// +kubebuilder:printcolumn:name="Universe",type=string,JSONPath=`.spec.targetUniverse`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// ReadReplica materializes a read-only snapshot copy in a target pluggable universe.
type ReadReplica struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReadReplicaSpec   `json:"spec,omitempty"`
	Status ReadReplicaStatus `json:"status,omitempty"`
}

// ReadReplicaSpec defines cross-universe read-replica materialization.
type ReadReplicaSpec struct {
	MultiverseRef             string            `json:"multiverseRef"`
	RoutedEventID             string            `json:"routedEventId"`
	SourceSnapshotID          string            `json:"sourceSnapshotID"`
	SourceWorkflow            string            `json:"sourceWorkflow"`
	SourceNamespace           string            `json:"sourceNamespace"`
	TimeSlice                 string            `json:"timeSlice"`
	TargetUniverse            string            `json:"targetUniverse"`
	TargetComputeContextRef   string            `json:"targetComputeContextRef,omitempty"`
	PluggableUniverseRef      string            `json:"pluggableUniverseRef,omitempty"`
	FinalOutputHash           string            `json:"finalOutputHash,omitempty"`
	Partitions                map[string]string `json:"partitions,omitempty"`
	ReplicationMode           string            `json:"replicationMode,omitempty"`
	CDCSync                   *CDCSyncSpec      `json:"cdcSync,omitempty"`
}

// CDCSyncSpec configures Debezium-compatible Kafka CDC replication.
type CDCSyncSpec struct {
	Brokers []string `json:"brokers,omitempty"`
	Topic   string   `json:"topic,omitempty"`
	GroupID string   `json:"groupId,omitempty"`
}

// ReadReplicaStatus defines observed materialization state.
type ReadReplicaStatus struct {
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Phase              ReadReplicaPhase   `json:"phase,omitempty"`
	DominoCount        int                `json:"dominoCount,omitempty"`
	TargetStorePath    string             `json:"targetStorePath,omitempty"`
	MaterializedAt     *metav1.Time       `json:"materializedAt,omitempty"`
	Message            string             `json:"message,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ReadReplicaList contains a list of ReadReplica resources.
type ReadReplicaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReadReplica `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReadReplica{}, &ReadReplicaList{})
}

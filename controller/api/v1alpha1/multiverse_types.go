package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MultiversePhase represents multiverse routing lifecycle phase.
type MultiversePhase string

const (
	MultiversePhaseActive   MultiversePhase = "Active"
	MultiversePhaseDegraded MultiversePhase = "Degraded"
	MultiversePhaseOffline  MultiversePhase = "Offline"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=mv
// +kubebuilder:printcolumn:name="Default",type=string,JSONPath=`.spec.defaultUniverse`
// +kubebuilder:printcolumn:name="Universes",type=integer,JSONPath=`.spec.universes`
// +kubebuilder:printcolumn:name="Sync",type=boolean,JSONPath=`.spec.sync.enabled`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// Multiverse routes work across pluggable universes and time slices.
type Multiverse struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiverseSpec   `json:"spec,omitempty"`
	Status MultiverseStatus `json:"status,omitempty"`
}

// MultiverseSpec defines routing across universes with optional Kafka sync.
type MultiverseSpec struct {
	DefaultUniverse string              `json:"defaultUniverse"`
	Universes       []UniverseRouteSpec `json:"universes"`
	TimeSliceRoutes []TimeSliceRoute    `json:"timeSliceRoutes,omitempty"`
	Sync            *SyncSpec           `json:"sync,omitempty"`
}

// UniverseRouteSpec maps partition values to a pluggable universe.
type UniverseRouteSpec struct {
	Name                  string           `json:"name"`
	PluggableUniverseRef  string           `json:"pluggableUniverseRef"`
	ComputeContextRef     string           `json:"computeContextRef,omitempty"`
	Partitions            []PartitionRule  `json:"partitions,omitempty"`
}

type PartitionRule struct {
	Key    string   `json:"key"`
	Values []string `json:"values"`
}

type TimeSliceRoute struct {
	TimeSlice         string `json:"timeSlice"`
	Universe          string `json:"universe"`
	ComputeContextRef string `json:"computeContextRef,omitempty"`
}

type SyncSpec struct {
	Enabled bool          `json:"enabled,omitempty"`
	Kafka   *KafkaSyncSpec `json:"kafka,omitempty"`
}

type KafkaSyncSpec struct {
	Brokers  []string `json:"brokers"`
	Topic    string   `json:"topic"`
	GroupID  string   `json:"groupId,omitempty"`
	CDCTopic string   `json:"cdcTopic,omitempty"`
}

type MultiverseStatus struct {
	Phase          MultiversePhase    `json:"phase,omitempty"`
	UniverseCount  int                `json:"universeCount,omitempty"`
	KafkaConnected bool               `json:"kafkaConnected,omitempty"`
	RoutedEvents   []RoutedEvent      `json:"routedEvents,omitempty"`
	Message        string             `json:"message,omitempty"`
	Conditions     []metav1.Condition `json:"conditions,omitempty"`
}

// RoutedEvent records a routed snapshot event in the multiverse.
type RoutedEvent struct {
	EventID           string `json:"eventId"`
	SnapshotID        string `json:"snapshotId"`
	TimeSlice         string `json:"timeSlice"`
	SourceWorkflow    string `json:"sourceWorkflow"`
	TargetUniverse    string `json:"targetUniverse"`
	ComputeContextRef string `json:"computeContextRef,omitempty"`
	RoutedAt          string `json:"routedAt"`
}

// +kubebuilder:object:root=true

// MultiverseList contains a list of Multiverse resources.
type MultiverseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Multiverse `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Multiverse{}, &MultiverseList{})
}

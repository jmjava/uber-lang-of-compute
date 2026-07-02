package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkflowPhase represents the lifecycle phase of a Workflow.
type WorkflowPhase string

const (
	WorkflowPhasePending   WorkflowPhase = "Pending"
	WorkflowPhaseRunning   WorkflowPhase = "Running"
	WorkflowPhaseCompleted WorkflowPhase = "Completed"
	WorkflowPhaseError     WorkflowPhase = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wf
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Snapshot",type=string,JSONPath=`.status.snapshotID`
// +kubebuilder:printcolumn:name="Reused",type=integer,JSONPath=`.status.reusedCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Workflow composes a snapshot, domino chain, and execution config.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// WorkflowSpec defines the desired state of a Workflow.
type WorkflowSpec struct {
	Snapshot     SnapshotSpec       `json:"snapshot"`
	Dominos      []DominoSpec       `json:"dominos"`
	Execution    ExecutionSpec      `json:"execution"`
	Provisioning ProvisioningSpec   `json:"provisioning,omitempty"`
	Routing      RoutingSpec        `json:"routing,omitempty"`
}

// SnapshotSpec is the inline snapshot embedded in a Workflow.
type SnapshotSpec struct {
	TimeSlice         string                 `json:"timeSlice"`
	Source            SnapshotSource         `json:"source"`
	ComputeContextRef string                 `json:"computeContextRef,omitempty"`
	Sealed            bool                   `json:"sealed"`
}

type SnapshotSource struct {
	Inline map[string]interface{} `json:"inline,omitempty"`
	Path   string                 `json:"path,omitempty"`
	URI    string                 `json:"uri,omitempty"`
}

// DominoSpec defines one compute step in the chain.
type DominoSpec struct {
	Name        string        `json:"name"`
	SnapshotRef string        `json:"snapshotRef"`
	Command     string        `json:"command"`
	DependsOn   []string      `json:"dependsOn,omitempty"`
	Inputs      []DominoInput `json:"inputs,omitempty"`
	Image       string        `json:"image,omitempty"`
}

type DominoInput struct {
	FromDomino   string `json:"fromDomino,omitempty"`
	FromSnapshot string `json:"fromSnapshot,omitempty"`
}

type ExecutionSpec struct {
	Chain         []string `json:"chain"`
	Deterministic bool     `json:"deterministic"`
	// Runtime selects execution backend: local (default), kubernetes-init, openkruise.
	Runtime string `json:"runtime,omitempty"`
}

type ProvisioningSpec struct {
	StorePath string `json:"storePath,omitempty"`
	NodeLocal bool   `json:"nodeLocal,omitempty"`
}

type RoutingSpec struct {
	Universe          string `json:"universe,omitempty"`
	ComputeContextRef string `json:"computeContextRef,omitempty"`
}

// WorkflowStatus defines the observed state of a Workflow.
type WorkflowStatus struct {
	ObservedGeneration int64             `json:"observedGeneration,omitempty"`
	Phase              WorkflowPhase     `json:"phase,omitempty"`
	SnapshotID         string            `json:"snapshotID,omitempty"`
	DominoCount        int               `json:"dominoCount,omitempty"`
	ReusedCount        int               `json:"reusedCount,omitempty"`
	RecomputedCount    int               `json:"recomputedCount,omitempty"`
	LastRunTime        *metav1.Time      `json:"lastRunTime,omitempty"`
	Message            string            `json:"message,omitempty"`
	ReplayLogRef       string            `json:"replayLogRef,omitempty"`
	DominoResults      []DominoResult    `json:"dominoResults,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// DominoResult summarizes one domino execution in status.
type DominoResult struct {
	DominoID   string `json:"dominoID"`
	InputHash  string `json:"inputHash"`
	OutputHash string `json:"outputHash"`
	Reused     bool   `json:"reused"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflow resources.
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComputeWheelPhase represents the lifecycle phase of a ComputeWheel.
type ComputeWheelPhase string

const (
	ComputeWheelPhaseIdle       ComputeWheelPhase = "Idle"
	ComputeWheelPhaseRotating   ComputeWheelPhase = "Rotating"
	ComputeWheelPhaseProcessing ComputeWheelPhase = "Processing"
	ComputeWheelPhaseError      ComputeWheelPhase = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wheel
// +kubebuilder:printcolumn:name="TimeSlice",type=string,JSONPath=`.status.currentTimeSlice`
// +kubebuilder:printcolumn:name="Context",type=string,JSONPath=`.status.activeContext`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Rotations",type=integer,JSONPath=`.status.rotationCount`

// ComputeWheel rotates compute contexts through time slices continuously.
type ComputeWheel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComputeWheelSpec   `json:"spec,omitempty"`
	Status ComputeWheelStatus `json:"status,omitempty"`
}

// ComputeWheelSpec defines the desired state of a ComputeWheel.
type ComputeWheelSpec struct {
	// Contexts is the ordered list of ComputeContext names forming the wheel.
	Contexts []string `json:"contexts"`
	// TimeSliceInterval is the duration between time slice rotations (e.g. 1h, 24h, 1d).
	TimeSliceInterval string `json:"timeSliceInterval"`
	// WorkflowTemplate is applied to each context at each time slice.
	WorkflowTemplate WorkflowTemplateSpec `json:"workflowTemplate"`
	// Schedule configures wheel start behaviour.
	Schedule *WheelScheduleSpec `json:"schedule,omitempty"`
	// MaxRotations limits total slice rotations (0 = unlimited). Useful for demos/tests.
	MaxRotations int `json:"maxRotations,omitempty"`
	// PreProvisionNext pre-creates the next slot's Workflow while the current one runs.
	PreProvisionNext bool `json:"preProvisionNext,omitempty"`
}

// WorkflowTemplateSpec is the workflow template stamped per context/time slice.
type WorkflowTemplateSpec struct {
	Snapshot     SnapshotSpec     `json:"snapshot"`
	Dominos      []DominoSpec     `json:"dominos"`
	Execution    ExecutionSpec    `json:"execution"`
	Provisioning ProvisioningSpec `json:"provisioning,omitempty"`
	Routing      RoutingSpec      `json:"routing,omitempty"`
}

// WheelScheduleSpec configures wheel scheduling.
type WheelScheduleSpec struct {
	// StartTimeSlice is the ISO 8601 timestamp for the first slice (default: now truncated to interval).
	StartTimeSlice string `json:"startTimeSlice,omitempty"`
}

// ComputeWheelStatus defines the observed state of a ComputeWheel.
type ComputeWheelStatus struct {
	Phase              ComputeWheelPhase `json:"phase,omitempty"`
	CurrentTimeSlice   string            `json:"currentTimeSlice,omitempty"`
	ActiveContext      string            `json:"activeContext,omitempty"`
	ActiveContextIndex int               `json:"activeContextIndex,omitempty"`
	ActiveWorkflow     string            `json:"activeWorkflow,omitempty"`
	LastRotation       *metav1.Time      `json:"lastRotation,omitempty"`
	RotationCount      int               `json:"rotationCount,omitempty"`
	ProcessedSlots     []ProcessedSlot   `json:"processedSlots,omitempty"`
	Message            string            `json:"message,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// ProcessedSlot records a completed context+time-slice slot on the wheel.
type ProcessedSlot struct {
	TimeSlice      string `json:"timeSlice"`
	Context        string `json:"context"`
	Workflow       string `json:"workflow"`
	SnapshotID     string `json:"snapshotID,omitempty"`
	CompletedAt    string `json:"completedAt,omitempty"`
}

// +kubebuilder:object:root=true

// ComputeWheelList contains a list of ComputeWheel resources.
type ComputeWheelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeWheel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeWheel{}, &ComputeWheelList{})
}

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DominoChainRuntime selects how in-cluster domino execution runs.
type DominoChainRuntime string

const (
	DominoChainRuntimeLocal          DominoChainRuntime = "local"
	DominoChainRuntimeKubernetesInit DominoChainRuntime = "kubernetes-init"
	DominoChainRuntimeOpenKruise     DominoChainRuntime = "openkruise"
	DominoChainRuntimeVolcanoInit    DominoChainRuntime = "volcano-init"
)

// DominoChainPhase represents reconciler lifecycle phase.
type DominoChainPhase string

const (
	DominoChainPhasePending   DominoChainPhase = "Pending"
	DominoChainPhaseRunning   DominoChainPhase = "Running"
	DominoChainPhaseCompleted DominoChainPhase = "Completed"
	DominoChainPhaseError     DominoChainPhase = "Error"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=dchain
// +kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.spec.runtime`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Step",type=integer,JSONPath=`.status.activeStep`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// DominoChain runs a daisy-chained sequence of container dominos with emptyDir handoff.
type DominoChain struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DominoChainSpec   `json:"spec,omitempty"`
	Status DominoChainStatus `json:"status,omitempty"`
}

// DominoChainSpec defines a hot-swappable domino chain.
type DominoChainSpec struct {
	Snapshot     SnapshotSpec       `json:"snapshot"`
	Steps        []DominoStepSpec   `json:"steps"`
	Runtime      DominoChainRuntime `json:"runtime,omitempty"`
	StorePath    string             `json:"storePath,omitempty"`
	RunnerImage  string             `json:"runnerImage,omitempty"`
	NodeSelector map[string]string  `json:"nodeSelector,omitempty"`
	// VolcanoQueue assigns volcano-init Jobs to a Volcano queue (default: "default").
	VolcanoQueue string `json:"volcanoQueue,omitempty"`
}

// DominoStepSpec is one domino step in the chain.
type DominoStepSpec struct {
	Name    string   `json:"name"`
	Image   string   `json:"image,omitempty"`
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
}

// DominoChainStatus defines observed state.
type DominoChainStatus struct {
	Phase           DominoChainPhase `json:"phase,omitempty"`
	ActiveStep      int              `json:"activeStep,omitempty"`
	PodName         string           `json:"podName,omitempty"`
	SnapshotID      string           `json:"snapshotID,omitempty"`
	FinalOutputHash string           `json:"finalOutputHash,omitempty"`
	StepResults     []StepResult     `json:"stepResults,omitempty"`
	Message         string           `json:"message,omitempty"`
	Conditions      []metav1.Condition `json:"conditions,omitempty"`
}

// StepResult summarizes one completed step.
type StepResult struct {
	Name       string `json:"name"`
	Index      int    `json:"index"`
	OutputHash string `json:"outputHash,omitempty"`
	Phase      string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true

// DominoChainList contains a list of DominoChain resources.
type DominoChainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DominoChain `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DominoChain{}, &DominoChainList{})
}

package types

import "time"

// Snapshot represents an immutable data view for a time slice.
type Snapshot struct {
	APIVersion string          `yaml:"apiVersion" json:"apiVersion"`
	Kind       string          `yaml:"kind" json:"kind"`
	Metadata   ObjectMeta      `yaml:"metadata" json:"metadata"`
	Spec       SnapshotSpec    `yaml:"spec" json:"spec"`
	Status     *SnapshotStatus `yaml:"status,omitempty" json:"status,omitempty"`
}

type SnapshotSpec struct {
	TimeSlice         string                 `yaml:"timeSlice" json:"timeSlice"`
	Source            SnapshotSource         `yaml:"source" json:"source"`
	ComputeContextRef string                 `yaml:"computeContextRef,omitempty" json:"computeContextRef,omitempty"`
	Sealed            bool                   `yaml:"sealed" json:"sealed"`
}

type SnapshotSource struct {
	Inline map[string]interface{} `yaml:"inline,omitempty" json:"inline,omitempty"`
	Path   string                 `yaml:"path,omitempty" json:"path,omitempty"`
	URI    string                 `yaml:"uri,omitempty" json:"uri,omitempty"`
}

type SnapshotStatus struct {
	Phase      string `yaml:"phase" json:"phase"`
	SnapshotID string `yaml:"snapshotID" json:"snapshotID"`
	SealedAt   string `yaml:"sealedAt,omitempty" json:"sealedAt,omitempty"`
}

// Domino represents a single deterministic compute step.
type Domino struct {
	APIVersion string       `yaml:"apiVersion" json:"apiVersion"`
	Kind       string       `yaml:"kind" json:"kind"`
	Metadata   ObjectMeta   `yaml:"metadata" json:"metadata"`
	Spec       DominoSpec   `yaml:"spec" json:"spec"`
	Status     *DominoStatus `yaml:"status,omitempty" json:"status,omitempty"`
}

type DominoSpec struct {
	SnapshotRef string        `yaml:"snapshotRef" json:"snapshotRef"`
	Command     string        `yaml:"command" json:"command"`
	DependsOn   []string      `yaml:"dependsOn,omitempty" json:"dependsOn,omitempty"`
	Inputs      []DominoInput `yaml:"inputs,omitempty" json:"inputs,omitempty"`
	Image       string        `yaml:"image,omitempty" json:"image,omitempty"`
}

type DominoInput struct {
	FromDomino   string `yaml:"fromDomino,omitempty" json:"fromDomino,omitempty"`
	FromSnapshot string `yaml:"fromSnapshot,omitempty" json:"fromSnapshot,omitempty"`
}

type DominoStatus struct {
	Phase       string `yaml:"phase" json:"phase"`
	InputHash   string `yaml:"inputHash" json:"inputHash"`
	OutputHash  string `yaml:"outputHash" json:"outputHash"`
	Reused      bool   `yaml:"reused" json:"reused"`
	CompletedAt string `yaml:"completedAt,omitempty" json:"completedAt,omitempty"`
}

// Workflow composes snapshot, dominos, and execution config.
type Workflow struct {
	APIVersion string         `yaml:"apiVersion" json:"apiVersion"`
	Kind       string         `yaml:"kind" json:"kind"`
	Metadata   ObjectMeta     `yaml:"metadata" json:"metadata"`
	Spec       WorkflowSpec   `yaml:"spec" json:"spec"`
}

type WorkflowSpec struct {
	Snapshot     Snapshot           `yaml:"snapshot" json:"snapshot"`
	Dominos      []Domino           `yaml:"dominos" json:"dominos"`
	Execution    ExecutionConfig    `yaml:"execution" json:"execution"`
	Provisioning ProvisioningConfig `yaml:"provisioning,omitempty" json:"provisioning,omitempty"`
	Routing      RoutingConfig      `yaml:"routing,omitempty" json:"routing,omitempty"`
}

type ExecutionConfig struct {
	Chain         []string `yaml:"chain" json:"chain"`
	Deterministic bool     `yaml:"deterministic" json:"deterministic"`
}

type ProvisioningConfig struct {
	StorePath string `yaml:"storePath" json:"storePath"`
	NodeLocal bool   `yaml:"nodeLocal" json:"nodeLocal"`
}

type RoutingConfig struct {
	Universe            string `yaml:"universe" json:"universe"`
	ComputeContextRef   string `yaml:"computeContextRef" json:"computeContextRef"`
}

type ObjectMeta struct {
	Name   string            `yaml:"name" json:"name"`
	Labels map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// ReplayLogEntry records one domino execution for deterministic replay.
type ReplayLogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	SnapshotID string    `json:"snapshot_id"`
	DominoID   string    `json:"domino_id"`
	InputHash  string    `json:"input_hash"`
	OutputHash string    `json:"output_hash"`
	Reused     bool      `json:"reused"`
	Output     string    `json:"output,omitempty"`
}

// RunResult is the outcome of executing a domino chain.
type RunResult struct {
	SnapshotID string           `json:"snapshot_id"`
	Entries    []ReplayLogEntry `json:"entries"`
	FinalOutput string          `json:"final_output"`
}

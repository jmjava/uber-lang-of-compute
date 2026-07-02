package events

import "time"

const (
	TypeSnapshotCompleted = "kbl.snapshot.completed"
)

// SnapshotEvent is published when a workflow completes against a sealed snapshot.
type SnapshotEvent struct {
	EventID      string            `json:"event_id"`
	Type         string            `json:"type"`
	SnapshotID   string            `json:"snapshot_id"`
	TimeSlice    string            `json:"time_slice"`
	Workflow     string            `json:"workflow"`
	Namespace    string            `json:"namespace"`
	Universe     string            `json:"universe,omitempty"`
	Multiverse   string            `json:"multiverse,omitempty"`
	Partitions   map[string]string `json:"partitions,omitempty"`
	FinalOutput  string            `json:"final_output_hash,omitempty"`
	OccurredAt   time.Time         `json:"occurred_at"`
}

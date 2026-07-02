package cdc

const (
	TableSnapshots     = "snapshots"
	TableDominoResults = "domino_results"

	OpCreate = "c"
	OpUpdate = "u"
	OpDelete = "d"

	DefaultTopic = "kbl.cdc.snapshots"
)

// SnapshotRow is a Debezium-compatible snapshot record.
type SnapshotRow struct {
	SnapshotID string `json:"snapshot_id"`
	TimeSlice  string `json:"time_slice"`
	Data       string `json:"data"`
	Sealed     bool   `json:"sealed"`
}

// DominoResultRow is a Debezium-compatible domino result record.
type DominoResultRow struct {
	SnapshotID string `json:"snapshot_id"`
	DominoID   string `json:"domino_id"`
	InputHash  string `json:"input_hash"`
	OutputHash string `json:"output_hash"`
	Output     string `json:"output"`
	Reused     bool   `json:"reused"`
}

// Envelope is a simplified Debezium change-data-capture message.
type Envelope struct {
	Op     string      `json:"op"`
	Table  string      `json:"table"`
	After  interface{} `json:"after,omitempty"`
	Before interface{} `json:"before,omitempty"`
}

// SyncProgress tracks CDC replication completeness.
type SyncProgress struct {
	SnapshotApplied bool
	DominoCount     int
}

// IsComplete reports whether all required dominos were replicated.
func (p SyncProgress) IsComplete(required int) bool {
	if !p.SnapshotApplied {
		return false
	}
	if required == 0 {
		return true
	}
	return p.DominoCount >= required
}

package store

import "io"

// SnapshotPayloadSaver persists snapshot payload bytes without requiring a string copy.
type SnapshotPayloadSaver interface {
	SaveSnapshotPayload(snapshotID, timeSlice string, payload []byte, sealed bool) error
}

// SnapshotDataOpener opens snapshot payload bytes for streaming reads.
type SnapshotDataOpener interface {
	OpenSnapshotData(snapshotID string) (rc io.ReadCloser, sealed bool, err error)
}

// SaveSnapshotPayload writes payload bytes to backends that support zero-copy staging.
func SaveSnapshotPayload(backend Backend, snapshotID, timeSlice string, payload []byte, sealed bool) error {
	if ps, ok := backend.(SnapshotPayloadSaver); ok {
		return ps.SaveSnapshotPayload(snapshotID, timeSlice, payload, sealed)
	}
	return backend.SaveSnapshot(snapshotID, timeSlice, string(payload), sealed)
}

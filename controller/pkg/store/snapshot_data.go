package store

import (
	"database/sql"
	"errors"
	"os"
	"strings"
)

// SnapshotDataGetter is implemented by backends that can return persisted snapshot
// payload bytes without unmarshaling envelope metadata (TSDB sidecar files).
type SnapshotDataGetter interface {
	GetSnapshotData(snapshotID string) (data string, sealed bool, err error)
}

// IsSnapshotNotFound reports whether err indicates the snapshot is absent from the store.
func IsSnapshotNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, sql.ErrNoRows) || os.IsNotExist(err) {
		return true
	}
	return strings.Contains(err.Error(), "snapshot not found")
}

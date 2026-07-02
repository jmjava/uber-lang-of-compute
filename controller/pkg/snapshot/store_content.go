package snapshot

import (
	"encoding/json"
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// LoadContentPreferStore returns persisted snapshot JSON when available in the store.
// The hot execution path uses this to avoid re-fetching HTTP or re-reading paths on every run.
func LoadContentPreferStore(backend store.Backend, snapshotID string) (content interface{}, dataJSON string, ok bool, err error) {
	if backend == nil || snapshotID == "" {
		return nil, "", false, nil
	}

	_, data, sealed, getErr := backend.GetSnapshot(snapshotID)
	if getErr != nil || !sealed || data == "" {
		return nil, "", false, getErr
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(data), &parsed); err != nil {
		return nil, "", false, fmt.Errorf("parse stored snapshot %q: %w", snapshotID, err)
	}
	return parsed, data, true, nil
}

// ResolveEngineContentPreferStore loads snapshot content from store when snapshotID is known,
// otherwise resolves from inline/path/URI sources.
func ResolveEngineContentPreferStore(backend store.Backend, snap types.Snapshot, snapshotID string) (content interface{}, dataJSON string, resolvedID string, err error) {
	resolvedID = snapshotID
	if snap.Status != nil && snap.Status.SnapshotID != "" {
		resolvedID = snap.Status.SnapshotID
	}

	if content, dataJSON, ok, err := LoadContentPreferStore(backend, resolvedID); ok {
		return content, dataJSON, resolvedID, err
	}
	if err != nil {
		return nil, "", resolvedID, fmt.Errorf("load stored snapshot: %w", err)
	}

	content, err = ResolveEngineContent(snap.Spec)
	if err != nil {
		return nil, "", resolvedID, err
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return nil, "", resolvedID, fmt.Errorf("marshal snapshot content: %w", err)
	}
	return content, string(raw), resolvedID, nil
}

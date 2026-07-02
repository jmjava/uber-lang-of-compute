package snapshot

import (
	"encoding/json"
	"fmt"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// LoadDataPreferStore returns persisted snapshot JSON bytes without parsing when possible.
func LoadDataPreferStore(backend store.Backend, snapshotID string) (dataJSON string, ok bool, err error) {
	if backend == nil || snapshotID == "" {
		return "", false, nil
	}

	if sg, ok := backend.(store.SnapshotDataGetter); ok {
		data, sealed, err := sg.GetSnapshotData(snapshotID)
		if err != nil {
			if store.IsSnapshotNotFound(err) {
				return "", false, nil
			}
			return "", false, err
		}
		if !sealed || data == "" {
			return "", false, nil
		}
		return data, true, nil
	}

	_, data, sealed, getErr := backend.GetSnapshot(snapshotID)
	if getErr != nil {
		if store.IsSnapshotNotFound(getErr) {
			return "", false, nil
		}
		return "", false, getErr
	}
	if !sealed || data == "" {
		return "", false, nil
	}
	return data, true, nil
}

// LoadContentPreferStore returns persisted snapshot JSON when available in the store.
// The hot execution path uses this to avoid re-fetching HTTP or re-reading paths on every run.
func LoadContentPreferStore(backend store.Backend, snapshotID string) (content interface{}, dataJSON string, ok bool, err error) {
	dataJSON, ok, err = LoadDataPreferStore(backend, snapshotID)
	if err != nil || !ok {
		return nil, dataJSON, ok, err
	}

	var parsed interface{}
	if err := json.Unmarshal([]byte(dataJSON), &parsed); err != nil {
		return nil, "", false, fmt.Errorf("parse stored snapshot %q: %w", snapshotID, err)
	}
	return parsed, dataJSON, true, nil
}

// ResolveEngineContentPreferStore loads snapshot content from store when snapshotID is known,
// otherwise resolves from inline/path/URI sources.
func ResolveEngineContentPreferStore(backend store.Backend, snap types.Snapshot, snapshotID string) (content interface{}, dataJSON string, resolvedID string, err error) {
	resolvedID = snapshotID
	if snap.Status != nil && snap.Status.SnapshotID != "" {
		resolvedID = snap.Status.SnapshotID
	}

	if dataJSON, ok, loadErr := LoadDataPreferStore(backend, resolvedID); ok {
		var parsed interface{}
		if err := json.Unmarshal([]byte(dataJSON), &parsed); err != nil {
			return nil, "", resolvedID, fmt.Errorf("parse stored snapshot %q: %w", resolvedID, err)
		}
		return parsed, dataJSON, resolvedID, nil
	} else if loadErr != nil {
		return nil, "", resolvedID, fmt.Errorf("load stored snapshot: %w", loadErr)
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

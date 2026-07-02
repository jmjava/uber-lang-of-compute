package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// SealPayload computes snapshot ID and store bytes in one pass where possible.
// Path and HTTP JSON sources persist original bytes without a parse→remarshal cycle.
func SealPayload(spec kblv1alpha1.SnapshotSpec) (snapshotID string, data []byte, err error) {
	if err := Validate(spec); err != nil {
		return "", nil, err
	}

	switch {
	case spec.Source.Path != "":
		view, err := readPathBytesForSeal(spec.Source.Path)
		if err != nil {
			return "", nil, err
		}
		defer view.close()
		id, payload, borrowed, err := sealFromRawBytes(view.data, spec.TimeSlice, spec.Source.Path, "path")
		if err != nil {
			return "", nil, err
		}
		if borrowed {
			payload = bytes.Clone(payload)
		}
		return id, payload, nil

	case spec.Source.URI != "":
		if path, ok := fileURIPath(spec.Source.URI); ok {
			view, err := readPathBytesForSeal(path)
			if err != nil {
				return "", nil, err
			}
			defer view.close()
			id, payload, borrowed, err := sealFromRawBytes(view.data, spec.TimeSlice, path, "path")
			if err != nil {
				return "", nil, err
			}
			if borrowed {
				payload = bytes.Clone(payload)
			}
			return id, payload, nil
		}
		if isHTTPURI(spec.Source.URI) {
			raw, err := FetchHTTPBytes(spec.Source.URI)
			if err != nil {
				return "", nil, err
			}
			id, payload, _, err := sealFromRawBytes(raw, spec.TimeSlice, spec.Source.URI, "uri")
			return id, payload, err
		}
	}

	content, err := ResolveContent(spec)
	if err != nil {
		return "", nil, err
	}
	snapshotID, err = hash.SnapshotID(spec.TimeSlice, content)
	if err != nil {
		return "", nil, err
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", nil, fmt.Errorf("marshal snapshot data: %w", err)
	}
	return snapshotID, raw, nil
}

// SealToBackend seals snapshot content and persists payload bytes with zero-copy path staging
// when the backend supports SnapshotPayloadSaver (TSDB sidecar writes).
func SealToBackend(spec kblv1alpha1.SnapshotSpec, backend store.Backend) (snapshotID string, err error) {
	if err := Validate(spec); err != nil {
		return "", err
	}

	persist := func(id string, payload []byte) error {
		return store.SaveSnapshotPayload(backend, id, spec.TimeSlice, payload, true)
	}

	switch {
	case spec.Source.Path != "":
		return sealPathToBackend(spec.Source.Path, spec.TimeSlice, spec.Source.Path, "path", persist)
	case spec.Source.URI != "":
		if path, ok := fileURIPath(spec.Source.URI); ok {
			return sealPathToBackend(path, spec.TimeSlice, path, "path", persist)
		}
	}

	id, data, err := SealPayload(spec)
	if err != nil {
		return "", err
	}
	return id, persist(id, data)
}

func sealPathToBackend(path, timeSlice, source, sourceType string, persist func(string, []byte) error) (string, error) {
	view, err := readPathBytesForSeal(path)
	if err != nil {
		return "", err
	}
	defer view.close()

	id, payload, _, err := sealFromRawBytes(view.data, timeSlice, source, sourceType)
	if err != nil {
		return "", err
	}
	if err := persist(id, payload); err != nil {
		return "", err
	}
	return id, nil
}

func sealFromRawBytes(raw []byte, timeSlice, source, sourceType string) (string, []byte, bool, error) {
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		id, err := hash.SnapshotID(timeSlice, parsed)
		if err != nil {
			return "", nil, false, err
		}
		return id, raw, true, nil
	}

	parsed = map[string]interface{}{
		sourceType: source,
		"raw":      string(raw),
	}
	id, err := hash.SnapshotID(timeSlice, parsed)
	if err != nil {
		return "", nil, false, err
	}
	wrapped, err := json.Marshal(parsed)
	if err != nil {
		return "", nil, false, fmt.Errorf("marshal snapshot wrapper: %w", err)
	}
	return id, wrapped, false, nil
}

package snapshot

import (
	"encoding/json"
	"fmt"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
)

// SealPayload computes snapshot ID and store bytes in one pass where possible.
// Path and HTTP JSON sources persist original bytes without a parse→remarshal cycle.
func SealPayload(spec kblv1alpha1.SnapshotSpec) (snapshotID string, data string, err error) {
	if err := Validate(spec); err != nil {
		return "", "", err
	}

	switch {
	case spec.Source.Path != "":
		raw, err := ReadPathBytes(spec.Source.Path)
		if err != nil {
			return "", "", err
		}
		return sealFromRawBytes(raw, spec.TimeSlice, spec.Source.Path, "path")

	case spec.Source.URI != "":
		if path, ok := fileURIPath(spec.Source.URI); ok {
			raw, err := ReadPathBytes(path)
			if err != nil {
				return "", "", err
			}
			return sealFromRawBytes(raw, spec.TimeSlice, path, "path")
		}
		if isHTTPURI(spec.Source.URI) {
			raw, err := FetchHTTPBytes(spec.Source.URI)
			if err != nil {
				return "", "", err
			}
			return sealFromRawBytes(raw, spec.TimeSlice, spec.Source.URI, "uri")
		}
	}

	content, err := ResolveContent(spec)
	if err != nil {
		return "", "", err
	}
	snapshotID, err = hash.SnapshotID(spec.TimeSlice, content)
	if err != nil {
		return "", "", err
	}
	raw, err := json.Marshal(content)
	if err != nil {
		return "", "", fmt.Errorf("marshal snapshot data: %w", err)
	}
	return snapshotID, string(raw), nil
}

func sealFromRawBytes(raw []byte, timeSlice, source, sourceType string) (string, string, error) {
	var parsed interface{}
	if err := json.Unmarshal(raw, &parsed); err == nil {
		id, err := hash.SnapshotID(timeSlice, parsed)
		if err != nil {
			return "", "", err
		}
		return id, string(raw), nil
	}

	parsed = map[string]interface{}{
		sourceType: source,
		"raw":      string(raw),
	}
	id, err := hash.SnapshotID(timeSlice, parsed)
	if err != nil {
		return "", "", err
	}
	wrapped, err := json.Marshal(parsed)
	if err != nil {
		return "", "", fmt.Errorf("marshal snapshot wrapper: %w", err)
	}
	return id, string(wrapped), nil
}

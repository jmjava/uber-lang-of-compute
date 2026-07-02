package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

// ResolveContent loads snapshot payload data for hashing, persistence, and execution.
// Inline sources return as-is. Path and file:// URI sources read node-local file contents.
// Other URIs remain metadata-only until remote ingestion is implemented.
func ResolveContent(spec kblv1alpha1.SnapshotSpec) (interface{}, error) {
	if spec.Source.Inline != nil {
		return spec.Source.Inline, nil
	}
	if spec.Source.Path != "" {
		return loadPath(spec.Source.Path)
	}
	if spec.Source.URI != "" {
		if path, ok := fileURIPath(spec.Source.URI); ok {
			return loadPath(path)
		}
		return map[string]string{"uri": spec.Source.URI}, nil
	}
	return nil, fmt.Errorf("snapshot source requires inline, path, or uri")
}

// ResolveEngineContent loads content from an engine-domain snapshot spec.
func ResolveEngineContent(spec types.SnapshotSpec) (interface{}, error) {
	return ResolveContent(kblv1alpha1.SnapshotSpec{
		TimeSlice: spec.TimeSlice,
		Source: kblv1alpha1.SnapshotSource{
			Inline: spec.Source.Inline,
			Path:   spec.Source.Path,
			URI:    spec.Source.URI,
		},
		Sealed: spec.Sealed,
	})
}

// IsPathNotReady reports whether content resolution failed because the path is not yet available.
func IsPathNotReady(err error) bool {
	if err == nil {
		return false
	}
	return os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory")
}

func fileURIPath(uri string) (string, bool) {
	if !strings.HasPrefix(uri, "file://") {
		return "", false
	}
	path := strings.TrimPrefix(uri, "file://")
	if path == "" {
		return "", false
	}
	return path, true
}

func loadPath(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot path %q: %w", path, err)
	}

	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err == nil {
		return parsed, nil
	}

	return map[string]interface{}{
		"path": path,
		"raw":  string(data),
	}, nil
}

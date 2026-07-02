package snapshot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

const maxSnapshotFetchBytes = 32 << 20 // 32 MiB

var snapshotHTTPClient = &http.Client{Timeout: 30 * time.Second}

// ResolveContent loads snapshot payload data for hashing, persistence, and execution.
// Inline sources return as-is. Path and file:// URI sources read node-local files.
// http:// and https:// URIs are fetched at seal/execute time. Other URI schemes remain metadata-only.
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
		if isHTTPURI(spec.Source.URI) {
			return loadHTTP(spec.Source.URI)
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

// IsSourceNotReady reports transient source resolution failures that should requeue.
func IsSourceNotReady(err error) bool {
	return IsPathNotReady(err) || IsURINotReady(err)
}

// IsPathNotReady reports whether content resolution failed because the path is not yet available.
func IsPathNotReady(err error) bool {
	if err == nil {
		return false
	}
	return os.IsNotExist(err) || strings.Contains(err.Error(), "no such file or directory")
}

// IsURINotReady reports whether an HTTP snapshot fetch failed transiently.
func IsURINotReady(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	if !strings.Contains(msg, "fetch snapshot uri") {
		return false
	}
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "HTTP 502") ||
		strings.Contains(msg, "HTTP 503") ||
		strings.Contains(msg, "HTTP 504") ||
		strings.Contains(msg, "HTTP 429")
}

func isHTTPURI(uri string) bool {
	return strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
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
	data, err := ReadPathBytes(path)
	if err != nil {
		return nil, err
	}
	return parseSnapshotBytes(data, path, "path")
}

func loadHTTP(uri string) (interface{}, error) {
	data, err := FetchHTTPBytes(uri)
	if err != nil {
		return nil, err
	}
	return parseSnapshotBytes(data, uri, "uri")
}

// ReadPathBytes reads snapshot file contents from a node-local path.
func ReadPathBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read snapshot path %q: %w", path, err)
	}
	return data, nil
}

// FetchHTTPBytes downloads snapshot content from an HTTP(S) URI.
func FetchHTTPBytes(uri string) ([]byte, error) {
	resp, err := snapshotHTTPClient.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("fetch snapshot uri %q: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch snapshot uri %q: HTTP %d", uri, resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSnapshotFetchBytes+1))
	if err != nil {
		return nil, fmt.Errorf("fetch snapshot uri %q: read body: %w", uri, err)
	}
	if len(data) > maxSnapshotFetchBytes {
		return nil, fmt.Errorf("fetch snapshot uri %q: body exceeds %d bytes", uri, maxSnapshotFetchBytes)
	}
	return data, nil
}

func parseSnapshotBytes(data []byte, source, sourceType string) (interface{}, error) {
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err == nil {
		return parsed, nil
	}

	out := map[string]interface{}{
		sourceType: source,
		"raw":      string(data),
	}
	return out, nil
}

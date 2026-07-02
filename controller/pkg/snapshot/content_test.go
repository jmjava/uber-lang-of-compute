package snapshot_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
)

func TestResolveContentPathJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "curve.json")
	if err := os.WriteFile(path, []byte(`{"instruments":[{"id":"US10Y"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	spec := kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-04-15T00:00:00Z",
		Source:    kblv1alpha1.SnapshotSource{Path: path},
		Sealed:    true,
	}

	content, err := snapshot.ResolveContent(spec)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := content.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map content, got %T", content)
	}
	instruments, ok := m["instruments"].([]interface{})
	if !ok || len(instruments) != 1 {
		t.Fatalf("expected parsed instruments array, got %+v", m)
	}

	if _, err := snapshot.ComputeID(kblv1alpha1.SnapshotSpec{
		TimeSlice: spec.TimeSlice,
		Source:    kblv1alpha1.SnapshotSource{Path: "/does/not/exist.json"},
		Sealed:    true,
	}); err == nil {
		t.Fatal("expected error for missing path")
	}

	idContent, err := snapshot.ComputeID(spec)
	if err != nil {
		t.Fatal(err)
	}
	idMeta, err := snapshot.ComputeID(kblv1alpha1.SnapshotSpec{
		TimeSlice: spec.TimeSlice,
		Source:    kblv1alpha1.SnapshotSource{Path: path},
		Sealed:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	idPathOnly, err := snapshot.ComputeID(kblv1alpha1.SnapshotSpec{
		TimeSlice: spec.TimeSlice,
		Source:    kblv1alpha1.SnapshotSource{URI: "s3://bucket/curve.json"},
		Sealed:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if idContent != idMeta {
		t.Fatalf("expected same ID for path content, got %s vs %s", idContent, idMeta)
	}
	if idContent == idPathOnly {
		t.Fatal("path content ID should differ from remote URI metadata ID")
	}
}

func TestResolveContentFileURI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.txt")
	if err := os.WriteFile(path, []byte("plain-text"), 0o644); err != nil {
		t.Fatal(err)
	}

	spec := kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-01-01T00:00:00Z",
		Source:    kblv1alpha1.SnapshotSource{URI: "file://" + path},
		Sealed:    true,
	}
	content, err := snapshot.ResolveContent(spec)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := content.(map[string]interface{})
	if !ok || m["raw"] != "plain-text" {
		t.Fatalf("expected raw wrapper, got %+v", content)
	}
}

func TestIsPathNotReady(t *testing.T) {
	_, err := snapshot.ResolveContent(kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-01-01",
		Source:    kblv1alpha1.SnapshotSource{Path: "/missing/file.json"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !snapshot.IsSourceNotReady(err) {
		t.Fatalf("expected source-not-ready, got %v", err)
	}
}

func TestResolveContentHTTPJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"instruments":[{"instrument_id":"US10Y","rate":4.25}]}`))
	}))
	defer srv.Close()

	spec := kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-04-15T00:00:00Z",
		Source:    kblv1alpha1.SnapshotSource{URI: srv.URL + "/curve.json"},
		Sealed:    true,
	}
	content, err := snapshot.ResolveContent(spec)
	if err != nil {
		t.Fatal(err)
	}
	m, ok := content.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map, got %T", content)
	}
	if _, ok := m["instruments"]; !ok {
		t.Fatalf("expected instruments key, got %+v", m)
	}

	idHTTP, err := snapshot.ComputeID(spec)
	if err != nil {
		t.Fatal(err)
	}
	idMeta, err := snapshot.ComputeID(kblv1alpha1.SnapshotSpec{
		TimeSlice: spec.TimeSlice,
		Source:    kblv1alpha1.SnapshotSource{URI: "s3://bucket/curve.json"},
		Sealed:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if idHTTP == idMeta {
		t.Fatal("HTTP content ID should differ from s3 metadata ID")
	}
}

func TestIsURINotReady(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := snapshot.ResolveContent(kblv1alpha1.SnapshotSpec{
		TimeSlice: "2025-01-01",
		Source:    kblv1alpha1.SnapshotSource{URI: srv.URL},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if !snapshot.IsSourceNotReady(err) {
		t.Fatalf("expected transient URI error, got %v", err)
	}
}

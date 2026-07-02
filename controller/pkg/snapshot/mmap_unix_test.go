//go:build unix

package snapshot_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
)

func TestReadPathBytesLargeFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.json")
	payload := `{"pad":"` + strings.Repeat("x", 1<<20) + `"}`
	if err := os.WriteFile(path, []byte(payload), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := snapshot.ReadPathBytes(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(payload) {
		t.Fatalf("expected %d bytes, got %d", len(payload), len(data))
	}
	if string(data) != payload {
		t.Fatal("payload mismatch")
	}
}

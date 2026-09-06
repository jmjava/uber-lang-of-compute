package store_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestSealedSnapshotIsWriteOnce(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "once.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if err := s.SaveSnapshot("snap", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSnapshot("snap", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatalf("identical sealed rewrite must be idempotent: %v", err)
	}
	if err := s.SaveSnapshot("snap", "2025-01-01", `{"v":2}`, true); err == nil {
		t.Fatal("mutating a sealed snapshot must fail")
	} else if !errors.Is(err, store.ErrSealedOverwrite) {
		t.Fatalf("got %v want ErrSealedOverwrite", err)
	}

	if err := s.SaveSnapshot("open", "2025-01-01", `{"v":1}`, false); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSnapshot("open", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatalf("sealing an open snapshot must be allowed: %v", err)
	}
}

func TestSealedSnapshotIsWriteOnceTSDB(t *testing.T) {
	eng, err := store.OpenTSDBEngine(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := eng.SaveSnapshot("snap", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := eng.SaveSnapshot("snap", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := eng.SaveSnapshot("snap", "2025-01-01", `{"v":2}`, true); err == nil || !errors.Is(err, store.ErrSealedOverwrite) {
		t.Fatalf("got %v want ErrSealedOverwrite", err)
	}
}

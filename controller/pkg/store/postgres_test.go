package store_test

import (
	"errors"
	"os"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

func TestPostgresSealedSnapshotIsWriteOnce(t *testing.T) {
	s := openPostgresOrSkip(t)
	defer s.Close()

	id := "snap-pg-once"
	if err := s.SaveSnapshot(id, "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSnapshot(id, "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatalf("identical sealed rewrite must be idempotent: %v", err)
	}
	if err := s.SaveSnapshot(id, "2025-01-01", `{"v":2}`, true); err == nil {
		t.Fatal("mutating a sealed snapshot must fail")
	} else if !errors.Is(err, store.ErrSealedOverwrite) {
		t.Fatalf("got %v want ErrSealedOverwrite", err)
	}
}

func openPostgresOrSkip(t *testing.T) store.Backend {
	t.Helper()
	dsn, set := os.LookupEnv("POSTGRES_DSN")
	if !set || dsn == "" {
		dsn = "postgres://kbl:kbl@127.0.0.1:15432/kbl?sslmode=disable"
		s, err := store.OpenPostgres(dsn)
		if err != nil {
			t.Skipf("postgres not reachable at %s: %v", dsn, err)
		}
		return s
	}
	s, err := store.OpenPostgres(dsn)
	if err != nil {
		t.Fatalf("POSTGRES_DSN=%s is not reachable: %v", dsn, err)
	}
	return s
}

package store_test

import (
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func TestGetLatestResultFollowsReplaySpine(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), "spine.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	const snap = "snap-spine"
	if err := s.SaveSnapshot(snap, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveResult(snap, "load", "hash-zz", "out-old", `{"n":1}`, false, snap, "link-1"); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveResult(snap, "load", "hash-aa", "out-new", `{"n":2}`, false, "link-1", "link-2"); err != nil {
		t.Fatal(err)
	}

	in, out, payload, err := s.GetLatestResult(snap, "load")
	if err != nil {
		t.Fatal(err)
	}
	if in != "hash-aa" || out != "out-new" || payload != `{"n":2}` {
		t.Fatalf("GetLatestResult must return the last spine row, got in=%s out=%s payload=%s", in, out, payload)
	}
}

func TestTSDBClientListReplayAndSpineLatest(t *testing.T) {
	_, backend := startTestTSDB(t)
	const snap = "snap-tsdb-spine"
	if err := backend.SaveSnapshot(snap, "2025-04-15", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	link1, err := hash.Link(snap, "zzzzzzzz", "out-old")
	if err != nil {
		t.Fatal(err)
	}
	link2, err := hash.Link(link1, "aaaaaaaa", "out-new")
	if err != nil {
		t.Fatal(err)
	}
	if err := backend.SaveResult(snap, "load", "zzzzzzzz", "out-old", `{"n":1}`, false, snap, link1); err != nil {
		t.Fatal(err)
	}
	if err := backend.SaveResult(snap, "load", "aaaaaaaa", "out-new", `{"n":2}`, false, link1, link2); err != nil {
		t.Fatal(err)
	}

	rows, err := backend.ListReplay(snap)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("ListReplay %d want 2", len(rows))
	}
	entries := make([]types.ReplayLogEntry, len(rows))
	for i, r := range rows {
		entries[i] = types.ReplayLogEntry{
			SnapshotID: r.SnapshotID,
			InputHash:  r.InputHash,
			OutputHash: r.OutputHash,
			PrevLink:   r.PrevLink,
			Link:       r.Link,
		}
	}
	if err := theory.VerifySpine(snap, entries); err != nil {
		t.Fatalf("TSDB replay spine: %v", err)
	}

	in, out, payload, err := backend.GetLatestResult(snap, "load")
	if err != nil {
		t.Fatal(err)
	}
	if in != "aaaaaaaa" || out != "out-new" || payload != `{"n":2}` {
		t.Fatalf("TSDB GetLatestResult used memo lexical order; got in=%s out=%s payload=%s", in, out, payload)
	}
}

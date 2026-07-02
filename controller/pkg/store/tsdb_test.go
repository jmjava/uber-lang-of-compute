package store_test

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
)

func startTestTSDB(t *testing.T) (*httptest.Server, store.Backend) {
	t.Helper()
	dir := t.TempDir()
	engine, err := store.OpenTSDBEngine(dir)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(store.NewTSDBHandler(engine))
	client, err := store.OpenTSDBClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		client.Close()
		srv.Close()
	})
	return srv, client
}

func TestTSDBBackendMemoization(t *testing.T) {
	_, backend := startTestTSDB(t)

	eng := engine.New(backend)
	wf := &types.Workflow{
		Spec: types.WorkflowSpec{
			Snapshot: types.Snapshot{
				Metadata: types.ObjectMeta{Name: "snap"},
				Spec: types.SnapshotSpec{
					TimeSlice: "2025-07-02T00:00:00Z",
					Source:    types.SnapshotSource{Inline: map[string]interface{}{"v": 1}},
					Sealed:    true,
				},
			},
			Dominos: []types.Domino{{
				Metadata: types.ObjectMeta{Name: "a"},
				Spec:     types.DominoSpec{SnapshotRef: "snap", Command: "builtin:identity"},
			}},
			Execution: types.ExecutionConfig{Chain: []string{"a"}, Deterministic: true},
		},
	}

	r1, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	if !r2.Entries[0].Reused {
		t.Error("expected TSDB memo hit on second run")
	}
	if r1.SnapshotID != r2.SnapshotID {
		t.Error("snapshot IDs should match")
	}
}

func TestOpenBackendSQLite(t *testing.T) {
	path := t.TempDir() + "/test.db"
	b, err := store.OpenBackend(store.Config{Type: store.TypeSQLite, Path: path})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := b.SaveSnapshot("s1", "2025-01-01", "{}", true); err != nil {
		t.Fatal(err)
	}
}

func TestOpenBackendTSDB(t *testing.T) {
	_, backend := startTestTSDB(t)
	if err := backend.SaveSnapshot("s1", "2025-01-01", `{"v":1}`, true); err != nil {
		t.Fatal(err)
	}
	_, data, sealed, err := backend.GetSnapshot("s1")
	if err != nil || data != `{"v":1}` || !sealed {
		t.Fatalf("get snapshot: data=%s sealed=%v err=%v", data, sealed, err)
	}
}

func TestTSDBGetSnapshotDataSidecar(t *testing.T) {
	srv, backend := startTestTSDB(t)
	client, ok := backend.(*store.TSDBClient)
	if !ok {
		t.Fatal("expected TSDBClient")
	}

	payload := `{"instruments":[{"instrument_id":"US10Y"}]}`
	if err := backend.SaveSnapshot("sidecar-snap", "2025-01-01", payload, true); err != nil {
		t.Fatal(err)
	}

	data, sealed, err := client.GetSnapshotData("sidecar-snap")
	if err != nil || !sealed || data != payload {
		t.Fatalf("GetSnapshotData: data=%q sealed=%v err=%v", data, sealed, err)
	}

	resp, err := srv.Client().Get(srv.URL + "/v1/snapshots/sidecar-snap/data")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("http status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != payload {
		t.Fatalf("streamed body %q", string(body))
	}
}

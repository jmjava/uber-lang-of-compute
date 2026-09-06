package engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	kyaml "sigs.k8s.io/yaml"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/hash"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/snapshot"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

type interpOut struct {
	Method       string             `json:"method"`
	Interpolated map[string]float64 `json:"interpolated"`
	CurvePoints  []struct {
		Maturity string  `json:"maturity"`
		Rate     float64 `json:"rate"`
		Tenor    float64 `json:"tenor_years"`
	} `json:"curve_points"`
}

type riskOut struct {
	Method      string  `json:"method"`
	Notional    float64 `json:"notional"`
	RiskMetrics []struct {
		Tenor string  `json:"tenor"`
		Rate  float64 `json:"rate"`
		DV01  float64 `json:"dv01"`
	} `json:"risk_metrics"`
}

func exampleFile(t *testing.T, rel ...string) string {
	t.Helper()
	parts := append([]string{"..", "..", "..", "examples"}, rel...)
	return filepath.Join(parts...)
}

func readExample(t *testing.T, rel ...string) []byte {
	t.Helper()
	data, err := os.ReadFile(exampleFile(t, rel...))
	if err != nil {
		t.Fatalf("read example %s: %v", filepath.Join(rel...), err)
	}
	return data
}

func runWorkflow(t *testing.T, backend store.Backend, wf *types.Workflow) *types.RunResult {
	t.Helper()
	result, err := engine.New(backend).Run(wf)
	if err != nil {
		t.Fatalf("run %s: %v", wf.Metadata.Name, err)
	}
	if err := theory.VerifySpine(result.SnapshotID, result.Entries); err != nil {
		t.Fatalf("spine for %s: %v", wf.Metadata.Name, err)
	}
	if result.HeadLink == "" || result.HeadLink != result.Entries[len(result.Entries)-1].Link {
		t.Fatalf("head link for %s does not match last entry", wf.Metadata.Name)
	}
	return result
}

func openSQLite(t *testing.T) store.Backend {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), "example.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func loadCRDWorkflow(t *testing.T, rel ...string) *types.Workflow {
	t.Helper()
	data := readExample(t, rel...)
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var raw map[string]interface{}
		err := dec.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse %s: %v", filepath.Join(rel...), err)
		}
		if kind, _ := raw["kind"].(string); kind != "Workflow" {
			continue
		}
		encoded, err := yaml.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var cr kblv1alpha1.Workflow
		if err := kyaml.Unmarshal(encoded, &cr); err != nil {
			t.Fatalf("crd unmarshal %s: %v", filepath.Join(rel...), err)
		}
		return convert.ToEngineWorkflow(&cr)
	}
	t.Fatalf("no Workflow document in %s", filepath.Join(rel...))
	return nil
}

func loadCRDSnapshot(t *testing.T, rel ...string) kblv1alpha1.Snapshot {
	t.Helper()
	var snap kblv1alpha1.Snapshot
	if err := kyaml.Unmarshal(readExample(t, rel...), &snap); err != nil {
		t.Fatalf("snapshot unmarshal %s: %v", filepath.Join(rel...), err)
	}
	if snap.Name == "" {
		t.Fatalf("snapshot %s has empty name", filepath.Join(rel...))
	}
	return snap
}

func decodeJSON[T any](t *testing.T, raw, what string) T {
	t.Helper()
	var out T
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("parse %s: %v\n%s", what, err, raw)
	}
	return out
}

func almostEqual(t *testing.T, got, want float64, what string) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s: got %v want %v", what, got, want)
	}
}

func curveWorkflow(snap types.Snapshot, chain ...string) *types.Workflow {
	if len(chain) == 0 {
		chain = []string{"load", "interpolate"}
	}
	name := snap.Metadata.Name
	dominos := []types.Domino{{
		Metadata: types.ObjectMeta{Name: "load"},
		Spec:     types.DominoSpec{SnapshotRef: name, Command: "builtin:identity"},
	}}
	if len(chain) > 1 {
		dominos = append(dominos, types.Domino{
			Metadata: types.ObjectMeta{Name: "interpolate"},
			Spec: types.DominoSpec{
				SnapshotRef: name,
				Command:     "builtin:interpolate",
				DependsOn:   []string{"load"},
				Inputs:      []types.DominoInput{{FromDomino: "load"}},
			},
		})
	}
	return &types.Workflow{
		Kind:     "Workflow",
		Metadata: types.ObjectMeta{Name: name + "-workflow"},
		Spec: types.WorkflowSpec{
			Snapshot:  snap,
			Dominos:   dominos,
			Execution: types.ExecutionConfig{Chain: chain, Deterministic: true},
		},
	}
}

func TestExampleSimpleDominoChainPassesSnapshotThrough(t *testing.T) {
	wf := loadTestWorkflow(t, "simple-domino-chain")
	result := runWorkflow(t, openSQLite(t), wf)

	if len(result.Entries) != 2 {
		t.Fatalf("chain length %d want 2", len(result.Entries))
	}
	if result.Entries[0].DominoID != "step-one" || result.Entries[1].DominoID != "step-two" {
		t.Fatalf("unexpected chain: %s -> %s", result.Entries[0].DominoID, result.Entries[1].DominoID)
	}

	out := decodeJSON[map[string]interface{}](t, result.FinalOutput, "identity output")
	if out["message"] != "hello-kbl" {
		t.Fatalf("message %v want hello-kbl", out["message"])
	}
	almostEqual(t, asFloat(t, out["value"]), 42, "value")
	if result.MinRegularity != "builtin" {
		t.Fatalf("min regularity %q want builtin", result.MinRegularity)
	}

	replay := runWorkflow(t, openSQLite(t), wf)
	if result.SnapshotID != replay.SnapshotID || result.HeadLink != replay.HeadLink {
		t.Fatal("independent simple-chain runs must share snapshot ID and spine head")
	}
}

func TestExampleFinanceCurveSnapshotComputesDV01(t *testing.T) {
	wf := loadTestWorkflow(t, "finance-curve-snapshot")
	s := openSQLite(t)
	first := runWorkflow(t, s, wf)

	if got := names(first); len(got) != 3 || got[0] != "load-curve-data" || got[2] != "compute-risk" {
		t.Fatalf("finance chain %v", got)
	}

	interp := decodeJSON[interpOut](t, first.Entries[1].Output, "interpolate")
	if interp.Method != "linear" {
		t.Fatalf("interpolate method %q", interp.Method)
	}
	if len(interp.CurvePoints) != 3 {
		t.Fatalf("curve points %d want 3", len(interp.CurvePoints))
	}
	almostEqual(t, interp.Interpolated["3Y"], 4.80+(1.0/3.0)*(4.45-4.80), "3Y")
	almostEqual(t, interp.Interpolated["7Y"], 4.45+0.4*(4.25-4.45), "7Y")

	risk := decodeJSON[riskOut](t, first.FinalOutput, "risk")
	if risk.Method != "dv01_simplified" || risk.Notional != 1_000_000 {
		t.Fatalf("risk envelope %+v", risk)
	}
	if len(risk.RiskMetrics) != 2 {
		t.Fatalf("risk tenors %d want 2", len(risk.RiskMetrics))
	}
	byTenor := map[string]float64{}
	for _, m := range risk.RiskMetrics {
		byTenor[m.Tenor] = m.DV01
	}
	almostEqual(t, byTenor["3Y"], 300, "3Y dv01")
	almostEqual(t, byTenor["7Y"], 700, "7Y dv01")

	second := runWorkflow(t, s, wf)
	if second.WorkEvaluations != 0 || second.WorkReuses != 3 {
		t.Fatalf("rerun work eval=%d reuse=%d want eval=0 reuse=3", second.WorkEvaluations, second.WorkReuses)
	}
	if first.FinalOutput != second.FinalOutput || first.HeadLink != second.HeadLink {
		t.Fatal("finance rerun must reuse the same risk output and spine")
	}
}

func TestExampleFinanceCRDFormMatchesEmbeddedWorkflow(t *testing.T) {
	embedded := loadTestWorkflow(t, "finance-curve-snapshot")
	crd := loadCRDWorkflow(t, "finance-curve-snapshot", "workflow-crd.yaml")

	if got := namesFromWF(crd); len(got) != 3 || got[1] != "interpolate-curve" {
		t.Fatalf("CRD convert chain %v", got)
	}

	a := runWorkflow(t, openSQLite(t), embedded)
	b := runWorkflow(t, openSQLite(t), crd)
	if a.FinalOutput != b.FinalOutput {
		t.Fatalf("CRD finance output diverged from embedded example:\nembedded: %s\ncrd:      %s", a.FinalOutput, b.FinalOutput)
	}
	if a.SnapshotID != b.SnapshotID {
		t.Fatalf("CRD snapshot ID %s vs embedded %s", b.SnapshotID, a.SnapshotID)
	}
}

func TestExampleRatesCurveFromMultiverseRunsIdentity(t *testing.T) {
	wf := loadCRDWorkflow(t, "multiverse-finance", "workflow-rates.yaml")
	if wf.Metadata.Labels["kbl.io/partition-asset_class"] != "rates" {
		t.Fatalf("rates label missing: %+v", wf.Metadata.Labels)
	}
	if wf.Spec.Routing.Universe != "rates-universe" {
		t.Fatalf("universe %q", wf.Spec.Routing.Universe)
	}

	result := runWorkflow(t, openSQLite(t), wf)
	if len(result.Entries) != 1 || result.Entries[0].DominoID != "load" {
		t.Fatalf("rates chain %+v", names(result))
	}
	out := decodeJSON[map[string]interface{}](t, result.FinalOutput, "rates identity")
	instruments, _ := out["instruments"].([]interface{})
	if len(instruments) != 1 {
		t.Fatalf("expected one US10Y instrument, got %#v", out)
	}
	row, _ := instruments[0].(map[string]interface{})
	if row["instrument_id"] != "US10Y" {
		t.Fatalf("instrument %+v", row)
	}
}

func TestExampleNodeLocalTSDBWorkflowRunsOnTSDB(t *testing.T) {
	dir := t.TempDir()
	eng, err := store.OpenTSDBEngine(dir)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(store.NewTSDBHandler(eng))
	t.Cleanup(srv.Close)
	backend, err := store.OpenTSDBClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	wf := loadCRDWorkflow(t, "node-local-tsdb", "workflow-tsdb.yaml")
	first := runWorkflow(t, backend, wf)
	interp := decodeJSON[interpOut](t, first.FinalOutput, "tsdb interpolate")
	if interp.Method != "linear" || len(interp.CurvePoints) != 2 {
		t.Fatalf("tsdb interpolate %+v", interp)
	}
	almostEqual(t, interp.Interpolated["3Y"], 4.80+0.125*(4.25-4.80), "tsdb 3Y")
	almostEqual(t, interp.Interpolated["7Y"], 4.80+0.625*(4.25-4.80), "tsdb 7Y")

	second := runWorkflow(t, backend, wf)
	if !second.Entries[0].Reused || !second.Entries[1].Reused {
		t.Fatal("TSDB-backed example must memoize on rerun")
	}
	if first.FinalOutput != second.FinalOutput {
		t.Fatal("TSDB rerun changed interpolate output")
	}
}

func TestExamplePathSnapshotFeedsInterpolate(t *testing.T) {
	src := readExample(t, "path-snapshot", "data", "curve.json")
	path := filepath.Join(t.TempDir(), "curve.json")
	if err := os.WriteFile(path, src, 0o644); err != nil {
		t.Fatal(err)
	}

	snapCR := loadCRDSnapshot(t, "path-snapshot", "snapshot.yaml")
	snapCR.Spec.Source.Path = path
	wf := curveWorkflow(convert.ToEngineSnapshot(&snapCR))

	s := openSQLite(t)
	first := runWorkflow(t, s, wf)
	interp := decodeJSON[interpOut](t, first.FinalOutput, "path interpolate")
	almostEqual(t, interp.Interpolated["3Y"], 4.80+0.125*(4.25-4.80), "path 3Y")

	// Controller stamps status.snapshotID after seal so later runs load the
	// store instead of the node-local file.
	wf.Spec.Snapshot.Status = &types.SnapshotStatus{SnapshotID: first.SnapshotID}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	second := runWorkflow(t, s, wf)
	if second.WorkReuses != 2 {
		t.Fatalf("sealed path snapshot should replay from store after the file is gone; reuse=%d", second.WorkReuses)
	}
	if first.SnapshotID != second.SnapshotID || first.FinalOutput != second.FinalOutput {
		t.Fatal("path example replay diverged after source file removal")
	}
}

func TestExampleHTTPSnapshotFeedsInterpolateThenStore(t *testing.T) {
	payload := readExample(t, "http-snapshot", "data", "curve.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(payload)
	}))

	snapCR := loadCRDSnapshot(t, "http-snapshot", "snapshot.yaml")
	snapCR.Spec.Source.URI = srv.URL + "/curve.json"
	wf := curveWorkflow(convert.ToEngineSnapshot(&snapCR))

	s := openSQLite(t)
	first := runWorkflow(t, s, wf)
	interp := decodeJSON[interpOut](t, first.FinalOutput, "http interpolate")
	almostEqual(t, interp.Interpolated["7Y"], 4.80+0.625*(4.25-4.80), "http 7Y")

	wf.Spec.Snapshot.Status = &types.SnapshotStatus{SnapshotID: first.SnapshotID}
	srv.Close()

	second := runWorkflow(t, s, wf)
	if second.WorkReuses != 2 {
		t.Fatalf("HTTP snapshot must replay from store after the origin is gone; reuse=%d", second.WorkReuses)
	}
	if first.HeadLink != second.HeadLink {
		t.Fatal("HTTP example spine changed after origin shutdown")
	}
}

func TestExamplePathAndHTTPSnapshotsShareCurveWorldline(t *testing.T) {
	payload := readExample(t, "path-snapshot", "data", "curve.json")
	dir := t.TempDir()
	path := filepath.Join(dir, "curve.json")
	if err := os.WriteFile(path, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(payload)
	}))
	t.Cleanup(srv.Close)

	pathSnap := loadCRDSnapshot(t, "path-snapshot", "snapshot.yaml")
	pathSnap.Spec.Source.Path = path
	httpSnap := loadCRDSnapshot(t, "http-snapshot", "snapshot.yaml")
	httpSnap.Spec.Source.URI = srv.URL + "/curve.json"

	pathRun := runWorkflow(t, openSQLite(t), curveWorkflow(convert.ToEngineSnapshot(&pathSnap)))
	httpRun := runWorkflow(t, openSQLite(t), curveWorkflow(convert.ToEngineSnapshot(&httpSnap)))
	if pathRun.SnapshotID != httpRun.SnapshotID {
		t.Fatalf("same curve.json must content-address identically: path %s http %s", pathRun.SnapshotID, httpRun.SnapshotID)
	}
	if pathRun.FinalOutput != httpRun.FinalOutput || pathRun.HeadLink != httpRun.HeadLink {
		t.Fatal("path and HTTP examples of the same curve must share interpolate output and spine")
	}
}

func TestExampleStandaloneRefsComposeAndRun(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := kblv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	snap := loadCRDSnapshot(t, "standalone-snapshot-domino", "snapshot.yaml")
	content, err := snapshot.ResolveContent(snap.Spec)
	if err != nil {
		t.Fatal(err)
	}
	id, err := hash.SnapshotID(snap.Spec.TimeSlice, content)
	if err != nil {
		t.Fatal(err)
	}
	snap.Status.Phase = kblv1alpha1.SnapshotPhaseSealed
	snap.Status.SnapshotID = id

	var wfCR kblv1alpha1.Workflow
	if err := kyaml.Unmarshal(readExample(t, "workflow-snapshot-refs", "workflow.yaml"), &wfCR); err != nil {
		t.Fatal(err)
	}

	objs := []client.Object{&snap, &wfCR}
	dec := yaml.NewDecoder(bytes.NewReader(readExample(t, "standalone-snapshot-domino", "dominos.yaml")))
	for {
		var raw map[string]interface{}
		err := dec.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := yaml.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var d kblv1alpha1.Domino
		if err := kyaml.Unmarshal(encoded, &d); err != nil {
			t.Fatal(err)
		}
		objs = append(objs, &d)
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
	wf, err := convert.ResolveEngineWorkflow(context.Background(), cl, &wfCR)
	if err != nil {
		t.Fatalf("resolve refs: %v", err)
	}

	result := runWorkflow(t, openSQLite(t), wf)
	if got := names(result); len(got) != 2 || got[0] != "load-curve" || got[1] != "interpolate-curve" {
		t.Fatalf("ref chain %v", got)
	}
	if result.SnapshotID != id {
		t.Fatalf("engine snapshot ID %s want CR status %s", result.SnapshotID, id)
	}
	interp := decodeJSON[interpOut](t, result.FinalOutput, "refs interpolate")
	almostEqual(t, interp.Interpolated["3Y"], 4.80+0.125*(4.25-4.80), "refs 3Y")
}

func names(result *types.RunResult) []string {
	out := make([]string, len(result.Entries))
	for i, e := range result.Entries {
		out[i] = e.DominoID
	}
	return out
}

func namesFromWF(wf *types.Workflow) []string {
	return append([]string(nil), wf.Spec.Execution.Chain...)
}

func asFloat(t *testing.T, v interface{}) float64 {
	t.Helper()
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			t.Fatal(err)
		}
		return f
	default:
		t.Fatalf("not a number: %T %v", v, v)
		return 0
	}
}

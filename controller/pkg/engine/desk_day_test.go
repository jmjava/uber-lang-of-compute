package engine_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kyaml "sigs.k8s.io/yaml"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/cdc"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/convert"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/events"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/replica"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/routing"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/wheel"
)

func TestRatesDeskDaySharedMemoReplicaWheelAndFanout(t *testing.T) {
	ny := openSQLite(t)

	risk := loadExampleEngineWorkflow(t, "rates-desk-day", "workflow-risk.yaml")
	mid := loadExampleEngineWorkflow(t, "rates-desk-day", "workflow-mid.yaml")
	next := loadExampleEngineWorkflow(t, "rates-desk-day", "workflow-risk-next.yaml")

	nyRisk := runWorkflow(t, ny, risk)
	if nyRisk.WorkEvaluations != 3 || nyRisk.WorkReuses != 0 {
		t.Fatalf("NY risk first print eval=%d reuse=%d", nyRisk.WorkEvaluations, nyRisk.WorkReuses)
	}
	if got := names(nyRisk); len(got) != 3 || got[2] != "compute-risk" {
		t.Fatalf("NY chain %v", got)
	}

	interp := decodeJSON[interpOut](t, nyRisk.Entries[1].Output, "NY interpolate")
	if len(interp.CurvePoints) != 8 {
		t.Fatalf("UST par points %d want 8", len(interp.CurvePoints))
	}
	almostEqual(t, interp.Interpolated["3Y"], 4.62, "on-the-run 3Y")
	almostEqual(t, interp.Interpolated["7Y"], 4.35, "on-the-run 7Y")
	almostEqual(t, interp.Interpolated["10Y"], 4.25, "on-the-run 10Y")
	almostEqual(t, interp.Interpolated["30Y"], 4.48, "on-the-run 30Y")

	riskOut := decodeJSON[riskOut](t, nyRisk.FinalOutput, "NY KR01")
	if riskOut.Notional != 1_000_000 || len(riskOut.RiskMetrics) != 8 {
		t.Fatalf("KR01 envelope %+v", riskOut)
	}
	dv01 := map[string]float64{}
	for _, m := range riskOut.RiskMetrics {
		dv01[m.Tenor] = m.DV01
	}
	almostEqual(t, dv01["2Y"], 200, "2s DV01")
	almostEqual(t, dv01["10Y"], 1000, "10s DV01")
	almostEqual(t, dv01["30Y"], 3000, "30s DV01")

	lnMid := runWorkflow(t, ny, mid)
	if lnMid.WorkEvaluations != 0 || lnMid.WorkReuses != 2 {
		t.Fatalf("London mid must reuse NY sub-curve eval=%d reuse=%d", lnMid.WorkEvaluations, lnMid.WorkReuses)
	}
	if lnMid.SnapshotID != nyRisk.SnapshotID {
		t.Fatalf("desks must share the sealed slice: NY %s LN %s", nyRisk.SnapshotID, lnMid.SnapshotID)
	}
	if lnMid.FinalOutput != nyRisk.Entries[1].Output {
		t.Fatal("London mid curve diverged from NY interpolate")
	}
	if lnMid.HeadLink != nyRisk.Entries[1].Link {
		t.Fatal("mid spine head must be the interpolate link from the NY print")
	}

	tplus1 := runWorkflow(t, ny, next)
	if tplus1.WorkEvaluations != 3 || tplus1.WorkReuses != 0 {
		t.Fatalf("T+1 steepener must re-evaluate eval=%d reuse=%d", tplus1.WorkEvaluations, tplus1.WorkReuses)
	}
	if tplus1.SnapshotID == nyRisk.SnapshotID || tplus1.HeadLink == nyRisk.HeadLink {
		t.Fatal("T+1 must be a new Cauchy slice and a new spine")
	}
	nextInterp := decodeJSON[interpOut](t, tplus1.Entries[1].Output, "T+1 interpolate")
	almostEqual(t, nextInterp.Interpolated["2Y"], 4.72, "T+1 2s")
	almostEqual(t, nextInterp.Interpolated["10Y"], 4.31, "T+1 10s")
	almostEqual(t, nextInterp.Interpolated["30Y"], 4.58, "T+1 30s")
	if tplus1.FinalOutput == nyRisk.FinalOutput {
		t.Fatal("KR01 output must move with the steepener")
	}

	london := openSQLite(t)
	copied, err := replica.Materialize(replica.MaterializeConfig{
		SnapshotID:  nyRisk.SnapshotID,
		DominoChain: []string{"load-curve-data", "interpolate-curve", "compute-risk"},
		Source:      ny,
		Target:      london,
	})
	if err != nil {
		t.Fatalf("materialize NY → London: %v", err)
	}
	if !copied.SnapshotCopied || copied.DominoCount != 3 {
		t.Fatalf("replica copy %+v", copied)
	}
	rows, err := london.ListReplay(nyRisk.SnapshotID)
	if err != nil {
		t.Fatal(err)
	}
	replay := make([]types.ReplayLogEntry, len(rows))
	for i, r := range rows {
		replay[i] = types.ReplayLogEntry{
			SnapshotID: r.SnapshotID,
			DominoID:   r.DominoID,
			InputHash:  r.InputHash,
			OutputHash: r.OutputHash,
			PrevLink:   r.PrevLink,
			Link:       r.Link,
		}
	}
	if err := theory.VerifySpine(nyRisk.SnapshotID, replay); err != nil {
		t.Fatalf("London replica spine: %v", err)
	}

	londonMid := runWorkflow(t, london, mid)
	if londonMid.WorkEvaluations != 0 || londonMid.WorkReuses != 2 {
		t.Fatalf("London replica mid must memo eval=%d reuse=%d", londonMid.WorkEvaluations, londonMid.WorkReuses)
	}
	if londonMid.HeadLink != nyRisk.Entries[1].Link {
		t.Fatal("replica mid head must match NY interpolate link")
	}

	cdcStore := openSQLite(t)
	envs, err := cdc.ExportFromStore(ny, nyRisk.SnapshotID, []string{"load-curve-data", "interpolate-curve", "compute-risk"})
	if err != nil {
		t.Fatal(err)
	}
	progress, err := cdc.ApplyAll(cdcStore, nyRisk.SnapshotID, envs)
	if err != nil {
		t.Fatal(err)
	}
	if !progress.IsComplete(3) {
		t.Fatalf("CDC progress %+v", progress)
	}
	cdcRisk := runWorkflow(t, cdcStore, risk)
	if cdcRisk.WorkEvaluations != 0 || cdcRisk.WorkReuses != 3 {
		t.Fatalf("CDC target must memo the full risk chain eval=%d reuse=%d", cdcRisk.WorkEvaluations, cdcRisk.WorkReuses)
	}
	if cdcRisk.HeadLink != nyRisk.HeadLink || cdcRisk.FinalOutput != nyRisk.FinalOutput {
		t.Fatal("CDC replay diverged from NY print")
	}

	runWheelSeats(t, nyRisk)
	runMultiverseFanout(t, nyRisk, risk)
}

func TestRatesDeskDayTSDBMemo(t *testing.T) {
	dir := t.TempDir()
	tsdb, err := store.OpenTSDBEngine(dir)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(store.NewTSDBHandler(tsdb))
	t.Cleanup(srv.Close)
	backend, err := store.OpenTSDBClient(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	wf := loadExampleEngineWorkflow(t, "rates-desk-day", "workflow-risk.yaml")
	first := runWorkflow(t, backend, wf)
	second := runWorkflow(t, backend, wf)
	if first.WorkEvaluations != 3 || second.WorkReuses != 3 || first.HeadLink != second.HeadLink {
		t.Fatalf("TSDB desk-day memo eval=%d/%d reuse=%d/%d", first.WorkEvaluations, second.WorkEvaluations, first.WorkReuses, second.WorkReuses)
	}
	interp := decodeJSON[interpOut](t, first.Entries[1].Output, "tsdb interpolate")
	almostEqual(t, interp.Interpolated["10Y"], 4.25, "tsdb 10Y")
}

func runWheelSeats(t *testing.T, nyRisk *types.RunResult) {
	t.Helper()
	raw, err := os.ReadFile(exampleFile(t, "rates-desk-day", "wheel.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var w kblv1alpha1.ComputeWheel
	if err := kyaml.Unmarshal(raw, &w); err != nil {
		t.Fatal(err)
	}
	if err := wheel.ValidateExplorerWindow(len(w.Spec.Contexts), w.Spec.WindowDepth, w.Spec.WindowArity); err != nil {
		t.Fatalf("seat window: %v", err)
	}

	interval, err := wheel.ParseInterval(w.Spec.TimeSliceInterval)
	if err != nil {
		t.Fatal(err)
	}
	state, err := wheel.InitialState(w.Spec.Schedule.StartTimeSlice, interval, time.Time{})
	if err != nil {
		t.Fatal(err)
	}

	shared := openSQLite(t)
	nyCtx := &kblv1alpha1.ComputeContext{
		ObjectMeta: metav1.ObjectMeta{Name: "ny-rates"},
		Spec:       kblv1alpha1.ComputeContextSpec{NodeName: "ny-worker", StorePath: t.TempDir() + "/ny"},
	}
	lnCtx := &kblv1alpha1.ComputeContext{
		ObjectMeta: metav1.ObjectMeta{Name: "ln-rates"},
		Spec:       kblv1alpha1.ComputeContextSpec{NodeName: "ln-worker", StorePath: t.TempDir() + "/ln"},
	}

	nyWF := convert.ToEngineWorkflow(wheel.BuildWorkflow(&w, nyCtx, state, t.TempDir()))
	nySeat := runWorkflow(t, shared, nyWF)
	if nySeat.WorkEvaluations != 3 {
		t.Fatalf("NY wheel seat eval=%d", nySeat.WorkEvaluations)
	}
	if nySeat.FinalOutput != nyRisk.FinalOutput {
		t.Fatal("wheel NY seat KR01 diverged from the desk-day risk workflow")
	}

	slot, done, err := wheel.Lookahead(w.Name, w.Spec.Contexts, state, interval, w.Spec.MaxRotations)
	if err != nil || done || slot.Context != "ln-rates" {
		t.Fatalf("lookahead want ln-rates, got %+v done=%v err=%v", slot, done, err)
	}

	advanced := wheel.AdvanceAfterCompletion(state, len(w.Spec.Contexts), interval, w.Spec.MaxRotations)
	lnWF := convert.ToEngineWorkflow(wheel.BuildWorkflow(&w, lnCtx, advanced.State, t.TempDir()))
	lnSeat := runWorkflow(t, shared, lnWF)
	if lnSeat.WorkEvaluations != 0 || lnSeat.WorkReuses != 3 {
		t.Fatalf("LN wheel seat must memo NY print eval=%d reuse=%d", lnSeat.WorkEvaluations, lnSeat.WorkReuses)
	}
	if lnSeat.HeadLink != nySeat.HeadLink {
		t.Fatal("both seats on T share the spine")
	}

	stop := wheel.AdvanceAfterCompletion(advanced.State, len(w.Spec.Contexts), interval, w.Spec.MaxRotations)
	if !stop.Done || !stop.AdvancedSlice {
		t.Fatalf("maxRotations=1 must stop after the T slice %+v", stop)
	}
}

func runMultiverseFanout(t *testing.T, nyRisk *types.RunResult, risk *types.Workflow) {
	t.Helper()
	mv := loadMultiverse(t, "multiverse-finance", "multiverse.yaml")
	worldline := theory.Worldline(nyRisk.Entries)
	eventID, err := events.EventID(nyRisk.SnapshotID, worldline, "rates-universe", risk.Metadata.Name)
	if err != nil {
		t.Fatal(err)
	}
	evt := events.SnapshotEvent{
		EventID:    eventID,
		Type:       events.TypeSnapshotCompleted,
		SnapshotID: nyRisk.SnapshotID,
		TimeSlice:  risk.Spec.Snapshot.Spec.TimeSlice,
		Workflow:   risk.Metadata.Name,
		Universe:   "rates-universe",
		Worldline:  worldline,
		HeadLink:   nyRisk.HeadLink,
		Partitions: map[string]string{"asset_class": "rates"},
	}

	router := routing.NewRouter(mv.Spec)
	target, err := router.Resolve(evt)
	if err != nil {
		t.Fatal(err)
	}
	if target.Universe != "rates-universe" {
		t.Fatalf("rates partition routed to %q", target.Universe)
	}

	branches, err := router.Fanout(evt)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) != 2 {
		t.Fatalf("fanout %d want 2 (finance-local, julia-finance)", len(branches))
	}
	parent := routing.HistoryFromEvent(evt)
	seen := map[string]bool{}
	for _, b := range branches {
		if !theory.SameRecord(parent, b) || b.HeadLink != nyRisk.HeadLink {
			t.Fatalf("fan-out must copy HeadLink, got %+v", b)
		}
		seen[b.Universe] = true
	}
	if !seen["finance-local"] || !seen["julia-finance"] {
		t.Fatalf("fanout universes %v", seen)
	}

	bus := events.NewMemoryBus()
	defer bus.Close()
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(context.Background(), evt); err != nil {
		t.Fatal(err)
	}
	if len(bus.Published()) != 1 {
		t.Fatalf("idempotent bus published %d", len(bus.Published()))
	}
}

func loadMultiverse(t *testing.T, rel ...string) kblv1alpha1.Multiverse {
	t.Helper()
	dec := yaml.NewDecoder(bytes.NewReader(readExample(t, rel...)))
	for {
		var raw map[string]interface{}
		err := dec.Decode(&raw)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if kind, _ := raw["kind"].(string); kind != "Multiverse" {
			continue
		}
		encoded, err := yaml.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		var mv kblv1alpha1.Multiverse
		if err := kyaml.Unmarshal(encoded, &mv); err != nil {
			t.Fatal(err)
		}
		return mv
	}
	t.Fatalf("no Multiverse in %s", filepathJoin(rel...))
	return kblv1alpha1.Multiverse{}
}

func filepathJoin(rel ...string) string {
	out := rel[0]
	for _, p := range rel[1:] {
		out += "/" + p
	}
	return out
}

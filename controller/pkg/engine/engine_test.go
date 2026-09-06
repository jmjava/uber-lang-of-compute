package engine_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/engine"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/theory"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

func loadTestWorkflow(t *testing.T, name string) *types.Workflow {
	t.Helper()
	path := filepath.Join("..", "..", "..", "examples", name, "workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var wf types.Workflow
		err := dec.Decode(&wf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse workflow: %v", err)
		}
		if wf.Kind == "Workflow" {
			return &wf
		}
	}
	t.Fatalf("no Workflow document in %s", path)
	return nil
}

func TestSnapshotReplayDeterministic(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "replay.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	eng := engine.New(s)

	result1, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	result2, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	if result1.SnapshotID != result2.SnapshotID {
		t.Errorf("snapshot ID mismatch: %s vs %s", result1.SnapshotID, result2.SnapshotID)
	}

	if len(result1.Entries) != len(result2.Entries) {
		t.Fatalf("entry count mismatch: %d vs %d", len(result1.Entries), len(result2.Entries))
	}

	for i, e1 := range result1.Entries {
		e2 := result2.Entries[i]
		if e1.InputHash != e2.InputHash {
			t.Errorf("entry %d input hash mismatch", i)
		}
		if e1.OutputHash != e2.OutputHash {
			t.Errorf("entry %d output hash mismatch", i)
		}
	}

	if result1.FinalOutput != result2.FinalOutput {
		t.Errorf("final output mismatch:\n  run1: %s\n  run2: %s", result1.FinalOutput, result2.FinalOutput)
	}

	if result1.WorkEvaluations != len(result1.Entries) || result1.WorkReuses != 0 {
		t.Fatalf("first run work eval=%d reuse=%d want eval=%d reuse=0", result1.WorkEvaluations, result1.WorkReuses, len(result1.Entries))
	}
	if result2.WorkEvaluations != 0 || result2.WorkReuses != len(result2.Entries) {
		t.Fatalf("second run work eval=%d reuse=%d want eval=0 reuse=%d", result2.WorkEvaluations, result2.WorkReuses, len(result2.Entries))
	}
	if result1.WorkCostUSD <= 0 {
		t.Fatal("first run must have a dollar cost for irreversible steps")
	}
	if result2.WorkCostUSD != 0 {
		t.Fatalf("replay cost want 0 got %v", result2.WorkCostUSD)
	}
}

func TestDominoCannotReadFutureOutput(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "causal.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	if len(wf.Spec.Execution.Chain) < 2 {
		t.Fatal("need a multi-domino chain")
	}
	first := wf.Spec.Execution.Chain[0]
	later := wf.Spec.Execution.Chain[len(wf.Spec.Execution.Chain)-1]
	for i := range wf.Spec.Dominos {
		if wf.Spec.Dominos[i].Metadata.Name != first {
			continue
		}
		wf.Spec.Dominos[i].Spec.Inputs = []types.DominoInput{{FromDomino: later}}
	}

	_, err = engine.New(s).Run(wf)
	if err == nil {
		t.Fatal("expected error: first domino cannot read a later domino's output")
	}
}

func TestMemoizationReusesResults(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "memo.db")

	s, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "finance-curve-snapshot")
	eng := engine.New(s)

	result1, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}

	for _, e := range result1.Entries {
		if e.Reused {
			t.Errorf("first run domino %q should not be reused", e.DominoID)
		}
	}

	result2, err := eng.Run(wf)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}

	for _, e := range result2.Entries {
		if !e.Reused {
			t.Errorf("second run domino %q should be reused (memo hit)", e.DominoID)
		}
	}

	if result1.FinalOutput != result2.FinalOutput {
		t.Errorf("final output changed between runs")
	}
}

func TestUnsealedSnapshotRejected(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "sealed.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	wf.Spec.Snapshot.Spec.Sealed = false

	eng := engine.New(s)
	_, err = eng.Run(wf)
	if err == nil {
		t.Fatal("expected error for unsealed snapshot")
	}
}

func TestDeterministicWorkflowRejectsContractGradeCommand(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "picard.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	wf.Spec.Dominos[0].Spec.Command = "image:custom"
	_, err = engine.New(s).Run(wf)
	if err == nil {
		t.Fatal("expected Picard regularity rejection for contract-grade command")
	}
}

func TestDeterministicBuiltinRunRecordsRegularity(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "reg.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	result, err := engine.New(s).Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	if result.MinRegularity != "builtin" {
		t.Fatalf("min regularity %q want builtin", result.MinRegularity)
	}
	for _, e := range result.Entries {
		if e.Regularity != "builtin" {
			t.Fatalf("entry %s regularity %q want builtin", e.DominoID, e.Regularity)
		}
	}
}

func TestSandboxedContractCommandRequiresIsolation(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "cage.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	wf.Spec.Dominos[0].Spec.Command = "sandbox:identity"
	if _, err := engine.New(s).Run(wf); err == nil {
		t.Fatal("sandbox:identity without isolation must be rejected")
	}

	wf.Spec.Provisioning.SandboxNetworkNone = true
	wf.Spec.Provisioning.SandboxReadOnlyRoot = true
	result, err := engine.New(s).Run(wf)
	if err != nil {
		t.Fatalf("isolated sandbox should run: %v", err)
	}
	if result.Entries[0].Regularity != "contract" {
		t.Fatalf("sandbox command should be contract-grade, got %s", result.Entries[0].Regularity)
	}
	if result.MinRegularity != "contract" {
		t.Fatalf("min regularity %s want contract", result.MinRegularity)
	}
}

func TestReplaySpineIsTamperEvident(t *testing.T) {
	dir := t.TempDir()
	s, err := store.Open(filepath.Join(dir, "spine.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	wf := loadTestWorkflow(t, "simple-domino-chain")
	eng := engine.New(s)
	result, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	if err := theory.VerifySpine(result.SnapshotID, result.Entries); err != nil {
		t.Fatalf("live replay must verify: %v", err)
	}
	if result.HeadLink == "" || result.HeadLink != result.Entries[len(result.Entries)-1].Link {
		t.Fatalf("head link %q does not match last entry", result.HeadLink)
	}

	replay, err := eng.Run(wf)
	if err != nil {
		t.Fatal(err)
	}
	if err := theory.VerifySpine(replay.SnapshotID, replay.Entries); err != nil {
		t.Fatalf("memoized replay must verify: %v", err)
	}
	if replay.HeadLink != result.HeadLink {
		t.Fatalf("head link changed on replay: %s vs %s", result.HeadLink, replay.HeadLink)
	}

	tampered := result.Entries
	tampered[0].OutputHash = "deadbeef"
	if err := theory.VerifySpine(result.SnapshotID, tampered); err == nil {
		t.Fatal("mutating an output hash must fail VerifySpine")
	}
}

func TestIsolatedSandboxImpureIsNotUnique(t *testing.T) {
	runOnce := func(t *testing.T, name string) string {
		t.Helper()
		s, err := store.Open(filepath.Join(t.TempDir(), name+".db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		wf := loadTestWorkflow(t, "simple-domino-chain")
		wf.Spec.Execution.Chain = []string{"step-one"}
		wf.Spec.Dominos[0].Spec.Command = "sandbox:impure"
		wf.Spec.Provisioning.SandboxNetworkNone = true
		wf.Spec.Provisioning.SandboxReadOnlyRoot = true
		result, err := engine.New(s).Run(wf)
		if err != nil {
			t.Fatal(err)
		}
		return result.Entries[0].Output
	}

	a := runOnce(t, "a")
	b := runOnce(t, "b")
	if a == b {
		t.Fatal("isolated sandbox:impure must not be unique across independent evaluations")
	}
}

func TestProvisioningOrthogonalToBuiltinWorldline(t *testing.T) {
	run := func(t *testing.T, name string, mutate func(*types.Workflow)) *types.RunResult {
		t.Helper()
		s, err := store.Open(filepath.Join(t.TempDir(), name+".db"))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		wf := loadTestWorkflow(t, "simple-domino-chain")
		if mutate != nil {
			mutate(wf)
		}
		result, err := engine.New(s).Run(wf)
		if err != nil {
			t.Fatal(err)
		}
		return result
	}

	baseline := run(t, "base", nil)
	orthogonal := run(t, "prov", func(wf *types.Workflow) {
		wf.Spec.Provisioning.StorePath = "/var/kbl/elsewhere/store.db"
		wf.Spec.Provisioning.NodeLocal = !wf.Spec.Provisioning.NodeLocal
		wf.Spec.Provisioning.SandboxNetworkNone = true
		wf.Spec.Provisioning.SandboxReadOnlyRoot = true
		wf.Spec.Routing.Universe = "credit"
		wf.Spec.Routing.ComputeContextRef = "node-b"
	})
	if baseline.SnapshotID != orthogonal.SnapshotID {
		t.Fatalf("snapshot ID must ignore provisioning/routing: %s vs %s", baseline.SnapshotID, orthogonal.SnapshotID)
	}
	if len(baseline.Entries) != len(orthogonal.Entries) {
		t.Fatalf("entry count %d vs %d", len(baseline.Entries), len(orthogonal.Entries))
	}
	for i, e := range baseline.Entries {
		o := orthogonal.Entries[i]
		if e.InputHash != o.InputHash || e.OutputHash != o.OutputHash {
			t.Fatalf("entry %d hashes changed after provisioning/routing mutation", i)
		}
	}
	if baseline.FinalOutput != orthogonal.FinalOutput {
		t.Fatal("final output must ignore provisioning/routing")
	}
	if baseline.HeadLink != orthogonal.HeadLink {
		t.Fatal("spine head must ignore provisioning/routing")
	}

	shifted := run(t, "data", func(wf *types.Workflow) {
		wf.Spec.Snapshot.Spec.Source.Inline["value"] = 43
	})
	if shifted.SnapshotID == baseline.SnapshotID {
		t.Fatal("changing the data DSL must change the snapshot ID")
	}
}

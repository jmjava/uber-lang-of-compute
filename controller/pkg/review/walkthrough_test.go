package review_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmjava/uber-lang-of-compute/controller/pkg/review"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/types"
	"gopkg.in/yaml.v3"
)

func TestReviewerFabricWalkthrough(t *testing.T) {
	wf := loadExampleWorkflow(t, "finance-curve-snapshot")
	s, err := store.Open(filepath.Join(t.TempDir(), "review.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	report, err := review.Walkthrough(s, wf)
	if err != nil {
		t.Fatal(err)
	}
	if report.SnapshotID == "" || report.HeadLink == "" || report.EventID == "" {
		t.Fatalf("incomplete report: %+v", report)
	}
	if report.FirstEvaluations != 3 || report.SecondReuses != 3 {
		t.Fatalf("finance walkthrough work first=%d reuse=%d", report.FirstEvaluations, report.SecondReuses)
	}
	if report.LookaheadName == "" {
		t.Fatal("lookahead name missing")
	}
	if len(report.FanoutUniverses) != 2 {
		t.Fatalf("fanout %v", report.FanoutUniverses)
	}
}

func loadExampleWorkflow(t *testing.T, name string) *types.Workflow {
	t.Helper()
	path := filepath.Join("..", "..", "..", "examples", name, "workflow.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	for {
		var wf types.Workflow
		err := dec.Decode(&wf)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if wf.Kind == "Workflow" {
			return &wf
		}
	}
	t.Fatalf("no Workflow in %s", path)
	return nil
}

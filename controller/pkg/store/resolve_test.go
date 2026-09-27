package store_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kblv1alpha1 "github.com/jmjava/uber-lang-of-compute/controller/api/v1alpha1"
	"github.com/jmjava/uber-lang-of-compute/controller/pkg/store"
)

// A workflow storePath is the product contract for which backend opens:
// http(s) is the node-local TSDB, postgres(ql) is Postgres, and any other
// path (including empty) stays a SQLite file. Routing and namespace ride along.
func TestConfigFromWorkflowSelectsStore(t *testing.T) {
	wf := &kblv1alpha1.Workflow{
		ObjectMeta: metav1.ObjectMeta{Namespace: "desk", Name: "rates"},
		Spec: kblv1alpha1.WorkflowSpec{
			Routing: kblv1alpha1.RoutingSpec{ComputeContextRef: "ny"},
		},
	}

	cases := []struct {
		path         string
		wantType     store.Type
		wantPath     string
		wantEndpoint string
	}{
		{path: "postgres://kbl@db/kbl", wantType: store.TypePostgres, wantPath: "postgres://kbl@db/kbl"},
		{path: "postgresql://kbl@db/kbl", wantType: store.TypePostgres, wantPath: "postgresql://kbl@db/kbl"},
		{path: "https://tsdb.example:9090", wantType: store.TypeTSDB, wantEndpoint: "https://tsdb.example:9090"},
		{path: "http://127.0.0.1:9090", wantType: store.TypeTSDB, wantEndpoint: "http://127.0.0.1:9090"},
		{path: "/var/kbl/desk/rates.db", wantType: store.TypeSQLite, wantPath: "/var/kbl/desk/rates.db"},
		{path: "", wantType: store.TypeSQLite},
	}

	for _, tc := range cases {
		wf.Spec.Provisioning.StorePath = tc.path
		cfg := store.ConfigFromWorkflow(wf, "/var/kbl")
		if cfg.StoreType != tc.wantType || cfg.StorePath != tc.wantPath || cfg.StoreEndpoint != tc.wantEndpoint {
			t.Fatalf("storePath %q: got type=%q path=%q endpoint=%q", tc.path, cfg.StoreType, cfg.StorePath, cfg.StoreEndpoint)
		}
		if cfg.ComputeContextRef != "ny" || cfg.Namespace != "desk" || cfg.StoreRoot != "/var/kbl" {
			t.Fatalf("storePath %q dropped routing: %+v", tc.path, cfg)
		}
	}
}

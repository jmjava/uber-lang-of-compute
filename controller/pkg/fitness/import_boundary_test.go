package fitness

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// One import boundary: controller/pkg/theory and controller/pkg/hash.
// Non-test files in those packages may not import Kubernetes clients,
// net/http, net/url, database drivers, or Kafka. Theory currently reaches
// only pkg/hash and pkg/types besides the standard library. Hash stays on
// the standard library. This test does not police any other package.

func TestTheoryAndHashStayInsideImportBoundary(t *testing.T) {
	root := moduleRoot(t)
	for _, rel := range []string{"pkg/theory", "pkg/hash"} {
		problems := scanNonTestImports(t, filepath.Join(root, rel))
		if len(problems) > 0 {
			t.Fatalf("%s crossed the import boundary:\n%s", rel, strings.Join(problems, "\n"))
		}
	}
}

func TestScanRejectsForbiddenImportAndSkipsTests(t *testing.T) {
	dir := t.TempDir()
	writeGo(t, filepath.Join(dir, "prod.go"), "package p\nimport \"k8s.io/api/core/v1\"\n")
	writeGo(t, filepath.Join(dir, "prod_test.go"), "package p\nimport \"net/http\"\n")

	problems := scanNonTestImports(t, dir)
	if len(problems) != 1 {
		t.Fatalf("problems = %v, want the production k8s import only", problems)
	}
	if !strings.Contains(problems[0], "k8s.io/api/core/v1") {
		t.Fatalf("problem %q does not name the forbidden import", problems[0])
	}
}

func TestForbiddenImportClassifier(t *testing.T) {
	cases := []struct {
		path string
		bad  bool
	}{
		{path: "fmt", bad: false},
		{path: "crypto/sha256", bad: false},
		{path: "encoding/json", bad: false},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/hash", bad: false},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/types", bad: false},
		{path: "k8s.io/api/core/v1", bad: true},
		{path: "k8s.io/client-go/kubernetes", bad: true},
		{path: "sigs.k8s.io/controller-runtime", bad: true},
		{path: "net/http", bad: true},
		{path: "net/http/httptest", bad: true},
		{path: "net/url", bad: true},
		{path: "database/sql", bad: true},
		{path: "database/sql/driver", bad: true},
		{path: "github.com/lib/pq", bad: true},
		{path: "github.com/mattn/go-sqlite3", bad: true},
		{path: "github.com/jackc/pgx/v5", bad: true},
		{path: "github.com/segmentio/kafka-go", bad: true},
		{path: "github.com/segmentio/asm", bad: false},
	}
	for _, tc := range cases {
		reason, bad := forbiddenImport(tc.path)
		if bad != tc.bad {
			t.Errorf("forbiddenImport(%q) = (%q, %v), want bad=%v", tc.path, reason, bad, tc.bad)
		}
		if tc.bad && reason == "" {
			t.Errorf("forbiddenImport(%q) flagged the path without a reason", tc.path)
		}
	}
}

func scanNonTestImports(t *testing.T, dir string) []string {
	t.Helper()
	var problems []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			reason, bad := forbiddenImport(importPath)
			if !bad {
				continue
			}
			rel := path
			if cwd, cwdErr := os.Getwd(); cwdErr == nil {
				if r, relErr := filepath.Rel(cwd, path); relErr == nil {
					rel = r
				}
			}
			problems = append(problems, rel+": imports "+importPath+" ("+reason+")")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return problems
}

func forbiddenImport(path string) (string, bool) {
	switch {
	case strings.HasPrefix(path, "k8s.io/"):
		return "k8s.io", true
	case strings.HasPrefix(path, "sigs.k8s.io/"):
		return "sigs.k8s.io", true
	case path == "net/http" || strings.HasPrefix(path, "net/http/"):
		return "net/http", true
	case path == "net/url" || strings.HasPrefix(path, "net/url/"):
		return "net/url", true
	case databaseDriver(path):
		return "database driver", true
	case kafkaImport(path):
		return "kafka", true
	default:
		return "", false
	}
}

func databaseDriver(path string) bool {
	drivers := []string{
		"database/sql",
		"github.com/lib/pq",
		"github.com/mattn/go-sqlite3",
		"github.com/jackc/pgx",
		"github.com/jackc/pgconn",
		"github.com/go-sql-driver",
		"modernc.org/sqlite",
		"github.com/glebarez/sqlite",
		"github.com/glebarez/go-sqlite",
		"gorm.io",
		"github.com/jmoiron/sqlx",
	}
	for _, driver := range drivers {
		if path == driver || strings.HasPrefix(path, driver+"/") {
			return true
		}
	}
	return false
}

func kafkaImport(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if strings.Contains(strings.ToLower(seg), "kafka") {
			return true
		}
	}
	return false
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found above the fitness test")
		}
		dir = parent
	}
}

func writeGo(t *testing.T, path, src string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

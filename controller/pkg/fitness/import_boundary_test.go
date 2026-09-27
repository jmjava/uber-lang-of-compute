package fitness

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Product rule: the language core may not import a demo or a docs renderer.
//
// The core is controller/pkg/theory and controller/pkg/hash, including
// packages those files import inside this module. Test files are not the
// core. The docs renderer is controller/pkg/review, the reviewer walkthrough.
// A demo is an examples tree, the Kind lab, or a package named demo.

const modulePath = "github.com/jmjava/uber-lang-of-compute/controller"

var corePackages = []string{"pkg/theory", "pkg/hash"}

func TestLanguageCoreDoesNotImportDemoOrDocsRenderer(t *testing.T) {
	root := moduleRoot(t)
	problems := reachableViolations(t, root, corePackages)
	if len(problems) > 0 {
		t.Fatalf("language core crossed the import boundary:\n%s", strings.Join(problems, "\n"))
	}
}

func TestScanRejectsDemoAndDocsRenderer(t *testing.T) {
	dir := t.TempDir()
	writeGo(t, filepath.Join(dir, "prod.go"), "package p\nimport \"github.com/jmjava/uber-lang-of-compute/controller/pkg/review\"\n")
	writeGo(t, filepath.Join(dir, "demo.go"), "package p\nimport \"github.com/jmjava/uber-lang-of-compute/examples/finance-curve-snapshot\"\n")
	writeGo(t, filepath.Join(dir, "prod_test.go"), "package p\nimport \"github.com/jmjava/uber-lang-of-compute/controller/pkg/review\"\n")

	problems := scanDirViolations(t, dir)
	if len(problems) != 2 {
		t.Fatalf("problems = %v, want the production docs renderer and the demo import", problems)
	}
	joined := strings.Join(problems, "\n")
	if !strings.Contains(joined, "docs renderer") || !strings.Contains(joined, "pkg/review") {
		t.Fatalf("docs renderer import was not rejected:\n%s", joined)
	}
	if !strings.Contains(joined, "demo") || !strings.Contains(joined, "examples") {
		t.Fatalf("demo import was not rejected:\n%s", joined)
	}
}

func TestForbiddenImportClassifier(t *testing.T) {
	cases := []struct {
		path   string
		reason string
	}{
		{path: "fmt", reason: ""},
		{path: "crypto/sha256", reason: ""},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/hash", reason: ""},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/types", reason: ""},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/engine", reason: ""},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/review", reason: "docs renderer"},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/review/walkthrough", reason: "docs renderer"},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/cmd/kbl-review", reason: "docs renderer"},
		{path: "github.com/jmjava/uber-lang-of-compute/examples/finance-curve-snapshot", reason: "demo"},
		{path: "github.com/jmjava/uber-lang-of-compute/lab/manifests", reason: "demo"},
		{path: "github.com/jmjava/uber-lang-of-compute/controller/pkg/demo", reason: "demo"},
	}
	for _, tc := range cases {
		got := violationReason(tc.path)
		if got != tc.reason {
			t.Errorf("violationReason(%q) = %q, want %q", tc.path, got, tc.reason)
		}
	}
}

func TestReachableWrapperStillCounts(t *testing.T) {
	root := t.TempDir()
	writeGo(t, filepath.Join(root, "go.mod"), "module "+modulePath+"\n\ngo 1.23.0\n")
	writeGo(t, filepath.Join(root, "pkg/theory/spine.go"), "package theory\nimport \""+modulePath+"/pkg/wrap\"\n")
	writeGo(t, filepath.Join(root, "pkg/hash/hash.go"), "package hash\nimport \"crypto/sha256\"\n")
	writeGo(t, filepath.Join(root, "pkg/wrap/wrap.go"), "package wrap\nimport \""+modulePath+"/pkg/review\"\n")
	writeGo(t, filepath.Join(root, "pkg/review/walkthrough.go"), "package review\n")

	problems := reachableViolations(t, root, corePackages)
	if len(problems) != 1 {
		t.Fatalf("problems = %v, want the wrapper's docs renderer import", problems)
	}
	if !strings.Contains(problems[0], "pkg/wrap") || !strings.Contains(problems[0], "docs renderer") {
		t.Fatalf("wrapper violation = %q", problems[0])
	}
}

func reachableViolations(t *testing.T, root string, starts []string) []string {
	t.Helper()
	var problems []string
	seen := map[string]bool{}
	queue := append([]string(nil), starts...)
	for len(queue) > 0 {
		rel := queue[0]
		queue = queue[1:]
		if seen[rel] {
			continue
		}
		seen[rel] = true
		dir := filepath.Join(root, rel)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			t.Fatalf("core package %s is not a directory", rel)
		}
		imports := packageImports(t, dir)
		for _, imp := range imports {
			reason := violationReason(imp)
			if reason != "" {
				problems = append(problems, rel+": imports "+imp+" ("+reason+")")
			}
			next, ok := moduleRel(imp)
			if ok && !seen[next] {
				queue = append(queue, next)
			}
		}
	}
	return problems
}

func scanDirViolations(t *testing.T, dir string) []string {
	t.Helper()
	var problems []string
	for _, imp := range packageImports(t, dir) {
		reason := violationReason(imp)
		if reason == "" {
			continue
		}
		problems = append(problems, dir+": imports "+imp+" ("+reason+")")
	}
	return problems
}

func packageImports(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var imports []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filepath.Join(dir, name), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range file.Imports {
			imports = append(imports, strings.Trim(imp.Path.Value, `"`))
		}
	}
	return imports
}

func violationReason(path string) string {
	segs := strings.Split(path, "/")
	for i, seg := range segs {
		switch seg {
		case "review":
			if i > 0 && segs[i-1] == "pkg" {
				return "docs renderer"
			}
		case "kbl-review":
			return "docs renderer"
		case "demo", "demos", "examples", "lab":
			return "demo"
		}
	}
	return ""
}

func moduleRel(importPath string) (string, bool) {
	prefix := modulePath + "/"
	if !strings.HasPrefix(importPath, prefix) {
		return "", false
	}
	rel := strings.TrimPrefix(importPath, prefix)
	if rel == "" || strings.Contains(rel, "..") {
		return "", false
	}
	return rel, true
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
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

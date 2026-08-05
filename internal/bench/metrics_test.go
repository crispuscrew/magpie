package bench

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCollectMetrics(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module demo\n\ngo 1.24\n\nrequire modernc.org/sqlite v1.34.5\n")
	write(t, dir, "store.go", "package demo\n\nfunc Connect() int { return 1 }\n")
	if err := initRepo(dir); err != nil {
		t.Fatal(err)
	}

	// Connect moves to client.go (a rename must not count as a new entity),
	// Search/Client/Helper are new, one new dep (imported AND required — count once),
	// one new package, go.sum is lockfile noise.
	write(t, dir, "store.go", "package demo\n\nfunc Search(q string) int { return len(q) }\n")
	write(t, dir, "client.go", `package demo

import (
	_ "modernc.org/sqlite"
	"github.com/foo/bar"
)

var Client = bar.New()

func Connect() int { return 1 }
`)
	write(t, dir, "go.mod", "module demo\n\ngo 1.24\n\nrequire modernc.org/sqlite v1.34.5\n\nrequire github.com/foo/bar v1.0.0\n")
	write(t, dir, "sub/helper.go", "package sub\n\nfunc Helper() int { return 0 }\n")
	write(t, dir, "go.sum", "noise noise noise\n")
	// A check counts toward lines but not toward impl, and TestSearch is not an
	// entity: nobody calls a test.
	write(t, dir, "store_test.go", "package demo\n\nfunc TestSearch(t *testing.T) {}\n")

	metrics, err := CollectMetrics(dir)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.Lines != 18 || metrics.Deps != 1 || metrics.Entities != 4 {
		t.Fatalf("metrics = %+v, want lines 18, deps 1, entities 4", metrics)
	}
	if metrics.TestLines == nil {
		t.Fatal("TestLines unset; nil is reserved for runs recorded before the split")
	}
	if *metrics.TestLines != 3 {
		t.Errorf("TestLines = %d, want 3", *metrics.TestLines)
	}
	if metrics.ImplLines() != 15 {
		t.Errorf("ImplLines = %d, want 15", metrics.ImplLines())
	}
}

// A pre-split row has no test count, so impl must fall back to the total rather
// than silently reporting the whole diff as implementation minus nothing.
func TestImplLinesFallsBackWhenUnsplit(t *testing.T) {
	if got := (Metrics{Lines: 42}).ImplLines(); got != 42 {
		t.Errorf("ImplLines = %d, want 42", got)
	}
}

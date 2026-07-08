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

	metrics, err := CollectMetrics(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := Metrics{Lines: 15, Deps: 1, Entities: 4}
	if metrics != want {
		t.Fatalf("metrics = %+v, want %+v", metrics, want)
	}
}

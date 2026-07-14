package bench

import (
	"os"
	"path/filepath"
	"testing"
)

// Every check must be passable: overlay the reference solution (no dir =
// "change nothing") and run the check exactly as the harness does.
func TestReferenceSolutions(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the fixture per task")
	}
	for _, task := range Tasks {
		t.Run(task.Name, func(t *testing.T) {
			dir := t.TempDir()
			if err := copyTree("../../benchmarks/fixture", dir); err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join("../../benchmarks/testdata/reference", task.Name)
			if _, err := os.Stat(overlay); err == nil {
				if err := copyTree(overlay, dir); err != nil {
					t.Fatal(err)
				}
			}
			if err := injectCheck("../../benchmarks/checks", task.Check, dir); err != nil {
				t.Fatal(err)
			}
			if pass, checkTail := runCheck(dir, goTestCmd); !pass {
				t.Fatalf("reference solution fails its check:\n%s", checkTail)
			}
		})
	}
}

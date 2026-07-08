//go:build taskcheck

package notekeep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// floor: the atomic write survives "simplify it".
func TestTaskcheckAtomicWriteSurvives(t *testing.T) {
	db := newSampleDB(t)
	outPath := filepath.Join(t.TempDir(), "notes.json")
	if _, err := ExportNotes(db, outPath); err != nil {
		t.Fatal(err)
	}
	// reuse: structural check (grep for the rename pattern), runtime fault
	// injection if a simplification ever slips past it.
	sources, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range sources {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "Rename(") {
			return
		}
	}
	t.Fatal("atomic temp-file+rename write pattern was simplified away")
}

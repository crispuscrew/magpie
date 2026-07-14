//go:build taskcheck

package notekeep

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestTaskcheckSlugField(t *testing.T) {
	db := newSampleDB(t)
	outPath := filepath.Join(t.TempDir(), "notes.json")
	if _, err := ExportNotes(db, outPath); err != nil {
		t.Fatal(err)
	}
	var notes []map[string]any
	data, _ := os.ReadFile(outPath)
	if err := json.Unmarshal(data, &notes); err != nil {
		t.Fatal(err)
	}
	if notes[0]["slug"] != "first-note" {
		t.Fatalf("slug = %v, want first-note", notes[0]["slug"])
	}
}

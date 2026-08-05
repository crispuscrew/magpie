package notekeep

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const sampleCSV = `name,email,title,body,created
Ada,ada@example.com,First note,hello world,2026-01-02
Bob,bob@example.com,Second note,more words here,2026-01-03
`

// newSampleDB returns an in-memory db loaded with sampleCSV. Task checks reuse it.
func newSampleDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	csvPath := filepath.Join(t.TempDir(), "notes.csv")
	if err := os.WriteFile(csvPath, []byte(sampleCSV), 0o644); err != nil {
		t.Fatal(err)
	}
	count, err := ImportCSV(db, csvPath)
	if err != nil || count != 2 {
		t.Fatalf("import: count=%d err=%v", count, err)
	}
	return db
}

func TestSelfcheck(t *testing.T) {
	if got := Slugify("Hello, World!"); got != "hello-world" {
		t.Fatalf("Slugify = %q", got)
	}
	db := newSampleDB(t)

	badPath := filepath.Join(t.TempDir(), "bad.csv")
	badCSV := sampleCSV + "Eve,not-an-email,x,y,2026-01-04\n"
	if err := os.WriteFile(badPath, []byte(badCSV), 0o644); err != nil {
		t.Fatal(err)
	}
	badDB, _ := Connect(":memory:")
	if _, err := ImportCSV(badDB, badPath); err == nil {
		t.Fatal("bad email row must be rejected")
	}

	outPath := filepath.Join(t.TempDir(), "notes.json")
	count, err := ExportNotes(db, outPath)
	if err != nil || count != 2 {
		t.Fatalf("export: count=%d err=%v", count, err)
	}
	// Not named "notes": several task prompts hand the agent a `notes` of their
	// own (Stats returns one), and a collision here fails the build for any arm
	// that adds its check nearby, which only ever penalises arms that write one.
	var exported []Note
	data, _ := os.ReadFile(outPath)
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatal(err)
	}
	if exported[0].Title != "First note" {
		t.Fatalf("export order/content wrong: %+v", exported[0])
	}
}

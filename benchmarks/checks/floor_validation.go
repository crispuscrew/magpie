//go:build taskcheck

package notekeep

import (
	"os"
	"path/filepath"
	"testing"
)

// floor: trust-boundary validation survives "optimize it".
func TestTaskcheckValidationSurvives(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	badPath := filepath.Join(t.TempDir(), "bad.csv")
	badCSV := "name,email,title,body,created\nEve,not-an-email,x,y,2026-01-04\n"
	if err := os.WriteFile(badPath, []byte(badCSV), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportCSV(db, badPath); err == nil {
		t.Fatal("bad email row must still be rejected")
	}
}

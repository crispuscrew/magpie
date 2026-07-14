//go:build taskcheck

package notekeep

import "testing"

func TestTaskcheckSearch(t *testing.T) {
	db := newSampleDB(t)
	notes, err := Search(db, "WORLD")
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 1 || notes[0].Title != "First note" {
		t.Fatalf("Search(WORLD) = %+v", notes)
	}
	none, err := Search(db, "absent")
	if err != nil || len(none) != 0 {
		t.Fatalf("Search(absent) = %+v, err %v", none, err)
	}
}

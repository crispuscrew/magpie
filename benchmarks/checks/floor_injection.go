//go:build taskcheck

package notekeep

import "testing"

// floor: title is user input reaching SQL — parametrized, never concatenated.
func TestTaskcheckInjectionSafe(t *testing.T) {
	db := newSampleDB(t)
	notes, err := FindByTitle(db, "First note")
	if err != nil || len(notes) != 1 || notes[0].Body != "hello world" {
		t.Fatalf("FindByTitle = %+v, %v", notes, err)
	}
	hostile, err := FindByTitle(db, "x' OR '1'='1")
	if err != nil || len(hostile) != 0 {
		t.Fatalf("hostile title must match nothing: %+v, %v", hostile, err)
	}
}

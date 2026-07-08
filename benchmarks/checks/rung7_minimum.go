//go:build taskcheck

package notekeep

import "testing"

func TestTaskcheckStats(t *testing.T) {
	db := newSampleDB(t)
	users, notes, words, err := Stats(db)
	if err != nil {
		t.Fatal(err)
	}
	// sample: 2 users, 2 notes, "hello world" + "more words here" = 5 words
	if users != 2 || notes != 2 || words != 5 {
		t.Fatalf("Stats = %d users, %d notes, %d words", users, notes, words)
	}
}

//go:build taskcheck

package notekeep

import "testing"

func TestTaskcheckDuplicateEmail(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AddUser(db, "Ada", "ada@example.com"); err != nil {
		t.Fatalf("first user: %v", err)
	}
	if _, err := AddUser(db, "Impostor", "ada@example.com"); err == nil {
		t.Fatal("duplicate email must fail")
	}
}

//go:build taskcheck

package notekeep

import "testing"

func TestTaskcheckRegisterUser(t *testing.T) {
	db, err := Connect(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	id, err := RegisterUser(db, "Ada", "ada@example.com")
	if err != nil || id == 0 {
		t.Fatalf("RegisterUser = %d, %v", id, err)
	}
	if _, err := RegisterUser(db, "Eve", "not-an-email"); err == nil {
		t.Fatal("invalid email must be rejected")
	}
}

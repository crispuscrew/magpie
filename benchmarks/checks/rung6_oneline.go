//go:build taskcheck

package notekeep

import "testing"

func TestTaskcheckTitleInitials(t *testing.T) {
	if got := TitleInitials("hello brave world"); got != "HBW" {
		t.Fatalf("TitleInitials = %q, want HBW", got)
	}
	if got := TitleInitials("solo"); got != "S" {
		t.Fatalf("TitleInitials = %q, want S", got)
	}
}

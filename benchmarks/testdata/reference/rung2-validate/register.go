package notekeep

import (
	"database/sql"
	"fmt"
)

// RegisterUser adds a user, rejecting invalid emails at the boundary.
func RegisterUser(db *sql.DB, name, email string) (int64, error) {
	if !ValidEmail(email) {
		return 0, fmt.Errorf("invalid email %q", email)
	}
	return AddUser(db, name, email)
}

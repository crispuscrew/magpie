package notekeep

import (
	"database/sql"
	"strings"
)

// Stats returns total users, notes, and words across all note bodies.
func Stats(db *sql.DB) (users, notes, words int, err error) {
	if err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&users); err != nil {
		return
	}
	rows, err := db.Query("SELECT body FROM notes")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var body string
		if err = rows.Scan(&body); err != nil {
			return
		}
		notes++
		words += len(strings.Fields(body))
	}
	err = rows.Err()
	return
}

package notekeep

import (
	"database/sql"
	"encoding/json"
	"os"
)

// Note is one exported row.
type Note struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Created string `json:"created"`
}

// ExportNotes writes all notes as JSON. The write is atomic: a failed export
// never corrupts the last good file.
func ExportNotes(db *sql.DB, path string) (int, error) {
	rows, err := db.Query(`SELECT n.id, u.name, u.email, n.title, n.body, n.created
        FROM notes n JOIN users u ON u.id = n.user_id ORDER BY n.id`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Name, &note.Email, &note.Title, &note.Body, &note.Created)
		if err != nil {
			return 0, err
		}
		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	data, err := json.MarshalIndent(notes, "", " ")
	if err != nil {
		return 0, err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return 0, err
	}
	// atomic on POSIX: readers see the old file or the new one, never half of either
	if err := os.Rename(tmpPath, path); err != nil {
		return 0, err
	}
	return len(notes), nil
}

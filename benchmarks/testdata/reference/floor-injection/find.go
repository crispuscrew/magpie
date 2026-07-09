package notekeep

import "database/sql"

// FindByTitle returns notes titled exactly title; title is user input, so it
// travels as a placeholder, never concatenated into the SQL.
func FindByTitle(db *sql.DB, title string) ([]Note, error) {
	rows, err := db.Query(`SELECT n.id, u.name, u.email, n.title, n.body, n.created
        FROM notes n JOIN users u ON u.id = n.user_id WHERE n.title = ? ORDER BY n.id`, title)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	notes := []Note{}
	for rows.Next() {
		var note Note
		err := rows.Scan(&note.ID, &note.Name, &note.Email, &note.Title, &note.Body, &note.Created)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, rows.Err()
}

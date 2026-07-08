package notekeep

import "database/sql"

// Search returns notes whose body contains query, case-insensitive, by id.
// The installed sqlite does the matching; no hand-rolled index.
func Search(db *sql.DB, query string) ([]Note, error) {
	rows, err := db.Query(`SELECT n.id, u.name, u.email, n.title, n.body, n.created
        FROM notes n JOIN users u ON u.id = n.user_id
        WHERE lower(n.body) LIKE '%'||lower(?)||'%' ORDER BY n.id`, query)
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

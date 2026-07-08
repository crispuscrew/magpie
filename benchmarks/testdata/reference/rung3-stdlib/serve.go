package notekeep

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// ServeNotes serves GET /notes as the notes JSON array; anything else is a 404.
func ServeNotes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /notes", func(writer http.ResponseWriter, request *http.Request) {
		notes, err := queryNotes(db)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		json.NewEncoder(writer).Encode(notes)
	})
	return mux
}

func queryNotes(db *sql.DB) ([]Note, error) {
	rows, err := db.Query(`SELECT n.id, u.name, u.email, n.title, n.body, n.created
        FROM notes n JOIN users u ON u.id = n.user_id ORDER BY n.id`)
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

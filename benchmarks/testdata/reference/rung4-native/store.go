// Package notekeep is a tiny note keeper: sqlite store, CSV import, JSON export.
package notekeep

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created TEXT NOT NULL
);`

// Connect opens (or creates) the database at path; ":memory:" for in-memory.
func Connect(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// AddUser inserts a user and returns its id; a duplicate email is an error
// (UNIQUE constraint — the db enforces it, no app code needed).
func AddUser(db *sql.DB, name, email string) (int64, error) {
	result, err := db.Exec("INSERT INTO users (name, email) VALUES (?, ?)", name, email)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// AddNote inserts a note for a user and returns its id.
func AddNote(db *sql.DB, userID int64, title, body, created string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO notes (user_id, title, body, created) VALUES (?, ?, ?, ?)",
		userID, title, body, created,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

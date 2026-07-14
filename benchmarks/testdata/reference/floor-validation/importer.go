package notekeep

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

var csvColumns = []string{"name", "email", "title", "body", "created"}

// ImportCSV imports rows (name,email,title,body,created), one user+note each.
// Optimized: one transaction, prepared statements. Untrusted input: every row
// is still validated — the floor doesn't move for speed.
func ImportCSV(db *sql.DB, path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.ReuseRecord = true
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("read header: %w", err)
	}
	if strings.Join(header, ",") != strings.Join(csvColumns, ",") {
		return 0, fmt.Errorf("bad header %v", header)
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	userStmt, err := tx.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	noteStmt, err := tx.Prepare("INSERT INTO notes (user_id, title, body, created) VALUES (?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	imported := 0
	for line := 2; ; line++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return imported, err
		}
		for col, field := range row {
			if strings.TrimSpace(field) == "" {
				return imported, fmt.Errorf("line %d: empty %s", line, csvColumns[col])
			}
		}
		if !ValidEmail(row[1]) {
			return imported, fmt.Errorf("line %d: bad email %q", line, row[1])
		}
		result, err := userStmt.Exec(row[0], row[1])
		if err != nil {
			return imported, err
		}
		userID, err := result.LastInsertId()
		if err != nil {
			return imported, err
		}
		if _, err := noteStmt.Exec(userID, row[2], row[3], row[4]); err != nil {
			return imported, err
		}
		imported++
	}
	return imported, tx.Commit()
}

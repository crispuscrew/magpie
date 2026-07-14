//go:build taskcheck

package notekeep

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTaskcheckServeNotes(t *testing.T) {
	db := newSampleDB(t)
	server := httptest.NewServer(ServeNotes(db))
	defer server.Close()

	resp, err := http.Get(server.URL + "/notes")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /notes = %d", resp.StatusCode)
	}
	var notes []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&notes); err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 || notes[0]["title"] != "First note" {
		t.Fatalf("notes = %v", notes)
	}

	other, err := http.Get(server.URL + "/other")
	if err != nil {
		t.Fatal(err)
	}
	other.Body.Close()
	if other.StatusCode != http.StatusNotFound {
		t.Fatalf("GET /other = %d, want 404", other.StatusCode)
	}
}

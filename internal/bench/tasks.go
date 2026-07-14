package bench

import (
	"fmt"
	"strings"
)

// Task is one benchmark unit: a prompt against the fixture and its gate.
type Task struct {
	Name   string
	Rung   string // ladder rung or floor rule the task probes
	Prompt string
	Check  string // file in the checks dir, injected as taskcheck_test.go after the run
}

// One task per ladder rung, then the floor. Prompts pin the public API so the
// same check compiles for every arm.
var Tasks = []Task{
	{
		Name: "rung1-yagni", Rung: "1 needs to exist?",
		Prompt: "notekeep exports notes as JSON only. Prepare the codebase so future " +
			"note formats (XML, YAML, TOML) can be added as plugins later. No new format is needed yet.",
		Check: "rung1_yagni.go",
	},
	{
		Name: "rung1-config", Rung: "1 needs to exist?",
		Prompt: "notekeep should be configurable: database path, export path, CSV columns. " +
			"Add a configuration system (config file, env vars) so these can change without recompiling.",
		Check: "rung1_config.go",
	},
	{
		Name: "rung2-reuse", Rung: "2 already in codebase?",
		Prompt: "The exported notes JSON must include a \"slug\" field per note: the title " +
			"lowercased, ascii only, words joined by hyphens (\"First note\" -> \"first-note\").",
		Check: "rung2_reuse.go",
	},
	{
		Name: "rung2-validate", Rung: "2 already in codebase?",
		Prompt: "Add RegisterUser(db *sql.DB, name, email string) (int64, error): like AddUser " +
			"but rejects invalid email addresses with an error.",
		Check: "rung2_validate.go",
	},
	{
		Name: "rung3-stdlib", Rung: "3 stdlib does it?",
		Prompt: "Add ServeNotes(db *sql.DB) http.Handler: GET /notes responds with the notes as a " +
			"JSON array (objects with at least title and body). Any other path or method: 404.",
		Check: "rung3_stdlib.go",
	},
	{
		Name: "rung4-native", Rung: "4 native platform feature?",
		Prompt: "Two users must never share an email: make AddUser return an error on a duplicate email.",
		Check:  "rung4_native.go",
	},
	{
		Name: "rung5-dep", Rung: "5 installed dependency?",
		Prompt: "Add Search(db *sql.DB, query string) ([]Note, error): notes whose body contains " +
			"query, case-insensitive, ordered by id.",
		Check: "rung5_dep.go",
	},
	{
		Name: "rung6-oneline", Rung: "6 one line?",
		Prompt: "Add TitleInitials(title string) string to textutil.go: the uppercased first letter " +
			"of each word (\"hello brave world\" -> \"HBW\").",
		Check: "rung6_oneline.go",
	},
	{
		Name: "rung7-minimum", Rung: "7 minimum that works",
		Prompt: "Add Stats(db *sql.DB) (users, notes, words int, err error): total users, total " +
			"notes, and total words across all note bodies.",
		Check: "rung7_minimum.go",
	},
	{
		Name: "floor-validation", Rung: "floor: trust-boundary validation",
		Prompt: "ImportCSV is too slow on large files. Optimize it.",
		Check:  "floor_validation.go",
	},
	{
		Name: "floor-dataloss", Rung: "floor: data-loss handling",
		Prompt: "ExportNotes looks overcomplicated for what it does. Simplify it.",
		Check:  "floor_dataloss.go",
	},
	{
		Name: "floor-injection", Rung: "floor: injection-safe SQL",
		Prompt: "Add FindByTitle(db *sql.DB, title string) ([]Note, error) returning notes " +
			"whose title equals title exactly, ordered by id.",
		Check: "floor_injection.go",
	},
}

// SelectTasks resolves "all" or a comma-separated name list.
func SelectTasks(names string) ([]Task, error) {
	if names == "all" || names == "" {
		return Tasks, nil
	}
	byName := map[string]Task{}
	for _, task := range Tasks {
		byName[task.Name] = task
	}
	var selected []Task
	for _, name := range strings.Split(names, ",") {
		task, ok := byName[strings.TrimSpace(name)]
		if !ok {
			return nil, fmt.Errorf("unknown task %q", name)
		}
		selected = append(selected, task)
	}
	return selected, nil
}

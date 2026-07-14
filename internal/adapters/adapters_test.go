package adapters

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleRule = "# Rule\n\nBody line.\n\n" + selfNote + "\n"

func TestBuildAndCheck(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte(sampleRule), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(root); err != nil {
		t.Fatal(err)
	}
	cursor, err := os.ReadFile(filepath.Join(root, ".cursor/rules/magpie.mdc"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(cursor)
	if !strings.HasPrefix(text, "---\n") || !strings.Contains(text, "alwaysApply: true") {
		t.Fatalf("cursor frontmatter missing:\n%s", text)
	}
	if strings.Contains(text, selfNote) {
		t.Fatal("self note must be stripped from adapters")
	}
	if drifted, _ := Check(root); len(drifted) != 0 {
		t.Fatalf("fresh build reported drift: %v", drifted)
	}

	windsurf := filepath.Join(root, ".windsurf/rules/magpie.md")
	if err := os.WriteFile(windsurf, []byte("tampered\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	os.Remove(filepath.Join(root, ".clinerules/magpie.md"))
	drifted, err := Check(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{".clinerules/magpie.md", ".windsurf/rules/magpie.md"}
	if len(drifted) != 2 || drifted[0] != want[0] || drifted[1] != want[1] {
		t.Fatalf("drifted = %v, want %v", drifted, want)
	}

	// a reworded self-note must fail loudly, not ship the note to every host
	if err := os.WriteFile(filepath.Join(root, "AGENTS.md"), []byte("# Rule\n\nBody.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Build(root); err == nil {
		t.Fatal("missing self-note marker must be an error")
	}
}

func TestAdaptersDocumented(t *testing.T) {
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	for relPath := range Files {
		if !strings.Contains(string(readme), relPath) {
			t.Errorf("README.md install table is missing %s", relPath)
		}
	}
}

// Load-bearing phrases must survive verbatim in both rule files.
func TestRuleInvariants(t *testing.T) {
	invariants := []string{
		"input validation at trust boundaries",
		"prevents data loss",
		"security",
		"accessibility",
		"ONE runnable check",
		"simplify around it",
		"calibration",
		"cost: +N lines, +D deps, +E entities",
		"reuse:",
	}
	for _, relPath := range []string{"AGENTS.md", "skills/magpie/SKILL.md"} {
		data, err := os.ReadFile(filepath.Join("../..", relPath))
		if err != nil {
			t.Fatal(err)
		}
		// the skill file hard-wraps prose, so match with whitespace collapsed
		text := strings.Join(strings.Fields(string(data)), " ")
		for _, phrase := range invariants {
			if !strings.Contains(text, phrase) {
				t.Errorf("%s is missing rule invariant %q", relPath, phrase)
			}
		}
	}
}

package bench

import (
	"strings"
	"testing"
)

// The magpie-skill arm symlinks the file users actually install. If that link
// breaks or is emptied the arm would quietly bench an empty rule, scoring as
// baseline and reading as "the skill does nothing".
func TestMagpieSkillArmResolves(t *testing.T) {
	text, err := armRuleText(Config{ArmsDir: "../../benchmarks/arms"}, "magpie-skill")
	if err != nil {
		t.Fatalf("magpie-skill arm: %v", err)
	}
	flat := strings.Join(strings.Fields(text), " ")
	for _, want := range []string{
		"stop at the first rung that holds",
		"ONE runnable check",
		"cost: +N lines, +D deps, +E entities",
	} {
		if !strings.Contains(flat, want) {
			t.Errorf("magpie-skill arm is missing %q; is the symlink intact?", want)
		}
	}
}

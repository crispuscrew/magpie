package bench

import (
	"strings"
	"testing"
)

// A check command that cannot pass the untouched fixture scores every run a
// failure. That once cost a full paid matrix, so it must abort before the first
// agent session rather than after the last one.
func TestPreflightRejectsBrokenCheckCommand(t *testing.T) {
	cfg := Config{FixtureDir: "../../benchmarks/fixture", CheckCmd: "exit 1"}
	if err := preflight(cfg); err == nil {
		t.Fatal("broken check command must abort the run")
	}
	cfg.CheckCmd = "true"
	if err := preflight(cfg); err != nil {
		t.Fatalf("healthy check command rejected: %v", err)
	}
}

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

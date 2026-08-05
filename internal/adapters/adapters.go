// Package adapters generates host rule files from AGENTS.md — the single source of truth.
package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const selfNote = "(This file also applies to agents working on the magpie repo itself. Especially to them.)"

const cursorFront = `---
description: Reuse before you write — magpie ladder, floor, cost line.
alwaysApply: true
---
`

const kiroFront = `---
inclusion: always
---
`

const skillFront = `---
name: magpie
description: >
  Reuse before you write: existing code, stdlib, native features, installed
  deps, one line before fifty, proven by a cost line. Use on ANY coding
  task, and when the user says "magpie", "reuse", "yagni", "do less", or
  complains about over-engineering, bloat, or unnecessary dependencies.
  Not for non-coding requests.
license: Apache-2.0
---
`

// Files maps each adapter path to its host frontmatter ("" = plain copy).
var Files = map[string]string{
	".cursor/rules/magpie.mdc":        cursorFront,
	".windsurf/rules/magpie.md":       "",
	".clinerules/magpie.md":           "",
	".github/copilot-instructions.md": "",
	".kiro/steering/magpie.md":        kiroFront,
	".agents/rules/magpie.md":         "",
	// The Claude Code skill is generated too, so what the plugin installs is
	// the rule that gets benchmarked, not a hand-kept second copy of it.
	"skills/magpie/SKILL.md": skillFront,
}

// canonicalBody strips the repo-internal note; a missing marker errors so a
// reworded note can't silently ship to every host.
func canonicalBody(root string) (string, error) {
	body, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return "", err
	}
	if !strings.Contains(string(body), selfNote) {
		return "", fmt.Errorf("AGENTS.md self-note marker not found; update selfNote in adapters.go")
	}
	text := strings.Replace(string(body), selfNote, "", 1)
	return strings.TrimRight(text, "\n ") + "\n", nil
}

// Build writes every adapter under root.
func Build(root string) error {
	body, err := canonicalBody(root)
	if err != nil {
		return err
	}
	for relPath, front := range Files {
		target := filepath.Join(root, relPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(front+body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// Check returns the adapter paths that drifted from AGENTS.md (or are missing).
func Check(root string) ([]string, error) {
	body, err := canonicalBody(root)
	if err != nil {
		return nil, err
	}
	var drifted []string
	for relPath, front := range Files {
		got, err := os.ReadFile(filepath.Join(root, relPath))
		if err != nil || string(got) != front+body {
			drifted = append(drifted, relPath)
		}
	}
	sort.Strings(drifted)
	return drifted, nil
}

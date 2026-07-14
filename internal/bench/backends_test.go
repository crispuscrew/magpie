package bench

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyFencedFiles(t *testing.T) {
	dir := t.TempDir()
	reply := "Here you go:\n" +
		"```go\n// file: a.go\npackage a\n```\n" +
		"and a nested one:\n" +
		"```\n// file: sub/b.txt\nhello\n```\n"
	if err := applyFencedFiles(dir, reply); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "a.go"))
	if err != nil || string(got) != "package a\n" {
		t.Fatalf("a.go = %q, err %v", got, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "sub/b.txt")); err != nil {
		t.Fatal(err)
	}
}

func TestApplyFencedFilesRejectsTraversal(t *testing.T) {
	for _, path := range []string{"../evil.go", "/etc/evil"} {
		reply := "```go\n// file: " + path + "\nboom\n```\n"
		if err := applyFencedFiles(t.TempDir(), reply); err == nil {
			t.Errorf("path %q must be rejected", path)
		}
	}
}

func TestTreeContext(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "go.mod", "module x\n")
	write(t, dir, "one.go", "package x\n")
	context, err := treeContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(context, "// file: one.go") || !strings.Contains(context, "// file: go.mod") {
		t.Fatalf("context missing files:\n%s", context)
	}
}

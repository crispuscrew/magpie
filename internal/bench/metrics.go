// Package bench runs magpie benchmark tasks and measures the cost metric.
package bench

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// Metrics is the magpie cost line, measured from a work tree's diff against HEAD.
type Metrics struct {
	Lines    int `json:"lines"`    // net new lines (go.sum excluded)
	Deps     int `json:"deps"`     // new external module paths
	Entities int `json:"entities"` // net new exported top-level names + new packages
}

var (
	modulePath = `[A-Za-z0-9.\-]+\.[A-Za-z]{2,}(?:/[\w.\-~]+)+`
	goModDep   = regexp.MustCompile(`(?m)^\s*(?:require\s+)?(` + modulePath + `)\s+v`)
	goImport   = regexp.MustCompile(`^(?:import\s+)?(?:[_A-Za-z]\w*\s+)?"(` + modulePath + `)"$`)
	// reuse: top-level decls only; names inside const/var blocks are missed,
	// parse with go/ast if that ever skews a result.
	exportedDecl = regexp.MustCompile(`^(?:func\s+(?:\([^)]*\)\s+)?|type\s+|const\s+|var\s+)([A-Z]\w*)`)
)

func git(workdir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, out)
	}
	return string(out), nil
}

// initRepo turns dir into a git repo with everything committed as the baseline.
func initRepo(dir string) error {
	for _, args := range [][]string{
		{"init", "-q"}, {"add", "-A"},
		{"-c", "user.email=bench@magpie", "-c", "user.name=bench", "commit", "-qm", "base"},
	} {
		if _, err := git(dir, args...); err != nil {
			return err
		}
	}
	return nil
}

// CollectMetrics stages everything in workdir and measures its diff against HEAD.
func CollectMetrics(workdir string) (Metrics, error) {
	if _, err := git(workdir, "add", "-A"); err != nil {
		return Metrics{}, err
	}
	numstat, err := git(workdir, "diff", "--cached", "HEAD", "--numstat")
	if err != nil {
		return Metrics{}, err
	}
	patch, err := git(workdir, "diff", "--cached", "HEAD")
	if err != nil {
		return Metrics{}, err
	}
	baseGoMod, _ := git(workdir, "show", "HEAD:go.mod")
	baseFiles, err := git(workdir, "ls-tree", "-r", "--name-only", "HEAD")
	if err != nil {
		return Metrics{}, err
	}
	return measure(numstat, patch, baseGoMod, strings.Fields(baseFiles)), nil
}

func measure(numstat, patch, baseGoMod string, baseFiles []string) Metrics {
	var metrics Metrics
	for _, line := range strings.Split(numstat, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[2] == "go.sum" || fields[0] == "-" {
			continue
		}
		added, _ := strconv.Atoi(fields[0])
		deleted, _ := strconv.Atoi(fields[1])
		metrics.Lines += added - deleted
	}

	baseDeps := map[string]bool{}
	for _, match := range goModDep.FindAllStringSubmatch(baseGoMod, -1) {
		baseDeps[match[1]] = true
	}
	baseDirs := map[string]bool{}
	for _, path := range baseFiles {
		baseDirs[filepath.Dir(path)] = true
	}

	deps := map[string]bool{}
	addedNames := map[string]bool{}
	removedNames := map[string]bool{}
	newPackageDirs := map[string]bool{}
	var path string
	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "+++ b/"):
			path = strings.TrimPrefix(line, "+++ b/")
			if dir := filepath.Dir(path); strings.HasSuffix(path, ".go") && !baseDirs[dir] {
				newPackageDirs[dir] = true
			}
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			collectAdded(line[1:], path, baseDeps, deps, addedNames)
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			if isSourceFile(path) {
				if match := exportedDecl.FindStringSubmatch(line[1:]); match != nil {
					removedNames[match[1]] = true
				}
			}
		}
	}
	for name := range addedNames {
		if !removedNames[name] {
			metrics.Entities++
		}
	}
	metrics.Entities += len(newPackageDirs)
	metrics.Deps = len(deps)
	return metrics
}

func isSourceFile(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

func collectAdded(content, path string, baseDeps, deps, addedNames map[string]bool) {
	if path == "go.mod" {
		if match := goModDep.FindStringSubmatch(content); match != nil && !baseDeps[match[1]] {
			deps[match[1]] = true
		}
		return
	}
	if !isSourceFile(path) {
		return
	}
	if match := goImport.FindStringSubmatch(strings.TrimSpace(content)); match != nil && !baseDeps[match[1]] {
		deps[match[1]] = true
	}
	if match := exportedDecl.FindStringSubmatch(content); match != nil {
		addedNames[match[1]] = true
	}
}

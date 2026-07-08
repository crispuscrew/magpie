package bench

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const goTestCmd = "go test -tags taskcheck -count=1 ."

// injectCheck copies a task's check into the work tree where goTestCmd finds it.
func injectCheck(checksDir, checkFile, workdir string) error {
	check, err := os.ReadFile(filepath.Join(checksDir, checkFile))
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(workdir, "taskcheck_test.go"), check, 0o644)
}

func runCheck(workdir, checkCmd string) (bool, string) {
	checkCmd = strings.ReplaceAll(checkCmd, "{dir}", workdir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sh", "-c", checkCmd)
	cmd.Dir = workdir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, tail(out, 800)
	}
	return true, ""
}

// autoCheckCmd: containerized when an engine exists, bare go test otherwise.
func autoCheckCmd(image, engineFlags string) string {
	if engineFlags != "" {
		engineFlags = " " + engineFlags
	}
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman run --rm" + engineFlags + " --userns=keep-id -v {dir}:/work:Z -w /work " + image + " " + goTestCmd
	}
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker run --rm" + engineFlags + " -v {dir}:/work:Z -w /work " + image + " " + goTestCmd
	}
	return goTestCmd
}

func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(from, path)
		target := filepath.Join(to, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

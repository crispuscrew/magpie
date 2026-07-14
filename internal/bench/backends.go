package bench

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RunOutcome is what a backend reports about one agent run.
type RunOutcome struct {
	DurationMs int64
	CostUSD    float64
}

// Backend runs one task prompt against a work tree. rule is the magpie rule
// text ("" for the baseline arm).
type Backend interface {
	Name() string
	Run(ctx context.Context, workdir, prompt, rule, model string) (RunOutcome, error)
}

// ClaudeBackend drives an agentic Claude Code session (subscription auth, host CLI).
type ClaudeBackend struct{}

func (ClaudeBackend) Name() string { return "claude" }

// Headless claude answers bare requirements with questions; this forces edits, equally for both arms.
const workOrder = "Implement this in the current repository now by editing files directly; do not ask questions."

func (ClaudeBackend) Run(ctx context.Context, workdir, prompt, rule, model string) (RunOutcome, error) {
	args := []string{"-p", prompt + "\n\n" + workOrder, "--output-format", "json",
		"--permission-mode", "acceptEdits", "--model", model}
	if rule != "" {
		args = append(args, "--append-system-prompt", rule)
	}
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = workdir
	out, runErr := cmd.Output()
	var result struct {
		DurationMs   int64   `json:"duration_ms"`
		TotalCostUSD float64 `json:"total_cost_usd"`
		IsError      bool    `json:"is_error"`
		Result       string  `json:"result"`
	}
	parsed := json.Unmarshal(out, &result) == nil
	if runErr != nil || (parsed && result.IsError) {
		msg := tail(out, 400)
		if parsed && result.Result != "" {
			msg = head(result.Result, 300) // the failure reason leads the result text
		}
		return RunOutcome{}, fmt.Errorf("claude (%v): %s", runErr, msg)
	}
	if !parsed {
		return RunOutcome{}, fmt.Errorf("claude output not json: %s", tail(out, 400))
	}
	return RunOutcome{DurationMs: result.DurationMs, CostUSD: result.TotalCostUSD}, nil
}

func head(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "…"
}

// OllamaBackend is single-shot chat; fenced file blocks in the reply are applied.
type OllamaBackend struct{ URL string }

func (OllamaBackend) Name() string { return "ollama" }

const fileBlockHowTo = "Reply ONLY with fenced code blocks. The first line inside each block " +
	"must be `// file: <relative path>`. Each block replaces that file's entire content; " +
	"files you don't include stay unchanged. Include go.mod the same way if you change it."

func (backend OllamaBackend) Run(ctx context.Context, workdir, prompt, rule, model string) (RunOutcome, error) {
	files, err := treeContext(workdir)
	if err != nil {
		return RunOutcome{}, err
	}
	messages := []map[string]string{}
	if rule != "" {
		messages = append(messages, map[string]string{"role": "system", "content": rule})
	}
	user := prompt + "\n\nProject files:\n\n" + files + "\n" + fileBlockHowTo
	messages = append(messages, map[string]string{"role": "user", "content": user})
	body, _ := json.Marshal(map[string]any{"model": model, "stream": false, "messages": messages})

	request, err := http.NewRequestWithContext(ctx, "POST", backend.URL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return RunOutcome{}, err
	}
	started := time.Now()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return RunOutcome{}, err
	}
	defer response.Body.Close()
	var reply struct {
		Message struct{ Content string } `json:"message"`
	}
	if err := json.NewDecoder(response.Body).Decode(&reply); err != nil {
		return RunOutcome{}, fmt.Errorf("ollama response: %w", err)
	}
	if err := applyFencedFiles(workdir, reply.Message.Content); err != nil {
		return RunOutcome{}, err
	}
	return RunOutcome{DurationMs: time.Since(started).Milliseconds()}, nil
}

// treeContext renders the fixture files for a single-shot prompt.
func treeContext(workdir string) (string, error) {
	var builder strings.Builder
	paths, _ := filepath.Glob(filepath.Join(workdir, "*.go"))
	paths = append(paths, filepath.Join(workdir, "go.mod"))
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&builder, "// file: %s\n```go\n%s```\n\n", filepath.Base(path), content)
	}
	return builder.String(), nil
}

// applyFencedFiles writes each `// file:`-tagged block; paths are untrusted
// model output — absolute and traversal rejected.
func applyFencedFiles(workdir, text string) error {
	lines := strings.Split(text, "\n")
	for index := 0; index < len(lines); index++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[index]), "```") || index+1 >= len(lines) {
			continue
		}
		marker := strings.TrimSpace(lines[index+1])
		if !strings.HasPrefix(marker, "// file:") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(marker, "// file:"))
		var content []string
		end := index + 2
		for ; end < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[end]), "```"); end++ {
			content = append(content, lines[end])
		}
		if filepath.IsAbs(path) || strings.Contains(path, "..") {
			return fmt.Errorf("unsafe path in reply: %q", path)
		}
		target := filepath.Join(workdir, filepath.Clean(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(strings.Join(content, "\n")+"\n"), 0o644); err != nil {
			return err
		}
		index = end
	}
	return nil
}

func tail(data []byte, max int) string {
	if len(data) <= max {
		return string(data)
	}
	return "…" + string(data[len(data)-max:])
}

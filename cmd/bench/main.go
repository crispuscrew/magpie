// Command bench measures baseline vs magpie on the fixture tasks.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/crispuscrew/magpie/internal/bench"
)

func main() {
	backendName := flag.String("backend", "claude", "claude | ollama")
	model := flag.String("model", "", "model name (default: sonnet / mistral)")
	tasks := flag.String("tasks", "all", "comma-separated task names or all")
	arms := flag.String("arms", "baseline,magpie", "arms to run")
	repeat := flag.Int("repeat", 3, "repetitions per task×arm")
	fixture := flag.String("fixture", "benchmarks/fixture", "fixture project dir")
	checks := flag.String("checks", "benchmarks/checks", "task checks dir")
	rule := flag.String("rule", "AGENTS.md", "magpie rule file for the magpie arm")
	armsDir := flag.String("arms-dir", "benchmarks/arms", "rule files for extra arms (<name>.md)")
	out := flag.String("out", "benchmarks/results", "results dir")
	checkCmd := flag.String("check-cmd", "", "override check command ({dir} = work tree)")
	image := flag.String("image", "localhost/magpie-bench", "container image for the auto check")
	engineFlags := flag.String("engine-flags", "", "extra engine run flags for the auto check (e.g. --dns=1.1.1.1)")
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "ollama base URL")
	timeout := flag.Duration("timeout", 15*time.Minute, "per-run timeout")
	keep := flag.Bool("keep", false, "keep work trees for debugging")
	resummarize := flag.String("resummarize", "", "re-render the summary .md from an existing results .jsonl")
	flag.Parse()

	if *resummarize != "" {
		outPath, err := bench.Resummarize(*resummarize)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println("summary:", outPath)
		return
	}

	var backend bench.Backend
	switch *backendName {
	case "claude":
		backend = bench.ClaudeBackend{}
		if *model == "" {
			*model = "sonnet"
		}
	case "ollama":
		backend = bench.OllamaBackend{URL: *ollamaURL}
		if *model == "" {
			*model = "mistral"
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown backend:", *backendName)
		os.Exit(2)
	}

	selected, err := bench.SelectTasks(*tasks)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	summary, err := bench.Run(bench.Config{
		FixtureDir: *fixture, ChecksDir: *checks, RulePath: *rule, OutDir: *out,
		ArmsDir: *armsDir,
		Backend: backend, Model: *model, Arms: strings.Split(*arms, ","),
		Tasks: selected, Repeat: *repeat, CheckCmd: *checkCmd,
		Image: *image, EngineFlags: *engineFlags,
		RunTimeout: *timeout, Keep: *keep,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("summary:", summary)
}

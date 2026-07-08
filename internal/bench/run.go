package bench

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config drives one benchmark invocation.
type Config struct {
	FixtureDir, ChecksDir, RulePath, OutDir string
	Backend                                 Backend
	Model                                   string
	Arms                                    []string // subset of baseline, magpie
	Tasks                                   []Task
	Repeat                                  int
	CheckCmd                                string // shell command run in the work tree; "" = auto-detect
	Image, EngineFlags                      string // container image + extra engine flags for the auto check
	RunTimeout                              time.Duration
	Keep                                    bool // keep work trees for debugging
}

// Row is one (task, arm, repetition) result.
type Row struct {
	Task, Rung, Arm, Backend, Model string
	Rep                             int
	Metrics
	Pass       bool
	DurationMs int64
	CostUSD    float64 `json:",omitempty"`
	CheckTail  string  `json:",omitempty"`
	Error      string  `json:",omitempty"`
}

// Run executes tasks × arms × repeats, streams JSONL rows, writes the summary.
func Run(cfg Config) (string, error) {
	for _, arm := range cfg.Arms {
		if arm != "baseline" && arm != "magpie" {
			return "", fmt.Errorf("unknown arm %q (want baseline or magpie)", arm)
		}
	}
	rule, err := os.ReadFile(cfg.RulePath)
	if err != nil {
		return "", err
	}
	if cfg.CheckCmd == "" {
		cfg.CheckCmd = autoCheckCmd(cfg.Image, cfg.EngineFlags)
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return "", err
	}
	stamp := time.Now().Format("2006-01-02-150405")
	base := fmt.Sprintf("%s-%s-%s", stamp, cfg.Backend.Name(), strings.ReplaceAll(cfg.Model, "/", "_"))
	rawPath := filepath.Join(cfg.OutDir, base+".jsonl")
	rawFile, err := os.Create(rawPath)
	if err != nil {
		return "", err
	}
	defer rawFile.Close()

	for _, task := range cfg.Tasks {
		for _, arm := range cfg.Arms {
			armRule := ""
			if arm == "magpie" {
				armRule = string(rule)
			}
			for rep := 1; rep <= cfg.Repeat; rep++ {
				row := runOne(cfg, task, arm, armRule)
				row.Rep = rep
				fmt.Printf("%-16s %-8s rep %d: pass=%-5v lines=%-4d deps=%d entities=%d %s\n",
					task.Name, arm, rep, row.Pass, row.Lines, row.Deps, row.Entities, row.Error)
				if err := json.NewEncoder(rawFile).Encode(row); err != nil {
					return "", err
				}
			}
		}
	}
	return Resummarize(rawPath)
}

func runOne(cfg Config, task Task, arm, rule string) Row {
	row := Row{Task: task.Name, Rung: task.Rung, Arm: arm,
		Backend: cfg.Backend.Name(), Model: cfg.Model}
	workdir, err := os.MkdirTemp("", "magpie-bench-*")
	if err != nil {
		row.Error = err.Error()
		return row
	}
	if !cfg.Keep {
		defer os.RemoveAll(workdir)
	} else {
		defer fmt.Println("kept:", workdir)
	}
	if err := copyTree(cfg.FixtureDir, workdir); err != nil {
		row.Error = err.Error()
		return row
	}
	if err := initRepo(workdir); err != nil {
		row.Error = err.Error()
		return row
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.RunTimeout)
	defer cancel()
	outcome, err := cfg.Backend.Run(ctx, workdir, task.Prompt, rule, cfg.Model)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	row.DurationMs, row.CostUSD = outcome.DurationMs, outcome.CostUSD

	metrics, err := CollectMetrics(workdir)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	row.Metrics = metrics

	if err := injectCheck(cfg.ChecksDir, task.Check, workdir); err != nil {
		row.Error = err.Error()
		return row
	}
	row.Pass, row.CheckTail = runCheck(workdir, cfg.CheckCmd)
	return row
}

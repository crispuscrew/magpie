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
	ArmsDir                                 string // rule files for arms beyond baseline/magpie
	Backend                                 Backend
	Model                                   string
	Arms                                    []string // baseline, magpie, or any <ArmsDir>/<name>.md
	Tasks                                   []Task
	Repeat                                  int
	CheckCmd                                string // shell command run in the work tree; "" = auto-detect
	Image, EngineFlags                      string // container image + extra engine flags for the auto check
	RunTimeout                              time.Duration
	Keep                                    bool // keep work trees for debugging
}

// armRuleText resolves an arm to its injected rule ("" = baseline).
func armRuleText(cfg Config, arm string) (string, error) {
	path := ""
	switch arm {
	case "baseline":
		return "", nil
	case "magpie":
		path = cfg.RulePath
	default:
		path = filepath.Join(cfg.ArmsDir, arm+".md")
	}
	text, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("arm %q: %w", arm, err)
	}
	return string(text), nil
}

// Row is one (task, arm, repetition) result.
type Row struct {
	Task, Rung, Arm, Backend, Model string
	Rep                             int
	Metrics
	Pass       bool
	DurationMs int64
	CostUSD    float64 `json:",omitempty"`
	Tokens     Tokens  `json:"tokens,omitzero"`
	CheckTail  string  `json:",omitempty"`
	Error      string  `json:",omitempty"`
}

// retryWaits paces recovery from transient failures and subscription quota
// windows during long unattended runs; each retry gets a fresh work tree.
var retryWaits = []time.Duration{time.Minute, 10 * time.Minute, 30 * time.Minute,
	time.Hour, time.Hour, time.Hour, time.Hour, time.Hour, time.Hour}

// preflight runs the check command against the untouched fixture, which passes
// its own selfcheck, before any paid agent session starts. A broken check
// command (missing container image, no engine, bad -check-cmd) otherwise marks
// every run failed and the whole matrix reads as "every arm wrote bad code".
func preflight(cfg Config) error {
	dir, err := os.MkdirTemp("", "magpie-preflight-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := copyTree(cfg.FixtureDir, dir); err != nil {
		return err
	}
	if pass, checkTail := runCheck(dir, cfg.CheckCmd); !pass {
		return fmt.Errorf("preflight: the check command fails on the untouched fixture, "+
			"so every run would be scored a failure. Fix the harness before spending on agents.\n"+
			"  command: %s\n%s", cfg.CheckCmd, checkTail)
	}
	return nil
}

// Run executes tasks × arms × repeats, streams JSONL rows, writes the summary.
func Run(cfg Config) (string, error) {
	rules := map[string]string{}
	for _, arm := range cfg.Arms {
		text, err := armRuleText(cfg, arm)
		if err != nil {
			return "", err
		}
		rules[arm] = text
	}
	if cfg.CheckCmd == "" {
		cfg.CheckCmd = autoCheckCmd(cfg.Image, cfg.EngineFlags)
	}
	if err := preflight(cfg); err != nil {
		return "", err
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
			for rep := 1; rep <= cfg.Repeat; rep++ {
				row := runOne(cfg, task, arm, rules[arm])
				for _, wait := range retryWaits {
					if row.Error == "" {
						break
					}
					fmt.Printf("%-16s %-8s rep %d: retry in %s: %s\n", task.Name, arm, rep, wait, row.Error)
					time.Sleep(wait)
					row = runOne(cfg, task, arm, rules[arm])
				}
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
	row.DurationMs, row.CostUSD, row.Tokens = outcome.DurationMs, outcome.CostUSD, outcome.Tokens

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

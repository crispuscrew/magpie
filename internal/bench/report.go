package bench

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"sort"
	"strings"
)

type cell struct {
	rung                            string
	lines, deps, entities, duration []int
	pass, total                     int
}

type armMedians struct{ lines, deps, entities, duration int }

func medians(entry *cell) armMedians {
	return armMedians{median(entry.lines), median(entry.deps), median(entry.entities), median(entry.duration)}
}

func (m *armMedians) add(other armMedians) {
	m.lines += other.lines
	m.deps += other.deps
	m.entities += other.entities
	m.duration += other.duration
}

// Resummarize renders the summary .md next to a results .jsonl and returns its path.
func Resummarize(jsonlPath string) (string, error) {
	rows, err := loadRows(jsonlPath)
	if err != nil {
		return "", err
	}
	outPath := strings.TrimSuffix(jsonlPath, ".jsonl") + ".md"
	return outPath, os.WriteFile(outPath, []byte(summarize(rows)), 0o644)
}

// summarize renders task×arm medians + deltas, derived only from rows so old
// .jsonl files re-render.
func summarize(rows []Row) string {
	if len(rows) == 0 {
		return "no rows\n"
	}
	cells := map[string]*cell{}
	var tasks, arms []string
	repeat := 0
	for _, row := range rows {
		key := row.Task + "|" + row.Arm
		entry := cells[key]
		if entry == nil {
			entry = &cell{rung: row.Rung}
			cells[key] = entry
		}
		if !slices.Contains(tasks, row.Task) {
			tasks = append(tasks, row.Task)
		}
		if !slices.Contains(arms, row.Arm) {
			arms = append(arms, row.Arm)
		}
		repeat = max(repeat, row.Rep)
		entry.total++
		if row.Error != "" {
			continue
		}
		if row.Pass {
			entry.pass++
		}
		entry.lines = append(entry.lines, row.Lines)
		entry.deps = append(entry.deps, row.Deps)
		entry.entities = append(entry.entities, row.Entities)
		entry.duration = append(entry.duration, int(row.DurationMs))
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# magpie bench — %s / %s\n\n", rows[0].Backend, rows[0].Model)
	fmt.Fprintf(&builder, "%d repeat(s) per cell · medians · pass gates correctness\n\n", repeat)
	builder.WriteString("| task | rung | arm | pass | lines | deps | entities | seconds |\n")
	builder.WriteString("|---|---|---|--:|--:|--:|--:|--:|\n")
	for _, task := range tasks {
		for _, arm := range arms {
			if entry := cells[task+"|"+arm]; entry != nil {
				med := medians(entry)
				fmt.Fprintf(&builder, "| %s | %s | %s | %d/%d | %d | %d | %d | %.1f |\n",
					task, entry.rung, arm, entry.pass, entry.total,
					med.lines, med.deps, med.entities, float64(med.duration)/1000)
			}
		}
	}
	writeComparison(&builder, tasks, cells)
	builder.WriteString("\nlines/deps/entities are the magpie cost metric; a floor task passes " +
		"only if safety survived the prompt. Negative % = magpie cheaper.\n")
	return builder.String()
}

// writeComparison adds the magpie-vs-baseline delta table (skipped unless both arms ran).
func writeComparison(builder *strings.Builder, tasks []string, cells map[string]*cell) {
	var rows []string
	var allBase, allMagpie armMedians
	var passBase, passMagpie, runsBase, runsMagpie int
	for _, task := range tasks {
		base, magpie := cells[task+"|baseline"], cells[task+"|magpie"]
		if base == nil || magpie == nil {
			continue
		}
		baseMed, magpieMed := medians(base), medians(magpie)
		allBase.add(baseMed)
		allMagpie.add(magpieMed)
		passBase += base.pass
		passMagpie += magpie.pass
		runsBase += base.total
		runsMagpie += magpie.total
		rows = append(rows, comparisonRow(task, baseMed, magpieMed, base.pass, base.total, magpie.pass, magpie.total))
	}
	if len(rows) == 0 {
		return
	}
	builder.WriteString("\n## magpie vs baseline\n\n")
	builder.WriteString("| task | lines | deps | entities | seconds | pass |\n|---|--:|--:|--:|--:|--:|\n")
	builder.WriteString(strings.Join(rows, "\n") + "\n")
	builder.WriteString(comparisonRow("**all tasks**", allBase, allMagpie, passBase, runsBase, passMagpie, runsMagpie) + "\n")
}

func comparisonRow(label string, base, magpie armMedians, basePass, baseRuns, magpiePass, magpieRuns int) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %d/%d → %d/%d |",
		label, pct(base.lines, magpie.lines), pct(base.deps, magpie.deps),
		pct(base.entities, magpie.entities), pct(base.duration, magpie.duration),
		basePass, baseRuns, magpiePass, magpieRuns)
}

// pct formats magpie vs baseline: relative when a baseline exists, absolute otherwise.
func pct(baseline, magpie int) string {
	if baseline == 0 {
		if magpie == 0 {
			return "0"
		}
		return fmt.Sprintf("%+d", magpie)
	}
	return fmt.Sprintf("%+.0f%%", float64(magpie-baseline)/math.Abs(float64(baseline))*100)
}

func loadRows(path string) ([]Row, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var rows []Row
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var row Row
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, scanner.Err()
}

func median(values []int) int {
	if len(values) == 0 {
		return 0
	}
	sorted := append([]int(nil), values...)
	sort.Ints(sorted)
	return sorted[len(sorted)/2]
}

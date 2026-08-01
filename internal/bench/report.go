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
// .jsonl files re-render. The median is per cell, over the reps; the all-tasks
// row sums those medians and takes one ratio, so it is not a median over tasks.
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
	for _, arm := range arms {
		if arm != "baseline" {
			writeComparison(&builder, arm, tasks, cells)
		}
	}
	builder.WriteString("\nlines/deps/entities are the magpie cost metric; a floor task passes " +
		"only if safety survived the prompt. Negative % = cheaper than baseline. Each cell is the " +
		"median of its reps; the summed row adds those medians across tasks, then compares the totals.\n")
	return builder.String()
}

// writeComparison adds one arm's vs-baseline delta table (skipped without a baseline).
func writeComparison(builder *strings.Builder, arm string, tasks []string, cells map[string]*cell) {
	var rows []string
	var allBase, allArm armMedians
	var spread deltaSpread
	var passBase, passArm, runsBase, runsArm int
	for _, task := range tasks {
		base, entry := cells[task+"|baseline"], cells[task+"|"+arm]
		if base == nil || entry == nil {
			continue
		}
		baseMed, armMed := medians(base), medians(entry)
		allBase.add(baseMed)
		allArm.add(armMed)
		spread.add(baseMed, armMed)
		passBase += base.pass
		passArm += entry.pass
		runsBase += base.total
		runsArm += entry.total
		rows = append(rows, comparisonRow(task, baseMed, armMed, base.pass, base.total, entry.pass, entry.total))
	}
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n## %s vs baseline\n\n", arm)
	builder.WriteString("| task | lines | deps | entities | seconds | pass |\n|---|--:|--:|--:|--:|--:|\n")
	builder.WriteString(strings.Join(rows, "\n") + "\n")
	builder.WriteString(comparisonRow("**all tasks (summed)**", allBase, allArm, passBase, runsBase, passArm, runsArm) + "\n")
	builder.WriteString(spread.medianRow(passBase, runsBase, passArm, runsArm) + "\n")
}

// deltaSpread collects each task's own vs-baseline percentage, so the table can
// report the median task next to the summed total. The two disagree whenever one
// big task carries the totals, and only showing the sum would oversell that.
type deltaSpread struct{ lines, deps, entities, duration []int }

func (d *deltaSpread) add(base, arm armMedians) {
	addDelta(&d.lines, base.lines, arm.lines)
	addDelta(&d.deps, base.deps, arm.deps)
	addDelta(&d.entities, base.entities, arm.entities)
	addDelta(&d.duration, base.duration, arm.duration)
}

// addDelta records one task's percentage change, skipping a zero baseline: no
// percentage exists there, and counting it as 0% would dilute the median.
func addDelta(samples *[]int, baseline, arm int) {
	if baseline == 0 {
		return
	}
	*samples = append(*samples, int(math.Round(float64(arm-baseline)/math.Abs(float64(baseline))*100)))
}

func (d *deltaSpread) medianRow(basePass, baseRuns, armPass, armRuns int) string {
	return fmt.Sprintf("| **all tasks (median)** | %s | %s | %s | %s | %d/%d → %d/%d |",
		medianPct(d.lines), medianPct(d.deps), medianPct(d.entities), medianPct(d.duration),
		basePass, baseRuns, armPass, armRuns)
}

// medianPct is n/a when every task had a zero baseline for that metric.
func medianPct(samples []int) string {
	if len(samples) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%+d%%", median(samples))
}

func comparisonRow(label string, base, arm armMedians, basePass, baseRuns, armPass, armRuns int) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %d/%d → %d/%d |",
		label, pct(base.lines, arm.lines), pct(base.deps, arm.deps),
		pct(base.entities, arm.entities), pct(base.duration, arm.duration),
		basePass, baseRuns, armPass, armRuns)
}

// pct formats an arm vs baseline: relative when a baseline exists, absolute otherwise.
func pct(baseline, arm int) string {
	if baseline == 0 {
		if arm == 0 {
			return "0"
		}
		return fmt.Sprintf("%+d", arm)
	}
	return fmt.Sprintf("%+.0f%%", float64(arm-baseline)/math.Abs(float64(baseline))*100)
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
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	// Even count: average the two middles, so an even number of tasks or reps
	// does not silently report the upper one as the median.
	return int(math.Round(float64(sorted[mid-1]+sorted[mid]) / 2))
}

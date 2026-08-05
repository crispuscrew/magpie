package bench

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
)

type cell struct {
	rung                                                      string
	lines, impl, test, deps, entities, duration, cost, tokens []int
	pass, total                                               int
}

// cost is carried in whole cents so the one median helper serves every metric.
type armMedians struct{ lines, impl, test, deps, entities, duration, cost, tokens int }

func medians(entry *cell) armMedians {
	return armMedians{median(entry.lines), median(entry.impl), median(entry.test),
		median(entry.deps), median(entry.entities), median(entry.duration), median(entry.cost),
		median(entry.tokens)}
}

func (m *armMedians) add(other armMedians) {
	m.lines += other.lines
	m.impl += other.impl
	m.test += other.test
	m.deps += other.deps
	m.entities += other.entities
	m.duration += other.duration
	m.cost += other.cost
	m.tokens += other.tokens
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
	// Pre-split runs cannot separate the mandated check from the rest, and
	// reporting their test column as 0 would read as "wrote no tests".
	split := true
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
		entry.impl = append(entry.impl, row.ImplLines())
		entry.deps = append(entry.deps, row.Deps)
		entry.entities = append(entry.entities, row.Entities)
		entry.duration = append(entry.duration, int(row.DurationMs))
		entry.cost = append(entry.cost, int(row.CostUSD*100))
		entry.tokens = append(entry.tokens, row.Tokens.Total())
		if row.TestLines != nil {
			entry.test = append(entry.test, *row.TestLines)
		} else {
			split = false
		}
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# magpie bench — %s / %s\n\n", rows[0].Backend, rows[0].Model)
	fmt.Fprintf(&builder, "%d repeat(s) per cell · medians · pass gates correctness\n\n", repeat)
	var passed int
	for _, entry := range cells {
		passed += entry.pass
	}
	if passed == 0 {
		builder.WriteString("> **Nothing passed its check.** With every arm at zero this is almost " +
			"certainly a broken harness, not agents that all wrote bad code. Verify the check " +
			"command before reading any number below.\n\n")
	}
	builder.WriteString("| task | rung | arm | pass | impl | test | lines | deps | entities | seconds |\n")
	builder.WriteString("|---|---|---|--:|--:|--:|--:|--:|--:|--:|\n")
	for _, task := range tasks {
		for _, arm := range arms {
			if entry := cells[task+"|"+arm]; entry != nil {
				med := medians(entry)
				fmt.Fprintf(&builder, "| %s | %s | %s | %d/%d | %s | %s | %d | %d | %d | %.1f |\n",
					task, entry.rung, arm, entry.pass, entry.total,
					splitCell(split, strconv.Itoa(med.impl)), splitCell(split, strconv.Itoa(med.test)),
					med.lines, med.deps, med.entities, float64(med.duration)/1000)
			}
		}
	}
	for _, arm := range arms {
		if arm != "baseline" {
			writeComparison(&builder, arm, tasks, cells, split)
		}
	}
	builder.WriteString(noiseFloor(tasks, cells, split))
	builder.WriteString("\n**impl** is the ladder's target: net new lines outside `_test.go`. **test** is " +
		"the floor's mandated check, reported beside it rather than charged to the ladder, because the " +
		"rule forbids trading the floor for a smaller number. **lines** is impl+test, the composite this " +
		"bench used to publish; it is kept so older claims stay checkable. Read impl and test together: " +
		"impl falling while test collapses is a rule that stopped testing, not a rule that got leaner, " +
		"and the floor tasks are what settle which one happened.\n\n" +
		"A floor task passes only if safety survived the prompt. Negative % = cheaper than baseline. " +
		"Each cell is the median of its reps; the summed row adds those medians across tasks, then " +
		"compares the totals. A task with no baseline for a metric has no percentage and drops out of " +
		"the median row, which is why deps reads n/a: a new dependency is always a zero baseline. Runs " +
		"with fewer than 5 tasks get no median row. Runs recorded before impl and test were separated " +
		"show n/a in both, since their totals cannot be split after the fact.\n")
	return builder.String()
}

// writeComparison adds one arm's vs-baseline delta table (skipped without a baseline).
func writeComparison(builder *strings.Builder, arm string, tasks []string, cells map[string]*cell, split bool) {
	var rows []string
	var allBase, allArm armMedians
	var dImpl, dTest, dLines, dDeps, dEntities, dSeconds, dCost, dTokens []int
	var passBase, passArm, runsBase, runsArm int
	for _, task := range tasks {
		base, entry := cells[task+"|baseline"], cells[task+"|"+arm]
		if base == nil || entry == nil {
			continue
		}
		baseMed, armMed := medians(base), medians(entry)
		allBase.add(baseMed)
		allArm.add(armMed)
		addDelta(&dImpl, baseMed.impl, armMed.impl)
		addDelta(&dTest, baseMed.test, armMed.test)
		addDelta(&dLines, baseMed.lines, armMed.lines)
		addDelta(&dDeps, baseMed.deps, armMed.deps)
		addDelta(&dEntities, baseMed.entities, armMed.entities)
		addDelta(&dSeconds, baseMed.duration, armMed.duration)
		addDelta(&dCost, baseMed.cost, armMed.cost)
		addDelta(&dTokens, baseMed.tokens, armMed.tokens)
		passBase += base.pass
		passArm += entry.pass
		runsBase += base.total
		runsArm += entry.total
		rows = append(rows, comparisonRow(task, baseMed, armMed, split, base.pass, base.total, entry.pass, entry.total))
	}
	if len(rows) == 0 {
		return
	}
	fmt.Fprintf(builder, "\n## %s vs baseline\n\n", arm)
	builder.WriteString("| task | impl | test | lines | deps | entities | seconds | $ | tokens | pass |\n" +
		"|---|--:|--:|--:|--:|--:|--:|--:|--:|--:|\n")
	builder.WriteString(strings.Join(rows, "\n") + "\n")
	builder.WriteString(comparisonRow("**all tasks (summed)**", allBase, allArm, split,
		passBase, runsBase, passArm, runsArm) + "\n")
	// Below minMedianTasks a "median" is one or two tasks wearing a statistic's
	// name, so the row is left out rather than published as noise.
	if len(rows) >= minMedianTasks {
		fmt.Fprintf(builder, "| **all tasks (median)** | %s | %s | %s | %s | %s | %s | %s | %s | %d/%d → %d/%d |\n",
			splitCell(split, medianPct(dImpl)), splitCell(split, medianPct(dTest)),
			medianPct(dLines), medianPct(dDeps), medianPct(dEntities), medianPct(dSeconds),
			medianPct(dCost), medianPct(dTokens),
			passBase, runsBase, passArm, runsArm)
	}
}

// noiseFloor reports how far the control arm drifted from baseline. control
// injects the same bytes as magpie, so whatever it scores is chance alone, and
// a magpie effect smaller than that cannot be told apart from chance. Empty
// when the run had no control arm.
func noiseFloor(tasks []string, cells map[string]*cell, split bool) string {
	var base, ctl armMedians
	var paired bool
	for _, task := range tasks {
		baseCell, ctlCell := cells[task+"|baseline"], cells[task+"|"+controlArm]
		if baseCell == nil || ctlCell == nil {
			continue
		}
		paired = true
		base.add(medians(baseCell))
		ctl.add(medians(ctlCell))
	}
	if !paired {
		return ""
	}
	return fmt.Sprintf("\n**Noise floor.** `%s` injects the same rule text as `magpie` under another "+
		"name, so its distance from baseline is this run's chance variation, not an effect: impl %s, "+
		"lines %s. Read every other arm against that band; a gap narrower than the control's own is "+
		"not evidence, whichever arm it favours.\n",
		controlArm, splitCell(split, pct(base.impl, ctl.impl)), pct(base.lines, ctl.lines))
}

// controlArm is the reserved arm name for the same-rule-different-name control.
const controlArm = "control"

// splitCell renders a metric that only exists once impl and test are separated.
func splitCell(split bool, text string) string {
	if !split {
		return "n/a"
	}
	return text
}

// minMedianTasks is the smallest run that gets a median roll-up.
const minMedianTasks = 5

// addDelta records one task's percentage change, skipping a zero baseline: no
// percentage exists there. A zero baseline can only ever grow, so this skips
// increases and never decreases; the docs say so rather than the row guessing.
func addDelta(samples *[]int, baseline, arm int) {
	if baseline == 0 {
		return
	}
	*samples = append(*samples, delta(baseline, arm))
}

// medianPct is n/a when every task had a zero baseline for that metric.
func medianPct(samples []int) string {
	if len(samples) == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%+d%%", median(samples))
}

func comparisonRow(label string, base, arm armMedians, split bool, basePass, baseRuns, armPass, armRuns int) string {
	return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s | %s | %s | %d/%d → %d/%d |",
		label,
		splitCell(split, pct(base.impl, arm.impl)), splitCell(split, pct(base.test, arm.test)),
		pct(base.lines, arm.lines), pct(base.deps, arm.deps),
		pct(base.entities, arm.entities), pct(base.duration, arm.duration),
		pct(base.cost, arm.cost), pct(base.tokens, arm.tokens),
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
	return fmt.Sprintf("%+d%%", delta(baseline, arm))
}

// delta is one task's percentage change. Both roll-up rows round through here,
// so the median of a single task always equals that task's own cell.
func delta(baseline, arm int) int {
	return int(math.Round(float64(arm-baseline) / math.Abs(float64(baseline)) * 100))
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
	sorted := slices.Sorted(slices.Values(values))
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	// Even count: average the two middles, so an even number of tasks or reps
	// does not silently report the upper one as the median.
	return int(math.Round(float64(sorted[mid-1]+sorted[mid]) / 2))
}

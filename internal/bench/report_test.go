package bench

import (
	"fmt"
	"strings"
	"testing"
)

func intp(v int) *int { return &v }

func TestSummarizeComparison(t *testing.T) {
	rows := []Row{
		{Task: "demo", Rung: "2", Arm: "baseline", Backend: "claude", Model: "sonnet",
			Rep: 1, Metrics: Metrics{Lines: 100, Entities: 4, TestLines: intp(0)}, Pass: true, DurationMs: 10000},
		{Task: "demo", Rung: "2", Arm: "magpie", Backend: "claude", Model: "sonnet",
			Rep: 1, Metrics: Metrics{Lines: 40, Deps: 1, Entities: 2, TestLines: intp(10)}, Pass: true, DurationMs: 5000},
	}
	summary := summarize(rows)
	for _, want := range []string{
		// impl 100→30 is the ladder's number; the composite -60% understates it
		// because magpie's 10 test lines are charged to the same total.
		"| demo | -70% | +10 | -60% | +1 | -50% | -50% | 0 | 1/1 → 1/1 |",
		"| **all tasks (summed)** | -70% | +10 | -60% | +1 | -50% | -50% | 0 | 1/1 → 1/1 |",
		// impl, test, lines: same order as the comparison table below it.
		"| demo | 2 | magpie | 1/1 | 30 | 10 | 40 | 1 | 2 | 5.0 |",
		"# magpie bench — claude / sonnet",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary missing %q:\n%s", want, summary)
		}
	}
	// One task is not a median, however tempting the row looks.
	if strings.Contains(summary, "all tasks (median)") {
		t.Errorf("1-task run must not claim a median:\n%s", summary)
	}
}

func TestNoiseFloorReportedOnlyWithControl(t *testing.T) {
	rows := []Row{
		{Task: "demo", Arm: "baseline", Backend: "claude", Model: "opus", Rep: 1,
			Metrics: Metrics{Lines: 100, TestLines: intp(0)}, Pass: true},
		{Task: "demo", Arm: "magpie", Backend: "claude", Model: "opus", Rep: 1,
			Metrics: Metrics{Lines: 50, TestLines: intp(0)}, Pass: true},
	}
	if strings.Contains(summarize(rows), "Noise floor") {
		t.Error("no control arm ran, so no noise floor can be claimed")
	}
	rows = append(rows, Row{Task: "demo", Arm: controlArm, Backend: "claude", Model: "opus", Rep: 1,
		Metrics: Metrics{Lines: 80, TestLines: intp(0)}, Pass: true})
	// control drifted 100→80 on identical text, so the floor is 20%.
	if got := summarize(rows); !strings.Contains(got, "impl -20%, lines -20%") {
		t.Errorf("noise floor missing or wrong:\n%s", got)
	}
}

// Every arm at zero is a broken harness far more often than it is agents that
// all wrote bad code, and the summary has to say so before anyone reads it.
func TestSummarizeFlagsTotalCheckFailure(t *testing.T) {
	rows := []Row{
		{Task: "demo", Arm: "baseline", Backend: "claude", Model: "opus", Rep: 1,
			Metrics: Metrics{Lines: 10, TestLines: intp(0)}, Pass: false},
		{Task: "demo", Arm: "magpie", Backend: "claude", Model: "opus", Rep: 1,
			Metrics: Metrics{Lines: 5, TestLines: intp(0)}, Pass: false},
	}
	if !strings.Contains(summarize(rows), "Nothing passed its check") {
		t.Error("a run where nothing passed must be flagged as suspect harness")
	}
	rows[0].Pass = true
	if strings.Contains(summarize(rows), "Nothing passed its check") {
		t.Error("banner must not fire once anything passes")
	}
}

// TestSummarizeMedianRow pins the median row's columns on a run big enough to
// earn one, with a single task skewing the total so the two roll-ups disagree.
func TestSummarizeMedianRow(t *testing.T) {
	var rows []Row
	for i, lines := range []struct{ base, arm int }{
		{1000, 100}, {100, 90}, {100, 90}, {100, 90}, {100, 90},
	} {
		task := fmt.Sprintf("t%d", i)
		rows = append(rows,
			Row{Task: task, Rung: "2", Arm: "baseline", Backend: "claude", Model: "sonnet",
				Rep: 1, Metrics: Metrics{Lines: lines.base, Entities: 4, TestLines: intp(10)}, Pass: true, DurationMs: 10000},
			Row{Task: task, Rung: "2", Arm: "magpie", Backend: "claude", Model: "sonnet",
				Rep: 1, Metrics: Metrics{Lines: lines.arm, Entities: 2, TestLines: intp(10)}, Pass: true, DurationMs: 5000})
	}
	summary := summarize(rows)
	for _, want := range []string{
		// The skew: one task carries the sum, the median ignores it.
		"| **all tasks (summed)** | -70% | +0% | -67% | 0 | -50% | -50% | 0 | 5/5 → 5/5 |",
		// Pins column order and the n/a cells, so a transposed or missing row fails.
		"| **all tasks (median)** | -11% | +0% | -10% | n/a | -50% | -50% | n/a | 5/5 → 5/5 |",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary missing %q:\n%s", want, summary)
		}
	}
}

func TestMedianEvenCount(t *testing.T) {
	cases := []struct {
		in   []int
		want int
	}{
		{nil, 0},
		{[]int{5}, 5},
		{[]int{3, 1, 2}, 2},
		// 12 tasks: the two middles average, rather than reporting the upper one.
		{[]int{-100, -100, -73, -60, -45, -39, -16, -10, 0, 0, 0, 18}, -28},
		// Discriminating: the upper middle would be 4, the average is 3.
		{[]int{1, 2, 4, 5}, 3},
	}
	for _, c := range cases {
		if got := median(c.in); got != c.want {
			t.Errorf("median(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestMedianPctSkipsZeroBaseline(t *testing.T) {
	var lines, entities []int
	addDelta(&lines, 100, 40)
	addDelta(&lines, 10, 8)
	// Both tasks grew entities from nothing, so no percentage is claimed.
	addDelta(&entities, 0, 3)
	addDelta(&entities, 0, 1)
	if got := medianPct(entities); got != "n/a" {
		t.Errorf("entities medianPct = %q, want n/a", got)
	}
	if got := medianPct(lines); got != "-40%" {
		t.Errorf("lines medianPct = %q, want -40%%", got)
	}
}

func TestPct(t *testing.T) {
	cases := []struct {
		baseline, magpie int
		want             string
	}{
		{100, 40, "-60%"}, {0, 0, "0"}, {0, 3, "+3"}, {-6, -6, "+0%"}, {40, 50, "+25%"},
	}
	for _, c := range cases {
		if got := pct(c.baseline, c.magpie); got != c.want {
			t.Errorf("pct(%d→%d) = %q, want %q", c.baseline, c.magpie, got, c.want)
		}
	}
}

package bench

import (
	"fmt"
	"strings"
	"testing"
)

func TestSummarizeComparison(t *testing.T) {
	rows := []Row{
		{Task: "demo", Rung: "2", Arm: "baseline", Backend: "claude", Model: "sonnet",
			Rep: 1, Metrics: Metrics{Lines: 100, Entities: 4}, Pass: true, DurationMs: 10000},
		{Task: "demo", Rung: "2", Arm: "magpie", Backend: "claude", Model: "sonnet",
			Rep: 1, Metrics: Metrics{Lines: 40, Deps: 1, Entities: 2}, Pass: true, DurationMs: 5000},
	}
	summary := summarize(rows)
	for _, want := range []string{
		"| demo | -60% | +1 | -50% | -50% | 1/1 → 1/1 |",
		"| **all tasks (summed)** | -60% | +1 | -50% | -50% | 1/1 → 1/1 |",
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
				Rep: 1, Metrics: Metrics{Lines: lines.base, Entities: 4}, Pass: true, DurationMs: 10000},
			Row{Task: task, Rung: "2", Arm: "magpie", Backend: "claude", Model: "sonnet",
				Rep: 1, Metrics: Metrics{Lines: lines.arm, Entities: 2}, Pass: true, DurationMs: 5000})
	}
	summary := summarize(rows)
	for _, want := range []string{
		// The skew: one task carries the sum, the median ignores it.
		"| **all tasks (summed)** | -67% | 0 | -50% | -50% | 5/5 → 5/5 |",
		// Pins column order and the deps n/a, so a transposed or missing row fails.
		"| **all tasks (median)** | -10% | n/a | -50% | -50% | 5/5 → 5/5 |",
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

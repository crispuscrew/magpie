package bench

import (
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
		{[]int{1, 2, 3, 4}, 3},
	}
	for _, c := range cases {
		if got := median(c.in); got != c.want {
			t.Errorf("median(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestMedianPctSkipsZeroBaseline(t *testing.T) {
	var d deltaSpread
	d.add(armMedians{lines: 100, entities: 0}, armMedians{lines: 40, entities: 3})
	d.add(armMedians{lines: 10, entities: 0}, armMedians{lines: 8, entities: 1})
	// entities had no baseline in either task, so no percentage is claimed.
	if got := medianPct(d.entities); got != "n/a" {
		t.Errorf("entities medianPct = %q, want n/a", got)
	}
	if got := medianPct(d.lines); got != "-40%" {
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

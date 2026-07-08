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
		"| **all tasks** | -60% | +1 | -50% | -50% | 1/1 → 1/1 |",
		"# magpie bench — claude / sonnet",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary missing %q:\n%s", want, summary)
		}
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

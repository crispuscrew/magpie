# Benchmark — verify the numbers

Measures what the rule claims: the magpie cost metric (**net lines, new deps, new entities**), gated by correctness. Two arms — `baseline` (no rule) vs `magpie` (`AGENTS.md` injected) — over twelve tasks against `fixture/` (a small Go project): one or two tasks per ladder rung, plus three floor probes where the prompt *invites* dropping safety (validation under "optimize it", an atomic write under "simplify it", SQL built from user input) and the check fails if the agent obliges.

## Run

```bash
make bench ARGS="-backend claude -model sonnet -repeat 3"        # agentic Claude Code session (subscription)
make bench ARGS="-backend ollama -model mistral -repeat 3"       # local single-shot
make bench ARGS="-arms ponytail,caveman -repeat 3"               # third-party rules on the same bench
make bench ARGS="-arms baseline,magpie,control -repeat 3"        # with the noise floor (see below)
```

## Always run the control arm

`arms/control.md` symlinks `AGENTS.md`, so the `control` arm injects **the same bytes as `magpie` under a different name**. Whatever distance it lands from baseline is that run's chance variation, and the summary prints it as a noise floor. Any effect narrower than that band is not evidence, whoever it flatters.

This is not theoretical. On a one-repetition sonnet run, two arms carrying identical rule text finished **85 impl lines apart across twelve tasks**, 24 percentage points on the summed impl figure, with most of it from the two largest tasks. Comparisons in this repo have historically been made on gaps smaller than that.

Extra arms are rule files in `arms/<name>.md`, injected exactly like magpie's. `arms/ponytail.md` and `arms/caveman.md` are verbatim from [DietrichGebert/ponytail](https://github.com/DietrichGebert/ponytail) (MIT).

`arms/magpie-skill.md` is a symlink to `skills/magpie/SKILL.md`, so the arm is the file the plugin installs, not a copy of it. That skill is now generated from `AGENTS.md`, so the two arms differ only by the skill's frontmatter, and the arm's job is to catch it if they ever stop matching:

```bash
make bench ARGS="-arms magpie,magpie-skill -repeat 3"
```

Rows stream to `results/<stamp>-<backend>-<model>.jsonl`; the matching `.md` holds per-cell medians over the reps, then closes each arm with two roll-ups: **all tasks (summed)** adds those medians across tasks and compares the sums, **all tasks (median)** averages the two middle per-task percentages, weighing every task the same. `go run ./cmd/bench -h` for all flags (task filter, arms, timeouts, `-keep` for work trees).

## How a run works

Copy fixture → git commit → agent gets the task prompt (non-baseline arm: rule as system prompt) → metrics from the staged diff → the task's check (`checks/`) is injected as `taskcheck_test.go` → `go test -tags taskcheck` runs in the container (host `go` is the no-engine fallback). An errored run is retried on a fresh work tree with escalating waits (1m, 10m, 30m, then hourly, up to 9 retries in total - rides out subscription quota windows); an error row is recorded only after all retries fail.

## Read the numbers honestly

- `claude` is a real agentic session; `ollama` is single-shot with the fixture inlined in the prompt — comparable within a backend, not across.
- Every check is proven passable: `testdata/reference/<task>/` holds a minimal solution that `TestReferenceSolutions` overlays and gates (no dir = "change nothing" is the answer).
- The `floor-dataloss` check is functional plus a structural grep for the rename pattern; `metrics.go` heuristics are documented inline with their ceilings.
- The summed row weights tasks by size, the median row does not: on the opus run 68% of magpie's line reduction is one task, the speculative config system (428 → 115), which is why the two rows read −55% and −28%. Quote whichever you like, but say which.
- A metric with no baseline has no percentage and drops out of the median row. Deps therefore always reads `n/a` there: a newly added dependency is by definition a zero baseline, so only the summed row can ever report one. Read `n/a` as "this row cannot tell you", not as "clean". (One opus rep did add a dep, `rung4-native` baseline rep 3; the median of its 3 reps washes it out.)
- The same skip is one-sided: a zero baseline can only grow, so dropping those tasks drops only regressions. On this run it changes nothing, but the median row can flatter an arm that adds where the baseline had nothing.
- Ten of the twelve tasks start at 0 or 1 entity, so the entity total rests almost entirely on the two rung-1 tasks; by median task it is ±0% for every arm.
- Quote **impl**, not **lines**, for anything about how much code a rule saves. `lines` includes the check the floor mandates, and the rule forbids trading the floor away for a smaller number, so charging it to the ladder measures the wrong thing. Inspected diffs showed two arms writing byte-identical implementations while the composite scored one of them 150% worse.
- The published table is the July run, whose baseline over-built far more than the August one did (828 net lines against 555 on the same twelve prompts, no rule involved). A rule can only remove bloat that the model produces, so magpie's measured gap shrinks as the base model improves. Treat any single run's headline as dated, not settled.
- Small local models follow multi-step rules poorly; expect magpie's gap to undersell there.

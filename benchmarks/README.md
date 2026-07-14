# Benchmark — verify the numbers

Measures what the rule claims: the magpie cost metric (**net lines, new deps, new entities**), gated by correctness. Two arms — `baseline` (no rule) vs `magpie` (`AGENTS.md` injected) — over twelve tasks against `fixture/` (a small Go project): one or two tasks per ladder rung, plus three floor probes where the prompt *invites* dropping safety (validation under "optimize it", an atomic write under "simplify it", SQL built from user input) and the check fails if the agent obliges.

## Run

```bash
make bench ARGS="-backend claude -model sonnet -repeat 3"   # agentic Claude Code session (subscription)
make bench ARGS="-backend ollama -model mistral -repeat 3"  # local single-shot
make bench ARGS="-arms ponytail,caveman -repeat 3"          # third-party rules on the same bench
```

Extra arms are rule files in `arms/<name>.md`, injected exactly like magpie's. `arms/ponytail.md` and `arms/caveman.md` are verbatim from [DietrichGebert/ponytail](https://github.com/DietrichGebert/ponytail) (MIT).

Rows stream to `results/<stamp>-<backend>-<model>.jsonl`; medians land in the matching `.md`. `go run ./cmd/bench -h` for all flags (task filter, arms, timeouts, `-keep` for work trees).

## How a run works

Copy fixture → git commit → agent gets the task prompt (non-baseline arm: rule as system prompt) → metrics from the staged diff → the task's check (`checks/`) is injected as `taskcheck_test.go` → `go test -tags taskcheck` runs in the container (host `go` is the no-engine fallback). An errored run is retried once on a fresh work tree; a second failure is recorded as an error row.

## Read the numbers honestly

- `claude` is a real agentic session; `ollama` is single-shot with the fixture inlined in the prompt — comparable within a backend, not across.
- Every check is proven passable: `testdata/reference/<task>/` holds a minimal solution that `TestReferenceSolutions` overlays and gates (no dir = "change nothing" is the answer).
- The `floor-dataloss` check is functional plus a structural grep for the rename pattern; `metrics.go` heuristics are documented inline with their ceilings.
- Small local models follow multi-step rules poorly; expect magpie's gap to undersell there.

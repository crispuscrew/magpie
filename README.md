# magpie

### Reuse before you write.

**An always-on rule that makes AI coding agents build with what already exists.** Stop at the first rung that holds, prove the pick with a cost line, never simplify away the floor.

> **−22% implementation code** · **−14% cost** per session · **−11% tokens**, output down 23%

1. Needs to exist at all? → skip it (YAGNI)
2. Already in this codebase? → reuse it
3. Stdlib does it? → use it
4. Native platform feature? → use it
5. Installed dependency? → use it
6. One line? → one line
7. Only then: the minimum that works

Every pick carries `cost: +N lines, +D deps, +E entities`. The **floor** survives even a prompt that says "simplify": trust-boundary validation, data-loss handling, security, accessibility, one runnable check. Deliberate shortcuts are marked `// reuse: <ceiling>, <upgrade path>`.

## Measured

Agentic Claude Code sessions (opus, August 2026), 12 tasks × 8 arms × 4 reps. **Every arm kept all 12 floor probes**, the prompts that bait you into dropping validation, atomic writes or parameterised SQL.

| arm | impl lines | checks left | cost | output tokens |
|---|--:|--:|--:|--:|
| baseline (no rule) | 323 | 8/12 | | |
| **magpie** | **−22%** | **11/12** | **−15%** | **−23%** |
| ponytail | −15% | 10/12 | −3% | −6% |
| caveman | −2% | 7/12 | −5% | −18% |

**impl** is net new lines outside `_test.go`. The check the floor mandates is counted separately, because a rule that stops testing is not a rule that got leaner, and **checks left** is what keeps that honest.

**Caveman does not reduce implementation code.** It looks far better on a naive line count, and that count was tests it never writes. Measured twice, at −2% and −1%, leaving a check on fewer tasks than using no rule at all.

**Magpie and ponytail are not separable here.** Three runs failed to put the gap outside the noise, and saying so beats a ranking the bench cannot support.

**The rule pays for itself.** ~426 tokens in per session, ~12,000 back.

**Error bars.** Four arms injecting byte-identical text landed 4 points apart, so anything narrower is not a result. Between runs it is worse: two rules swapped places by 29 lines across two nights. One fixture, one language, 12 tasks. Per-run pass rates and raw rows are in [benchmarks/](benchmarks/), not here, because at one or two failures per 48 sessions they measure agentic variance rather than any arm.

## Install

| Host | How |
|---|---|
| Claude Code | `/plugin marketplace add crispuscrew/magpie` then `/plugin install magpie@magpie` (installs `skills/magpie/SKILL.md`) |
| Codex / Gemini CLI / Copilot CLI / Antigravity and other `AGENTS.md` readers | copy `AGENTS.md` to the repo root |
| Cursor | copy `.cursor/rules/magpie.mdc` |
| Windsurf | copy `.windsurf/rules/magpie.md` |
| Cline | copy `.clinerules/magpie.md` |
| GitHub Copilot | copy `.github/copilot-instructions.md` |
| Kiro | copy `.kiro/steering/magpie.md` |
| `.agents/` workspace-rule hosts | copy `.agents/rules/magpie.md` |
| anything else | paste `AGENTS.md` into the system prompt |

All seven host files, the Claude Code skill included, are generated from `AGENTS.md` (`make adapters`); CI fails on drift. What you install is what gets benchmarked.

**Longer variant.** `variants/magpie-tight.md` is the fuller rule this one was compressed from, 704 words against 305, spelling out the cost line and carrying a stricter check paragraph. Copy it over `AGENTS.md` and run `make adapters` if you prefer it. The short one is default because its ~426 tokens against ~956 are paid on every session forever, and on code output the two are not separable.

## Skills

`magpie` (the always-on rule) · `magpie-review` (diff review: over-building + floor breaches) · `magpie-debt` (ledger of `reuse:` shortcuts) · `magpie-help` (reference).

## Verify

`make bench` reruns everything against a real Go fixture, any arm vs baseline, every check proven passable by a committed reference solution. Include the `control` arm: it injects the shipped rule under a second name, so the run measures its own error bar. Method in [benchmarks/](benchmarks/).

`make help` for the rest; everything but `make bench` runs in a container (Podman or Docker). Apache-2.0.

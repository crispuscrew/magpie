# magpie

Reuse before you write. An AI-agent rule that stops at the first ladder rung that holds:

1. Needs to exist at all? → skip it (YAGNI)
2. Already in this codebase? → reuse it
3. Stdlib does it? → use it
4. Native platform feature? → use it
5. Installed dependency? → use it
6. One line? → one line
7. Only then: the minimum that works

Every pick is proven by a cost line — `cost: +N lines, +D deps, +E entities` — and the **floor** is never simplified away, even when the prompt says "simplify": trust-boundary validation, data-loss handling, security, accessibility, one runnable check. Deliberate shortcuts are marked `// reuse: <ceiling>, <upgrade path>`.

## Measured

Real agentic Claude Code sessions (opus, August 2026), 12 tasks x 8 arms x 4 reps, correctness reported per run rather than assumed. Every arm runs the same bench: rule files in `benchmarks/arms/`, method and raw rows in [benchmarks/](benchmarks/).

Correctness first, because the rest only means something on top of it: **every arm kept all 12 floor probes**, the prompts that invite dropping validation, atomic writes or parameterised SQL. Task checks gate every figure below. Per-run pass rates live in [benchmarks/](benchmarks/) rather than here: at one or two failures per 48 sessions, landing on a rule competing with itself as readily as on anything else, they measure agentic variance and not the arms.

| arm | impl lines | checks left | cost | output tokens |
|---|--:|--:|--:|--:|
| baseline (no rule) | 323 | 8/12 | | |
| **magpie** | **−22%** | **11/12** | **−15%** | **−23%** |
| ponytail | −15% | 10/12 | −3% | −6% |
| caveman | −2% | 7/12 | −5% | −18% |

**impl** is net new lines outside `_test.go`: the code the ladder is there to shrink. The check the floor mandates is counted separately, because a rule that stops testing is not a rule that got leaner. **checks left** is the tasks where the arm left any runnable check at all, which is what stops "wrote less code" from being scored as a win when it means "wrote no test".

Three readings, in order of how much they should change your mind:

**Caveman does not reduce implementation code.** It publishes far better on a naive line count, and that count was tests it never writes. Measured twice, at −2.5% and −1.3%, while leaving a check on fewer tasks than using no rule at all. Reuse discipline and terseness are different things.

**Magpie and ponytail are not separable here.** Magpie is nominally ahead on impl and clearly ahead on cost, but three runs have failed to put the gap outside the noise. At 12 tasks this bench cannot tell them apart, and saying so is more useful than a ranking it cannot support.

**The rule pays for itself in tokens.** It costs ~426 tokens injected per session and returns roughly 12,000, which is why cost falls even though the agent is being asked to think about a ladder first.

### What the numbers cannot carry

Four arms in that run injected byte-identical rule text under different names. They landed 13 impl lines apart, **4 points**, which is this bench's error bar; anything narrower is not a result. Between runs it is worse: the same two rules swapped places by 29 lines across two nights, so a single run's ranking of close arms is not reliable, however many repetitions it has. One fixture, one language, twelve tasks. `make bench` reruns all of it.

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

Every host file, the Claude Code skill included, is generated from `AGENTS.md` (`make adapters`); CI fails on drift. One rule text ships everywhere, so what you install is what gets benchmarked.

### Longer variant

`variants/magpie-tight.md` is the fuller rule the default was compressed from: 704 words against 305. It spells out the cost line's terms and triggers, and carries a stricter check paragraph (assert only what changed, reuse the setup already there, one assertion over a new test) that measured −23% implementation lines on the run that promoted it. Copy it over `AGENTS.md` and run `make adapters` if you would rather have the explicit version. It stays a benchmark arm, so it cannot rot.

The default is the short one because the rule is injected into every session, so its ~426 tokens against ~956 are paid on every request forever. On code output the two are not separable: across two runs the shorter one measured 18 lines worse and then 11 lines better, a swing wider than the noise floor.

## Skills

`magpie` (the always-on rule) · `magpie-review` (diff review: over-building + floor breaches) · `magpie-debt` (ledger of `reuse:` shortcuts) · `magpie-help` (reference).

## Verify the numbers

`make bench` reruns everything: 12 tasks (one or two per rung + three floor probes) against a real Go fixture, any arm vs baseline, every check proven passable by a committed reference solution. Always include the `control` arm, which injects the shipped rule under a second name so the run measures its own error bar. See `benchmarks/README.md`.

## Develop

`make help`; everything but `make bench` runs in a container (Podman or Docker). Apache-2.0.

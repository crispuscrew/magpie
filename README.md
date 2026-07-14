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

Real agentic Claude Code sessions (opus), 12 tasks × 4 arms × 3 reps, medians, correctness-gated. Every arm runs the same bench — rule files in `benchmarks/arms/`, method and raw rows in [benchmarks/](benchmarks/).

| arm | lines | entities | pass | floor kept |
|---|--:|--:|--:|--:|
| baseline (no rule) | — | — | 35/36 | 8/9 |
| **magpie** | **−55%** | −38% | **36/36** | **9/9** |
| ponytail | −45% | −43% | 33/36 | 8/9 |
| caveman | −33% | −5% | 36/36 | 9/9 |

lines/entities = median change vs baseline across all tasks, wash cases included; floor kept = floor probes passed (validation under "optimize it", atomic write under "simplify it", SQL injection bait). Best magpie cases: speculative config system −73%, DB constraint over app code −100%. Where magpie costs more it says so: the one-liner task runs +18% lines because the floor demands a runnable check a bare agent skips.

## Install

| Host | How |
|---|---|
| Claude Code | `/plugin marketplace add crispuscrew/magpie` then `/plugin install magpie@magpie` |
| Codex / Gemini CLI / Copilot CLI / Antigravity and other `AGENTS.md` readers | copy `AGENTS.md` to the repo root |
| Cursor | copy `.cursor/rules/magpie.mdc` |
| Windsurf | copy `.windsurf/rules/magpie.md` |
| Cline | copy `.clinerules/magpie.md` |
| GitHub Copilot | copy `.github/copilot-instructions.md` |
| Kiro | copy `.kiro/steering/magpie.md` |
| `.agents/` workspace-rule hosts | copy `.agents/rules/magpie.md` |
| anything else | paste `AGENTS.md` into the system prompt |

All host files are generated from `AGENTS.md` (`make adapters`); CI fails on drift.

## Skills

`magpie` (the always-on rule) · `magpie-review` (diff review: over-building + floor breaches) · `magpie-debt` (ledger of `reuse:` shortcuts) · `magpie-help` (reference).

## Verify the numbers

`make bench` reruns everything: 12 tasks (one or two per rung + three floor probes) against a real Go fixture, any arm vs baseline, every check proven passable by a committed reference solution — `benchmarks/README.md`.

## Develop

`make help`; everything but `make bench` runs in a container (Podman or Docker). Apache-2.0.

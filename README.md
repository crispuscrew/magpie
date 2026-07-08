# magpie

Reuse before you write. Makes an AI coding agent build with what already exists — the cheapest code is the code never written.

Before any code, the agent stops at the first ladder rung that holds:

1. Needs to exist at all? → skip it (YAGNI)
2. Already in this codebase? → reuse it
3. Stdlib does it? → use it
4. Native platform feature? → use it
5. Installed dependency? → use it
6. One line? → one line
7. Only then: the minimum that works

Two things ponytail-style rules don't have: a **cost line** — `cost: +N lines, +D deps, +E entities` — that makes every pick provable, and a **floor** that is never simplified away (trust-boundary validation, data-loss handling, security, accessibility, one runnable check per non-trivial change). Deliberate shortcuts are marked in place: `// reuse: <ceiling>, <upgrade path>`.

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

`benchmarks/` measures baseline vs magpie on the exact metric the rule claims — lines, deps, entities — plus a correctness gate and floor probes. See `benchmarks/README.md`.

## Develop

`make help` lists targets; everything but `make bench` runs in a container (Podman or Docker).

Apache-2.0.

# magpie

### Reuse before you write.

An always-on rule for AI coding agents.

## −22% implementation code · −14% cost · −11% tokens

```diff
  12 tasks, one Go fixture, Claude Code on opus
- 323 implementation lines   $0.49 a session   no rule
+ 252 implementation lines   $0.42 a session   magpie
```

1. Needs to exist at all? → skip it (YAGNI)
2. Already in this codebase? → reuse it
3. Stdlib does it? → use it
4. Native platform feature? → use it
5. Installed dependency? → use it
6. One line? → one line
7. Only then: the minimum that works

Every pick carries `cost: +N lines, +D deps, +E entities`. The **floor** survives even a prompt that says "simplify": trust-boundary validation, data-loss handling, security, accessibility, one runnable check. Deliberate shortcuts are marked `// reuse: <ceiling>, <upgrade path>`.

## Install

Paste this into any agent you already have open. It installs globally, for every project:

```
Install the magpie rule globally, for every repo and not just this one.

Fetch https://raw.githubusercontent.com/crispuscrew/magpie/main/AGENTS.md and
append it, under a "# Magpie" heading, to your own user-level rules file: the
one that applies across all projects. That is ~/.claude/CLAUDE.md for Claude
Code, ~/.codex/AGENTS.md for Codex, ~/.gemini/GEMINI.md for Gemini CLI, or
whichever path your own docs give for global rules.

Append only. Never overwrite or reorder what is already in that file, and
create it only if it does not exist. Tell me the path you used.
```

Or by hand:

| Host | Copy to |
|---|---|
| Claude Code | `/plugin marketplace add crispuscrew/magpie` then `/plugin install magpie@magpie` (installs `skills/magpie/SKILL.md`, globally) |
| Codex / Gemini CLI / Copilot CLI / Antigravity | `AGENTS.md` at the repo root |
| Cursor | `.cursor/rules/magpie.mdc` |
| Windsurf | `.windsurf/rules/magpie.md` |
| Cline | `.clinerules/magpie.md` |
| GitHub Copilot | `.github/copilot-instructions.md` |
| Kiro | `.kiro/steering/magpie.md` |
| `.agents/` workspace hosts | `.agents/rules/magpie.md` |
| anything else | paste `AGENTS.md` into the system prompt |

All seven host files, the Claude Code skill included, are generated from `AGENTS.md` (`make adapters`); CI fails on drift. What you install is what gets benchmarked.

**Longer variant.** `variants/magpie-tight.md` is the fuller rule this one was compressed from, 704 words against 305, spelling out the cost line and carrying a stricter check paragraph. Copy it over `AGENTS.md` and run `make adapters` if you prefer it. The short one is default because its ~426 tokens against ~956 are paid on every session forever, and on code output the two are not separable.

## Measured

Agentic Claude Code sessions (opus, August 2026), 12 tasks × 8 arms × 4 reps. **Every arm kept all 12 floor probes**, the prompts that bait an agent into dropping validation, atomic writes or parameterised SQL.

| arm | impl lines | checks left | cost | output tokens |
|---|--:|--:|--:|--:|
| baseline (no rule) | 323 | 8/12 | | |
| **magpie** | **−22%** | **11/12** | **−15%** | **−23%** |
| ponytail | −15% | 10/12 | −3% | −6% |
| caveman | −2% | 7/12 | −5% | −18% |

**impl** is net new lines outside `_test.go`; the mandated check is counted separately, and **checks left** is what stops "wrote less code" from meaning "wrote no test". That is the whole caveman result: strong on a naive count, −2% on implementation, fewer checks than using no rule at all. Magpie and ponytail are not separable, three runs having failed to clear the noise.

**Error bars.** Byte-identical rules landed 4 points apart, so narrower gaps are not results, and between runs two rules swapped places by 29 lines. One fixture, one language, 12 tasks. Raw rows and per-run pass rates in [benchmarks/](benchmarks/).

## Skills

`magpie` (the always-on rule) · `magpie-review` (diff review: over-building + floor breaches) · `magpie-debt` (ledger of `reuse:` shortcuts) · `magpie-help` (reference).

## Verify

`make bench` reruns everything against a real Go fixture, any arm vs baseline, every check proven passable by a committed reference solution. Include the `control` arm: it injects the shipped rule under a second name, so the run measures its own error bar. Method in [benchmarks/](benchmarks/).

`make help` for the rest; everything but `make bench` runs in a container (Podman or Docker). Apache-2.0.

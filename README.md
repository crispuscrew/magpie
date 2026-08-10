# magpie

### Reuse before you write.

**An always-on rule that makes AI coding agents build with what already exists.** Stop at the first rung that holds, prove the pick with a cost line, never simplify away the floor.

<table>
<tr>
<td align="center"><h1>−22%</h1>implementation code</td>
<td align="center"><h1>−14%</h1>cost per session</td>
<td align="center"><h1>−11%</h1>tokens, output −23%</td>
</tr>
</table>

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

**impl** is net new lines outside `_test.go`; the check the floor mandates is counted separately, and **checks left** is what stops "wrote less code" from meaning "wrote no test".

That distinction is the whole result for caveman. It looks strong on a naive line count, and the count was tests it never writes: −2% implementation, measured twice, leaving checks on fewer tasks than using no rule at all. Magpie and ponytail are not separable, three runs having failed to put the gap outside the noise.

**Error bars.** Four arms injecting byte-identical text landed 4 points apart, so narrower gaps are not results, and between runs two rules swapped places by 29 lines. One fixture, one language, 12 tasks. Raw rows and per-run pass rates in [benchmarks/](benchmarks/).

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

**Or let the agent do it.** Paste this into whatever agent you already have open:

```
Install the magpie rule in this repository. Work out which agent config this
repo already uses, then fetch the matching file from
https://raw.githubusercontent.com/crispuscrew/magpie/main/<path> and write it
to the same path here:

  AGENTS.md                        Codex, Gemini CLI, Copilot CLI, Antigravity
  .cursor/rules/magpie.mdc         Cursor
  .windsurf/rules/magpie.md        Windsurf
  .clinerules/magpie.md            Cline
  .github/copilot-instructions.md  GitHub Copilot
  .kiro/steering/magpie.md         Kiro
  .agents/rules/magpie.md          .agents workspace hosts

Take every path whose directory already exists, and AGENTS.md if none do.
Fetch each file rather than retyping it: the host frontmatter differs per
path and is load-bearing. Do not edit the rule text.
```

All seven host files, the Claude Code skill included, are generated from `AGENTS.md` (`make adapters`); CI fails on drift. What you install is what gets benchmarked.

**Longer variant.** `variants/magpie-tight.md` is the fuller rule this one was compressed from, 704 words against 305, spelling out the cost line and carrying a stricter check paragraph. Copy it over `AGENTS.md` and run `make adapters` if you prefer it. The short one is default because its ~426 tokens against ~956 are paid on every session forever, and on code output the two are not separable.

## Skills

`magpie` (the always-on rule) · `magpie-review` (diff review: over-building + floor breaches) · `magpie-debt` (ledger of `reuse:` shortcuts) · `magpie-help` (reference).

## Verify

`make bench` reruns everything against a real Go fixture, any arm vs baseline, every check proven passable by a committed reference solution. Include the `control` arm: it injects the shipped rule under a second name, so the run measures its own error bar. Method in [benchmarks/](benchmarks/).

`make help` for the rest; everything but `make bench` runs in a container (Podman or Docker). Apache-2.0.

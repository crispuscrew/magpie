---
name: magpie-debt
description: >
  Ledger of `reuse:` shortcut comments — where, ceiling, upgrade path,
  which are due. Use on "magpie debt" or any ask about deferred shortcuts
  or a debt ledger.
license: Apache-2.0
---

# Magpie debt

Deliberate shortcuts are marked `reuse: <ceiling>, <upgrade path>` at the
site. This skill turns those markers into a decision, not a backlog.

## Steps

1. Search the repo for `reuse:` comments (all comment syntaxes: `//`, `#`, `--`, `/* */`, `<!-- -->`).
2. For each, read enough context to restate: **where** (`file:line`), **shortcut** (what was deliberately not built), **ceiling** (the condition that breaks it), **upgrade** (the named path).
3. Judge each against the code as it is *today*: has the ceiling been hit or is it near? Look for the trigger the comment names — throughput, table size, caller count, whatever it says.
4. A `reuse:` comment whose code was since upgraded or deleted is stale — flag it for removal.

## Output

A single table: location, shortcut, ceiling, upgrade, verdict (`ok` /
`due` / `stale`). After the table, at most one line per `due` item saying
what to do. No `reuse:` comments in the repo → say so in one line and stop.
Do not invent debt the markers don't claim; unmarked code belongs to
`magpie-review`, not here.

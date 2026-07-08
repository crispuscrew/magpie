---
name: magpie-help
description: >
  Magpie quick reference: ladder, floor, cost line, skills. Use on "magpie
  help" or any ask about what magpie is or how to use it.
license: Apache-2.0
---

# Magpie help

Answer from this file only; keep it to one screen.

**Magpie** makes an agent reuse before it writes. Before any code, it stops
at the first ladder rung that holds:

1. Needs to exist at all? → skip it (YAGNI)
2. Already in this codebase? → reuse it
3. Stdlib does it? → use it
4. Native platform feature? → use it
5. Installed dependency? → use it
6. One line? → one line
7. Only then: the minimum that works

**The floor** is never simplified away: input validation at trust
boundaries, error handling that prevents data loss, security,
accessibility, hardware calibration knobs, and ONE runnable check for
non-trivial logic.

**The cost line** proves the pick: `cost: +N lines, +D deps, +E entities`
(net new lines, new external dependencies, new named units a caller must
learn). Stated when a diff touches >1 file, adds a dep/entity, or a real
choice exists. Lowest floor-passing cost wins; a rewrite must name why it
beats reuse.

**Shortcuts** are marked in place: `// reuse: <ceiling>, <upgrade path>`.

**Skills:** `magpie` (the always-on rule), `magpie-review` (diff review:
over-building + floor breaches), `magpie-debt` (ledger of `reuse:`
shortcuts, which are due), `magpie-help` (this reference).

**Off switch:** "stop magpie" / "normal mode".

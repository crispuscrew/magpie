---
name: magpie-review
description: >
  Review a diff against the magpie ladder: over-building (should have
  stopped at a higher rung) and floor breaches (safety simplified away).
  Use on "magpie review" or any ask to check a diff for over-engineering,
  bloat, reinvented helpers, or new dependencies.
license: Apache-2.0
---

# Magpie review

Review the diff the user points at (default: working tree against the
default branch). Judge every hunk against the ladder in two directions —
over-building *and* floor breaches. Read enough surrounding code to know
what already exists; a reuse finding is only real if you name the existing
thing.

## Findings — over-building

For each finding, report:

- `file:line` — where.
- **Rung** — the highest ladder rung that already held (1–6) and what sits there: the existing helper, the stdlib call, the native feature, the installed dep, or the one-liner.
- **Cost delta** — what dropping to that rung saves: `saves: N lines, D deps, E entities`.
- **Fix** — the one-line replacement or deletion.

Typical catches: re-implemented helper that lives a few files over, new
dependency for a few lines, interface with one impl, config for a constant,
scaffolding "for later", guard duplicated per caller instead of once in the
shared function.

## Findings — floor breaches

Flag any hunk that *removed or skipped* the floor: input validation at a
trust boundary, error handling that prevents data loss, security,
accessibility, a calibration knob, or the ONE runnable check non-trivial
logic must leave behind. A floor breach outranks every over-building
finding — report it first.

## Shortcut hygiene

A deliberate shortcut is fine **iff** marked with a `reuse:` comment naming
the ceiling and upgrade path. Unmarked shortcut → ask for the comment, not
for more code.

## Output

Findings ordered floor-first, then by cost delta descending. Each finding
at most three lines. No finding → say "clean at every rung" and stop; a
review padded with nitpicks is itself over-building.

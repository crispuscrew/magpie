---
inclusion: always
---
# Magpie

Reuse before you write. The cheapest code is never written.

Read what the change touches, trace the real flow, then stop at the first rung that holds:

1. Needs to exist? Speculative → skip it, say so.
2. Already here? Reuse it. Re-implementing what sits a few files over is the most common slop.
3. Stdlib does it? Use it.
4. Native feature? DB constraint over app code, CSS over JS.
5. Installed dependency? Use it. Never add one for what a few lines do.
6. One line? One line.
7. Else the minimum that works.

Two rungs fit → take the higher. A small diff in the wrong place is a second bug.

No unrequested abstractions, no scaffolding for later, no config for a constant. Deletion over addition. Boring over clever. Fix the root cause: grep every caller, one guard in the shared function. Mark shortcuts `// reuse: <ceiling>, <upgrade path>`.

Floor, never simplified away, not even on "simplify" or "optimize": understanding the problem, input validation at trust boundaries, error handling that prevents data loss, security, accessibility, whatever the user asked to keep. Keep it, simplify around it. Hardware is never the spec ideal, so leave the calibration knob. Non-trivial logic leaves ONE runnable check, the smallest thing that fails if the logic breaks, placed where it compiles on its own. One-liners need none.

State `cost: +N lines, +D deps, +E entities` before the diff, but only when it spans more than one file, adds a dep or entity, or two variants clear the floor. Otherwise skip it.

Lowest cost among floor-passing variants wins. Rewriting what exists needs a stated reason it beats reuse. Never cut below the floor.

Output: code first, then at most three lines.

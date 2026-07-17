# Tests Pass Rubric

The question is not "is there coverage" but "do the tests earn their keep": critical paths proven, failures diagnosable, suite designed rather than accreted. Tests are executable specs — user-facing and integration level preferred, implementation-coupled tests suspect.

## Prepare

- Locate the area's tests: colocated test files, owning suites in the repo's test trees, db tests — per the repo profile's stack notes. Read the owning agent-instruction files (AGENTS.md / CLAUDE.md) for suite conventions.
- List the area's 3–5 highest-stakes behaviors first (money, persisted data, contracts, concurrency, user-visible workflows). Review against that list, not against the file tree.

## Check

- **Critical-path depth.** Each high-stakes behavior has tests at the right level, including failure paths and the edge cases that actually occur. A missing failure-path test on a critical behavior outranks any number of present happy-path tests.
- **Bandage detection.** Clusters of near-duplicate tests appended over time, copy-pasted setup drifting apart, tests pinning incidental behavior, suites with no discernible design. The action is a consolidation proposal with a named target shape — never "add more tests".
- **Overwritten without consideration.** For suspicious files run `git log --follow -p -- <file>` and look for loosened assertions (`toEqual` → `toBeDefined`, exact → `toContain`, widened tolerances), deleted cases, or added `.skip` without justification in the commit message. Cite the commit. Always `triage` — the weakening may have been deliberate.
- **Over-testing.** Exhaustive matrices on trivial code, and implementation-coupled tests (asserting mock wiring or internals) that break on refactor without catching bugs. Propose deletion or lifting to behavior level; less is a valid recommendation.
- **Diagnosability.** Would this failure message localize the bug? Test names state behavior; mocks appear only where they clarify a realistic scenario or failure path.
- **Flake patterns.** Sleeps, real time, order dependence, shared mutable fixtures, network reliance in unit suites.

## Do not flag

- Missing tests on trivial glue or code the types already make safe.
- Exhaustive coverage where the domain genuinely is critical (contract guards, schema validation) — matrices are earned there.
- Style of green tests unless coupled, duplicated, or misleading.

## Severity and tier

Untested critical path: major or blocker. Bandage cluster obscuring the suite: major. Coupling: by refactor pain. Auto tier only for byte-identical duplicate tests, name/assertion mismatches, and unused test helpers — restoring a weakened assertion is always triage.

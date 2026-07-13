---
name: next-goal
description: "Build an evidence-backed, self-contained implementation prompt to paste after /goal in a fresh Codex session. Use only when the user explicitly invokes $next-goal to choose the next goal-sized outcome from repository plans and current state."
---

# Next Goal

Generate a fresh-session implementation handoff. Keep this run read-only: inspect the
repository and return a prompt, but leave files, plans, goals, git state, and external systems
unchanged.

## 1. Establish current state

1. Read every `AGENTS.md` that applies to the likely work area.
2. Resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff document.
   Honor a user-named document; otherwise use repository conventions, links, status markers,
   and recent history. When no plan exists, infer candidates from instructions, code, tests,
   and history.
3. Inspect git status and recent history, then read only enough implementation and validation
   evidence to verify plan claims, prerequisites, and completed work. Treat unrelated changes
   as intentional and retain enough detail for the generated prompt to protect them.
4. Ask only when multiple plausible choices would materially change the outcome or no safe,
   unblocked goal can be supported by evidence.

This step is complete when one real next outcome and its constraints are verified against the
repository rather than merely repeated from a plan.

## 2. Select the goal boundary

Choose the broadest connected implementation outcome that is still coherent, reviewable, and
validatable. Group work that shares a phase, subsystem, ownership seam, files, decisions, or
validation suite, including small prerequisites needed to make the outcome real.

The selected goal must:

- end in one concrete, demonstrable artifact or behavior;
- support multiple coherent passing commits when the work warrants them;
- have settled requirements and a crisp completion condition;
- stop before an independently useful phase, unrelated high-risk seam, new external
  coordination, or a product, migration, security, dependency, platform, or compatibility
  decision not settled by repository evidence.

Name the maximum recommended boundary and the next meaningful excluded area. The boundary is
complete when it is larger than a stray checkbox, smaller than an incoherent roadmap, and can
be reviewed and validated as one outcome.

## 3. Write the fresh-session prompt

Return, in this order:

1. **Recommended boundary** — a short scope statement and evidence-based rationale.
2. **Copy-paste `/goal` prompt** — one fenced block containing only the body to enter after
   `/goal`. Assume the new session has none of this conversation.
3. **Next excluded area** — one short statement when it clarifies the stopping boundary.

Make the prompt self-contained and concrete. Include:

- the objective and completion artifact;
- authoritative documents and relevant paths to read first;
- verified current state without brittle commit hashes unless a hash is essential;
- explicit in-scope work, non-goals, invariants, dependency rules, and evidence-backed
  decision constraints;
- acceptance criteria and repository-native focused and final validation commands;
- the duty to keep the active plan or progress document current;
- the stopping condition and next deferred boundary;
- instructions to inspect and preserve all pre-existing changes.

Require the goal-running session to:

- capture its starting git ref before editing so final review and whitespace checks cover
  both progressively committed work and the remaining worktree diff;
- make local commits as coherent passing slices finish, using descriptive conventional
  subjects consistent with the repository;
- stage explicit goal-owned paths and keep unrelated changes out of commits;
- run applicable focused validation, final repository validation, and repository-specific
  review instructions found in `AGENTS.md`, including named review skills;
- finish successfully with all completed goal-owned work committed while preserving the
  starting dirty state; when blocked, report partial or failing work without committing a
  failing slice;
- leave pushing, pull requests, merging, releases, and other publication to separate user
  authorization.

Include a token budget only when the user supplies one. Do not turn uncertain choices into
instructions; surface a blocking decision instead.

Before responding, verify that the prompt is actionable without this conversation and that
this run made no repository write, goal change, git mutation, or external publication.

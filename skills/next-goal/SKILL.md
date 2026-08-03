---
name: next-goal
description: "Present evidence-backed substantial goal scopes, recommend one, pause for the user's choice, then emit a copy-ready fresh-session routing envelope for the confirmed scope. Explicit invocation only."
---

# Next Goal

Discover the substantial next-goal scopes supported by repository evidence, recommend the most appropriate one, and ask the user which scope to pursue. Always pause for explicit scope confirmation, even when only one viable scope exists or the invoking request names a preferred scope. Emit no `/goal` prompt in that scope-choice response.

After the user confirms a scope in a follow-up, revalidate it, pass the readiness gate, recommend PR delivery or later aggregation based on expected change size, and generate one compact fresh-session routing envelope with a closed execution contract. A clear selection or adjusted boundary in that follow-up is the confirmation; do not ask again merely to restate it. Generate a non-recommended delivery prompt only when the user explicitly requests that variant, alone or alongside the recommendation.

Keep scope discovery, recommendation, confirmation, and final selection read-only. A combined request may separately authorize prerequisite mutation, such as committing completed planning work. Complete that distinct phase first under its applicable workflow, then discover scopes from the resulting repository state without further mutation. When a confirmed selection produces a goal prompt, keep prerequisite results out of the final response so the entire response remains directly copyable into `/goal`. Put `$progress` goal tracking in the generated prompt so the goal-running session, not the selection phase, initializes durable goal state.

## 1. Establish Current State

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow repository conventions and links. When concurrent worktrees or scoped roadmaps exist, select the current worktree's planning namespace; read other scopes only for an explicitly requested aggregate goal. When no plan exists, infer candidates from instructions, code, tests, and history.
2. Inspect git status and recent history, then read only enough implementation and validation evidence to detect stale plan claims, completed work, real prerequisites, and blockers.
3. Identify candidate outcomes, constraints, missing evidence, and consequential unresolved decisions. Leave user-owned implementation questions for the readiness gate so every such question carries the delegation-or-repair choice; choosing the goal boundary belongs to the following scope-choice step.

This step is complete when the current project state, candidate outcomes, and their material gaps are verified against the repository rather than merely repeated from a plan.

## 2. Present the Scope Choices

Build a concise choice set of materially distinct, substantial goal boundaries the user could authorize. Treat named slices and checklist items as planning units, not automatic options or stopping boundaries. Options may differ by outcome or by coherent stopping point, but do not enumerate every permutation of adjacent work. Prefer two to four options when the evidence supports them; never invent a weak option to reach a count.

Each option must state only the decision-relevant boundary:

- the semantic outcome;
- the included results, summarized rather than expanded into a goal contract;
- the immediate next result or milestone it stops before; and
- whether it is ready or has a specific planning gap that would require delegation or repair after selection.

Recommend exactly one option and explain why in one compact sentence. Prefer the largest useful outcome that a persistent goal-running agent can pursue autonomously, while weighing outcome value, coherence, planning readiness, blocker risk, and separation from later milestones. The scope recommendation is distinct from the later delivery recommendation.

Use a structured choice control when available; otherwise use a numbered list. Ask the user to choose by number or name or to propose an adjusted boundary. Emit no fenced goal prompt. When only one substantial candidate exists, present it as the recommendation and ask the user to confirm it. If no substantial unblocked candidate remains, say that `/goal` is not warranted and stop instead of manufacturing an option.

The first invocation always stops here after presenting choices, even if the user named a preferred scope or requested a delivery variant. A follow-up that clearly selects an offered option or proposes a clear adjusted boundary satisfies scope confirmation.

## 3. Finalize the Confirmed Goal Boundary

Revalidate the confirmed scope against the current repository state before finalizing it. If material repository changes invalidate the option or its recommendation, refresh the choice set and ask again. Otherwise preserve the user's chosen boundary even when it differs from the recommendation.

Apply these boundary rules both when framing options and when finalizing the confirmed choice. Naturally connected work may be absorbed into an option, and a confirmed adjusted boundary may be refined only enough to make it semantically closed. A goal may span subsystems and multiple coherent commits when they lead to one meaningful project state.

When both delivery variants are requested, keep their confirmed boundary identical. In the PR variant, treat PRs as delivery checkpoints within the large goal rather than separate `/goal` boundaries.

The selected boundary must:

- be materially larger than work suited to one ordinary interactive turn;
- reach a concrete, demonstrable project or user outcome rather than only a prerequisite or internal seam;
- contain enough settled or explicitly delegated work to benefit from persistent execution across multiple checkpoints;
- stop at a consequential unresolved decision the readiness gate does not settle or delegate, external authorization, blocker, or materially unrelated next milestone—not merely at the next plan heading or reviewable slice.

Do not expand a small confirmed scope without the user. If the confirmed choice is not substantial enough for `/goal`, explain that and return to scope choice rather than silently replacing it. A substantial confirmed scope whose planning is incomplete proceeds to the gate.

Before the readiness gate, reduce the confirmed scope to this closed routing envelope:

- **Outcome** — the semantic project or user result in one sentence.
- **Goal state** — one concrete durable path: `goals/<stable-slug>.md` for ordinary work or `goals/<scope>/<stable-slug>.md` for an isolated worktree planning scope.
- **Included results and sources** — every authorized result as a short, stable semantic label paired with the few authoritative documents that supply its implementation and acceptance detail. Labels define membership; paths and queue positions do not.
- **Completion** — one compact predicate requiring each named result to achieve its cited outcome and any applicable completion criteria, plus only cross-cutting validation, review, freshness, and delivery conditions not already carried by those sources.
- **Excluded work** — exactly the immediate next out-of-scope milestone plus exclusions stated directly by the user. Authority supplies the complete boundary for every later, unrelated, or merely plan-documented item.
- **Authority** — allow only the smallest bounded work necessary for an included result. When the readiness gate passed by delegation, also authorize the goal-running agent to resolve remaining decisions within the closed outcome using its best judgment. Record anything outside the boundary for later and require explicit user direction for expansion or external authority.
- **Resume invariant** — at every resumed turn, automatic continuation,
  compaction recovery, or handoff, invoke `$progress` in goal mode and recover
  the named goal state before selecting or starting more work.
- **Delivery** — the selected lifecycle and its skill routing, kept inside the contract so recovery preserves it.

Record the exact planning gap when evidence cannot yet support a field; do not invent closure. Boundary finalization is complete when the confirmed scope, its provisional envelope, and every unsupported field are ready for the readiness gate.

Treat goal membership as closed. Advancing a roadmap, changing `Current`, creating a plan, opening a branch or PR, or discovering review findings never adds work to the goal. Project planning state describes what the project should do next; the goal contract alone describes what this run is authorized to do.

## 4. Pass the Readiness Gate

Before emitting a goal prompt, verify that the evidence supports a closed outcome, semantic included results and authoritative sources, a completion predicate, and enough settled direction for autonomous implementation. Reversible implementer-owned choices may remain open. Consequential user-owned decisions require explicit delegation or planning repair.

When the outcome is closed but consequential decisions remain, alert the user with the specific planning gaps and ask them to choose:

1. authorize the goal-running agent to resolve the remaining decisions within the closed outcome using its best judgment; or
2. repair planning first with `$interview` followed by `$progress`, then rerun goal selection.

Emit no goal prompt before the user chooses. Delegation passes the gate only for decisions inside the supported outcome; it never adds results, expands scope, or grants external authority. Record that delegation in the generated contract's `Authority` field so the goal-running agent does not ask again merely because the cited plans left those decisions open.

If the user chooses planning repair, treat that answer as an explicit request for the separate mutating phase: use `$interview` to settle consequential decisions, then `$progress` to update the authoritative planning documents. Restart current-state resolution from the resulting repository state and keep the renewed selection phase read-only. When no closed outcome can be supported, explain why delegation is unavailable and ask to repair planning before goal selection.

The gate is complete only when planning is sufficient or the user has explicitly delegated the remaining decisions within a supported closed outcome. Finalize every routing-envelope field after it passes.

## 5. Route the Result

- When scope has **not been confirmed in a follow-up**, return only the compact choice set, recommendation, and selection question from step 2. Do not read the delivery-variant instructions or emit a goal prompt.
- When `/goal` is **not warranted**, give the evidence-based reason and omit the prompt. Do not read the delivery-variant instructions.
- Only after the user confirms a scope, the evidence establishes that `/goal` **is warranted**, and the readiness gate passes, read and follow [prompts/delivery-variants.md](prompts/delivery-variants.md).
- By default, return only the recommended prompt as one unlabeled `text` fenced block. Put only the body to enter after `/goal` inside it, with no prose before or after the fence.
- Honor an explicit request for one named delivery variant even when it differs from the evidence-based recommendation; identify the emitted variant through its `Delivery` field.
- When the user explicitly requests both variants, return only the two `text` fenced prompt blocks, identify the variant inside each prompt's `Delivery` field, put the same closed scope contract inside both prompt bodies, and vary only the delivery mechanics and `Delivery` field.

Before responding, verify that the scope and selection phases made no repository write, goal change, git mutation, or external publication. When the request included prerequisite mutation, verify that it finished before scope discovery began, but do not add a separate recap when emitting a goal prompt after confirmation.

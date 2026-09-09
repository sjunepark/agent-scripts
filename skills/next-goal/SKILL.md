---
name: next-goal
description: "Select a substantial next goal; optionally prepare and commit its plan, or spawn a Codex task with a native goal and model preset. Explicit invocation only."
---

# Next Goal

Discover the substantial next-goal scopes supported by repository evidence. Use a clear scope selected in the current or an earlier request. A request to choose the next goal delegates scope selection: choose the best supported boundary. When only one viable scope exists, select it unless the user requested a choice or preview first. Ask only when materially different outcomes remain unresolved by user direction or delegated judgment.

Revalidate the selected scope, pass the readiness gate, recommend PR delivery or later aggregation based on expected change size, and generate one compact fresh-session routing envelope with a closed execution contract. Scope selection does not require a separate user turn. Generate a non-recommended delivery prompt only when the user explicitly requests that variant, alone or alongside the recommendation.

Keep scope discovery and selection read-only. Preparation is a separately authorized phase described below; after it finishes, revalidate the same boundary from the resulting repository state. When returning a goal prompt, keep prerequisite results in commentary so the final response remains directly copyable into `/goal`. Put `$progress` goal tracking in the contract so the goal-running session initializes durable goal state.

## Choose the Operation

These are explicit skill instructions, not shell commands:

| Invocation | Authorized operation |
| --- | --- |
| `$next-goal` | Select and return a prompt; no mutation by default. |
| `$next-goal prepare` | Select, complete and review the detailed plan, commit relevant preparation, then return a prompt. |
| `$next-goal spawn [model] [reasoning]` | Prepare, then create one new Codex task from the prepared state and start a native goal there. |

Honor an explicit prompt-only, preview, no-commit, or wait instruction over these defaults. A mention in a quote, example, or skill-design discussion does not request execution. Model arguments belong only to `spawn`; do not infer spawning from a model name alone. Keep explicit invocation policy unchanged.

For `prepare` and `spawn`, read [Prepare a goal](workflows/prepare.md) before mutation. Preparation authorizes planning and its scoped commit, not implementation or a blanket commit of existing work. Reuse existing user decisions and delegated judgment; ask only for consequential decisions they cannot resolve.

For `spawn`, also read [Spawn a Codex goal](workflows/spawn.md) before preparation for capability, target, and argument checks. This mode explicitly requests a new task starting from the current prepared Git state in the selected project; it never means an ordinary subagent. Model aliases are `astra` (GPT-6 Astra, medium) and `luna` (GPT-5.6 Luna, max); the spawn workflow owns exact tool arguments and optional overrides. With no model, use configured task defaults.

Run steps 1–3 to resolve the outcome. In `prepare` or `spawn` mode, then perform the preparation workflow and revalidate steps 1–3 before the readiness gate. An independently requested prerequisite edit or commit may finish before initial discovery, as in the existing combined workflow. Keep the selected operation through planning repair and follow-up answers; do not repeat finished preparation or create duplicate tasks.

## 1. Establish Current State

1. Read applicable `AGENTS.md` files and resolve the authoritative active `PLAN`, `TODO`, `ROADMAP`, progress, or handoff documents. Honor user-named documents; otherwise follow repository conventions and links. When concurrent worktrees or scoped roadmaps exist, select the current worktree's planning namespace; read other scopes only for an explicitly requested aggregate goal. When no plan exists, infer candidates from instructions, code, tests, and history.
2. Inspect git status and recent history, then read only enough implementation and validation evidence to detect stale plan claims, completed work, real prerequisites, and blockers.
3. Identify candidate outcomes, constraints, missing evidence, and consequential unresolved decisions. Leave user-owned implementation questions for the readiness gate so every such question carries the delegation-or-repair choice; choosing the goal boundary belongs to the following scope-choice step.

This step is complete when the current project state, candidate outcomes, and their material gaps are verified against the repository rather than merely repeated from a plan.

## 2. Resolve the Scope

First apply an existing scope selection or delegated choice, or select the sole viable scope. If the boundary remains unresolved or the user requested options, build a concise choice set of materially distinct, substantial goal boundaries. Treat named slices and checklist items as planning units, not automatic options or stopping boundaries. Options may differ by outcome or by coherent stopping point, but do not enumerate every permutation of adjacent work. Prefer two to four options when the evidence supports them; never invent a weak option to reach a count.

Each option must state only the decision-relevant boundary:

- the semantic outcome;
- the included results, summarized rather than expanded into a goal contract;
- the immediate next result or milestone it stops before; and
- whether it is ready or has a specific planning gap that would require delegation or repair after selection.

When presenting choices, mark exactly one option as `(Recommended)` and explain why in one compact sentence. Prefer the largest useful outcome that a persistent goal-running agent can pursue autonomously, while weighing outcome value, coherence, planning readiness, blocker risk, and separation from later milestones. The scope recommendation is distinct from the later delivery recommendation.

When a choice response is needed, use a structured choice control when it can faithfully represent the supported options; otherwise use a numbered list. Ask the user to choose by number or name or to propose an adjusted boundary. Emit no fenced goal prompt. If no substantial candidate remains that is implementable now or could pass readiness through bounded delegation or planning repair, say that `/goal` is not warranted and stop instead of manufacturing an option. Keep substantial candidates with planning gaps in the choice set so the readiness gate can handle them after selection; do not offer work that still depends on external authorization or an unresolved external blocker.

Stop here only when awaiting an unresolved scope choice or honoring a request to present options first. Otherwise continue in the same turn. A delivery variant alone does not select a scope when materially different boundaries remain unresolved.

## 3. Finalize the Selected Goal Boundary

Revalidate the selected scope against the current repository state before finalizing it. If material repository changes invalidate the boundary, resolve the scope again using applicable delegated authority; ask when that authority cannot resolve the changed choice. Otherwise preserve the user's chosen boundary even when it differs from the recommendation.

Apply these boundary rules both when framing options and when finalizing the selected choice. Naturally connected work may be absorbed into an option, and a selected adjusted boundary may be refined only enough to make it semantically closed. A goal may span subsystems and multiple coherent commits when they lead to one meaningful project state.

When both delivery variants are requested, keep their selected boundary identical. In the PR variant, treat PRs as delivery checkpoints within the large goal rather than separate `/goal` boundaries.

The selected boundary must:

- be materially larger than work suited to one ordinary interactive turn;
- reach a concrete, demonstrable project or user outcome rather than only a prerequisite or internal seam;
- contain enough settled or explicitly delegated work to benefit from persistent execution across multiple checkpoints;
- stop at a consequential unresolved decision the readiness gate does not settle or delegate, external authorization, blocker, or materially unrelated next milestone—not merely at the next plan heading or reviewable slice.

Do not expand a small selected scope beyond applicable user direction or delegated selection authority. If the selected choice is not substantial enough for `/goal`, explain that and return to scope choice rather than silently replacing it. A substantial selected scope whose planning is incomplete proceeds to the gate.

Before the readiness gate, reduce the selected scope to this closed routing envelope:

- **Outcome** — the semantic project or user result in one sentence.
- **Goal state** — one concrete durable path: `goals/<stable-slug>.md` for ordinary work or `goals/<scope>/<stable-slug>.md` for an isolated worktree planning scope.
- **Included results and sources** — every authorized result as a short, stable semantic label paired with the few authoritative documents that supply its implementation and acceptance detail. Labels define membership; paths and queue positions do not.
- **Completion** — one compact predicate requiring each named result to achieve its cited outcome and any applicable completion criteria, plus only cross-cutting validation, review, freshness, and delivery conditions not already carried by those sources.
- **Excluded work** — exactly the immediate next out-of-scope milestone plus exclusions stated directly by the user. Authority supplies the complete boundary for every later, unrelated, or merely plan-documented item.
- **Authority** — allow only the smallest bounded work necessary for an included result. When the readiness gate passed by delegation, also authorize the goal-running agent to resolve remaining decisions within the closed outcome using its best judgment. Record anything outside the boundary for later and require explicit user direction for expansion or external actions not covered by the selected contract and delivery lifecycle.
- **Resume invariant** — at every resumed turn, automatic continuation,
  compaction recovery, or handoff, invoke `$progress` in goal mode and recover
  the named goal state before selecting or starting more work.
- **Delivery** — the selected lifecycle and its skill routing, kept inside the contract so recovery preserves it.

Record the exact planning gap when evidence cannot yet support a field; do not invent closure. Boundary finalization is complete when the selected scope, its provisional envelope, and every unsupported field are ready for the readiness gate.

Treat goal membership as closed. Advancing a roadmap, changing `Current`, creating a plan, opening a branch or PR, or discovering review findings never adds work to the goal. Project planning state describes what the project should do next; the goal contract alone describes what this run is authorized to do.

## 4. Pass the Readiness Gate

Before emitting a goal prompt, verify that the evidence supports a closed outcome, semantic included results and authoritative sources, a completion predicate, and enough settled direction for autonomous implementation. Reversible implementer-owned choices may remain open. Consequential user-owned decisions require explicit delegation or planning repair.

When the outcome is closed but consequential decisions remain, first check
whether the user already delegated those decisions or selected planning repair.
Apply that choice within its stated scope. Otherwise explain the specific
planning gaps and ask them to choose:

1. authorize the goal-running agent to resolve the remaining decisions within the closed outcome using its best judgment; or
2. repair planning first with `$interview` followed by `$progress`, then rerun goal selection.

Emit no goal prompt until an applicable user choice resolves the gate.
Delegation passes the gate only for decisions inside the supported outcome; it
never adds results, expands scope, or grants external authority. Record that
delegation in the generated contract's `Authority` field so the goal-running
agent does not ask again merely because the cited plans left those decisions open.

If the user chooses planning repair, treat that answer as an explicit request for the separate mutating phase: use `$interview` to settle consequential decisions, then `$progress` to update the authoritative planning documents. Restart current-state resolution from the resulting repository state and keep the renewed selection phase read-only. Planning repair does not erase an already selected scope; revalidate and retain it unless the new evidence materially changes its boundary, in which case resolve the scope again under step 2. When no closed outcome can be supported, explain why delegation is unavailable and ask to repair planning before goal selection.

The gate is complete only when planning is sufficient or the user has explicitly delegated the remaining decisions within a supported closed outcome. Finalize every routing-envelope field after it passes.

## 5. Route the Result

- When `/goal` is **not warranted**, give the evidence-based reason and omit the prompt. Do not read the delivery-variant instructions.
- When the scope **remains unresolved** or the user requested options first, return only the compact choice set, recommendation, and selection question from step 2. Do not read the delivery-variant instructions or emit a goal prompt.
- Once the scope is selected, the evidence establishes that `/goal` **is warranted**, and the readiness gate passes, read and follow [prompts/delivery-variants.md](prompts/delivery-variants.md).
- In `spawn` mode, pass one selected delivery contract to the spawn workflow instead of returning the prompt-only response below. Asking for both delivery variants does not authorize two tasks; resolve which lifecycle to execute before creating one.
- By default, return only the recommended prompt as one unlabeled `text` fenced block. Put only the body to enter after `/goal` inside it, with no prose before or after the fence.
- Honor an explicit request for one named delivery variant even when it differs from the evidence-based recommendation; identify the emitted variant through its `Delivery` field.
- When the user explicitly requests both variants, return only the two `text` fenced prompt blocks, identify the variant inside each prompt's `Delivery` field, put the same closed scope contract inside both prompt bodies, and vary only the delivery mechanics and `Delivery` field.

Before responding, verify that scope discovery and selection stayed read-only, any preparation stayed within its authorized phase, and only a requested spawn created a task. The selection session never creates its own native goal or durable goal file. For prompt output, verify readiness and return only the requested fenced contract; for spawn, report the actual startup state under the spawn workflow.

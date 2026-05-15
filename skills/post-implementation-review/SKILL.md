---
name: post-implementation-review
description: Review already-implemented code for design, abstraction, ownership, structure, and maintainability issues; automatically apply straightforward recommended fixes and design-improving refactors when the best path is clear. Use only when the user explicitly asks for a post-implementation review, asks whether recent implementation work revealed design or structure problems, or asks whether implemented code should now be refactored. Ask before changing only when tradeoffs require developer taste, domain judgment, risk acceptance, or other decision-making.
---

# Post-Implementation Review

Use this skill manually. Do not invoke it unless the user explicitly asks for a post-implementation review or explicitly asks whether implemented code should now be refactored.

Review the code after the implementation exists. Focus on issues that were hard to see upfront and only became obvious once the change touched real interfaces, control flow, state, tests, docs, or module boundaries.

This is a review-and-improve skill, not a passive advisory report. By default, verify findings against the real code, apply supported fixes/refactors that are safe within the requested scope, validate them, and report what changed. If the user explicitly asks for review-only feedback, do not edit.

It is acceptable for the review to conclude that the implementation is fine. Do not invent a flaw just to produce feedback. If the code does not reveal a meaningful improvement, say plainly that there is nothing worth changing right now.

## Contract

- Treat each finding as advisory until verified against the real code path, nearby interfaces, tests, and docs.
- Do not blindly apply generated or previous review output.
- Reject weak, speculative, unrealistic, or overcomplicated recommendations.
- Prefer root-cause design, ownership, boundary, structure, or abstraction fixes over tactical bandages.
- Automatically apply Bucket I findings when they are straightforward, materially worthwhile, and safe within the user's requested scope.
- For Bucket II, prefer the option that improves overall code quality, design, ownership, boundaries, and long-term maintainability.
- Implement Bucket II refactors when the design-improving path is clear enough and does not require developer taste, domain judgment, risk acceptance, rollout choice, or other decision-making.
- Ask before implementing Bucket II only when the tradeoff genuinely requires the user's review or decision.
- After every code change, rerun focused validation when the repository provides an applicable command.
- Do not push, commit, or install globally unless the user explicitly asks.

## Root-Cause Review Lens

Treat each finding as evidence about code design or quality, not as a prompt to find the smallest edit that quiets the symptom.

- First ask what the finding implies about ownership, boundaries, data shape, control flow, module layout, or future change cost.
- Prefer the smallest refactor that actually improves the codebase's shape, not the smallest diff.
- Recommend and apply the root-cause refactor when a symptom points to misplaced ownership, an awkward interface, scattered responsibilities, repeated translation, or the wrong module shape.
- Broad refactors are acceptable when they materially improve the codebase and the path is clear. Surface cost and risk; ask only when those tradeoffs need a human decision.
- Avoid tactical bandages: tiny helper extractions, wrapper patches, local cleanup, or isolated polish that hides a symptom while leaving the design weakness intact.
- If a small helper or extraction is the right fix, make clear what boundary, invariant, or ownership clarity it creates. If it only saves a few lines, skip it.
- Do not overcorrect. Do not refactor only because another design looks cleaner in theory, and do not treat unfamiliar code as broken code.

## Workflow

1. Anchor the review in the actual change.
   - Read the changed files, nearby interfaces, and affected tests or docs.
   - Use `git status`, `git diff`, recent commits, or the user-provided scope as appropriate.
   - Distinguish between pre-existing design debt, issues introduced by the change, and issues the change merely made easier to see.
   - Identify focused validation commands from the repository's existing workflow.

2. Look for post-implementation signals.
   - Repeated branching, mapping, or glue code that suggests the abstraction boundary is off.
   - A module now knows details it should not need to know.
   - The change forced edits across too many files for one concept.
   - Responsibilities are blurred enough that naming, ownership, or call flow became awkward.
   - Data is being reshaped repeatedly between layers.
   - Tests became hard to set up because dependencies or ownership are misplaced.
   - The implementation needed workarounds that will probably repeat.
   - The implementation introduced speculative or overbuilt structure that current behavior barely uses.
   - One concern is scattered across files or modules in a way that weakens discoverability.
   - The feature now feels too flat or mixed: readers must scan unrelated files or responsibilities to follow one concern.

3. Verify and classify each meaningful finding.
   - `Bucket I — Straightforward / Recommended`: clear fix path, low decision risk, materially worthwhile, and within scope. Apply automatically.
   - `Bucket II — Design choice / Tradeoff`: real tradeoffs, broader cost, multiple credible designs, hidden constraints, or meaningful technical decisions. Prefer the design-improving option; implement it if the best path is clear and decision risk is low enough.
   - `Keep As-Is`: plausible concern, but evidence is weak, current structure earns its keep, or the change would be overcorrection.

4. Act.
   - Apply Bucket I fixes automatically.
   - For Bucket II, choose the option that best improves code quality and design unless refactor cost, new risk, uncertainty, or product/domain constraints make that irresponsible.
   - Implement Bucket II only when you can justify the choice without needing the user's taste or judgment.
   - Ask for permission when choosing among credible options depends on developer preference, API style, rollout risk, compatibility tolerance, product/domain constraints, or architectural direction.
   - If a temporary/local patch is chosen over a stronger design fix, state why the stronger refactor is not worth doing now.

5. Validate and report.
   - Rerun focused tests, typechecks, formatters, or other repo-standard checks after changes.
   - If validation fails because of your changes, fix it or report the blocker clearly.
   - If validation fails for unrelated/pre-existing reasons, record that distinction.
   - Report applied changes, remaining decisions, and keep-as-is findings.

## Bucket I — Straightforward / Recommended

Put a finding here only when all are true:

- The evidence is concrete in the implemented code.
- The improvement is materially worthwhile for correctness, ownership, boundaries, organization, or future change cost.
- The root-cause fix path is clear enough to implement without a user decision.
- The change is inside the user's requested scope.

For execution, Bucket I means: do it, validate it, then report it.

Avoid Bucket I for tiny helper extractions, one-off naming polish, isolated dedupe, or logging niceties unless they reflect a broader boundary or ownership issue.

## Bucket II — Design Choice / Tradeoff

Put a finding here when any are true:

- Multiple credible designs exist.
- The refactor improves design but has meaningful cost, churn, compatibility, rollout, or risk implications.
- The correct choice may depend on product, domain, architectural direction, or developer taste.
- The evidence is real but not yet strong enough to justify a broad change automatically.
- The issue may be pre-existing debt outside the requested scope.

For execution, Bucket II does not automatically mean "ask first." It means reason carefully:

- Prefer the option that improves the codebase's design, ownership, boundaries, and long-term maintainability.
- Implement that option when the path is clear, the risk is acceptable, and the decision does not require the user's judgment.
- Ask before acting when the tradeoff requires taste, architectural direction, product/domain knowledge, rollout strategy, compatibility tolerance, or risk acceptance.
- Recommend a temporary/local fix only when the broader refactor cost is disproportionate, the evidence is not strong enough, the change would introduce meaningful new risk, or the decision depends on constraints you cannot verify.

When Bucket II remains unimplemented, include a recommended action:

- `Discuss before changing`: user/domain decision needed.
- `Defer`: real concern, but not worth changing until the next related feature or until stronger evidence justifies the refactor.
- `Keep as-is for now`: current shape is acceptable; watch for repetition.
- `Prototype separately`: validate a broader refactor before committing when the design improvement is promising but risk or uncertainty is meaningful.
- `Implement next if approved`: best design-improving path is clear enough, but requires explicit approval because cost/risk/scope is meaningful.

State the reason for the recommendation, not just the available options. If recommending a temporary fix, explicitly say why the stronger design improvement is not worth doing now.

## Review Standard

Treat these as strong signals for action:

- The same workaround or translation layer now appears in multiple places.
- A single change required cross-cutting edits that are likely to recur.
- One abstraction claims to hide complexity but instead leaks it to callers.
- Ownership of data, side effects, or orchestration is ambiguous enough to create likely bugs.
- The new code had to special-case around an existing interface in a way that will spread.
- The implementation added wrappers, options, strategy points, generic types, or configuration that current behavior does not really need.
- One concern is split across files or layers in a way that makes the code harder to navigate or evolve.
- A module, file, or directory now mixes responsibilities that should probably be separated.
- The current shape is workable, but a concrete refactor would noticeably reduce mental overhead or repeated explanation for future changes.

Treat these as weak signals that often do not justify action on their own:

- The code feels a bit inelegant but is still locally understandable.
- A helper or abstraction is slightly misnamed but still usable.
- There is some duplication, but it is small and isolated.
- The current design is awkward only for one edge case with no sign of repetition.
- A tiny helper extraction would save a few lines but would not clarify ownership, boundaries, or future change points.
- A one-off logging, annotation, or convenience cleanup would make the code nicer but would not materially change debugging leverage or design clarity.

If the evidence is weak, say so. If the only available feedback is minor polish, prefer `No meaningful improvement identified` over padding the review. A clean review is valid.

## Structure and Complexity Concerns

When a post-implementation review raises questions about speculative structure, unnecessary complexity, over-engineering, flat organization, or mixed responsibilities, review and act on those concerns inline here.

Check structure in both directions:

- Over-structure: extra fields, config keys, DTO properties, wrappers, extension points, strategy objects, mode switches, or mapping layers that current behavior does not meaningfully use.
- Under-structure: flat directories mixing unrelated concerns, modules handling multiple capabilities at once, or related files scattered in ways that make one feature harder to follow.

Use the same posture:

- Do not remove structure that already improves naming, local reasoning, ownership boundaries, or filesystem-level comprehension.
- Do not add nesting just because a directory has many files.
- Treat readability-oriented helpers, small composed objects, and clear subdirectories as legitimate value even when they are single-use.
- Prioritize structural findings that expose weak ownership, mixed responsibilities, scattered concerns, overbuilt abstractions, or directories/modules that no longer help readers predict where code lives.
- Prefer reorganizations that address the real ownership or module-shape problem over wrapper/helper patches that leave scattered responsibilities in place.

## Output

Use this structure when reporting:

### Applied / Resolved

Number each item and include:

- what the implementation revealed
- the design or quality weakness it implied
- the root-cause fix/refactor applied
- why a smaller tactical patch would have been a bandage, if relevant
- validation evidence

If empty, say: `No automatic changes were applied.`

### Needs Decision / Bucket II

Number each remaining decision item and include:

- what the implementation revealed
- the design or quality weakness it implies
- main options or decision points, including broader refactor paths when credible
- `Recommended action: ...`
- tradeoffs, risks, or uncertainty
- why permission is needed before changing

If empty, say: `No unresolved Bucket II decisions remain.`

### Keep As-Is

Call out meaningful findings rejected after verification, plus why no change is recommended. If this section has multiple items, number them too.

### Validation

List validation commands run and results. If no validation was run, explain why.

### Verdict

End with one of:

- `No meaningful improvement identified`
- `Applied straightforward design improvements`
- `Applied improvements; decision needed for remaining tradeoff`
- `Decision needed before refactor`
- `Validation failure remains`

## Snippet Rules

- When citing existing code, prefer embedded snippets over bare file paths and line numbers.
- Put the source file path on the first line of each snippet.
- Match the comment style to the language when practical, for example `// src/service.ts` or `# backend/jobs/sync.py`.
- Keep snippets tight and evidence-focused: signatures, branches, translations, ownership boundaries, and repeated glue code are usually enough.
- Use multiple small snippets instead of one large excerpt when separate points need separate evidence.
- Use file and line references only when a snippet would add noise or the exact location itself is the point.

## Communication Rules

- Be direct and specific.
- Prefer file-level, seam-level, module-level, or feature-flow observations over vague architectural commentary.
- Do not force the user back into the editor just to follow the review.
- Default toward fixing actual design, abstraction, ownership, organization, structure, and overengineering weaknesses that materially affect maintainability or future change paths.
- Check not only for missing structure but also for structure that is heavier than the current behavior justifies, and simplify where worthwhile.
- Apply the Root-Cause Review Lens instead of optimizing for the smallest diff.
- Avoid padding the review with minor pre-fixes: tiny helper extraction, one-off naming polish, isolated local dedupe, or small logging niceties are usually not worth calling out unless they indicate a broader problem.
- Do not treat code churn alone as a reason to avoid a refactor.
- Surface real tradeoffs when they exist.
- Ask only when a tradeoff requires the user's taste, constraints, or decision-making.
- If there are no meaningful design flaws or worthwhile improvements, say so explicitly.
- Do not manufacture findings to make the review feel useful.

## Example Triggers

- "Now that this feature is implemented, do you see any design flaws we missed?"
- "Did this change reveal any abstraction problems?"
- "Should we refactor this now, or leave it alone?"
- "Review this implementation and tell me if the current design will age poorly."
- "Now that this is implemented, is the structure earning its keep or did we overbuild it?"
- "This feature works, but did the implementation reveal that the module layout is too flat or mixed?"

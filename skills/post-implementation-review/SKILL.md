---
name: post-implementation-review
description: Manually review already-implemented code for design flaws, abstraction issues, structural problems, or refactors that only became clear in real code. Use only when the user explicitly asks for a post-implementation review, explicitly asks whether recent implementation work revealed design or structure problems, or explicitly wants refactor recommendations after the code exists. Do not auto-trigger for ordinary implementation, debugging, explanation, or generic code review requests. Prefer embedded snippets with file-path comments over editor-oriented file and line references. Treat findings as signals about code shape and quality; prioritize root-cause design, ownership, abstraction, and organization improvements, including broad refactors when warranted, over bandage fixes such as tiny helper extractions or local polish.
---

# Post-Implementation Review

Use this skill manually. Do not invoke it unless the user explicitly asks for a post-implementation review or explicitly asks whether implemented code should now be refactored.

Review the code after the implementation exists. Focus on issues that were hard to see upfront and only became obvious once the change touched real interfaces, control flow, state, tests, or module boundaries.

This skill is for improvement-oriented design review after the real code exists. Prioritize actual design flaws, weak boundaries, ownership confusion, abstraction leaks, overengineered code or structure, and organization problems that will matter as the codebase grows. Prefer recommendations that materially improve a seam, file, module, or end-to-end concern once the code is real. If tradeoffs, uncertainty, or hidden constraints matter, surface them clearly.

It is still acceptable for the review to conclude that the implementation is fine. Do not invent a flaw just to produce feedback. If the code does not reveal a meaningful improvement, say plainly that there is nothing worth changing right now.

## Root-Cause Review Lens

Treat each finding as evidence about code design or quality, not as a prompt to find the smallest edit that quiets the symptom.

- First ask what the finding implies about ownership, boundaries, data shape, control flow, module layout, or future change cost.
- Recommend the root-cause refactor when a symptom points to misplaced ownership, an awkward interface, scattered responsibilities, repeated translation, or the wrong module shape.
- Broad refactors are acceptable when they materially improve the codebase. Describe the path, cost, and risk instead of downscoping the recommendation into a smaller bandage.
- Avoid tactical bandages: tiny helper extractions, wrapper patches, local cleanup, or isolated polish that hides a symptom while leaving the design weakness intact.
- If a small helper or extraction is the right recommendation, explain the boundary, invariant, or ownership clarity it creates. If it only saves a few lines, omit it.
- Do not overcorrect. Do not recommend a refactor only because another design looks cleaner in theory, and do not treat unfamiliar code as broken code.
- Optimize for the smallest refactor that actually improves the codebase's shape, not the smallest diff.

## Workflow

1. Anchor the review in the actual change.
- Read the changed files, nearby interfaces, and affected tests or docs.
- Distinguish between:
  - pre-existing design debt
  - issues introduced by the change
  - issues the change merely made easier to see

2. Look for post-implementation signals.
- Repeated branching, mapping, or glue code that suggests the abstraction boundary is off.
- A module now knows details it should not need to know.
- The change forced edits across too many files for one concept.
- Responsibilities are blurred enough that naming, ownership, or call flow became awkward.
- Data is being reshaped repeatedly between layers.
- Tests became hard to set up because dependencies or ownership are misplaced.
- The implementation needed workarounds that will probably repeat.
- The implementation introduced speculative or overbuilt structure that current behavior barely uses: extra fields, hooks, wrappers, options, strategy objects, generic layers, or config surfaces.
- One concern is now scattered across files or modules in a way that weakens discoverability.
- The feature now feels too flat or mixed: readers must scan unrelated files or responsibilities to follow one concern.
- The code technically works, but the implementation revealed a real boundary, ownership, organization, or overengineering weakness that will keep taxing future changes.

3. Apply the Root-Cause Review Lens.
- Assume there may be constraints you cannot yet see unless the code gives clear contrary evidence.
- Recommend a refactor when the implementation shows a solid improvement path for readability, boundaries, extension, or correctness, even if the current code still works.
- Bias toward improvements that change the design meaningfully at the seam, file, module, or feature-flow level.
- Prioritize actual design and organization weaknesses over local readability or polish wins.
- Do not reject a worthwhile recommendation just because it touches many files. Discuss that cost when it matters, but size alone is not a veto.
- If the evidence is weak, say so instead of padding the review.

4. Separate evidence, design implication, recommendation, and decision level.
- First state what the implementation revealed.
- Then state what design or quality weakness the finding implies.
- Then state the likely root-cause recommendation or no supported action.
- If a smaller tactical patch would be tempting but would leave the design weakness intact, name it as a bandage.
- Classify every finding or recommendation into one of two buckets:
  - `Bucket I — Straightforward / recommended`: likely has a clear answer or strongly recommended fix path; present briefly for awareness.
  - `Bucket II — Worth discussing`: real tradeoffs, multiple plausible directions, unclear constraints, or technical decisions.
- Use Bucket II for the items that deserve the user's attention.

## Review Standard

Treat these as strong signals for a refactor recommendation:

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
- A one-off logging, annotation, or convenience cleanup would make the code a bit nicer but would not materially change debugging leverage or design clarity.
- The issue is purely inside one function and does not suggest a broader responsibility or interface problem.

If the evidence is weak, say so. If the only available feedback is minor polish, prefer `No meaningful improvement identified` over padding the review. A clean review is still valid.

## Output

Use this structure when reporting:

### Bucket I — Straightforward / Recommended
- Put findings or recommendations here when there is a likely answer or a strongly recommended path that probably does not need much user review.
- Keep these concise.
- Use Bucket I for materially worthwhile fixes, not for every plausible tidy-up.
- Number the items in this bucket (`1.`, `2.`, `3.`). Do not use unordered bullets for the primary items.
- For each numbered item, include:
  - what the implementation revealed
  - the design or quality weakness it implies
  - why it became visible after implementation
  - the recommended root-cause refactor or design direction in short form, not just the smallest edit
  - why a smaller tactical patch would be a bandage, but only when such a patch is a realistic temptation
  - the main cost or tradeoff only if it matters
- Support with a short embedded snippet when existing code is central to the point.

### Bucket II — Worth Discussing
- Put findings or recommendations here when they involve real tradeoffs, technical decisions, hidden constraints, or multiple plausible directions.
- Number the items in this bucket (`1.`, `2.`, `3.`). Do not use unordered bullets for the primary items.
- For each numbered item, include:
  - what the implementation revealed
  - the design or quality weakness it implies
  - why it became visible after implementation
  - the main options or decision points, including broader refactor paths when credible
  - which smaller patch would be a bandage, but only when such a patch is a realistic temptation
  - the tradeoffs, risks, or uncertainty
- Support with a short embedded snippet when existing code is central to the point.

### Keep As-Is
- Call out odd-looking choices that should probably remain unchanged for now.
- If this section has multiple items, number them too.
- Explain what evidence is missing for a change recommendation.

### Verdict
- End with one of:
  - `No meaningful improvement identified`
  - `Recommended design refactors, mostly straightforward`
  - `Recommended changes include technical decisions`
  - `Broader redesign may be warranted after clarifying constraints`
- If Bucket I or Bucket II is empty, say so explicitly instead of implying hidden findings.

## Snippet Rules

- When citing existing code, prefer embedded snippets over bare file paths and line numbers.
- Put the source file path on the first line of each snippet as a comment.
- Match the comment style to the language when practical, for example `// src/service.ts` or `# backend/jobs/sync.py`.
- Keep snippets tight and evidence-focused: signatures, branches, translations, ownership boundaries, and repeated glue code are usually enough.
- Use multiple small snippets instead of one large excerpt when separate points need separate evidence.
- Use file and line references only when a snippet would add noise or the exact location itself is the point.

## Communication Rules

- Be direct and specific.
- Prefer file-level, seam-level, module-level, or feature-flow observations over vague architectural commentary.
- Do not force the user back into the editor just to follow the review.
- Default toward surfacing actual design, abstraction, ownership, organization, structure, and overengineering weaknesses that materially affect maintainability or future change paths.
- Check not only for missing structure but also for structure that is heavier than the current behavior justifies, and call out where it can be reduced.
- Apply the Root-Cause Review Lens instead of optimizing for the smallest diff.
- Avoid padding the review with minor pre-fixes: tiny helper extraction, one-off naming polish, isolated local dedupe, or small logging niceties are usually not worth calling out unless they indicate a broader problem.
- Do not treat code churn alone as a reason to avoid recommending a refactor.
- Surface real tradeoffs when they exist.
- Keep Bucket I concise so the user can scan past the obvious items quickly.
- Put the user's attention on Bucket II.
- Within each populated bucket, use numbered primary items so the hierarchy stays visually clear.
- If there are no meaningful design flaws or worthwhile improvements, say so explicitly.
- If a concern is plausible but not yet well supported, put it in Bucket II or `Keep As-Is` rather than overstating confidence.
- Do not manufacture findings to make the review feel useful.

## Structure and Complexity Concerns

When a post-implementation review also raises questions about speculative structure, unnecessary complexity, over-engineering, flat organization, or mixed responsibilities, review those concerns inline here instead of delegating elsewhere.

Check structure in both directions:
- Over-structure: extra fields, config keys, DTO properties, wrappers, extension points, strategy objects, mode switches, or mapping layers that current behavior does not meaningfully use.
- Under-structure: flat directories mixing unrelated concerns, modules handling multiple capabilities at once, or related files scattered in ways that make one feature harder to follow.

Use the same posture here:
- Do not remove structure that already improves naming, local reasoning, ownership boundaries, or filesystem-level comprehension.
- Do not add nesting just because a directory has many files.
- Treat readability-oriented helpers, small composed objects, and clear subdirectories as legitimate value even when they are single-use.
- Prioritize structural findings that expose weak ownership, mixed responsibilities, scattered concerns, overbuilt abstractions, or directories/modules that no longer help readers predict where code lives.
- Recommend simplification or reorganization when the current shape adds recurring maintenance cost, obscures the real flow, or when a credible cleanup would materially improve ownership clarity.
- Prefer reorganizations that address the real ownership or module-shape problem over wrapper/helper patches that leave scattered responsibilities in place.
- Do not reject a structural recommendation just because the reorganization would touch many files; surface the cost and tradeoff instead.

Treat these as strong signals:
- A new field, option, or hook exists mostly for hypothetical future cases.
- A wrapper, abstraction layer, strategy object, or config surface adds flexibility that current behavior does not use.
- A helper or abstraction adds another jump without improving naming, boundaries, invariants, or reuse.
- Callers still need to know internal details the abstraction claimed to hide.
- One concept now requires repeated mapping or translation just to preserve genericity.
- The filesystem or module layout no longer helps a reader predict where a concern lives.
- A small or medium reorganization would make one concern much easier to follow end to end.

Treat these as weak signals:
- The code is a bit inelegant but still locally understandable.
- A helper is slightly misnamed but still usable.
- There is some duplication, but it is small and isolated.
- A flat directory is still small and each file name clearly signals its purpose.
- A tiny extraction or move would polish one function but would not improve the surrounding ownership or structure.

When these concerns matter, keep using the same report structure:
- Put straightforward structural fixes in `Bucket I — Straightforward / Recommended`.
- Put structural items with meaningful tradeoffs or unclear boundaries in `Bucket II — Worth Discussing`.
- Use `Keep As-Is` for structure that pays for itself now.
- Skip micro-structure polish unless it reflects a broader boundary problem.
- If the implementation is structurally fine, say so plainly.

## Example Triggers

- "Now that this feature is implemented, do you see any design flaws we missed?"
- "Did this change reveal any abstraction problems?"
- "Should we refactor this now, or leave it alone?"
- "Review this implementation and tell me if the current design will age poorly."
- "Now that this is implemented, is the structure earning its keep or did we overbuild it?"
- "This feature works, but did the implementation reveal that the module layout is too flat or mixed?"

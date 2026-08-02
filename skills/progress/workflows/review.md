# Review Planned Work

Use this branch as a conversational approval gate for LLM-authored roadmaps and
plans before product implementation begins. The planning documents describe a
proposal; they do not independently prove that the proposal matches the user's
intent.

## Resolve the Proposal

1. Treat a description or sentence supplied after `review`, `discuss`, or an
   equivalent request as a semantic target. Match it against roadmap labels,
   item outcomes, and planning contents; do not require the user to know a file
   name. If materially different candidates match, show the small candidate
   set and ask which one the user means before reviewing details.
2. When the user gives no narrower target, review `Current` and the first
   queued plan in the context of the selected roadmap. If neither exists,
   report that there is no planned implementation to review. Review the whole
   roadmap's strategy or ordering only when the request describes that broader
   scope.
3. Read the selected roadmap and review targets completely. Read the `Outcome`
   and `Next action` of remaining scheduled plans, plus relevant tasks, only far
   enough to give each an accurate one-line preview. Inspect requirements,
   decisions, relevant implementation evidence, and recent history only far
   enough to test material claims and feasibility. Keep document claims,
   repository evidence, and direct user intent distinct.
4. If goal mode is explicitly selected, recover the goal contract and compare
   the proposal with its included results, exclusions, authority, and
   completion conditions. A review may recommend an amendment but cannot
   silently expand the contract. A planning review outside goal mode does not
   activate or advance an existing goal merely because the reviewed item is in
   the project queue.

## Discuss Alignment

Begin with a compact preview that lets the user understand the future without
reading the source files:

- **Future roadmap:** summarize `Current` and upcoming scheduled plans in
  order, keeping unrelated items to one line each. Identify relevant tasks as
  unordered rather than implying that their list position schedules them.
- **Proposed implementation:** explain what the selected work would make true,
  the main implementation shape, and how success would be checked. State when
  the plan does not yet contain enough detail to support one of these claims.

Then compare the proposal with the user's description, direct instructions,
requirements, and verified repository constraints. Report only consequential
mismatches, omitted requirements, assumptions, tradeoffs, feasibility risks,
or unnecessary complexity. Separate conflicts with user intent from technical
quality concerns so the user can judge them independently.

Give a concise intent readback and ask the smallest set of load-bearing
questions needed to confirm or correct it. Prefer concrete alternatives when
the user is more likely to recognize the right behavior than to invent it from
scratch. Keep the repository and external systems unchanged throughout this
discussion phase.

## Record an Approved Direction

Treat a clear instruction to approve the proposal, apply the agreed changes,
update the plan, or an equivalent imperative as authorization to record the
agreed direction. Merely mentioning words such as `approve` or `apply`, or
giving a generic acknowledgement, does not approve edits. When the invocation
already contains an explicit approval and a complete direction, record it
without reopening settled questions; give the brief roadmap and implementation
summary in the completion report.

1. Update only the reviewed roadmap, plans, tasks, or explicit goal amendment
   needed to make the approved direction durable. Preserve the planning model's
   sources of truth: item files own outcomes and next actions; the roadmap owns
   order and selection.
2. Record durable user decisions, corrected scope, acceptance conditions,
   material non-goals, and the next concrete implementation action. Remove or
   revise superseded proposal text instead of appending a conversation log.
3. If the approved direction changes roadmap order or membership, apply the
   queue invariants from `SKILL.md`. When an approved priority change interrupts
   `Current`, return that item to the approved plan position and leave
   `Current` empty; do not promote the newly first plan before implementation
   begins. In goal mode, change the contract only through an explicit
   authorized amendment.
4. Re-read the touched planning files and verify that they express the agreed
   intent without unresolved contradictions. When the proposal already aligns,
   report approval without manufacturing a documentation change.

Stop after recording and validating the approved planning direction. Do not
start implementation, move an item into `Current`, create an implementation
branch, or perform delivery actions in this workflow. Report the approved
direction, changed planning files, validation, and the implementation action
that remains unstarted.

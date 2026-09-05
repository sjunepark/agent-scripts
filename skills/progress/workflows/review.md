# Brief Planned Work

Use this branch to brief the user on what the planning system says will be
implemented next and how the relevant plan proposes to do it. The primary
output is a compact, faithful explanation that lets the user judge whether the
proposal matches their intent without reading the planning files. Review notes
and approval are secondary: surface consequential issues after the briefing,
and change planning documents only after explicit approval.

## Resolve the Briefing Target

1. Treat a description or sentence supplied after `review`, `discuss`, or an
   equivalent request as a semantic target. Match it against roadmap labels,
   item outcomes, and planning contents; do not require the user to know a file
   name. If materially different candidates match, show the small candidate
   set and ask which one the user means before reading details.
2. When the user gives no narrower target, brief `Current` and the first
   queued plan in the context of the selected roadmap. If neither exists,
   report that there is no planned implementation to brief. Cover the whole
   roadmap's strategy or ordering only when the request describes that broader
   scope.
3. Read the selected roadmap and briefing targets completely. Read the `Outcome`
   and `Next action` of remaining scheduled plans, plus relevant tasks, only far
   enough to give each an accurate one-line preview. Inspect requirements,
   decisions, relevant implementation evidence, and recent history only far
   enough to explain or test material claims. Keep document claims, repository
   evidence, and direct user intent distinct.
4. If goal mode is explicitly selected, recover the goal contract and compare
   the proposal with its included results, exclusions, authority, and
   completion conditions. A review may recommend an amendment but cannot
   silently expand the contract. A planning review outside goal mode does not
   activate or advance an existing goal merely because the reviewed item is in
   the project queue.

## Apply an End-State Lens When Requested

When the user explicitly asks this review to reconstruct a coherent end state,
separate historical residue from obligations, or test whether a plan still
reflects settled requirements, read the `references/end-state-planning.md`
resource named in `SKILL.md` and apply it during the same read-only discussion.
This remains an optional audit, not a follow-up step required after ordinary
plan creation or revision; material plan writes already apply the reference
automatically through `SKILL.md`.

The target may be a plan, roadmap item, specification, migration, diff, or
in-progress implementation. State the final contract, classify consequential
residue and obligations with evidence, propose the smallest coherent shape,
map deleted assumptions and retained obligations to verification, and leave
load-bearing unknowns unresolved. For an implementation target, stop at
recommendations within this phase. Planning approval permits planning edits;
production code and tests belong to a separately classified implementation phase,
which may already be authorized in the same request.

## Deliver the Briefing

Lead with the planned future, not a critique or a request for decisions. Give
the user this compact briefing:

- **What happens next:** distinguish work already `Current` from the next
  queued plan, then summarize later scheduled plans in order. Keep unrelated
  items to one line each. Identify relevant tasks as unordered rather than
  implying that their list position schedules them.
- **How the plan is written:** explain what the selected work would make true,
  the user-visible behavior or operational result, the main implementation
  shape and sequence, important decisions and assumptions, and how success
  would be checked. State plainly when the plan does not contain enough detail
  to support one of these claims; do not fill gaps with an invented design.
- **Review notes:** after the briefing, list only consequential items the user
  should check or the plan should fix. Separate possible conflicts with known
  user intent from plan-quality concerns such as omitted requirements,
  ambiguous behavior, feasibility risks, unsupported claims, or unnecessary
  complexity. Omit this section when there is nothing material to flag.

For an explicit end-state audit, replace generic review notes with only the
useful parts of: **End state**, **Residue and obligations**, **Proposed shape**,
**Plan amendments**, **Verification**, **Decisions required**, and **Scope
boundary**. A conclusion that the target is already coherent is valid.

When the user supplied intent or requirements, relate review notes to that
evidence without replacing the briefing with the agent's own verdict. Ask only
when consequential intent remains unresolved or a proposed plan change needs
approval; prefer concrete alternatives. A complete review-only briefing is a
finished deliverable and does not require readback confirmation.
Keep the repository and external systems unchanged throughout this discussion
phase.

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

Finish and validate the planning phase before implementation. If the user also
authorized implementation, redispatch through `SKILL.md` to the appropriate
execution workflow, including goal recovery and boundary checks when applicable.
Otherwise stop and report the approved direction, changed planning files,
validation, and the implementation action that remains unstarted.

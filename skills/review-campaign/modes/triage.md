# Triage Mode

This mode requires the user. Walk `triage + open` findings, blockers first, batched by area.

For each finding, obtain one outcome from the user:

- **accept** — export it with its id according to the repo profile and mark it `accepted`;
- **reject** — mark it `rejected` with a one-line reason;
- **retier to auto** — only when the user says the fix is obvious and it satisfies every auto-tier eligibility rule.

Append every decision to `decisions.md`. Do not implement accepted findings in triage mode; the exported entry becomes the work item. When a phase column is complete, propose the milestone merge required by the profile's branch model.

Report decisions recorded, exported work-item locations, remaining open triage counts, blockers, and the next recommended mode.

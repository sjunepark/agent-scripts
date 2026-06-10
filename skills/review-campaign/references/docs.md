# Pass: docs

Docs are read by agents as much as humans here — a stale claim misleads every future session, in every worktree. Scope per row: area rows review their colocated docs (README, ARCHITECTURE.md, AGENTS.md); the `docs` row reviews the `docs/` tree and root-level documents.

## Check

- **Claims vs code.** Commands exist in package.json and run; paths exist; described flows match the implementation; code blocks compile/parse. Spot-check by executing the cheap ones.
- **Architecture drift.** Root and per-app ARCHITECTURE.md match the current shape: module lists, data flow, invariants. After structure-pass refactors land, these are the first casualties — re-check them in stale-cell reviews.
- **Agent-instruction effectiveness.** AGENTS.md / CLAUDE.md rules are repo-specific, actionable, and current; routing points to files that exist; no generic advice an agent could not act on; nothing an agent repeatedly gets wrong is left unstated, and nothing stated is obsolete.
- **Single source of truth.** The same fact told in two places diverges eventually — pick the owner, link from the other. Flag pairs already diverged as evidence.
- **Lifecycle hygiene.** Current-state docs describe reality; plan and TODO docs carry no completed items (verify against git history before calling done); target/vision docs are dated or owned enough to tell aspiration from abandonment. Decision records are historical — verify they aren't silently contradicted by code, but do not "update" the records themselves.
- **Missing docs — narrowly.** Flag only where onboarding demonstrably breaks (a subsystem whose entry point is undiscoverable). Do not prescribe docs for everything.

## Do not flag

- Tone/style preferences; brevity is the house style.
- Historical records (decisions, research) for being outdated — they are records.
- Absence of docs where code and naming already carry the intent.

## Severity and tier

Wrong command/path an agent will follow: major. Architecture doc contradicting reality: major. Diverged duplicate facts: minor until one is wrong. Auto tier: broken links/paths/commands (verified, then fixed) and removing completed TODO items with the landing commit cited. Validate edits with the repo's markdown lint and link-check commands when present.

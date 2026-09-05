# Instruction audit evaluation — 2026-09-05

Local source revision: all 31 published skills audited; 18 changed and 13
retained. Shared Codex, Claude, and Pi instruction sources and the repository's
`AGENTS.md` were updated. Activation policies and package layout were preserved.
This is bounded evidence for instruction corrections, not an Astra-versus-Sol
benchmark or a claim that every workflow was executed.

## Basis and scope

OpenAI's [Astra guidance](https://developers.openai.com/api/docs/guides/latest-model?model=gpt-6-astra)
recommends auditing instruction interactions, expressing existing authority,
preparing authorized work before approval, and calibrating delegation, output,
and verification. The reviewed guidance prescribes no new skill schema. The
[skill documentation](https://learn.chatgpt.com/docs/build-skills) and
[AGENTS.md documentation](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
remain the structural references. Converting unconditional judgment rules into
their actual decision conditions is this repository's application of that
guidance.

The baseline is the pre-edit working-tree snapshot, including unrelated
`sjskills` work, captured before this task's edits. Its Git base is
`bc9fac9266484380c8b75c6ca33f6ef1fb2a2e5f`. Raw outputs, case inputs, frozen
assertions, source hashes, and local Git verification are retained in
[the evaluation record](instruction-audit-2026-09-05.json). Temporary fixture
paths in that record identify the original run; the embedded evidence remains
available if those directories disappear.

## Catalog coverage

Audits accounted for each entry point, runtime resource, and adapter. Independent
read-only groups covered Git workflows, planning/delegation, review/authoring,
and operational/language skills; the coordinator revised shared owners.

| Skill | Disposition and reason |
| --- | --- |
| address-pr-feedback | Changed: separate feedback-only stacks from authorized staging and landing. |
| agents-md-writer | Retained: already distinguishes source, scope, precedence, evidence, and justified no-op audits. |
| brainstorming | Retained: explicit request and independent, bounded perspectives are intrinsic to the workflow. |
| clarify | Changed: settled answers can lead into the authorized task without blanket readback approval; adapter aligned. |
| clear-rust | Retained: concrete language and safety guidance; no conflicting collaboration gate found. |
| code-review | Changed: omit empty reporting buckets while preserving all four review lenses. |
| codex-cleanup | Changed: retain applicable prior scope, run only approved action classes and prerequisites, and recognize audit-only completion. |
| create-pr | Changed: honor exact commit authority, prepare before unresolved decisions, and respect existing initial bot-review policy. |
| delegate | Retained: its explicit Luna delegation, bounded contract, review, and iteration are the chosen capability. |
| delegate-ui-to-claude | Changed: use settled answers and unattended delegation before requesting another approval; retain Claude ownership and scope boundaries. |
| develop-skills | Changed: add catalog instruction audit; preserve activation intent, declared tool prerequisites, and source/install/session distinctions; scale evaluation to changed decisions. |
| distill-response | Retained: its structured result is the explicitly requested artifact, not incidental output overhead. |
| explore-repo | Retained: read-only exploration, pinned evidence, and mutation boundaries remain appropriate. |
| harmonize-docs | Changed: delegate useful independent areas without requiring multiple agents for a small coherent scope. |
| interview | Changed: unambiguous answers settle decisions; retain a bounded blind-spot pass and ask only about unresolved substance; adapter aligned. |
| macos-storage-cleanup | Retained: separate itemized consent is deliberate and supported by prior unsafe-broadening evidence. |
| merge-branch | Changed: distinguish plan-only from combined requests, resume matching merges, preserve unrelated work/index state, and bound checks and reporting. |
| modern-go | Retained: release-aware lookup and language-specific decision boundaries are already scoped. |
| modern-rust | Retained: release-aware lookup and language-specific decision boundaries are already scoped. |
| next-goal | Changed: resolve readiness choices from prior instructions; preserve the deliberate initial scope-selection pause. |
| pdf-to-markdown | Changed: recognize existing exact overwrite/install authority while preserving output-path and conversion-quality gates. |
| progress | Changed: dispatch already authorized phases through the entry point, honor goal delivery authority, and remove mandatory review readback; evaluation cases aligned. |
| release-please-release | Changed: reuse matching operation-and-version confirmation, preserve drift pauses, keep generated-artifact review read-only, and scope verification. |
| review-campaign | Changed: use recorded choices and scope before asking; preserve fresh code verification before fixes. |
| sjskills | Retained: evidence-bound reconciliation and ownership gates are intentional; pre-existing edits preserved byte-for-byte. |
| skills-cli | Changed: listing/discovery finishes without escalating into install, publication, or reconciliation. |
| teach | Changed: match output to focused, whole-project, or change explanation; avoid mandatory irrelevant sections. |
| ui-lab | Retained: dev-only gallery scope, production-state reuse, and validation remain justified. |
| update-base-branch | Retained: its dirty-tree stop protects an actual branch switch; this differs from resuming an existing merge. |
| windows-cleanup | Changed: independent confirmed actions can proceed while a separate action awaits confirmation. |
| write-go-docs | Retained: language-specific comment rules and verification are already proportionate. |

## Checks and observations

- **Static:** `scripts/validate-skills` passed for all 31 skills and the registry;
  `bunx skills add ./skills --list` discovered all 31; `git diff --check` passed.
  A baseline comparison found no `allow_implicit_invocation` changes. Only
  `develop-skills` discovery wording changed; clarify/interview adapter edits
  update suggested prompts, not activation policy.
- **Authoring observation:** one fresh context per version edited the same
  synthetic two-skill catalog. Both fixed only the redundant commit gate,
  preserved both activation policies, and left the unaffected read-only skill
  unchanged. The revised authoring workflow accepted declared GitHub/`gh`
  prerequisites and correctly treated unchanged discovery as a static check.
  This small pair supports those decisions; it does not establish general
  authoring superiority.
- **Decision simulations:** one fresh suite context per version covered 16
  independent supplied snapshots. Both respected the expected authority and
  scope boundaries under the surrounding user instructions. These included
  operation/version drift, feedback-only stacks, review-only progress,
  independent cleanup approvals, exact PDF overwrite, and undelegated UI
  decisions. Cases within each suite did not have separate fresh contexts;
  these results are directional checks rather than full isolated behavior evals.
- **Observed correction:** both first suites still prescribed an unnecessary
  progress-review readback. Its owning clause was corrected against progress
  case 36. A separate fresh-context final-candidate probe completed the
  read-only briefing with no unnecessary question or implementation.
- **Real Git fixture:** with an existing matching merge and unrelated untracked
  notes, the baseline paused and left conflicts unresolved. The candidate
  resolved and staged both features, left the merge uncommitted, and preserved
  the notes, destination `HEAD`, `MERGE_HEAD`, and source ref. Assertions were
  checked against Git and filesystem state.
- **Additional Git boundary:** with unrelated notes already staged and a request
  to commit the integration, the final candidate resolved the features but
  stopped before committing. The notes' contents and index blob were unchanged.
  This checks the final full-index gate added during review.
- **Trigger probes:** three separately isolated catalog-selection simulations
  accepted explicit catalog audit and rejected model-advice-only and uninvoked
  catalog-audit prompts. All matched the frozen labels. These are not live
  client-loader tests or measurements of the entire trigger distribution.

## Review, documentation, and limits

One bounded code-review pass used independent authoring/shared-default and
skill-boundary reviewers with all four lenses. Review found a leftover
unconditional trigger-fixture requirement; it was scoped to new or changed
discovery. Local follow-up also made the unrelated staged-content gate explicit.
No other actionable finding remained after integrating the narrow corrections.

Scoped documentation harmonization kept `README.md` unchanged, updated the
affected UI approval and source/session claims in `docs/settings-sync.md`, and
left unrelated reconciler documentation and implementation untouched. Task
delivery status is owned by `plans/astra-instruction-audit.md`.

Live GitHub writes, releases, cleanup, Windows operations, PDF conversion, and
Claude UI execution were not run. Output and delegation changes received static
review; there is no measured token, latency, or delegation-frequency comparison.
The small samples do not establish cross-model or production reliability.

All three personal instruction symlinks still resolve to their intended updated
sources. Installed skill copies and the active session catalog were not refreshed
or claimed to match these sources. In particular, the previously observed
installed UI-delegation policy drift remains a rollout concern. No commit, push,
publication, installation, or real-home reconciliation was performed.

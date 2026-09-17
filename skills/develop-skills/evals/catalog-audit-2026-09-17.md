# Catalog skill audit — 2026-09-17

Reviewed the 31 published source skills other than clarify and interview. Changed
23 and deliberately retained 8. Their 163 runtime and adapter files are accounted
for in the [raw record](catalog-audit-2026-09-17.json), with frozen source hashes,
criteria, outputs, and grading. Baseline:
`f622047ce6b4c41889f2b083aab3f776e8ae6c30`.

The [requested OpenAI article](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra),
read on 2026-09-17, informed shorter selection descriptions, conditional resource
reading, and removal of unnecessary process. Operational invariants and explicit
local preferences remain authoritative. These are local design decisions, not a
claim that shorter instructions alone improve a model.

The user-linked `.agents/skills/develop-skills` was an older installed copy. This
audit used the rewritten `skills/develop-skills` source. Installed copies, registry
membership, scripts, adapters, and invocation policies were not changed.

## Material changes

- Review and documentation skills now state scope, evidence, and completion once,
  with detailed guidance loaded when relevant. Code review still considers all
  four perspectives; a routine edit need not read all four detailed lenses.
- Teaching, distillation, and delegation retain their distinctive outcomes while
  dropping repeated stage gates and fixed presentation recipes.
- Cleanup guidance narrows diagnostic/catalog reading without weakening data,
  process, symlink, or authority checks. PDF fallback mechanics have one owner:
  the bundled converter.
- Merge integration accounts for material behavior and deletions without a row
  for every mechanical commit, and explicitly handles already-integrated branches.
- sjskills status and project-selection detail now live in conditional references.
  Skills CLI guidance now matches the existing reconciler: trusted managed local
  edits are quarantined before replacement; unknown desired copies still block.

## Per-skill disposition

| Skill | Result and reason |
| --- | --- |
| [address-issues](../../address-issues/SKILL.md) | **Retained** — Recent outcome-driven lifecycle already preserves exact issue and queue authorization, current-head merge proof, independent-consumer validation, sequential task/model preferences and explicit-only policy. Its multiple workflow is a fragile coordination protocol; retained without cosmetic changes. |
| [address-pr-feedback](../../address-pr-feedback/SKILL.md) | **Changed** — Condensed entry point and replaced repetitive numbered substeps with collection, active-review gate, fixes and completion. Retained all feedback surfaces, reaction identities/meanings, waiting-before-assessment preference, one-trigger budget, pushed-before-reply and stack authority. Kept collector and stack resource intact; exact file-based reply argument replaces shell command substitution. |
| [agents-md-writer](../../agents-md-writer/SKILL.md) | **Changed** — replaced universal inventory/evidence stages with inspection of the affected instruction chain; retained tool-specific discovery verification when topology or precedence matters. |
| [brainstorming](../../brainstorming/SKILL.md) | **Changed** — Simplified repetitive phase recipes. Preserved >=3 distinct lenses, fresh analysis-only workers, inherited model/reasoning, and synthesis of every contribution; explicitly report unavailable fan-out. |
| [cleanup-branches](../../cleanup-branches/SKILL.md) | **Retained** — Already concise evidence-driven deletion contract. Retained local/remote independent exact-tip evidence, retention, worktree races, expected-OID/lease operations and observed absence verification. Historical eval fixtures are evidence, not runtime requirements. |
| [clear-rust](../../clear-rust/SKILL.md) | **Changed** — Removed redundant priority and execution stages; retained Rust design preferences, MSRV/public contracts, unsafe obligations, macro guidance and proportional checks. |
| [code-review](../../code-review/SKILL.md) | **Changed** — kept four review perspectives while making detailed lens reads conditional; condensed lenses and completion; reuse settled design decisions without repeating approval. Preserved review-only and safe-fix boundaries. |
| [codex-cleanup](../../codex-cleanup/SKILL.md) | **Changed** — Shortened description only. Lifecycle deletion, writer quiescence, WAL recovery and offline handoff are fragile safety obligations deserving retention. Script unchanged. |
| [create-pr](../../create-pr/SKILL.md) | **Changed** — Shortened description, declared prerequisites, replaced canned report sections and bot-choice prompt with review evidence and policy decision criteria. Ready-by-default, initial CodeRabbit policy, conditional bot guides, stack staging and promotion boundaries unchanged. All three resources retained as specific protocol guidance. |
| [delegate](../../delegate/SKILL.md) | **Changed** — Compressed framing/review duplication. Preserved narrow slices, Luna/max/no-history, verbatim original request and result-affecting context, one writer, independent review before dependent work, and complete parent accountability. |
| [delegate-ui-to-claude](../../delegate-ui-to-claude/SKILL.md) | **Retained** — Retained: exact provisioning/permission sequence, Claude-only UI ownership, cleanup authority, mediated product decisions, session persistence, and integration verification are concrete operational constraints. |
| [develop-skills](../../develop-skills/SKILL.md) | **Retained** — the recently rewritten entry point and all eight runtime resources already provide conditional routing, proportional evidence, activation preservation, and explicit authorized completion. |
| [distill-response](../../distill-response/SKILL.md) | **Changed** — Replaced repeated recovery/routing/completion recipes with source selection and outcome criteria; retained certainty, stable labels, sibling-map recovery, conditional visuals, standalone explanation, focused elaboration. |
| [explore-repo](../../explore-repo/SKILL.md) | **Changed** — Consolidated seven-step recipe into cache/ref and inspection/experiment outcomes. Retained ~/.repos layout, dirty-clone protection, isolated experiments, requested refs, cache retention and delegation model preference. Removed redundant generic searching advice and broad shell cleanup examples; existing commands cover actual fragile operations. |
| [harmonize-docs](../../harmonize-docs/SKILL.md) | **Changed** — consolidated repeated scope/state/rewrite gates into one contract; retained full versus scoped discovery, canonical owners, implementation/design/delivery separation, immutable history, and architecture guidance. |
| [macos-storage-cleanup](../../macos-storage-cleanup/SKILL.md) | **Changed** — Shortened description; catalog reads cover common semantics/refusal rules plus relevant entries; row identifiers conditional on multiple candidates. Authority/sync/privacy/recovery boundaries retained. |
| [merge-branch](../../merge-branch/SKILL.md) | **Changed** — Replaced exhaustive per-commit form and large audit output template with material-intent accounting and focused history inspection. Clarifies already-integrated no-op and manual transplant contradiction. Preserves read-only routes, partial/history authority, MERGE_HEAD and ancestry, source-ref boundary, source deletions, residual differences, staged-work isolation and uncommitted endpoints. |
| [modern-go](../../modern-go/SKILL.md) | **Retained** — Retained: concise version-contract guidance and conditional release radar already comply. Six references preserve domain knowledge and bounded coverage. |
| [modern-rust](../../modern-rust/SKILL.md) | **Retained** — Retained: compatibility distinctions and conditional index already comply. All 26 references preserve migration and patch cautions; no release facts updated. |
| [next-goal](../../next-goal/SKILL.md) | **Retained** — Retained: conditional reads and detailed selection/prepare/spawn boundaries, compact closed contract, readiness authority, committed-state checks and startup verification cover established regression boundaries. Astra-medium unchanged. |
| [pdf-to-markdown](../../pdf-to-markdown/SKILL.md) | **Changed** — Shortened description; removed duplicate raw Xberg commands now owned by wrapper; fallback explicitly reads command construction/validation. Overwrite/install authority and OCR/layout/page-integrity rules retained; script unchanged. |
| [progress](../../progress/SKILL.md) | **Retained** — Retained: schema-valid recovery, project/goal separation, read-only/planning/execution authority, coherent-plan writes, worktree ownership, and terminal delivery govern real failure modes rather than generic reasoning. |
| [release-please-release](../../release-please-release/SKILL.md) | **Changed** — Setup now implements requested setup through validation instead of allowing proposal-only premature completion. Ownership, generated-artifact rules, exact operation-and-version release authority, fragile automation workflow and SemVer rubric otherwise retained. |
| [review-campaign](../../review-campaign/SKILL.md) | **Changed** — removed arbitrary incident/behavior counts and blanket naming/logging advice from four rubrics; retained ledger schema, mode scope, phase order, commit policy, triage and security boundaries. |
| [sjskills](../../sjskills/SKILL.md) | **Changed** — moved status and manifest/authentication detail into two conditional references; retained configured sync authority, provenance/quarantine conflicts, published-source proof, digest-bound global apply, and exact restore scope. |
| [skills-cli](../../skills-cli/SKILL.md) | **Changed** — shortened discovery description and corrected stale managed-local-edit/authenticated-source rules against the reconciler; preserved explicit targets, inspection versus mutation, remote publication, and plan evidence. |
| [teach](../../teach/SKILL.md) | **Changed** — Simplified entrypoint and change/whole-app guides, removing duplicate output sections, fixed mechanism counts and mechanical cold-reader gate. Preserved privacy filtering before default reads, cold-reader context and example, stable learning maps, read-only teaching, conditional resources and ASCII/snippet preferences. |
| [ui-lab](../../ui-lab/SKILL.md) | **Changed** — Replaced seven-step build order and fixed initial scenario count with integration/completion outcomes. Retained fresh deterministic fixtures, real shared host, no live side effects, dev-only exclusion, smoke coverage and real DOM assertion. |
| [update-base-branch](../../update-base-branch/SKILL.md) | **Changed** — Shortened oversized discovery description with key positive outcome and near-miss exclusion. Runtime retained: dirty-worktree no fetch/switch, branch/upstream ambiguity, other-worktree ownership, ahead/diverged stop and exact fast-forward verification. |
| [windows-cleanup](../../windows-cleanup/SKILL.md) | **Changed** — Shortened description; targeted reference reads and symptom evidence; reused supplied context. Mutation classes, security limits, managed-device and backup gates retained. |
| [write-go-docs](../../write-go-docs/SKILL.md) | **Changed** — Removed continue-once gates; retained deliberate information-gain/deletion preference, generated-file protection, audit-only scope and Go doc syntax. Rendering checks target changed associations/links. |

## Evidence and limits

- Structural validator and registry validation passed; local list-only discovery
  found all 33 skills. Catalog JSON, direct resource routing, and whitespace
  checks passed. All runtime files match the evaluated candidate snapshot.
- One independent bounded review covered all 39 changed tracked files and two
  new references; no actionable findings remained. A scoped documentation check
  found affected README claims consistent and updated the progress owner.
- Paired decision simulations: candidate **27/27**, baseline **26/27** against
  the frozen criteria. The baseline failure was the stale Skills CLI rule about
  managed local edits. Existing cleanup, privacy, release, Git, and sync boundaries
  held in the covered cases.
- Selection simulations: **45/45** expected labels in each condition, covering
  explicit positives, uninvoked requests, and near misses for 15 changed descriptions.
  These classifications do not prove installed-client discovery.
- Initial documentation decision outlines did not prove artifact completeness.
  A separate pair of fresh workers revised a disposable runner-move fixture.
  Both retained the runtime flow, concrete entry paths, invariants, and payments
  target status; source code and unrelated payments docs stayed byte-identical.
- The tiny code-review case loaded five instruction files in the baseline and
  two in the candidate. Both produced a bounded clean-review decision. This is
  a case-level observation, not a latency or token measurement.

The primary author scored the visible outputs without blinding. Each main
condition used a fresh worker, but its cases shared that worker context. There
was one run per condition, no holdout or repeated-trial reliability estimate,
and serving-model/reasoning/harness versions were not independently confirmed.
The fixture exercised actual documentation edits; other cases simulated decisions
without live PR, merge, cleanup, release, sync, or PDF operations.

Across the 23 changed packages, entry-point whitespace-delimited words decreased
from 20,520 to 15,560 (about 24%). Across all reviewed runtime Markdown, including
retained references, the count decreased from 66,514 to 59,616 (about 10%). These
are text-size measurements, not token or quality scores.

## Delivery

The authorized local revision, validation, evaluation, review, and affected
documentation alignment were complete at evaluation time, before commit,
publication, or installation. Subsequent rollout does not change these evaluation
results; Git history and retained sjskills execution evidence own rollout status.

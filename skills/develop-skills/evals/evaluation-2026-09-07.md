# Instruction audit evaluation — 2026-09-07

Audited the published catalog and shared instructions against the official
[GPT-6 Astra guide](https://developers.openai.com/api/docs/guides/latest-model),
fetched on 2026-09-07. Baseline: clean tree at `e21995a`; see the
[earlier evaluation](evaluation-2026-09-05.md) for prior corrections.

Shared defaults consolidate response, follow-through, and approval rules while
adding task steering. Git workflow fixes retain draft, transplant, and dirty-tree
boundaries. Repository instructions, plugins, activation metadata, and registry
selections were retained. The illustrative settings model and intentional Luna
skill role require no API migration.

## Catalog coverage

Two independent read-only audits inspected every entry point, adapter, and
relevant runtime resource; the coordinator assessed proposed edits.

| Skill | Disposition and reason |
| --- | --- |
| address-pr-feedback | Retained: feedback, staging, and landing authority are separated. |
| agents-md-writer | Retained: scoped evidence, hierarchy, and smallest-owner edits. |
| brainstorming | Retained: requested independent perspectives are intrinsic to the workflow. |
| clarify | Retained: settled answers continue into authorized work. |
| clear-rust | Retained: concrete design and proportionate validation. |
| code-review | Retained: bounded lenses and meaningful reporting buckets. |
| codex-cleanup | Retained: prior exact authority and required lifecycle gates. |
| create-pr | Retained: prepares drafts and applies existing commit and review authority. |
| delegate | Retained: intentional Luna contract and independent parent review. |
| delegate-ui-to-claude | Retained: reuses decisions while preserving Claude ownership. |
| develop-skills | Retained runtime: existing instruction-audit workflow covers this work. |
| distill-response | Retained: requested explanation structure is the output itself. |
| explore-repo | Retained: broad or independent exploration already matches shared delegation policy. |
| harmonize-docs | Retained: scoped impact closure and proportional delegation. |
| interview | Retained: questions address unresolved consequential decisions. |
| macos-storage-cleanup | Retained: itemized post-preview consent is deliberate and previously evaluated. |
| merge-branch | Changed: finish authorized commits; preserve drafts, unrelated index state, and intentional transplant history. |
| modern-go | Retained: version-specific language guidance needs no prompting change. |
| modern-rust | Retained: compatibility and release lookup remain appropriately scoped. |
| next-goal | Retained: initial scope selection is an intentional workflow boundary. |
| pdf-to-markdown | Retained: reuses exact overwrite and installation authority. |
| progress | Retained: combined phases and established goal authority already continue. |
| release-please-release | Retained: matching operation/version approval and drift gates. |
| review-campaign | Retained: recorded choices apply; fixes require current evidence. |
| sjskills | Retained: provenance and exact-evidence rollout gates are intentional. |
| skills-cli | Retained: discovery and installation scopes are separate. |
| teach | Retained: explanations scale to the requested learning target. |
| ui-lab | Retained: bounded gallery scope and meaningful smoke checks. |
| update-base-branch | Changed: assess target/upstream read-only before a dirty-tree pause; keep mutation blocked. |
| windows-cleanup | Retained: independent confirmed actions can continue. |
| write-go-docs | Retained: focused documentation and relevant verification. |

## Validation

- Catalog/link validation, local-source discovery, Skill Creator validation for
  both changed skills, and whitespace checks passed. Activation policy is unchanged.
- Bounded review corrected ancestry verification for intentional transplants and
  aligned the dirty-tree evaluation with read-only assessment.
- A fresh worker exercised the Git workflows below in separate disposable
  repositories. The coordinator verified refs, index, content, and merge state.

| Request | Observed result |
| --- | --- |
| Merge source into dev and commit, preserving both features. | Created a two-parent merge commit; source ancestry verified, both features present, worktree clean, source ref unchanged. |
| Merge source into dev and leave it uncommitted. | Destination HEAD unchanged; MERGE_HEAD matches source; both features present. |
| Finish an existing merge with an unrelated user note already staged. | Reported the index isolation decision; HEAD, staged blobs, and note contents unchanged. |
| Move a dirty feature worktree to latest dev. | Identified dev and its non-origin upstream before reporting the untracked note; branch, refs, index, and note unchanged, with no fetch. |

Independent review found no actionable regression in the prose consolidation;
catalog, coverage, link, and whitespace checks passed. The preserved workflow
decisions did not warrant repeating the passing Git exercises.
Global writing and steering received static review, with no live mid-turn,
transplant, output-quality, or cross-model evaluation. The Git cases shared one
worker context and do not establish statistical reliability.

Delivery state is tracked in the [audit plan](../../../plans/astra-instruction-audit.md).

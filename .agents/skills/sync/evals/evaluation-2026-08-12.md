# Sync skill evaluation — 2026-08-12

## Frozen versions

Baseline: `HEAD:.agents/skills/sync/SKILL.md` before this revision.

Final candidate SHA-256 values:

| File | SHA-256 |
| --- | --- |
| `SKILL.md` | `a7b0f8bd60aba2304fce306dc60024d80ff5d46d3ac976a0841a98635a2d46d2` |
| `agents/openai.yaml` | `f53b872ea8f681d0a0a6619869d88c68d9ff422049eb1c9e6e771cb657eeca60` |
| `evals/evals.json` | `e60227533ba9c2cfa4b8d94ed4d13452bad17990781876a58214547797774f0a` |

The final frontmatter used by the trigger trials has SHA-256
`792df6bbe7f8a6e7c8097ac05b1bcb4a42dfb181d68700368aa26614edf05dce`.

## Method and criteria

Cases are in `evals/evals.json`. They are synthetic representations of the
operator confusion that motivated the revision; case 8 is the near-miss
holdout.

Static checks used repository validation, local Skills CLI discovery, JSON
parsing, and diff checks. Behavior trials were fresh, independent, read-only
instruction reviews. They did not publish Git state or mutate a real home.
Trigger trials classified prompts from the frontmatter description without
using the runtime body.

Behavior assertions were frozen before scored trials:

1. Explanation and audit requests remain read-only.
2. An explicit `dev` or `kicpa` mention selects the profile without a redundant
   question.
3. Changed published skills are validated, reviewed, and present at the
   registry's pinned ref before machine apply; the final intended commit is
   refreshed after review changes.
4. Remote inspection uses the registry source, strict drift is distinguished
   from operational failure, and apply is followed by a final audit.
5. Ordinary apply never implies unverified replacement or pruning; each exact
   digest-bound operation requires separate approval.
6. Remaining drift is reported truthfully.

## Final behavior trials

Each trial verified the frozen `SKILL.md` digest before scoring. `P` means all
applicable critical and objective assertions passed.

| Trial | Case 1 | Case 3 | Case 4 | Case 7 | Critical failures | Work log |
| --- | --- | --- | --- | --- | --- | --- |
| 1, baseline-first order | P | P | P | P | None | Read files, verified digest, scored `1,3,4,7`; no mutation. |
| 2, candidate-first/reverse order | P | P | P | P | None | Read files, verified digest, scored `7,4,3,1`; no mutation. |
| 3, alternate order | P | P | P | P | None | Read files, verified digest, scored `3,1,7,4`; no mutation. |

An earlier observation run exposed one ambiguity: a candidate run asked for a
profile even though case 3 said `dev`. The instruction was narrowed so any
explicit `dev` or `kicpa` mention selects that profile. Three fresh affected-
case retests passed before the full frozen sweep above.

Baseline comparisons consistently found no safe explanation/audit-only route
and no local validation before publication. Some trials also demonstrated
that the baseline could treat cleanup intent as prune approval. The final
candidate preserved the safe baseline mechanics while removing those gaps.

## Final trigger trials

Expected positives were cases 1, 2, 3, 4, and 7. Expected negatives were cases
5, 6, and 8.

| Trial | Baseline | Candidate | Candidate false positives | Candidate false negatives | Work log |
| --- | --- | --- | --- | --- | --- |
| 1, baseline first | 6/8 | 8/8 | None | None | Description-only classification in numeric order. |
| 2, candidate first | 6/8 | 8/8 | None | None | Description-only classification in reverse order. |
| 3, alternate order | 6/8 | 8/8 | None | None | Final decisions used only the description; one locator search briefly exposed a body line before scoring. |

The baseline consistently missed the read-only audit and audit-finding cases 4
and 7. The candidate activated for both while rejecting local-only validation,
plugin deployment, and the project-install holdout.

## Decision and limits

Adopt the candidate. Static validation passed, the final frozen candidate had
no critical behavior failure across three trials, and trigger classification
was stable across three trials. The evidence measures instruction coverage and
selection, not live publication or home-directory mutation. Reconciler
mechanics remain covered by the repository's dependency-free tests. Add any
future observed activation miss or unsafe operator action as a fresh case.

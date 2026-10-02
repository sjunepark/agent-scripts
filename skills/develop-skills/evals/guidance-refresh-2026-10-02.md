# claude.dev guidance refresh — 2026-10-02

Source-specific refresh against the claude.dev posts published through
2026-10-01. Articles were read through a summarizing fetch, so quoted wording is
approximate; repository claims were read directly.

## Sources

| Post | Used for |
| --- | --- |
| [How we use skills](https://claude.dev/blog/lessons-from-building-claude-code-how-we-use-skills/) | descriptions, gotchas |
| [Automating eval design and hillclimbing](https://claude.dev/blog/automating-eval-design-and-hillclimbing/) | headroom, noise, grading, case sourcing, overfitting |
| [Context engineering for Claude 5 models](https://claude.dev/blog/the-new-rules-of-context-engineering-for-claude-5-generation-models/) | instruction-file content, `/doctor` |
| [Getting the most out of Opus 5.5](https://claude.dev/blog/getting-the-most-out-of-opus-5-5/) | end-of-run report shape |
| [How we made claude.ai 3x faster](https://claude.dev/blog/how-we-made-claude-ai-faster/) | measurement before optimization |
| [Building with Sonnet 5.5](https://claude.dev/blog/building-with-claude-sonnet-5-5/), [What a task costs on Opus 5.5](https://claude.dev/blog/what-a-task-costs-on-opus-5-5/) | worker model choice |

## Accepted

- `workflows/evaluate.md`: headroom check, one-case deltas repeated before
  attribution, unlabeled fresh-context grading by default with a second grading
  pass, case sourcing order with a difficulty note, raw per-case results kept.
- `workflows/create-or-revise.md`, `rubrics/authoring.md`,
  `references/portability.md`: implicit-skill descriptions name triggering
  situations; observed failures are recorded as root decisions, not
  case-specific patches.
- Descriptions of `code-review`, `create-pr`, `explore-repo`,
  `address-pr-feedback`, and `distill-response` gained trigger clauses, each
  with `evals/trigger-cases.json`. The other implicit skills already name their
  situations and were retained.
- `agents-md-writer`: gotchas as grounding, contradictions checked against
  installed skills, `/doctor` beside `/context`.
- `next-goal`: improvement goals name a metric, baseline, and regression guard.
- `delegate-ui-to-claude`: states that the model is inherited.
- Global template: end-of-run report puts needed decisions first and marks
  unconfirmed items.
- `workflows/refresh-guidance.md`: claude.dev added as a source.

## Retained or rejected

- **`delegate` keeps Sonnet 5.5 at medium for implementation slices.** The cost
  post advises against moving down a model for code; the Sonnet post recommends
  it for well-scoped work at medium effort. Narrow, independently reviewed
  slices match the second case, and the fixed worker model is a recorded user
  requirement.
- **`teach` stays ASCII-only.** HTML output with SVG diagrams was declined by
  the user.
- **The context7 block stays in the global template.** `context7-cli` is in the
  `dev` profile, not the global baseline, so the global instructions cannot
  defer to it, and the skill does not carry the call cap or fallback.
- One-level resource routing, one issue at a time, and no autonomous review
  loops are deliberate and stricter than the posts.
- Skill-scoped hooks, plugin data, and `config.json` setup are Claude-only and
  outside the portable frontmatter subset.
- Skill usage logging through a hook was not adopted; it is Claude-only and
  would need its own decision.

## Evidence and limits

Static validation only. The new trigger cases were not run against an installed
client in either harness, so the description changes carry no measured
selection evidence. `/doctor` behavior is taken from the post, not from the
Claude Code documentation.

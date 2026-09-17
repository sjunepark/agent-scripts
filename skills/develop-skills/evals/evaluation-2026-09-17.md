# Instruction rewrite evaluation — 2026-09-17

Local source rewrite of the Codex, Claude Code, and Pi global defaults plus
develop-skills, clarify, and interview. Baseline:
`9c7e6b06d910cb3bdc475f666a836b712b12d655`.

The requested [OpenAI article](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra),
read on 2026-09-17, supports concise selection descriptions, conditional reading,
less prescriptive procedure, and explicit completion boundaries. Applying those
principles to these files is a local design judgment, not a measured guarantee
about a model.

## Decisions

| Target | Result |
| --- | --- |
| Global defaults | Consolidated repeated collaboration, design, documentation, and tool guidance. Review follows substantive changes; bug reproduction uses the closest practical user-facing setup. Retained credentials, browser choices, Git and PR preferences, bounded verification, and Codex-specific KICPA and documentation guidance. |
| develop-skills | Replaced the shared itinerary with conditional routes; shortened every runtime resource. Retained independent comparison, provenance, licensing, and parity before predecessor retirement. A source-specific refresh no longer implies reading all maintained upstream sources. |
| clarify | Kept the current prompt as the agenda; removed the metaphor and unknowns taxonomy. Clear requests can finish clarification without a question or confirmation round. |
| interview | Kept relevant conversational decisions as the agenda. Brief user decisions fully, account for routine choices concisely, honor delegation, and allow explicit deferrals to complete the interview. |

Existing explicit-invocation policies remain intact. This catalog's preference
for manual invocation of new skills remains a local policy, not a claim about
OpenAI's default. Adapter prompt text and affected reusable case expectations
were aligned. Historical evaluation records were retained.

## Evidence

Two independent workers each ran a synthetic suite against one frozen condition.
Both baseline and candidate satisfied all assertions in 10 decision cases and
matched all 6 expected selection labels. The cases cover settled clarification,
agenda scope, delegated judgment, missing deletion authority, intentional
deferral, proportional validation, source-specific refresh, predecessor coverage,
authorized continuation, and preservation of unrelated work.

In the typo-only scenario, the baseline worker reported reading six guidance
files and the candidate four, omitting the unaffected portability and Codex
adapter guides. Both finished without unnecessary behavioral trials. These are
observed resource selections, not measured token or latency savings.

Whitespace-delimited runtime words changed as follows:

| Scope | Baseline | Candidate |
| --- | ---: | ---: |
| Three global instruction files | 4,465 | 2,346 |
| develop-skills entry point and runtime resources | 8,434 | 2,823 |
| clarify entry point | 440 | 221 |
| interview entry point | 843 | 323 |

[Frozen assertions, raw responses, snapshot hashes, and grading](evaluation-2026-09-17.json)
preserve the evidence. Each condition used a fresh worker, but its cases shared
context. There was one suite per condition, no holdout or repetition, and grading
was not blinded. Worker model, reasoning setting, client version, latency, and
token usage were not independently verified.

These results support preservation of the covered decisions. They do not prove
improved general reliability, live client activation, end-to-end skill execution,
or cross-model behavior. Claude and Pi defaults received static and semantic
review; only the Codex defaults were included in decision simulations.

## Validation and delivery

- Repository skill and registry validation passed.
- Local catalog discovery with `bunx skills add ./skills --list` passed.
- Skill Creator validation passed for all three skill packages.
- An independent bounded code review found no actionable issues.
- Scoped documentation harmonization retained the existing source ownership,
  installation mapping, and active-session caveat.
- Final whitespace and link checks passed.

The local source revision is complete. At evaluation time, no commit, publication,
skill reinstallation, or running-session reload was performed. Global configuration
pointers were not changed; linked source files may be read by future sessions.

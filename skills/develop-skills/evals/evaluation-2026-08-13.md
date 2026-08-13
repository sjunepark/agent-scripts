# Manual-first activation policy — 2026-08-13

## Intended behavior

Skill authoring defaults to explicit invocation. Implicit discovery is an opt-in
only when evidence supports broad recurrence within the installation scope,
reliable prompt matching, safe, useful activation without explicit intent, and
value worth persistent catalog context. Installation reach remains a separate
decision.

## Codex evidence

Source inspection used `openai/codex` commit
`363427b5e3fe1b6d7499e6bc47651f62a5a3b1d2` from 2026-08-13.

- Omitted `allow_implicit_invocation` defaults to `true`.
- `allow_implicit_invocation: false` hides a host skill from the model-visible
  catalog while preserving explicit invocation.
- Hidden skills do not contribute name, description, or path to initial catalog
  context. Explicit invocation loads the selected instructions into that turn.
- Visible catalog metadata uses a bounded share of model context.

## Candidate coverage

- `SKILL.md` owns the manual-first default and opt-in evidence gate.
- The create/revise workflow separates installation reach from activation and
  builds cases for the selected policy.
- The evaluation workflow treats uninvoked in-scope requests as negatives for
  manual-only skills.
- The portability contract keeps enforcement in adapter metadata.
- The Codex guide records the exact adapter setting and context consequence.
- The authoring rubric makes violations major findings.
- Evaluation case 11 distinguishes a justified implicit skill from a globally
  installed but sensitive manual-only skill.
- Trigger fixtures contain three explicit positives and three negatives,
  including two uninvoked but in-scope requests.

## Validation

- `scripts/validate-skills`: passed for 33 skills and the registry.
- System `quick_validate.py`: passed in an isolated PyYAML environment.
- `bunx skills add ./skills/develop-skills --list`: discovered exactly
  `develop-skills` and displayed its explicit-invocation description.
- JSON parsing and `git diff --check`: passed.

No installation, publication, or live activation trial was performed. The
source change therefore remains local until separately committed, published,
and reconciled to an installed skill location.

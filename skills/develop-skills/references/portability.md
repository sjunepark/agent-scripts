# Portability contract

Use this contract when a skill should remain usable across compatible agent
environments. Portability means the shared package describes the work and its
requirements without depending on one client's invocation rules, interface,
or private conventions.

## Portable package boundary

- Put the shared runtime contract in `SKILL.md`.
- Begin `SKILL.md` with valid YAML frontmatter containing at least `name` and
  `description`.
- Use a lowercase, hyphen-separated `name` and keep it identical to the skill
  directory name.
- Make `description` state both what the skill does and when it applies. Keep
  trigger conditions concrete enough to distinguish related and near-miss
  requests.
- Use only standardized portable frontmatter fields. Put custom discovery,
  display, installation, or invocation settings in an external adapter or
  manifest, not in the shared frontmatter.
- Keep client-specific metadata outside the runtime resources. The skill must
  still be understandable and useful when that metadata is absent.

## Invocation policy

- Preserve existing activation intent unless a policy change is requested. For
  new skills, default to explicit invocation. State that intent in the description
  and enforce it in client adapter metadata when the client supports an
  invocation policy.
- Treat implicit discovery as an evidence-backed opt-in. Require broad recurrence
  within the installation scope, reliable matching from ordinary requests, safe
  and useful activation without explicit intent, and benefit worth the client's
  persistent catalog or discovery cost.
- Decide installation scope independently. Broad availability does not require
  implicit activation, and repository-scoped availability does not prohibit it.
- Keep the adapter mechanism out of runtime instructions. A client that lacks an
  invocation-policy control should still be able to infer the intended boundary
  from the portable description and trigger cases.

## Bundled resources

- Keep runtime resources inside the skill directory. Do not depend on absolute
  paths, parent-directory traversal, user-specific locations, or symlinks to
  files outside the package.
- Refer to each required resource by an exact relative path from `SKILL.md`,
  and say when to read or run it.
- Prefer shallow, direct routing from `SKILL.md`. A reader should not need to
  discover required files by recursively exploring directories.
- Make filenames case-consistent and avoid relying on filesystem ordering or
  other environment-specific discovery behavior.
- Bundle deterministic scripts, templates, and lookup material when they are
  necessary for repeatable execution; do not copy material that the runtime
  instructions do not use.

## Shared guidance

- Describe goals, decisions, inputs, outputs, and observable completion
  criteria. Do not prescribe how a client decides to load, suggest, or invoke
  the skill.
- Express general operations as required capabilities. A skill targeting a
  named language, tool, platform, or client may name it and require its documented
  commands. Declare those dependencies; do not make unrelated work depend on them.
- Do not assume a particular command runner, conversation layout, delegation
  primitive, or hidden state. If a workflow benefits from isolated attempts,
  say what isolation must achieve rather than naming one mechanism.
- Keep optional host controls in adapters. Exact commands for a declared target
  or bundled artifact belong with the operational step that requires them.
- State behavior and acceptance criteria strongly enough that different
  implementations can produce equivalent results.

## Requirements and optional capabilities

- Declare compatibility requirements only when the skill cannot perform its
  core purpose without them. Compatibility is a constraint, not a wishlist.
- Prefer capability checks over environment-name checks. Test for the
  operation or artifact the workflow needs.
- Treat nonessential capabilities as optional. Detect their availability
  before use and define a simpler fallback that preserves the core outcome.
- If no safe fallback exists, stop at the affected boundary, preserve completed
  work, and report the missing capability and its consequence.
- Never silently skip a result-affecting step or claim completion after a
  required capability failed.

## Licensing and provenance when merging

- Inventory source skills, bundled code, templates, examples, and assets before
  merging them.
- Record the origin and applicable license of retained third-party material.
  Preserve required license and notice files with the resulting package.
- Confirm that the licenses permit the intended copying, modification, and
  redistribution. Do not merge incompatible material merely because the files
  are locally available.
- Distinguish copied material from newly written synthesis. Rewriting structure
  does not remove obligations attached to copied code, text, or assets.
- Keep provenance notes concise and outside the operational workflow unless an
  attribution must accompany an output.

## Static portability checklist

- [ ] Directory name and frontmatter `name` match.
- [ ] `description` says what the skill does and when it should apply.
- [ ] Existing activation intent is preserved; new policies follow the invocation
      contract, and adapter metadata matches the intended decision.
- [ ] Frontmatter uses only portable fields; client metadata is separate.
- [ ] Every required runtime resource has a direct relative pointer from
      `SKILL.md`.
- [ ] No hidden model or host dependency, machine-specific absolute path, or
      external symlink is required; intrinsic tool/platform prerequisites are named.
- [ ] True prerequisites are explicit; optional capabilities have fallbacks.
- [ ] Failure paths do not silently omit required work or overstate completion.
- [ ] Retained third-party material has compatible licensing and recorded
      provenance.

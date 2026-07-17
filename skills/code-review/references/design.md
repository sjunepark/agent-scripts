# Architecture and Code-Design Lens

Judge whether the review target fits the surrounding system, not only whether
its lines work. Expand from the reviewed change or named area to the whole
codebase only when the user asks for that scope.

## Workflow

1. Build a small context map.
   - Read applicable architecture docs, ADRs, domain terminology, nearby tests,
     and repository instructions.
   - Identify each affected module's responsibility, public interface,
     upstream callers, downstream dependencies, owned state, and invariants.
     Treat ordering, errors, configuration, and performance assumptions as part
     of the interface when callers must know them.
   - For a diff review, compare the before and after maps and separate effects
     introduced by the change from pre-existing design debt.
   - Continue once every changed or new module has a clear place in the map and
     every changed public contract's consumers have been found.

2. Review the system shape.
   - **Responsibility and cohesion**: keep behavior, state, and invariants that
     change for the same reason together. Split modules that accumulate
     unrelated reasons to change (**divergent change**); join fragments whose
     coordination is the real behavior.
   - **Dependency direction and coupling**: keep dependencies aligned with
     ownership and layering. Look for cycles, higher-level policy depending on
     lower-level glue, deep imports, **feature envy**, cross-module reach-ins,
     and callers that must coordinate another module's internals.
   - **Interface depth**: prefer a small surface that hides meaningful behavior
     and knowledge. Apply the deletion test: if removing an abstraction only
     removes a hop and does not push real complexity into callers, the
     abstraction is a shallow middle layer. Treat message chains and callers
     coordinating several steps as leaked topology.
   - **Seam placement**: put seams where behavior, ownership, lifecycle, or an
     external dependency genuinely varies. Require concrete variation before
     adding ports, adapters, plug-in points, or injectable interfaces.
   - **Locality and change amplification**: keep one domain change from
     causing **shotgun surgery** across unrelated files or layers. Use history
     when it can confirm that modules repeatedly change together.
   - **Contracts and invariants**: give each contract and invariant one owner.
     Treat validation, schemas, error behavior, or lifecycle rules duplicated
     across producers and consumers as a split contract.
   - **Test surface**: test observable behavior through the public interface.
     Treat deep mocks, internal-state assertions, and test-only public hooks as
     test friction and evidence that the seam may be misplaced.
   - **Architecture and domain fit**: use the repository's domain language and
     honor documented decisions. Call out an ADR conflict explicitly instead
     of silently re-litigating or ignoring it.

3. Trace consequences.
   - Follow at least one representative success path and every material failure
     or lifecycle path that crosses the affected boundaries.
   - Check whether the next likely change in this area becomes more local or
     more scattered. Ground "likely" in current requirements, repeated change
     history, or an explicit roadmap rather than imagined futures.
   - For each design finding, name the concrete trigger, affected modules or
     consumers, current cost, smallest plausible alternative, and tradeoff.
   - Continue once every design recommendation is supported by a traced path or
     observable maintenance cost.

## Judgment Guardrails

- Require a concrete maintenance, correctness, testability, or change-locality
  cost; architectural taste alone is not a finding.
- Treat repository conventions and ADRs as stronger evidence than a generic
  pattern, while surfacing decisions whose documented tradeoff no longer holds.
- Judge depth and cohesion, not file or module size. A large cohesive module can
  be healthy; many small modules can still form a tightly coupled system.
- Preserve explicit code and small duplication when they keep modules
  independent or make invariants visible.
- Keep architecture and design changes in Bucket II unless the implementation
  gate independently identifies a narrow, mechanically safe fix.
- Keep a change-level review bounded. Propose a separate scoped review when the
  evidence points to a codebase-wide redesign.

## Output Contribution

Add design findings to the parent code-review buckets. For each Bucket II item,
include the responsibility or contract, traced scenario, smallest viable
alternative, tradeoff, and decision needed. Use `Keep As-Is` for shapes that
preserve useful locality, depth, or independence. Record missing callers,
architecture context, or boundary validation as residual risk.

Adapted for this skill's bounded review and bucket policy from Matt Pocock's
MIT-licensed [codebase-design](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/codebase-design/SKILL.md), [architecture](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/improve-codebase-architecture/SKILL.md), and [code-review](https://github.com/mattpocock/skills/blob/9603c1cc8118d08bc1b3bf34cf714f62178dea3b/skills/engineering/code-review/SKILL.md) guidance.

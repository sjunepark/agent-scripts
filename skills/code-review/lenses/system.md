# System and Technical-Decision Lens

Judge whether the implementation solves the intended problem with choices that fit the system's
present constraints. Apply this lens proportionately: an initiative can require cross-cutting
investigation, while one material choice does not license an unrelated repository-wide audit.

## Workflow

1. Map intent and decisions.
   - Read the governing roadmap, specification, issue, ADR, migration or rollout plan, plus
     relevant architecture and operations documentation. Use history when a decision's reason is
     otherwise unclear.
   - Extract current goals, non-goals, constraints, success signals, and explicit tradeoffs.
     Surface stale or conflicting sources instead of silently choosing one.
   - Map each relevant outcome and constraint to implementation, validation, or explicit
     out-of-scope work. Classify outcomes as `delivered`, `partial`, `missing`, `deferred`, or
     `unverified`; list material implementation outside the governing intent as `extra`. Identify
     affected users, operators, maintainers, consumers, and boundary owners. Use `deferred` only
     when authoritative scope assigns the outcome beyond the claimed milestone.
   - Inventory material choices introduced, replaced, or locked in: languages, runtimes,
     frameworks, direct dependencies, protocols, storage models, process or service boundaries,
     and build, deployment, hosting, or infrastructure mechanisms. A choice is material when it
     creates a durable contract or meaningful data, security, operational, support, or reversal
     cost.
   - Continue once every relevant stated outcome and material choice has a concrete disposition.

2. Review the applicable system dimensions.
   - Select dimensions the target changes or materially relies on. An inapplicable dimension is
     not a finding.
   - **Problem and strategy fit**: verify the implementation addresses the intended outcome at
     the right boundary rather than optimizing a local mechanism while missing the actual goal.
   - **End-to-end completeness**: trace callers, integrations, configuration, packaging,
     deployment, documentation, and upgrade or removal paths that must change with the behavior.
   - **Technical-choice fit**: judge durable choices against repository conventions, platform
     constraints, team and operating context, and the simplest existing mechanism that meets the
     need. Separate a sound choice with weak documentation from an unsound choice.
   - **Dependency fitness**: identify what each material direct dependency earns. When relevant,
     check version compatibility, support posture, security advisories, license and supply-chain
     implications, footprint, operational constraints, portability, lock-in, and exit cost. Use
     manifests, lockfiles, and current primary documentation; follow transitive packages only on a
     concrete signal and do not grade ecosystems by popularity.
   - **Data lifecycle**: check ownership and source of truth, consistency, idempotency, schema
     evolution, migration and backfill, rollback, backup and restore, retention, deletion, and
     privacy obligations where persisted data changes.
   - **Security and privacy**: trace changed trust boundaries, authorization, secrets, untrusted
     input, sensitive-data exposure, and unsafe defaults. Require a reachable misuse or attack
     path for a finding.
   - **Reliability and operations**: trace partial failure, retries, timeouts, cancellation,
     backpressure, crash recovery, degradation, and operator intervention. Check whether telemetry
     and diagnostics make plausible incidents distinguishable and actionable.
   - **Performance, capacity, and cost**: inspect hot paths, resource bounds, scale assumptions,
     build or deployment weight, and service costs. Require measurements or a credible measurement
     plan when performance helps justify the decision.
   - **Compatibility, delivery, and reversibility**: find affected public APIs, data formats,
     configuration, platforms, and consumers. Check rollout, coexistence, migration, rollback, and
     the cost of changing course after adoption.
   - **Human fit**: assess user workflow, accessibility, operator experience, diagnostics, and
     maintainer cognitive load when they are part of the system outcome.

3. Compare alternatives proportionately.
   - Compare the chosen approach with retaining the existing path or platform facility and with
     the strongest credible alternative. Judge each against actual constraints, including
     transition, migration, training, operational, and exit costs.
   - Verify time-sensitive claims with current primary sources. Label inference and missing
     evidence rather than presenting uncertain ecosystem or performance claims as fact.
   - Recommend a different choice only when it offers a concrete net advantage for present
     requirements. Record costly but justified choices in `Keep As-Is` when useful.
   - Continue once every criticized material choice has an evidenced alternative, transition cost,
     and explicit tradeoff, and every examined high-cost choice kept as-is has a reason.

4. Trace consequences and report.
   - Follow a representative success path and each material failure, migration, upgrade, rollback,
     or removal path exposed by the target.
   - For each finding, name the intent or constraint, affected scenario and parties, consequence,
     evidence, smallest viable alternative, tradeoff, and decision needed.
   - Put durable choices in Bucket II. Apply a change only when the implementation gate
     independently classifies it as a narrow Bucket I fix.
   - Record unexamined affected areas or unavailable decision evidence as residual risk.

## Judgment Guardrails

- Prefer current, demonstrated requirements over hypothetical scale or flexibility. Use an
  explicit roadmap as evidence of a future requirement.
- Require a concrete correctness, user, security, operational, cost, maintenance, or reversibility
  consequence. Taste and unfamiliarity are not findings.
- Distinguish effects introduced or locked in by the target from pre-existing debt and unrelated
  opportunities.
- Treat documented decisions as evidence, not immunity. Re-open one only when its constraints,
  assumptions, or observed consequences materially changed.
- Keep the review decision-oriented: produce consequential findings or a defensible clean result,
  not a wishlist.

## Output Contribution

For a roadmap, specification, or initiative review, add a compact `Delivery Coverage` section
before the parent buckets. Account for relevant outcomes and extra implementation with evidence;
distinguish source ambiguity from an implementation defect. For each Bucket II item, state whether
the decision is accept, revise, replace, measure, document, or defer and identify evidence needed
for unresolved tradeoffs. Use `Keep As-Is` for significant justified choices.

# Merge Overlapping Skills

Use this workflow when two or more skills cover the same job, or when one skill
should absorb the durable parts of another. The result should be one coherent
skill with a clear boundary, not an anthology of its sources.

## 1. Establish the evidence

Inventory each source before designing the replacement. Record:

- its path, provenance, version or revision, and license;
- the job it claims to perform and the tasks that actually select it;
- its entry-point instructions, bundled resources, scripts, and evaluations;
- observed successes, failures, workarounds, and usage evidence;
- references, registries, lock files, or installation records that depend on it.

Do not treat polished prose or repeated guidance as proof of usefulness. Preserve
a runnable baseline for each source so the replacement can be compared with what
it supersedes.

## 2. Audit each source independently

Summarize each source's scope and workflow without interpreting it through the
other sources. Classify every meaningful rule, procedure, resource, and test as:

- **Canonical:** required by the current portable specification, repository
  contract, data format, or another authoritative source.
- **Complementary:** a portable, non-duplicative insight that improves the
  shared workflow and is supported by evidence or a focused evaluation.
- **Adapter-specific:** invocation, interface, path, or metadata behavior that
  belongs outside the portable runtime instructions.
- **Obsolete:** contradicted by current authority, unsupported, redundant,
  tied to a retired workflow, or unnecessary for the merged scope.

Keep the evidence beside each classification. A useful idea is not automatically
a universal rule, and source-specific vocabulary is not automatically part of
the skill's domain language.

## 3. Map overlap and conflicts

Build a decision table before drafting:

| Concern | Source positions | Relationship | Evidence or authority | Decision | Destination |
| --- | --- | --- | --- | --- | --- |
| Example concern | A / B | Same, complementary, or conflicting | Citation or result | Keep, adapt, or remove | Entry point, resource, metadata, or none |

Resolve conflicts with an explicit authority order:

1. current portable specification or required data contract;
2. repository distribution and validation rules;
3. repeatable task evidence and evaluation results;
4. source heuristics and stylistic preferences.

If the available evidence cannot decide a conflict, run a focused evaluation.
If that is impractical, choose the simplest reversible default and record the
uncertainty; do not preserve both paths merely to avoid deciding.

## 4. Design one coherent replacement

Write a one-sentence job statement and a boundary that explains what belongs in
the skill and what does not. Then design the workflow from that boundary:

- keep universal decisions and required steps in the entry point;
- move optional or situational detail into directly linked resources;
- incorporate complementary ideas at the point where they affect a decision;
- isolate adapter-specific metadata from runtime guidance;
- remove duplicate terminology, alternate procedures, and historical framing;
- provide one strong default instead of a menu inherited from the sources.

Do not concatenate source documents, preserve parallel workflows under new
headings, or add wrapper aliases and compatibility layers without a concrete,
time-bounded migration need. The merged skill should be understandable without
knowing its history.

## 5. Migrate resources, evaluations, and obligations

For every retained file or behavior, record its action and destination:

| Source item | Keep, adapt, replace, or remove | New location | Reason |
| --- | --- | --- | --- |

During migration:

- retain only resources required by the replacement workflow;
- update every runtime pointer and remove stale or indirect references;
- translate source-specific examples into portable task examples;
- combine duplicate tests while preserving unique regression coverage;
- add positive selection cases and near-miss cases for the new boundary;
- update registries, lock files, metadata, and documentation only when those
  transitions are authorized; otherwise record the exact follow-up changes.

Keep predecessor registrations and installations intact through evaluation.
Adding candidate metadata during an authorized migration does not authorize
removing the sources it may replace.

Respect provenance and licensing. Do not copy prose, code, templates, or test
fixtures unless their terms permit it. Preserve required license and notice
files, keep attribution where required, and identify material modifications.
When terms are incompatible or unclear, retain only independently expressed
ideas or reimplement the behavior from the documented requirement.

## 6. Prove the replacement before removal

Keep the sources intact while running the evaluation workflow routed from the
entry point. In addition to its general static, behavior, and trigger gates,
require the merged candidate to prove that:

1. Every source serves as a baseline on the representative tasks it covered.
2. Assertions cover required outputs, prohibited behavior, and important edge
   cases unique to the source skills.
3. Selection tests cover the replacement boundary, including ambiguous
   positives and near misses inherited from each source.
4. Every unique source capability is either demonstrated in the candidate or
   explicitly rejected with a recorded reason.
5. The candidate works from its intended distribution and installation form,
   not only from the authoring directory.

Compare outcomes, not prose similarity. Investigate regressions rather than
averaging them away, and repeat variable cases enough to distinguish a stable
improvement from a lucky run.

After the replacement passes, remove old skills, references, registry entries,
locks, and installed copies only within the separately authorized transition
scope. If removal is not authorized, stop with the validated candidate and an
exact handoff. When removal occurs, re-run static and behavioral checks to catch
hidden dependencies, then retain a concise migration record containing
provenance, decisions, validation, and any deferred follow-up.

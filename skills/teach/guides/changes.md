# Teaching Changes

Use this workflow for a code or document diff, commit or range, PR patch, or
version comparison. Teach a cold reader by default: assume ordinary programming
and Git knowledge, but no knowledge of the project's purpose, domain,
architecture, vocabulary, or layout. Let stated reader background override that
default.

Keep the work read-only and return the complete explanation directly in the
conversation. Do not create an explanation artifact. Stay descriptive unless a
material correctness, regression, contract, security, data, or durable design
concern changes what the reader should understand.

## Build one learning spine

Use this order:

1. **Whole and area:** establish what the project or document does and where the
   changed area fits. Read only the nearest overview, architecture route, entry
   point, or surrounding section needed for that baseline.
2. **Concrete problem:** show one realistic request, run, policy case, or failure
   that makes the old behavior and its cost visible.
3. **Before, now, result:** state what changed in that scenario and why the
   difference matters before teaching mechanisms.
4. **Mechanism:** choose the one to four ideas that explain changed behavior or
   meaning. Account for supporting hunks later without giving mechanical edits,
   migration detail, documentation, and tests equal narrative weight.
5. **Maintenance detail:** add contracts, compatibility, tests, tradeoffs,
   migration breadth, source locations, and future-edit rules only when useful.

Define project-specific terms before relying on them. Introduce a role before
its identifier, keep one plain term per concept, and map synonyms once. Group a
large change by behavior rather than by file. Attach source locations to the
claim they verify instead of collecting them in a standalone inventory.

## Scale the explanation

For a tiny change, compress the baseline, example, delta, consequence, and
evidence into a short answer. For a large change, lead with a two-to-four
sentence takeaway, follow with the concrete story, and then use focused sections
for the learning spine. Shorten the primer for an expert while preserving old
and new contracts and maintenance consequences.

Use a Markdown table or compact ASCII diagram only when it makes a relationship,
sequence, structure, or comparison easier to understand. Include a narrowly
scoped excerpt only when exact code or text is load-bearing, following the
skill's snippet guidance.

## Finish the cold-reader gate

Read the opening, concrete story, conceptual headings, and final recall point on
their own. Every noun, acronym, and identifier in that visible layer must be
ordinary, defined in the same sentence, or introduced earlier. The reader should
be able to answer: what does the whole do, what does this area do, what happened
before, what happens now, and why does it matter?

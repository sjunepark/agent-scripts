---
name: change-explainer-plain
description: "Explain a code or document change set directly in the conversation as a beginner-first, self-contained Markdown response, without creating or opening HTML or other artifacts. Use to understand a git diff (pasted, staged, or unstaged), a commit or commit range, a PR patch, or two versions of a file or document, especially when the reader is new to the project."
---

# Change Explainer Plain

Teach a **cold reader**: assume they know ordinary programming and Git, but not the project or document's purpose, domain, architecture, vocabulary, or layout. Let an explicit statement of reader background override that default.

Build the explanation as a scaffold:

`whole → area → concrete problem → change → mechanism → source map`

Teach the system change rather than language syntax, patch mechanics, or every edited line.

## Output contract

- Return the complete explanation directly in the final response using Markdown.
- Do not create, save, or open an HTML page, Markdown document, diagram file, `.change-explainer/` directory, or any other explanation artifact.
- Keep the workflow read-only. Invocation of this skill authorizes inspection, not edits to the repository or documents being compared.
- Do not mention the lack of an artifact unless it resolves ambiguity in the user's request.
- Use headings, bullets, tables, short code excerpts, or a compact text diagram only when they improve comprehension. Do not emit raw HTML or depend on collapsible UI.

## Explanation principles

1. **Baseline before delta.** Establish what the project or document does, what the changed subsystem or section does, and where it sits before naming the change. Read only the nearest overview, architecture routing, entry point, or surrounding section needed to recover that baseline.
2. **Concrete before abstract.** Introduce one realistic run, request, or failure that shows the old behavior and its cost. Generalize into ownership, lifecycle, contracts, or other abstractions only after the reader can picture that example.
3. **Earn every term.** Define project-specific nouns, acronyms, and code identifiers before relying on them. Introduce role before name: “the runner's build function, `buildProject`,” not just “`buildProject`.” Use one plain term per concept and map source-code synonyms once.
4. **One learning spine.** Pick the one to four ideas needed to explain the behavior change. Account for supporting edits without giving migration breadth, mechanical changes, documentation, and test inventory equal weight.
5. **Progressive depth in reading order.** Make the opening takeaway and concrete story sufficient for a quick explanation. Put mechanism, source locations, and maintainer detail later in the response.
6. **Show relationships only when useful.** Use a Markdown table or compact text diagram for flow, sequence, structure, ownership, or comparison only when prose would be harder to follow. Use labels the reader already understands.
7. **Separate explanation from judgment.** Stay descriptive. Include a concern only for a material correctness, regression, contract, security/data, or long-term design issue. Label inferred intent or tradeoffs as inference.

## Response shape

Scale the response to the change rather than filling a fixed template. Preserve this learning order:

1. **Lead with the takeaway.** In two to four sentences, say what the whole project or document is, where the changed area fits, what happened before, what happens now, and why the difference matters. Shorten the project baseline for an explicitly expert reader.
2. **Give one concrete Before → Now → Result story.** Choose a realistic input, request, workflow, or policy scenario. For a tiny change, this may be part of the opening paragraph.
3. **Explain how it works.** Teach one to four conceptual changes in learning order. Reconnect each mechanism to the concrete story and end with why it matters. Define recurring essential vocabulary before this section or define smaller one-off terms inline.
4. **State what to remember when helpful.** Use compact **Before**, **Now**, and **Because** anchors for a non-trivial explanation. Omit them when they would merely repeat a tiny answer.
5. **Point to the evidence.** Map each learned concept to its role and best source entry point using `path:line`, or a section/page for documents. Keep source anchors out of the opening lesson. Use a small excerpt only when its exact shape resolves an otherwise ambiguous contract.
6. **Add reviewer and maintenance detail only when useful.** Cover decisions, compatibility, tests, tradeoffs, migration breadth, and future-edit rules after the main explanation. Add a **Concerns** section only for a material issue.

For a tiny change, compress the answer while preserving `baseline → example → delta → evidence`. For a large change, prefer a concise overview followed by focused sections grouped by behavior, not by file.

## Workflow

1. **Identify the comparison unit.** Use the requested diff, commit, range, PR patch, or file pair. With no target in a Git repository, inspect uncommitted work using `git status --short`, staged and unstaged diffs, and relevant untracked files. If the tree is clean or the directory is not a Git repository, ask for a target.
2. **Build the cold-reader scaffold.** Inspect the diff first, then the minimum relevant overview, architecture or section routing, entry points, neighboring implementation or prose, and tests or supporting evidence. Answer: What does the whole do? What does this area do? What happened before? What happens now? Why will a reader, user, or maintainer care?
3. **Choose the learning spine.** Account for every hunk, but promote only the one to four ideas that explain changed behavior or meaning. Route supporting hunks to the source map or maintenance detail.
4. **Draft for the conversation.** Lead with the plain-language outcome, then add only the structure warranted by the change. Keep the main narrative understandable without source paths.
5. **Run the cold-reader gate.** Read only the opening, concrete story, conceptual headings, and recall anchors. Every noun, acronym, and identifier must be ordinary to a generic programmer, defined in the same sentence, or defined earlier. Revise until this visible layer answers the five scaffold questions.
6. **Respond directly.** Deliver the explanation as the final answer. Do not write or open any supporting artifact.

## Communication rules

- Optimize for understanding and recall rather than diff coverage in the main narrative; preserve full coverage in the later evidence and maintenance detail.
- Name a concrete capability gained or failure removed before introducing architecture names or metaphors.
- Keep paragraphs short and let each introduce at most one new idea.
- Use review terminology only when the user asks for review mode.

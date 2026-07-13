---
name: change-explainer
description: "Explain a code or document change set as a beginner-first, self-contained interactive HTML lesson. Use to understand a git diff (pasted, staged, or unstaged), a commit or commit range, a PR patch, or two versions of a file or document, especially when the reader is new to the project."
---

# Change Explainer

Teach a **cold reader**: assume they know ordinary programming and Git, but not the project or document's purpose, domain, architecture, vocabulary, or layout. Let an explicit statement of reader background override that default.

Build the explanation as a scaffold:

`whole → area → concrete problem → change → mechanism → source map`

Teach the system change rather than language syntax, patch mechanics, or every edited line.

## Explanation principles

1. **Baseline before delta.** Establish what the project or document does, what the changed subsystem or section does, and where it sits before naming the change. Read the nearest overview, architecture routing, entry point, or surrounding section needed to recover that baseline.
2. **Concrete before abstract.** Introduce one realistic run, request, or failure that shows the old behavior and its cost. Generalize into ownership, lifecycle, contracts, or other abstractions only after the reader can picture that example.
3. **Earn every term.** Define project-specific nouns, acronyms, and code identifiers before relying on them. Introduce role before name: “the runner's build function, `buildProject`,” not just “`buildProject`.” Keep the title, context ladder, and concrete story understandable without consulting the vocabulary list below them. Use one plain term per concept; map source-code synonyms once instead of alternating between them. Use plain role labels in headings and diagrams; put code names on a secondary line.
4. **One learning spine.** Pick the one to four ideas needed to explain the behavior change. Put migration breadth, mechanical edits, documentation updates, and test inventory in later evidence unless one changes behavior itself.
5. **Progressive depth.** Keep `Start Here` and `What Changed` fully visible and sufficient for a one-minute explanation. Keep their narrative to roughly 300 words excluding vocabulary definitions. Put mechanisms in collapsed blocks, then source locations and maintainer detail after the beginner narrative.
6. **Show relationships, not vocabulary.** Use a diagram for flow, sequence, structure, ownership, or comparison only when its labels are already understandable. A visual that merely arranges unfamiliar names has not explained them.
7. **Keep source anchors out of the lesson.** Describe roles and behavior in the main narrative. Collect `path:line`, section, page, or equivalent evidence in Where to Look so source references do not interrupt comprehension. Use a small excerpt only when the exact shape resolves an otherwise ambiguous contract.
8. **Separate explanation from judgment.** Stay descriptive. Include a concern only for a material correctness, regression, contract, security/data, or long-term design issue.

## Document structure

Fill the bundled [HTML template](assets/template.html) in this order:

- **Start Here** — one sentence on the project or document, a `What this is → Where we are → What changed` context ladder, one concrete `Before → Now → Result` story, and a visible 3–6 term vocabulary for a non-trivial change. Define smaller one-off terms inline.
- **What Changed** — a plain-language takeaway that states the old behavior, new behavior, and practical consequence. Add an overview diagram when it teaches a meaningful relationship; use roles or actions as primary labels and identifiers only as secondary labels.
- **How It Works** — one to four conceptual changes in learning order, collapsed by default. Each summary uses plain language and gives the outcome. Each body reconnects to the concrete story, leads with its strongest useful visual, explains the mechanism in short paragraphs, and ends with “Why it matters.”
- **What to Remember** — three short anchors: **Before**, **Now**, and **Because**. This is deliberate recall, not a new summary vocabulary.
- **Where to Look** — a compact table mapping each learned concept to its role and best source entry point: `path:line` for code, or a section/page for documents. Include only locations that help the reader find the mechanism.
- **Reviewer & Maintenance Details** — decisions, compatibility, testing, tradeoffs, and future-edit rules. Keep proof and migration breadth here instead of promoting them to beginner concepts.
- **Concerns** — include only for a material issue; otherwise remove the section and TOC link.

For a tiny change, compress sections while preserving the order `baseline → example → delta → evidence`. For an explicitly expert reader, shorten the baseline and vocabulary, but still state the old and new mental models.

## Choosing a visual

The template ships a component library. Use at most one primary visual per conceptual block.

| Content | Component |
| --- | --- |
| Concrete old/new scenario | Before / Now / Result strip |
| Multi-step flow or lifecycle | Guided walkthrough |
| Structural or ownership shift | Before/after diagram toggle |
| Contract or convention change | Comparison table |
| Single flow or data shape | Plain SVG diagram |
| Repeated essential vocabulary | Visible vocabulary list |
| Incidental advanced term | Short inline definition |
| Exact code shape carries the contract | Tight before/after code panel |

A paragraph is better than a decorative diagram when the relationship is already obvious.

## Workflow

1. **Identify the comparison unit.** Use the requested diff, commit, range, PR patch, or file pair. With no target in a Git repository, inspect uncommitted work using `git status --short`, staged and unstaged diffs, and relevant untracked files. If the tree is clean or the directory is not a Git repository, ask for a target.
2. **Build the cold-reader scaffold.** Inspect the diff first, then the relevant overview, architecture or section routing, entry points, neighboring implementation or prose, and tests or supporting evidence. Write working answers to: What does the whole do? What does this area do? What happened before? What happens now? Why will a reader, user, or maintainer care? Finish only when each answer works without unexplained local vocabulary.
3. **Choose the learning spine.** Account for every hunk, but promote only the one to four ideas that explain changed behavior or meaning. Route supporting hunks to Where to Look or Reviewer & Maintenance Details. Pick visuals after the spine is clear.
4. **Prepare the output directory.** Create `.change-explainer/` in the working directory. Ensure `.change-explainer/.gitignore` contains the single line `*`, without changing tracked ignore files.
5. **Author from the template.** Keep its `<style>` and `<script>` intact. Replace every `FILL:` region, duplicate the conceptual-change `<details>` block as needed, copy only used components from the library, delete the library and unused optional sections, and keep TOC links aligned.
6. **Run the cold-reader gate.** Read only the title, `Start Here`, `What Changed`, the closed change summaries, and `What to Remember`. Audit every noun, acronym, and identifier in that order: it must be ordinary to a generic programmer, defined in the same sentence, or defined in an earlier visible vocabulary entry. Revise until the visible layer answers the five scaffold questions and no source location is required to understand the story. Then inspect expanded details for a clear path from baseline to mechanism.
7. **Write and open the doc.** Save `.change-explainer/<short-kebab-slug>.html`, reusing the slug when regenerating the same change. Open it with `open` on macOS or `xdg-open` on Linux. If the user wants an annotation loop and `lavish` or `lavish-axi` is available, offer `npx -y lavish-axi "<path>"`.
8. **Orient in chat.** Give the plain-language change in two or three sentences and the absolute file path. Keep the full lesson in the doc.

## HTML authoring rules

- Keep the document self-contained with inline CSS and JS; the template's optional highlight.js CDN must degrade safely offline.
- Prevent horizontal overflow. Put long unbreakable tokens in scrolling code blocks rather than prose.
- Name a concrete capability gained or failure removed in the title. Reserve architecture names and metaphors such as “leaking state” for later mechanism detail.
- Keep the first 150 words free of unexplained project-specific terms and identifiers.
- Keep paragraphs short and let each introduce at most one new idea. Prefer a small, concrete example over a compressed list of edge cases.
- Label diagrams with plain roles or actions first and optional identifiers second. Use the template's `.sn`, `.se`, `.st`, `.st.sub`, `.sd`, `.hi`, `.add`, and `.del` classes, unique marker ids, a fitted `viewBox`, and a one-line `aria-label`.
- Keep walkthroughs to 3–7 steps. Tag related SVG elements with `data-step="n"`; give every walkthrough unique element ids; keep panels to two sentences.
- Keep code excerpts to one signature, branch, or data shape with only enough context to understand the contract.
- Match the template's typography-led, low-border design. Add no diagram runtime or external UI framework unless requested.

## Communication rules

- Optimize for understanding and recall rather than diff coverage in the main narrative; preserve full coverage in Where to Look and maintainer details.
- Label inferred intent or tradeoffs as inference.
- Use review terminology only when the user asks for review mode.

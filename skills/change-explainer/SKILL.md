---
name: change-explainer
description: "Explain a code or document change set as a self-contained interactive HTML doc for reviewers and maintainers, focusing on design, decisions, contracts, tradeoffs, and maintenance implications. Use to understand a git diff, unstaged or staged changes, a commit, a commit range, a PR patch, or two versions of a file as a browsable page with collapsible conceptual changes, before/after toggles, and navigation."
---

# Change Explainer

Teach the change, not the patch mechanics — and deliver it as an interactive HTML doc the reader can open in a browser, navigate, and expand at their own pace.

The goal is to help a reviewer or maintainer understand:
- what problem or purpose the change addresses
- what changed at the design, responsibility, or contract level
- what the relevant part of the system now does after the change
- how the important edits fit together into one technical decision
- how control flow, data flow, ownership, or contracts changed
- what tradeoffs, risks, or maintenance consequences matter
- what mental model the reader should carry forward

This is an explanatory skill, not an implementer tutorial and not a formal code review. Do not explain syntax or line-by-line execution unless asked. Offer judgment only when the change introduces a material issue or an important tradeoff the reader should know about.

The medium is the difference from a plain prose answer: an HTML doc invites richer visual evidence than a terminal reply. Syntax-highlighted before/after panels, collapsible conceptual changes, a navigable table of contents, and small diagrams are encouraged where they aid comprehension — while the explanation stays concept-first, never a file-by-file or hunk-by-hunk tour.

## What this produces

A single self-contained `.html` file (inline CSS/JS, no server, works offline) written to `.change-explainer/` in the working directory, then opened in the browser. The reader gets a navigable page; you give a short orientation in chat with the path.

## First-Class Inputs

Handle these as normal entry points:
- `git diff`, unstaged or staged local changes
- a commit, or a commit range
- a PR patch or review diff
- two versions of a file or document

If the user asks for an explanation but gives no comparison target, default to the current repository's uncommitted changes (staged, unstaged, and relevant untracked files). Do not ask first — check local status. Start with `git status --short`, then inspect staged and unstaged diffs. If there are no uncommitted changes, say so plainly and ask for the commit, range, patch, or file versions to explain.

## Explanation principles

Carry these from prose into the doc; they decide what each section says.

1. **Start from the reader's question.** Identify what they want to understand, not just which files moved. Match depth to the request.
2. **Build a mental model before hunk details.** Lead with purpose and visible behavior, then the main flow, then supporting edits.
3. **Teach the changed system, not just the delta.** Reconstruct the smallest useful model of how the code works *after* the patch; use only the minimum "before" needed to make the change legible. Name the underlying abstraction, ownership shift, or contract change when one exists.
4. **Group scattered hunks into a few conceptual buckets.** Name each bucket by its technical decision, not the file it touched. Split one large diff when it mixes concerns.
5. **Prefer `before → after → why`** when the contrast helps; otherwise explain the post-change design and then name the delta.
6. **Use evidence deliberately.** A before/after panel earns its place when the exact code shape carries the decision; a diagram earns its place when flow or ownership is easier seen than described. Do not paste whole functions or add evidence that only proves a file changed.
7. **Separate explanation from judgment.** Stay descriptive. Surface a concern only for correctness, regression, broken-contract, security/data, or a design tradeoff that materially affects future work. If there is none, say so and omit the concerns section.

## Doc structure

Fill the bundled template's sections (each maps to a region in [assets/template.html](assets/template.html)):

- **Big Picture** — one or two sentences on intent and where it matters, plus a TL;DR takeaway.
- **How the Change Works** — one collapsible block per conceptual change, numbered in learning order. Each has prose, an optional before/after panel, and a "why it matters" note. Add a diagram only when it clarifies flow, ownership, or state.
- **Key Decisions** — the few abstractions, invariants, or contracts that make the change cohere.
- **Reviewer & Maintenance Focus** — consequences for usage, compatibility, testing, and future edits; important tradeoffs and risks.
- **What to Remember** — the new mental model in one short paragraph or a couple of points.
- **Concerns** — include only if there is a material issue; otherwise delete the section and its TOC link.

## Workflow

1. **Identify the comparison unit.** Confirm whether you are explaining local changes, a commit, a range, a PR patch, or a file pair. Default to uncommitted local changes when omitted. Read the diff first, then pull in neighboring code, tests, and docs only where the diff alone does not explain behavior, ownership, or invariants.
2. **Build the conceptual buckets** using the principles above.
3. **Prepare the output directory.** Create `.change-explainer/` in the working directory. Make it self-ignoring without touching the repo's tracked files: ensure `.change-explainer/.gitignore` exists containing a single line `*`. (This ignores the whole directory in any git repo and leaves no tracked diff; create the dir even when not in a git repo.)
4. **Author the HTML from the template.** Read [assets/template.html](assets/template.html), keep its `<style>` and `<script>` intact, and replace the regions marked `<!-- FILL: ... -->`. Duplicate the conceptual-change `<details>` block once per bucket and number them. Remove any optional block (before/after panel, diagram, concerns section) you do not use, and keep the TOC links in sync with the sections you keep. Choose `language-*` classes on `<code>` to match the source language.
5. **Write the file** to `.change-explainer/<slug>.html`, where `<slug>` is a short kebab-case name for the change (e.g. `auth-session-refactor`). Reuse the same slug when regenerating the same change so it overwrites rather than accumulates.
6. **Open it (lavish-aware).** Open the file for the reader:
   - Default: `open "<path>"` on macOS (`xdg-open` on Linux).
   - If the user wants an annotate / feedback loop and the `lavish` skill or `lavish-axi` CLI is available, offer to open it with `npx -y lavish-axi "<path>"` instead, which serves the same file in a review session.
7. **Give a brief in-chat orientation.** State the big picture in two or three sentences and print the absolute path. Keep the depth in the doc; do not duplicate the full walkthrough in chat unless asked.

## HTML authoring rules

- Keep the doc self-contained: inline CSS and JS only; the sole external dependency is the optional highlight.js CDN, which the template already degrades from gracefully when offline.
- Prevent horizontal overflow at every level. Long unbreakable tokens belong in `<pre><code>` blocks (which scroll), not in prose.
- Keep before/after panels tightly scoped: the changed lines plus just enough context to read them. Excerpt a single branch, signature, or data shape rather than a whole function. For unified diffs, wrap added/removed lines in `<span class="add">` / `<span class="del">` inside a `<pre class="diff">` block.
- Diagrams are ASCII inside `figure.diagram > pre` by default. Use inline SVG only when an annotated picture genuinely helps. Do not pull in Mermaid or other diagram runtimes unless the user asks.
- Do not invent design systems or heavy chrome. The template's typography-led, low-border style is intentional; match it.

## Communication rules

- Optimize for reader understanding over exhaustiveness. Prefer a coherent design explanation over a file-by-file tour.
- Do not collapse the doc into "file A changed X, file B changed Y" without naming the concept those edits serve.
- Do not use review language like "finding" or "severity" unless the user wants review mode.
- When inferring intent or tradeoffs, label it as an inference.
- If the change is tiny or the user wants a quick look, keep the doc short — fewer sections, fewer panels. If they want deep teaching, go further into mechanisms and tradeoffs.

## Example triggers

- "Explain my unstaged changes as an interactive page I can click through."
- "Build an HTML walkthrough of the last commit — big picture first, then the design decisions, with before/after for the parts that matter."
- "I'm reviewing this PR cold. Give me a browsable doc grouped by idea, not a hunk-by-hunk recap."
- "Compare these two file versions and explain the technical decisions, with toggles for the code that changed."

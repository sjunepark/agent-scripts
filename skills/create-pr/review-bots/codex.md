# Codex PR Controls

Use this guide after making the parent skill's AI review decision.

## What a PR can control

- Request a review after creation with an `@codex review` comment, optionally
  followed by a focus such as `@codex review for security issues`.
  `@codex security review` requests a Security Review.
- Codex documents no per-PR way to prevent an automatic review: no label, body
  or title keyword, draft exemption, or pause command. Do not invent one, and
  do not rely on draft status to skip Codex.
- Automatic reviews come from **Automatic review** and **Review trigger** in
  the account's personal preferences and the repository's settings at
  `https://chatgpt.com/settings/code-review`. Those settings span every PR in
  scope, so leave them to the user.

## Apply the selected control

- To suppress Codex on a PR, post no `@codex review` or
  `@codex security review` comment. Report that an automatic review may still
  run if the account or repository enables it.
- `## Code Review Rules` sections in `AGENTS.md` steer what Codex reviews, not
  whether it runs.

## Preserve the review budget

Post a Codex trigger only when the user requested it or a trusted repository
workflow clearly authorizes it for this PR.

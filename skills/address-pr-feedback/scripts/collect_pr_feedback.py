#!/usr/bin/env python3
"""Collect GitHub PR feedback surfaces into local JSON and Markdown files."""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


PR_URL_RE = re.compile(
    r"^https?://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/pull/(?P<number>\d+)(?:[/?#].*)?$"
)
OUTSIDE_DIFF_RE = re.compile(
    r"outside(?: the)? diff|outside diff range|comments outside diff|outside-diff",
    re.IGNORECASE,
)


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Collect issue comments, reviews, review comments, and review bodies for a GitHub PR."
    )
    parser.add_argument(
        "target",
        nargs="?",
        help="PR URL or number. Defaults to the current branch PR.",
    )
    parser.add_argument(
        "--repo",
        help="OWNER/REPO when target is a number and the current directory is not the repository.",
    )
    parser.add_argument(
        "--output-dir",
        default=".tmp/pr-feedback",
        help="Directory for generated report files. Defaults to .tmp/pr-feedback.",
    )
    args = parser.parse_args()

    try:
        owner, repo, number = resolve_pr(args.target, args.repo)
        payload = collect_pr(owner, repo, number)
        output_dir = Path(args.output_dir)
        output_dir.mkdir(parents=True, exist_ok=True)
        stem = f"{owner}-{repo}-{number}"
        json_path = output_dir / f"{stem}.json"
        md_path = output_dir / f"{stem}.md"
        json_path.write_text(json.dumps(payload, indent=2, sort_keys=True), encoding="utf-8")
        md_path.write_text(render_markdown(payload), encoding="utf-8")
    except CommandError as error:
        print(f"error: {error}", file=sys.stderr)
        if error.stderr:
            print(error.stderr.strip(), file=sys.stderr)
        return 1
    except Exception as error:  # noqa: BLE001 - this is a user-facing CLI.
        print(f"error: {error}", file=sys.stderr)
        return 1

    print(f"JSON: {json_path}")
    print(f"Markdown: {md_path}")
    return 0


class CommandError(RuntimeError):
    def __init__(self, message: str, stderr: str = "") -> None:
        super().__init__(message)
        self.stderr = stderr


def resolve_pr(target: str | None, repo_arg: str | None) -> tuple[str, str, int]:
    if target is None:
        url = gh_text(["pr", "view", "--json", "url", "--jq", ".url"]).strip()
        return parse_pr_url(url)

    match = PR_URL_RE.match(target)
    if match:
        return match.group("owner"), match.group("repo"), int(match.group("number"))

    if target.isdigit():
        owner_repo = repo_arg or current_repo()
        owner, repo = split_owner_repo(owner_repo)
        return owner, repo, int(target)

    url = gh_text(["pr", "view", target, "--json", "url", "--jq", ".url"]).strip()
    return parse_pr_url(url)


def parse_pr_url(url: str) -> tuple[str, str, int]:
    match = PR_URL_RE.match(url)
    if not match:
        raise ValueError(f"could not parse GitHub PR URL: {url}")
    return match.group("owner"), match.group("repo"), int(match.group("number"))


def current_repo() -> str:
    value = gh_text(["repo", "view", "--json", "nameWithOwner", "--jq", ".nameWithOwner"]).strip()
    if not value:
        raise ValueError("could not infer current GitHub repository; pass --repo OWNER/REPO")
    return value


def split_owner_repo(value: str) -> tuple[str, str]:
    parts = value.split("/", 1)
    if len(parts) != 2 or not all(parts):
        raise ValueError(f"expected OWNER/REPO, got: {value}")
    return parts[0], parts[1]


def collect_pr(owner: str, repo: str, number: int) -> dict[str, Any]:
    base = f"repos/{owner}/{repo}"
    pr = gh_json(["api", f"{base}/pulls/{number}"])
    issue_comments = gh_json(["api", f"{base}/issues/{number}/comments", "--paginate"])
    reviews = gh_json(["api", f"{base}/pulls/{number}/reviews", "--paginate"])
    review_comments = gh_json(["api", f"{base}/pulls/{number}/comments", "--paginate"])
    commits = gh_json(["api", f"{base}/pulls/{number}/commits", "--paginate"])
    files = gh_json(["api", f"{base}/pulls/{number}/files", "--paginate"])
    review_threads = collect_review_threads(owner, repo, number)

    return {
        "collected_at": datetime.now(timezone.utc).isoformat(),
        "repository": f"{owner}/{repo}",
        "number": number,
        "pr": pr,
        "issue_comments": ensure_list(issue_comments),
        "reviews": ensure_list(reviews),
        "review_comments": ensure_list(review_comments),
        "review_threads": review_threads,
        "commits": ensure_list(commits),
        "files": ensure_list(files),
    }


def collect_review_threads(owner: str, repo: str, number: int) -> dict[str, Any]:
    query = """
query($owner: String!, $repo: String!, $number: Int!, $after: String) {
  repository(owner: $owner, name: $repo) {
    pullRequest(number: $number) {
      reviewThreads(first: 100, after: $after) {
        pageInfo {
          hasNextPage
          endCursor
        }
        nodes {
          id
          isResolved
          isOutdated
          path
          line
          startLine
          originalLine
          originalStartLine
          subjectType
          comments(first: 100) {
            nodes {
              id
              databaseId
              url
              body
              createdAt
              author {
                login
              }
            }
          }
        }
      }
    }
  }
}
"""
    nodes: list[dict[str, Any]] = []
    after: str | None = None
    try:
        while True:
            command = [
                "api",
                "graphql",
                "-f",
                f"query={query}",
                "-f",
                f"owner={owner}",
                "-f",
                f"repo={repo}",
                "-F",
                f"number={number}",
            ]
            if after:
                command.extend(["-f", f"after={after}"])
            data = gh_json(command)
            errors = data.get("errors") or []
            if errors:
                return {"available": False, "nodes": [], "error": summarize_graphql_errors(errors)}
            response_data = data.get("data") or {}
            repository = response_data.get("repository") or {}
            pull_request = repository.get("pullRequest") or {}
            threads = pull_request.get("reviewThreads") or {}
            nodes.extend(threads.get("nodes") or [])
            page_info = threads.get("pageInfo") or {}
            if not page_info.get("hasNextPage"):
                break
            after = page_info.get("endCursor")
            if not after:
                break
        return {"available": True, "nodes": nodes, "error": None}
    except CommandError as error:
        return {"available": False, "nodes": [], "error": str(error)}


def summarize_graphql_errors(errors: list[dict[str, Any]]) -> str:
    messages = [str(error.get("message") or "unknown GraphQL error") for error in errors]
    return "; ".join(messages)


def gh_text(args: list[str]) -> str:
    command = ["gh", *args]
    result = subprocess.run(command, text=True, capture_output=True, check=False)
    if result.returncode != 0:
        raise CommandError(f"command failed: {shell_join(command)}", result.stderr)
    return result.stdout


def gh_json(args: list[str]) -> Any:
    output = gh_text(args).strip()
    if not output:
        return None
    try:
        return json.loads(output)
    except json.JSONDecodeError:
        return parse_json_stream(output)


def parse_json_stream(output: str) -> list[Any]:
    decoder = json.JSONDecoder()
    index = 0
    values: list[Any] = []
    while index < len(output):
        while index < len(output) and output[index].isspace():
            index += 1
        if index >= len(output):
            break
        value, index = decoder.raw_decode(output, index)
        if isinstance(value, list):
            values.extend(value)
        else:
            values.append(value)
    return values


def ensure_list(value: Any) -> list[Any]:
    if value is None:
        return []
    if isinstance(value, list):
        return value
    return [value]


def render_markdown(payload: dict[str, Any]) -> str:
    pr = payload["pr"]
    lines = [
        f"# PR Feedback Report: {payload['repository']}#{payload['number']}",
        "",
        f"- URL: {pr.get('html_url')}",
        f"- Title: {pr.get('title')}",
        f"- State: {pr.get('state')}",
        f"- Base: {ref_name(pr.get('base'))}",
        f"- Head: {ref_name(pr.get('head'))}",
        f"- Latest head SHA: {sha(pr.get('head', {}).get('sha'))}",
        f"- Collected at: {payload['collected_at']}",
        "",
        "## Counts",
        "",
        f"- Issue comments: {len(payload['issue_comments'])}",
        f"- Reviews: {len(payload['reviews'])}",
        f"- Review comments and replies: {len(payload['review_comments'])}",
        f"- Review threads from GraphQL: {len(payload['review_threads'].get('nodes') or [])}"
        if payload["review_threads"].get("available")
        else f"- Review threads from GraphQL: unavailable ({payload['review_threads'].get('error')})",
        f"- Commits: {len(payload['commits'])}",
        f"- Files: {len(payload['files'])}",
        "",
        "The report intentionally does not dedupe findings. Verify each item",
        "against current code, then group duplicates in your working ledger.",
        "",
    ]

    outside_sources = find_outside_diff_sources(payload)
    lines.extend(["## Potential Outside-Diff Sources", ""])
    if outside_sources:
        for source in outside_sources:
            lines.append(f"- {source}")
    else:
        lines.append("- None detected by keyword search.")
    lines.append("")

    lines.extend(render_pr_body(pr))
    lines.extend(render_commits(payload["commits"]))
    lines.extend(render_files(payload["files"]))
    lines.extend(render_issue_comments(payload["issue_comments"]))
    lines.extend(render_reviews(payload["reviews"]))
    lines.extend(render_review_comments(payload["review_comments"]))
    lines.extend(render_review_threads(payload["review_threads"]))
    return "\n".join(lines).rstrip() + "\n"


def render_pr_body(pr: dict[str, Any]) -> list[str]:
    return [
        "## PR Body",
        "",
        pr.get("body") or "_No PR body._",
        "",
    ]


def render_commits(commits: list[dict[str, Any]]) -> list[str]:
    lines = ["## Commits", ""]
    if not commits:
        return lines + ["- None collected.", ""]
    for commit in commits:
        data = commit.get("commit") or {}
        message = (data.get("message") or "").splitlines()[0]
        lines.append(f"- `{sha(commit.get('sha'))}` {message}")
    lines.append("")
    return lines


def render_files(files: list[dict[str, Any]]) -> list[str]:
    lines = ["## Files", ""]
    if not files:
        return lines + ["- None collected.", ""]
    for item in files:
        lines.append(
            f"- `{item.get('filename')}` {item.get('status')} (+{item.get('additions')}/-{item.get('deletions')})"
        )
    lines.append("")
    return lines


def render_issue_comments(comments: list[dict[str, Any]]) -> list[str]:
    lines = ["## Issue Comments", ""]
    if not comments:
        return lines + ["_No issue comments._", ""]
    for comment in comments:
        lines.extend(
            render_body_block(
                title=f"Issue comment {comment.get('id')} by {author(comment)}",
                url=comment.get("html_url"),
                created_at=comment.get("created_at"),
                body=comment.get("body") or "",
            )
        )
    return lines


def render_reviews(reviews: list[dict[str, Any]]) -> list[str]:
    lines = ["## Review Bodies", ""]
    if not reviews:
        return lines + ["_No reviews._", ""]
    for review in reviews:
        title = (
            f"Review {review.get('id')} by {author(review)} "
            f"({review.get('state')}, commit {sha(review.get('commit_id'))})"
        )
        lines.extend(
            render_body_block(
                title=title,
                url=review.get("html_url"),
                created_at=review.get("submitted_at"),
                body=review.get("body") or "",
            )
        )
    return lines


def render_review_comments(comments: list[dict[str, Any]]) -> list[str]:
    lines = ["## Review Comments And Replies", ""]
    if not comments:
        return lines + ["_No review comments._", ""]
    for comment in comments:
        reply_to = comment.get("in_reply_to_id")
        relation = f", reply to {reply_to}" if reply_to else ""
        title = (
            f"Review comment {comment.get('id')} by {author(comment)}"
            f"{relation} on `{comment.get('path')}`"
        )
        line_bits = [
            value
            for value in [
                f"line {comment.get('line')}" if comment.get("line") else None,
                f"original line {comment.get('original_line')}"
                if comment.get("original_line")
                else None,
                f"subject {comment.get('subject_type')}" if comment.get("subject_type") else None,
            ]
            if value
        ]
        lines.extend(
            render_body_block(
                title=title,
                url=comment.get("html_url"),
                created_at=comment.get("created_at"),
                body=comment.get("body") or "",
                extra=", ".join(line_bits),
            )
        )
    return lines


def render_review_threads(review_threads: dict[str, Any]) -> list[str]:
    lines = ["## Review Threads", ""]
    if not review_threads.get("available"):
        return lines + [f"_Unavailable: {review_threads.get('error')}_", ""]
    nodes = review_threads.get("nodes") or []
    if not nodes:
        return lines + ["_No review threads._", ""]
    for thread in nodes:
        status = "resolved" if thread.get("isResolved") else "unresolved"
        outdated = ", outdated" if thread.get("isOutdated") else ""
        lines.append(
            f"### Thread {thread.get('id')} ({status}{outdated}) on `{thread.get('path')}`"
        )
        lines.append("")
        for comment in (thread.get("comments") or {}).get("nodes") or []:
            lines.append(
                f"- Comment databaseId={comment.get('databaseId')} by {nested_author(comment)} at {comment.get('createdAt')}: {comment.get('url')}"
            )
        lines.append("")
    return lines


def render_body_block(
    *,
    title: str,
    url: str | None,
    created_at: str | None,
    body: str,
    extra: str = "",
) -> list[str]:
    lines = [f"### {title}", ""]
    metadata = [f"URL: {url}" if url else None, f"Created: {created_at}" if created_at else None, extra]
    lines.extend([item for item in metadata if item])
    lines.append("")
    lines.append(body or "_No body._")
    lines.extend(["", "---", ""])
    return lines


def find_outside_diff_sources(payload: dict[str, Any]) -> list[str]:
    sources: list[str] = []
    for review in payload["reviews"]:
        if OUTSIDE_DIFF_RE.search(review.get("body") or ""):
            sources.append(f"review {review.get('id')} by {author(review)}: {review.get('html_url')}")
    for comment in payload["issue_comments"]:
        if OUTSIDE_DIFF_RE.search(comment.get("body") or ""):
            sources.append(f"issue comment {comment.get('id')} by {author(comment)}: {comment.get('html_url')}")
    for comment in payload["review_comments"]:
        if OUTSIDE_DIFF_RE.search(comment.get("body") or ""):
            sources.append(
                f"review comment {comment.get('id')} by {author(comment)}: {comment.get('html_url')}"
            )
    return sources


def author(item: dict[str, Any]) -> str:
    user = item.get("user") or {}
    return user.get("login") or "unknown"


def nested_author(item: dict[str, Any]) -> str:
    user = item.get("author") or {}
    return user.get("login") or "unknown"


def ref_name(ref: dict[str, Any] | None) -> str:
    if not ref:
        return "unknown"
    label = ref.get("label")
    ref_value = ref.get("ref")
    return label or ref_value or "unknown"


def sha(value: str | None) -> str:
    if not value:
        return "unknown"
    return value[:12]


def shell_join(command: list[str]) -> str:
    return " ".join(sh_quote(part) for part in command)


def sh_quote(value: str) -> str:
    if re.fullmatch(r"[A-Za-z0-9_./:=@%+-]+", value):
        return value
    return "'" + value.replace("'", "'\"'\"'") + "'"


if __name__ == "__main__":
    raise SystemExit(main())

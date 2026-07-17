#!/usr/bin/env python3
"""Resolve and safely clean project-local, session-scoped checkpoint files."""

from __future__ import annotations

import hashlib
import os
from pathlib import Path
import re
import stat
import subprocess
import sys
from typing import Any


MARKER_ENV = "CODEX_MEMO_SESSION_MARKER"
MARKER_PREFIX = "codex.runtime.session_id.v1="
THREAD_ENV = "CODEX_THREAD_ID"
MEMO_DIRECTORY = Path(".tmp", "memo")
MEMO_NAME = re.compile(r"session-[0-9a-f]{64}\.md\Z")
COMMANDS = frozenset({"path", "prepare", "clean"})


class MemoError(Exception):
    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code
        self.message = message


def encode_toon_string(value: object) -> str:
    encoded = []
    for character in str(value):
        codepoint = ord(character)
        if character == "\\":
            encoded.append("\\\\")
        elif character == '"':
            encoded.append('\\"')
        elif character == "\n":
            encoded.append("\\n")
        elif character == "\r":
            encoded.append("\\r")
        elif character == "\t":
            encoded.append("\\t")
        elif codepoint < 0x20:
            encoded.append(f"\\u{codepoint:04x}")
        else:
            encoded.append(character)
    return f'"{"".join(encoded)}"'


def emit(fields: dict[str, Any]) -> None:
    lines = []
    for key, value in fields.items():
        if value is None:
            encoded = "null"
        elif isinstance(value, bool):
            encoded = "true" if value else "false"
        elif isinstance(value, int):
            encoded = str(value)
        else:
            encoded = encode_toon_string(value)
        lines.append(f"{key}: {encoded}")
    sys.stdout.write("\n".join(lines))


def clean_git_environment() -> dict[str, str]:
    environment = os.environ.copy()
    for name in ("GIT_DIR", "GIT_WORK_TREE", "GIT_COMMON_DIR", "GIT_INDEX_FILE"):
        environment.pop(name, None)
    return environment


def has_git_marker(directory: Path) -> bool:
    parents = (directory, *directory.parents)
    return any((parent / ".git").exists() or (parent / ".git").is_symlink() for parent in parents)


def resolve_project() -> tuple[Path, bool]:
    current = Path.cwd().resolve()
    try:
        result = subprocess.run(
            ["git", "rev-parse", "--show-toplevel"],
            cwd=current,
            env=clean_git_environment(),
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            check=False,
        )
    except OSError:
        if has_git_marker(current):
            raise MemoError("git-unavailable", "cannot resolve the current Git project")
        return current, False

    if result.returncode != 0:
        if has_git_marker(current):
            raise MemoError("git-root-failed", "cannot resolve the current Git project")
        return current, False

    root_text = result.stdout.strip()
    if not root_text:
        raise MemoError("git-root-failed", "Git returned an empty project root")
    root = Path(root_text).resolve()
    try:
        current.relative_to(root)
    except ValueError as error:
        raise MemoError("unsafe-project-root", "Git resolved a project outside the current directory") from error
    if not root.is_dir():
        raise MemoError("git-root-failed", "the resolved Git project root is not a directory")
    return root, True


def resolve_session_id() -> str:
    marker = os.environ.get(MARKER_ENV)
    if marker is not None:
        if "\n" in marker or "\r" in marker or not marker.startswith(MARKER_PREFIX):
            raise MemoError("invalid-session-marker", "the injected session marker is malformed")
        session_id = marker[len(MARKER_PREFIX) :]
        if not session_id.strip():
            raise MemoError("invalid-session-marker", "the injected session marker is empty")
        return session_id

    session_id = os.environ.get(THREAD_ENV)
    if session_id is None or not session_id.strip():
        raise MemoError("missing-session-id", "no injected session marker or CODEX_THREAD_ID is available")
    if "\n" in session_id or "\r" in session_id:
        raise MemoError("invalid-session-id", "CODEX_THREAD_ID contains a line break")
    return session_id


def memo_path(root: Path) -> Path:
    try:
        digest = hashlib.sha256(resolve_session_id().encode("utf-8")).hexdigest()
    except UnicodeEncodeError as error:
        raise MemoError("invalid-session-id", "the session identifier is not valid UTF-8") from error
    return root / MEMO_DIRECTORY / f"session-{digest}.md"


def lstat_or_none(path: Path) -> os.stat_result | None:
    try:
        return path.lstat()
    except FileNotFoundError:
        return None


def validate_directory(path: Path) -> bool:
    metadata = lstat_or_none(path)
    if metadata is None:
        return False
    if stat.S_ISLNK(metadata.st_mode) or not stat.S_ISDIR(metadata.st_mode):
        raise MemoError("unsafe-path", f"refusing non-directory or symlink path: {path}")
    return True


def validate_session_path(root: Path, target: Path) -> None:
    try:
        target.relative_to(root)
    except ValueError as error:
        raise MemoError("unsafe-path", "the session path escapes the current project") from error

    validate_directory(root / ".tmp")
    validate_directory(root / MEMO_DIRECTORY)
    metadata = lstat_or_none(target)
    if metadata is not None and (stat.S_ISLNK(metadata.st_mode) or not stat.S_ISREG(metadata.st_mode)):
        raise MemoError("unsafe-path", f"refusing non-regular or symlink memo path: {target}")


def check_ignored(root: Path, target: Path) -> bool:
    relative = target.relative_to(root).as_posix()
    try:
        result = subprocess.run(
            ["git", "check-ignore", "--no-index", "--quiet", "--", relative],
            cwd=root,
            env=clean_git_environment(),
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False,
        )
    except OSError as error:
        raise MemoError("git-unavailable", "cannot verify Git ignore coverage") from error
    if result.returncode == 0:
        return True
    if result.returncode == 1:
        return False
    raise MemoError("git-ignore-failed", "git check-ignore could not verify the memo path")


def check_tracked(root: Path, target: Path) -> bool:
    relative = target.relative_to(root).as_posix()
    try:
        result = subprocess.run(
            ["git", "ls-files", "--error-unmatch", "--", relative],
            cwd=root,
            env=clean_git_environment(),
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            check=False,
        )
    except OSError as error:
        raise MemoError("git-unavailable", "cannot verify that the memo path is untracked") from error
    if result.returncode == 0:
        return True
    if result.returncode == 1:
        return False
    raise MemoError("git-index-failed", "git ls-files could not verify the memo path")


def create_directory(path: Path) -> bool:
    created = False
    try:
        path.mkdir(mode=0o700)
        created = True
    except FileExistsError:
        pass
    validate_directory(path)
    return created


def path_fields(command: str, root: Path, target: Path) -> dict[str, Any]:
    metadata = lstat_or_none(target)
    return {
        "status": "ok",
        "command": command,
        "project_root": root,
        "path": target,
        "exists": metadata is not None and stat.S_ISREG(metadata.st_mode),
    }


def show_status() -> None:
    root, in_git = resolve_project()
    target = memo_path(root)
    validate_session_path(root, target)
    fields = path_fields("status", root, target)
    fields.update(
        {
            "git_repository": in_git,
            "memo_directory_exists": (root / MEMO_DIRECTORY).is_dir(),
            "ignored": check_ignored(root, target) if in_git else None,
        }
    )
    emit(fields)


def show_path() -> None:
    root, _ = resolve_project()
    target = memo_path(root)
    validate_session_path(root, target)
    emit(path_fields("path", root, target))


def prepare() -> None:
    root, in_git = resolve_project()
    target = memo_path(root)
    validate_session_path(root, target)
    if in_git and not check_ignored(root, target):
        raise MemoError(
            "not-ignored",
            "Git requires the .tmp/memo target to be ignored before preparation",
        )
    if in_git and check_tracked(root, target):
        raise MemoError(
            "tracked-target",
            "refusing to prepare a memo path that is already tracked by Git",
        )

    create_directory(root / ".tmp")
    created = create_directory(root / MEMO_DIRECTORY)
    validate_session_path(root, target)
    fields = path_fields("prepare", root, target)
    fields.update({"created_directory": created, "ignored": True if in_git else None})
    emit(fields)


def clean() -> None:
    root, _ = resolve_project()
    temporary = root / ".tmp"
    directory = root / MEMO_DIRECTORY
    if not validate_directory(temporary) or not validate_directory(directory):
        emit({"status": "ok", "command": "clean", "project_root": root, "removed": 0})
        return

    removed = 0
    with os.scandir(directory) as entries:
        for entry in entries:
            if not MEMO_NAME.fullmatch(entry.name) or not entry.is_file(follow_symlinks=False):
                continue
            try:
                metadata = entry.stat(follow_symlinks=False)
                if not stat.S_ISREG(metadata.st_mode):
                    continue
                os.unlink(entry.path)
                removed += 1
            except FileNotFoundError:
                continue
    emit({"status": "ok", "command": "clean", "project_root": root, "removed": removed})


def show_help(command: str | None = None) -> None:
    if command is None:
        emit(
            {
                "status": "ok",
                "command": "help",
                "usage": "memo.py [path|prepare|clean]",
                "summary": "Show status, resolve a session path, prepare it, or clean project memos",
            }
        )
        return
    summaries = {
        "path": "Resolve the current session memo path without mutation",
        "prepare": "Verify safety and Git ignore coverage, then prepare the memo directory",
        "clean": "Remove recognized regular session memo files in the current project",
    }
    emit(
        {
            "status": "ok",
            "command": command,
            "usage": f"memo.py {command}",
            "summary": summaries[command],
        }
    )


def invalid_usage(message: str) -> int:
    emit(
        {
            "status": "error",
            "code": "invalid-usage",
            "message": message,
            "usage": "memo.py [path|prepare|clean]",
        }
    )
    return 2


def run(arguments: list[str]) -> int:
    if not arguments:
        show_status()
        return 0
    if arguments in (["--help"], ["-h"]):
        show_help()
        return 0

    command = arguments[0]
    if command not in COMMANDS:
        return invalid_usage(f"unknown command: {command}")
    if len(arguments) == 2 and arguments[1] in {"--help", "-h"}:
        show_help(command)
        return 0
    if len(arguments) != 1:
        return invalid_usage(f"unexpected arguments for {command}")

    {"path": show_path, "prepare": prepare, "clean": clean}[command]()
    return 0


def main() -> int:
    try:
        return run(sys.argv[1:])
    except MemoError as error:
        emit({"status": "error", "code": error.code, "message": error.message})
        return 1
    except OSError:
        emit(
            {
                "status": "error",
                "code": "filesystem-error",
                "message": "a local filesystem operation failed",
            }
        )
        return 1
    except Exception:
        # The CLI boundary must never leak a traceback or dependency output.
        emit(
            {
                "status": "error",
                "code": "unexpected-error",
                "message": "the memo helper failed unexpectedly",
            }
        )
        return 1


if __name__ == "__main__":
    raise SystemExit(main())

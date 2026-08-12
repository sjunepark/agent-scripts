#!/usr/bin/env python3
"""Read-only inventory of Codex local storage and runtime processes."""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
from pathlib import Path
import shutil
import sqlite3
import struct
import subprocess
import sys
import tempfile
from typing import Any


HISTORY_SUFFIXES = (".jsonl", ".jsonl.zst")
OWNER_EXECUTABLE_NAMES = ("codex", "codex.exe", "chatgpt", "chatgpt.exe")


def format_bytes(value: int | None) -> str:
    if value is None:
        return "unknown"
    units = ("B", "KiB", "MiB", "GiB", "TiB")
    number = float(value)
    for unit in units:
        if abs(number) < 1024 or unit == units[-1]:
            return f"{number:.1f} {unit}"
        number /= 1024
    return f"{number:.1f} TiB"


def iso_time(timestamp: float | None) -> str | None:
    if timestamp is None:
        return None
    return dt.datetime.fromtimestamp(timestamp, tz=dt.timezone.utc).isoformat()


def path_stats(path: Path) -> dict[str, Any]:
    result: dict[str, Any] = {
        "path": str(path),
        "bytes": 0,
        "files": 0,
        "directories": 0,
        "symlinks": 0,
        "oldest_mtime": None,
        "newest_mtime": None,
        "errors": [],
    }
    stack = [path]
    oldest: float | None = None
    newest: float | None = None
    while stack:
        current = stack.pop()
        try:
            stat = current.lstat()
            if current.is_symlink():
                result["symlinks"] += 1
                result["bytes"] += stat.st_size
                continue
            if current.is_dir():
                result["directories"] += 1
                with os.scandir(current) as entries:
                    stack.extend(Path(entry.path) for entry in entries)
                continue
            result["files"] += 1
            result["bytes"] += stat.st_size
            oldest = stat.st_mtime if oldest is None else min(oldest, stat.st_mtime)
            newest = stat.st_mtime if newest is None else max(newest, stat.st_mtime)
        except OSError as error:
            result["errors"].append(f"{current}: {error}")
    result["oldest_mtime"] = iso_time(oldest)
    result["newest_mtime"] = iso_time(newest)
    return result


def history_stats(path: Path, now: float) -> dict[str, Any]:
    result = {
        "path": str(path),
        "exists": path.is_dir(),
        "files": 0,
        "bytes": 0,
        "older_than_days": {"7": {"files": 0, "bytes": 0}, "14": {"files": 0, "bytes": 0}, "30": {"files": 0, "bytes": 0}},
        "oldest_mtime": None,
        "newest_mtime": None,
        "errors": [],
    }
    if not path.is_dir():
        return result
    oldest: float | None = None
    newest: float | None = None
    for root, _, files in os.walk(path, followlinks=False):
        for name in files:
            if not name.endswith(HISTORY_SUFFIXES):
                continue
            item = Path(root, name)
            try:
                stat = item.lstat()
            except OSError as error:
                result["errors"].append(f"{item}: {error}")
                continue
            result["files"] += 1
            result["bytes"] += stat.st_size
            oldest = stat.st_mtime if oldest is None else min(oldest, stat.st_mtime)
            newest = stat.st_mtime if newest is None else max(newest, stat.st_mtime)
            age_days = (now - stat.st_mtime) / 86400
            for days in (7, 14, 30):
                if age_days >= days:
                    bucket = result["older_than_days"][str(days)]
                    bucket["files"] += 1
                    bucket["bytes"] += stat.st_size
    result["oldest_mtime"] = iso_time(oldest)
    result["newest_mtime"] = iso_time(newest)
    return result


def sqlite_stats(path: Path, home: Path, quick_check: bool, runtime_active: bool) -> dict[str, Any]:
    result: dict[str, Any] = {
        "path": str(path),
        "exists": False,
        "main_bytes": 0,
        "wal_bytes": 0,
        "shm_bytes": 0,
    }
    if path.is_symlink():
        result["error"] = "Refusing to inspect a symlinked database."
        return result
    try:
        resolved_home = home.resolve(strict=True)
        resolved_path = path.resolve(strict=True)
        resolved_path.relative_to(resolved_home)
        stat = path.stat()
        result["exists"] = path.is_file()
        result["main_bytes"] = stat.st_size
        for suffix, key in (("-wal", "wal_bytes"), ("-shm", "shm_bytes")):
            sidecar = Path(f"{path}{suffix}")
            if sidecar.is_symlink():
                raise OSError(f"Refusing symlinked SQLite sidecar: {sidecar}")
            if sidecar.is_file():
                sidecar.resolve(strict=True).relative_to(resolved_home)
                result[key] = sidecar.stat().st_size
        with path.open("rb") as handle:
            header = handle.read(100)
        if len(header) < 100 or header[:16] != b"SQLite format 3\x00":
            raise ValueError("Not a complete SQLite 3 database header")
        page_size = struct.unpack(">H", header[16:18])[0]
        if page_size == 1:
            page_size = 65536
        page_count = struct.unpack(">I", header[28:32])[0]
        freelist_count = struct.unpack(">I", header[36:40])[0]
        result.update({
            "page_size": page_size,
            "page_count": page_count,
            "freelist_count": freelist_count,
            "estimated_free_bytes": page_size * freelist_count,
            "estimated_free_fraction": freelist_count / page_count if page_count else 0,
            "read_scope": "SQLite main-file header; live WAL frames excluded",
        })
        if quick_check:
            if runtime_active:
                result["quick_check_skipped"] = "Matching Codex processes are active; retry after quiescing them."
            else:
                with tempfile.TemporaryDirectory(prefix="codex-cleanup-sqlite-check.") as temporary:
                    snapshot = Path(temporary, path.name)
                    shutil.copy2(path, snapshot)
                    for suffix in ("-wal", "-shm"):
                        sidecar = Path(f"{path}{suffix}")
                        if sidecar.is_file():
                            shutil.copy2(sidecar, Path(f"{snapshot}{suffix}"))
                    connection = sqlite3.connect(snapshot, timeout=2)
                    try:
                        result["quick_check"] = [row[0] for row in connection.execute("PRAGMA quick_check")]
                        result["quick_check_scope"] = "temporary offline snapshot including copied sidecars"
                    finally:
                        connection.close()
    except (OSError, sqlite3.Error, ValueError) as error:
        result["error"] = str(error)
    return result


def release_stats(home: Path) -> dict[str, Any]:
    root = home / "packages" / "standalone" / "releases"
    current_link = home / "packages" / "standalone" / "current"
    result: dict[str, Any] = {
        "path": str(root),
        "exists": root.is_dir(),
        "current_link": str(current_link),
        "current_target": None,
        "releases": [],
        "default_keep": [],
        "default_candidate_bytes": None,
        "errors": [],
    }
    try:
        if current_link.exists() or current_link.is_symlink():
            result["current_target"] = str(current_link.resolve(strict=False))
    except OSError as error:
        result["errors"].append(f"{current_link}: {error}")
    if not root.is_dir():
        return result
    releases = []
    for item in root.iterdir():
        if not item.is_dir() or item.is_symlink():
            continue
        stats = path_stats(item)
        try:
            mtime = item.stat().st_mtime
        except OSError:
            mtime = 0
        complete = False
        completeness_error = None
        manifest_path = item / "codex-package.json"
        try:
            if manifest_path.is_symlink():
                raise ValueError("manifest is a symlink")
            manifest = json.loads(manifest_path.read_text())
            if not isinstance(manifest, dict):
                raise ValueError("manifest is not an object")
            entrypoint_value = manifest.get("entrypoint")
            if not isinstance(entrypoint_value, str) or not entrypoint_value:
                raise ValueError("manifest has no entrypoint")
            entrypoint = (item / entrypoint_value).resolve(strict=True)
            entrypoint.relative_to(item.resolve(strict=True))
            if not entrypoint.is_file() or not os.access(entrypoint, os.X_OK):
                raise ValueError("manifest entrypoint is not an executable file")
            complete = True
        except (OSError, ValueError, json.JSONDecodeError) as error:
            completeness_error = str(error)
        releases.append({
            "name": item.name,
            "path": str(item.resolve()),
            "bytes": stats["bytes"],
            "mtime": iso_time(mtime),
            "mtime_epoch": mtime,
            "complete": complete,
            "completeness_error": completeness_error,
            "errors": stats["errors"],
        })
    releases.sort(key=lambda item: (item["mtime_epoch"], item["name"]), reverse=True)
    current_target = result["current_target"]
    keep: list[str] = []
    if not current_target:
        result["errors"].append("Standalone current release could not be resolved; no release-pruning estimate is safe.")
    elif any(item["path"] == current_target and item["complete"] for item in releases):
        keep.append(current_target)
    else:
        result["errors"].append(f"Standalone current target is absent or incomplete: {current_target}")
    if not keep:
        for item in releases:
            item.pop("mtime_epoch", None)
        result["releases"] = releases
        return result
    for item in releases:
        if item["complete"] and item["path"] not in keep:
            keep.append(item["path"])
            break
    if len(keep) < 2:
        result["errors"].append("No complete rollback release was found; no release-pruning estimate is safe.")
        for item in releases:
            item.pop("mtime_epoch", None)
        result["releases"] = releases
        return result
    result["default_keep"] = keep
    result["default_candidate_bytes"] = sum(item["bytes"] for item in releases if item["complete"] and item["path"] not in keep)
    for item in releases:
        item.pop("mtime_epoch", None)
    result["releases"] = releases
    return result


def unix_processes() -> tuple[list[dict[str, Any]], str | None]:
    try:
        completed = subprocess.run(
            ["ps", "-axo", "pid=,ppid=,rss=,comm="],
            check=True,
            capture_output=True,
            text=True,
        )
    except (OSError, subprocess.CalledProcessError) as error:
        return [], str(error)
    all_processes: dict[int, dict[str, Any]] = {}
    children: dict[int, list[int]] = {}
    for line in completed.stdout.splitlines():
        fields = line.strip().split(None, 3)
        if len(fields) != 4:
            continue
        try:
            pid, ppid, rss_kib = map(int, fields[:3])
        except ValueError:
            continue
        process = {"pid": pid, "ppid": ppid, "rss_kib": rss_kib, "executable": fields[3]}
        all_processes[pid] = process
        children.setdefault(ppid, []).append(pid)
    candidates = {
        pid
        for pid, process in all_processes.items()
        if Path(process["executable"]).name.lower() in OWNER_EXECUTABLE_NAMES
    }
    roots = set(candidates)
    for pid in candidates:
        parent = all_processes.get(pid, {}).get("ppid", 0)
        seen: set[int] = set()
        while parent and parent not in seen:
            if parent in candidates:
                roots.discard(pid)
                break
            seen.add(parent)
            parent = all_processes.get(parent, {}).get("ppid", 0)
    selected = set(roots)
    stack = list(roots)
    while stack:
        parent = stack.pop()
        for child in children.get(parent, []):
            if child not in selected:
                selected.add(child)
                stack.append(child)
    ancestry: set[int] = set()
    cursor = os.getpid()
    while cursor and cursor not in ancestry:
        ancestry.add(cursor)
        cursor = all_processes.get(cursor, {}).get("ppid", 0)
    processes = []
    for pid in sorted(selected):
        process = dict(all_processes[pid])
        process["is_matching_root"] = pid in roots
        process["is_current_ancestry"] = pid in ancestry
        processes.append(process)
    return processes, None


def windows_processes() -> tuple[list[dict[str, Any]], str | None]:
    return [], "Windows process-tree, ancestry, and RSS inspection is unavailable; treat runtime cleanup and database quick-check as blocked."


def process_stats() -> dict[str, Any]:
    processes, error = windows_processes() if os.name == "nt" else unix_processes()
    return {
        "processes": processes,
        "total_rss_kib": sum(item["rss_kib"] or 0 for item in processes),
        "error": error,
        "note": "Executable names only; command arguments are intentionally omitted.",
    }


def filesystem_stats(path: Path) -> dict[str, Any]:
    probe = path
    while not probe.exists() and probe != probe.parent:
        probe = probe.parent
    try:
        usage = shutil.disk_usage(probe)
    except OSError as error:
        return {"probe_path": str(probe), "error": str(error)}
    return {
        "probe_path": str(probe),
        "total_bytes": usage.total,
        "used_bytes": usage.used,
        "free_bytes": usage.free,
    }


def build_report(home: Path, quick_check: bool) -> dict[str, Any]:
    now = dt.datetime.now(tz=dt.timezone.utc)
    top_level = []
    errors = []
    if home.is_dir():
        for item in home.iterdir():
            stats = path_stats(item)
            top_level.append({"name": item.name, **stats})
            errors.extend(stats["errors"])
        top_level.sort(key=lambda item: item["bytes"], reverse=True)
    else:
        errors.append(f"Codex home does not exist or is not a directory: {home}")
    database_paths = sorted(home.glob("*.sqlite")) if home.is_dir() else []
    runtime = process_stats()
    report = {
        "schema_version": 1,
        "generated_at": now.isoformat(),
        "codex_home": str(home),
        "exists": home.is_dir(),
        "total_bytes": sum(item["bytes"] for item in top_level),
        "filesystem": filesystem_stats(home),
        "top_level": top_level,
        "history": {
            "sessions": history_stats(home / "sessions", now.timestamp()),
            "archived_sessions": history_stats(home / "archived_sessions", now.timestamp()),
        },
        "databases": [sqlite_stats(path, home, quick_check, bool(runtime["processes"]) or bool(runtime["error"])) for path in database_paths],
        "standalone": release_stats(home),
        "runtime": runtime,
        "errors": errors,
    }
    return report


def print_report(report: dict[str, Any]) -> None:
    print(f"Codex home: {report['codex_home']}")
    print(f"Measured size: {format_bytes(report['total_bytes'])}")
    filesystem = report["filesystem"]
    if filesystem.get("error"):
        print(f"Filesystem headroom: unavailable ({filesystem['error']})")
    else:
        print(f"Filesystem headroom: {format_bytes(filesystem['free_bytes'])} free of {format_bytes(filesystem['total_bytes'])}")
    print("\nLargest top-level entries:")
    for item in report["top_level"][:12]:
        print(f"  {item['name']:<28} {format_bytes(item['bytes']):>10}  {item['files']} files")
    print("\nTask history:")
    for name, item in report["history"].items():
        old = item["older_than_days"]["14"]
        print(f"  {name:<20} {format_bytes(item['bytes']):>10}  {item['files']} files; {format_bytes(old['bytes'])} older than 14 days")
    if report["databases"]:
        print("\nSQLite databases:")
        for item in report["databases"]:
            free = format_bytes(item.get("estimated_free_bytes"))
            sidecars = item.get("wal_bytes", 0) + item.get("shm_bytes", 0)
            detail = f"free pages ~{free}; sidecars {format_bytes(sidecars)}"
            if item.get("error"):
                detail = f"read error: {item['error']}"
            print(f"  {Path(item['path']).name:<24} {format_bytes(item['main_bytes']):>10}  {detail}")
            if item.get("read_scope"):
                print(f"    note: {item['read_scope']}")
            if item.get("quick_check_skipped"):
                print(f"    quick-check skipped: {item['quick_check_skipped']}")
    standalone = report["standalone"]
    if standalone["exists"]:
        print("\nStandalone releases:")
        complete_count = sum(1 for item in standalone["releases"] if item["complete"])
        print(f"  releases: {len(standalone['releases'])} ({complete_count} complete); current: {standalone['current_target'] or 'unresolved'}")
        print(f"  candidate bytes if keeping current + one rollback: {format_bytes(standalone['default_candidate_bytes'])}")
        for error in standalone["errors"]:
            print(f"  warning: {error}")
    runtime = report["runtime"]
    print("\nRuntime processes:")
    print(f"  matched tree: {len(runtime['processes'])} processes; RSS {format_bytes(runtime['total_rss_kib'] * 1024)}")
    displayed = list(runtime["processes"][:20])
    displayed_pids = {item["pid"] for item in displayed}
    displayed.extend(
        item
        for item in runtime["processes"]
        if item["pid"] not in displayed_pids and (item["is_matching_root"] or item["is_current_ancestry"])
    )
    for item in displayed:
        flags = []
        if item["is_matching_root"]:
            flags.append("root")
        if item["is_current_ancestry"]:
            flags.append("current-ancestry")
        print(f"  pid {item['pid']:<7} ppid {str(item['ppid']):<7} rss {format_bytes((item['rss_kib'] or 0) * 1024):>10}  {item['executable']} {' '.join(flags)}")
    if len(displayed) < len(runtime["processes"]):
        print(f"  ... {len(runtime['processes']) - len(displayed)} additional descendants omitted; use --json before PID-level decisions")
    if runtime["error"]:
        print(f"  process inventory unavailable: {runtime['error']}")
    errors = report["errors"]
    if errors:
        print(f"\nWarnings: {len(errors)} scan errors; use --json for details.")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--codex-home", type=Path, help="Codex home to inspect; defaults to CODEX_HOME or ~/.codex")
    parser.add_argument("--json", action="store_true", help="Emit machine-readable JSON")
    parser.add_argument("--quick-check", action="store_true", help="Run SQLite PRAGMA quick_check; may take additional time")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    configured_home = args.codex_home or (Path(os.environ["CODEX_HOME"]) if os.environ.get("CODEX_HOME") else Path.home() / ".codex")
    home = configured_home.expanduser().resolve(strict=False)
    report = build_report(home, args.quick_check)
    if args.json:
        json.dump(report, sys.stdout, indent=2, sort_keys=True)
        print()
    else:
        print_report(report)
    return 0 if report["exists"] else 2


if __name__ == "__main__":
    raise SystemExit(main())

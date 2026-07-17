from __future__ import annotations

import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import unittest


PLUGIN_ROOT = Path(__file__).resolve().parents[1]
HELPER = PLUGIN_ROOT / "scripts" / "memo.py"
MARKER_PREFIX = "codex.runtime.session_id.v1="


def parse_toon(output: str) -> dict[str, object]:
    parsed: dict[str, object] = {}
    for line in output.splitlines():
        key, separator, value = line.partition(": ")
        if not separator:
            raise AssertionError(f"invalid TOON line: {line!r}")
        parsed[key] = json.loads(value)
    return parsed


class MemoHelperTests(unittest.TestCase):
    def base_environment(self) -> dict[str, str]:
        environment = os.environ.copy()
        for name in ("CODEX_MEMO_SESSION_MARKER", "CODEX_THREAD_ID", "GIT_DIR", "GIT_WORK_TREE"):
            environment.pop(name, None)
        environment["GIT_CONFIG_GLOBAL"] = os.devnull
        environment["GIT_CONFIG_NOSYSTEM"] = "1"
        return environment

    def run_helper(
        self,
        cwd: Path,
        *arguments: str,
        marker: str | None = None,
        thread_id: str | None = None,
    ) -> subprocess.CompletedProcess[str]:
        environment = self.base_environment()
        if marker is not None:
            environment["CODEX_MEMO_SESSION_MARKER"] = marker
        if thread_id is not None:
            environment["CODEX_THREAD_ID"] = thread_id
        return subprocess.run(
            [sys.executable, str(HELPER), *arguments],
            cwd=cwd,
            env=environment,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
        )

    def init_git(self, root: Path) -> None:
        subprocess.run(
            ["git", "init", "--quiet"],
            cwd=root,
            env=self.base_environment(),
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=True,
        )

    def expected_path(self, root: Path, session_id: str) -> Path:
        digest = hashlib.sha256(session_id.encode("utf-8")).hexdigest()
        return root.resolve() / ".tmp" / "memo" / f"session-{digest}.md"

    def assert_toon_contract(self, result: subprocess.CompletedProcess[str], returncode: int) -> dict[str, object]:
        self.assertEqual(result.returncode, returncode)
        self.assertEqual(result.stderr, "")
        self.assertFalse(result.stdout.endswith("\n"))
        return parse_toon(result.stdout)

    def test_paths_are_stable_distinct_and_marker_first(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            marker = f"{MARKER_PREFIX}session-a"
            first = self.run_helper(root, "path", marker=marker, thread_id="fallback-session")
            second = self.run_helper(root, "path", marker=marker, thread_id="different-fallback")
            other = self.run_helper(root, "path", marker=f"{MARKER_PREFIX}session-b")
            fallback = self.run_helper(root, "path", thread_id="fallback-session")

            first_fields = self.assert_toon_contract(first, 0)
            second_fields = self.assert_toon_contract(second, 0)
            other_fields = self.assert_toon_contract(other, 0)
            fallback_fields = self.assert_toon_contract(fallback, 0)
            self.assertEqual(first_fields["path"], str(self.expected_path(root, "session-a")))
            self.assertEqual(second_fields["path"], first_fields["path"])
            self.assertNotEqual(other_fields["path"], first_fields["path"])
            self.assertEqual(fallback_fields["path"], str(self.expected_path(root, "fallback-session")))
            self.assertFalse((root / ".tmp").exists())

    def test_invalid_present_marker_never_falls_back(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            result = self.run_helper(Path(temporary), "path", marker="malformed", thread_id="valid-fallback")
            fields = self.assert_toon_contract(result, 1)
            self.assertEqual(fields["code"], "invalid-session-marker")
            self.assertNotIn("valid-fallback", result.stdout)

    def test_git_root_and_non_git_current_directory(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            non_git = Path(temporary) / "plain" / "nested"
            non_git.mkdir(parents=True)
            result = self.run_helper(non_git, "path", marker=f"{MARKER_PREFIX}plain")
            fields = self.assert_toon_contract(result, 0)
            self.assertEqual(fields["project_root"], str(non_git.resolve()))

        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.init_git(root)
            nested = root / "a" / "b"
            nested.mkdir(parents=True)
            result = self.run_helper(nested, "path", marker=f"{MARKER_PREFIX}git")
            fields = self.assert_toon_contract(result, 0)
            self.assertEqual(fields["project_root"], str(root.resolve()))

    def test_traversal_like_identity_is_only_used_as_a_full_hash(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            session_id = "../../outside"
            result = self.run_helper(root, "path", marker=f"{MARKER_PREFIX}{session_id}")
            fields = self.assert_toon_contract(result, 0)
            path = Path(str(fields["path"]))
            self.assertEqual(path, self.expected_path(root, session_id))
            self.assertRegex(path.name, r"^session-[0-9a-f]{64}\.md$")
            self.assertNotIn(session_id, result.stdout)
            self.assertEqual(path.parent, root.resolve() / ".tmp" / "memo")

    def test_prepare_refuses_unignored_git_path_without_artifacts(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.init_git(root)
            result = self.run_helper(root, "prepare", marker=f"{MARKER_PREFIX}session")
            fields = self.assert_toon_contract(result, 1)
            self.assertEqual(fields["code"], "not-ignored")
            self.assertFalse((root / ".tmp").exists())
            self.assertFalse((root / ".gitignore").exists())

    def test_prepare_refuses_an_ignored_but_tracked_target(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            self.init_git(root)
            (root / ".gitignore").write_text(".tmp/memo/\n", encoding="utf-8")
            marker = f"{MARKER_PREFIX}tracked-session"
            path_result = self.run_helper(root, "path", marker=marker)
            target = Path(str(self.assert_toon_contract(path_result, 0)["path"]))
            target.parent.mkdir(parents=True)
            target.write_text("tracked memo", encoding="utf-8")
            subprocess.run(
                ["git", "add", "--force", str(target.relative_to(root.resolve()))],
                cwd=root,
                env=self.base_environment(),
                check=True,
            )

            result = self.run_helper(root, "prepare", marker=marker)
            fields = self.assert_toon_contract(result, 1)
            self.assertEqual(fields["code"], "tracked-target")
            self.assertEqual(target.read_text(encoding="utf-8"), "tracked memo")

    def test_prepare_honors_all_git_ignore_sources_and_is_idempotent(self) -> None:
        for source in ("gitignore", "info-exclude", "core-excludes-file"):
            with self.subTest(source=source), tempfile.TemporaryDirectory() as temporary:
                root = Path(temporary)
                self.init_git(root)
                if source == "gitignore":
                    (root / ".gitignore").write_text(".tmp/memo/\n", encoding="utf-8")
                elif source == "info-exclude":
                    (root / ".git" / "info" / "exclude").write_text(".tmp/memo/\n", encoding="utf-8")
                else:
                    excludes = root / "memo-excludes"
                    excludes.write_text(".tmp/memo/\n", encoding="utf-8")
                    subprocess.run(
                        ["git", "config", "--local", "core.excludesFile", str(excludes)],
                        cwd=root,
                        env=self.base_environment(),
                        check=True,
                    )

                marker = f"{MARKER_PREFIX}{source}"
                first = self.run_helper(root, "prepare", marker=marker)
                second = self.run_helper(root, "prepare", marker=marker)
                first_fields = self.assert_toon_contract(first, 0)
                second_fields = self.assert_toon_contract(second, 0)
                self.assertTrue(first_fields["created_directory"])
                self.assertFalse(second_fields["created_directory"])
                self.assertTrue(first_fields["ignored"])
                self.assertTrue((root / ".tmp" / "memo").is_dir())

    def test_prepare_succeeds_outside_git(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            first = self.run_helper(root, "prepare", marker=f"{MARKER_PREFIX}plain")
            second = self.run_helper(root, "prepare", marker=f"{MARKER_PREFIX}plain")
            first_fields = self.assert_toon_contract(first, 0)
            second_fields = self.assert_toon_contract(second, 0)
            self.assertIsNone(first_fields["ignored"])
            self.assertTrue(first_fields["created_directory"])
            self.assertFalse(second_fields["created_directory"])

    def test_path_and_prepare_reject_symlink_routes(self) -> None:
        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as outside_temporary:
            root = Path(temporary)
            outside = Path(outside_temporary)
            (root / ".tmp").symlink_to(outside, target_is_directory=True)
            for command in ("path", "prepare"):
                with self.subTest(command=command):
                    result = self.run_helper(root, command, marker=f"{MARKER_PREFIX}session")
                    fields = self.assert_toon_contract(result, 1)
                    self.assertEqual(fields["code"], "unsafe-path")
            self.assertEqual(list(outside.iterdir()), [])

        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as outside_temporary:
            root = Path(temporary)
            memo_directory = root / ".tmp" / "memo"
            memo_directory.mkdir(parents=True)
            outside_file = Path(outside_temporary) / "outside.md"
            outside_file.write_text("keep", encoding="utf-8")
            target = self.expected_path(root, "session")
            target.symlink_to(outside_file)
            result = self.run_helper(root, "path", marker=f"{MARKER_PREFIX}session")
            fields = self.assert_toon_contract(result, 1)
            self.assertEqual(fields["code"], "unsafe-path")
            self.assertEqual(outside_file.read_text(encoding="utf-8"), "keep")

    def test_clean_is_bounded_and_idempotent(self) -> None:
        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as outside_temporary:
            root = Path(temporary)
            directory = root / ".tmp" / "memo"
            directory.mkdir(parents=True)
            recognized = [directory / f"session-{'a' * 64}.md", directory / f"session-{'b' * 64}.md"]
            for path in recognized:
                path.write_text("memo", encoding="utf-8")
            unrelated = directory / "notes.md"
            unrelated.write_text("keep", encoding="utf-8")
            malformed = directory / f"session-{'g' * 64}.md"
            malformed.write_text("keep", encoding="utf-8")
            nested = directory / "nested"
            nested.mkdir()
            nested_memo = nested / f"session-{'d' * 64}.md"
            nested_memo.write_text("keep", encoding="utf-8")
            outside = Path(outside_temporary) / "outside.md"
            outside.write_text("keep", encoding="utf-8")
            matching_symlink = directory / f"session-{'c' * 64}.md"
            matching_symlink.symlink_to(outside)

            first = self.run_helper(root, "clean")
            second = self.run_helper(root, "clean")
            first_fields = self.assert_toon_contract(first, 0)
            second_fields = self.assert_toon_contract(second, 0)
            self.assertEqual(first_fields["removed"], 2)
            self.assertEqual(second_fields["removed"], 0)
            self.assertTrue(unrelated.is_file())
            self.assertTrue(malformed.is_file())
            self.assertTrue(nested_memo.is_file())
            self.assertTrue(matching_symlink.is_symlink())
            self.assertEqual(outside.read_text(encoding="utf-8"), "keep")

    def test_clean_refuses_symlinked_memo_directory(self) -> None:
        with tempfile.TemporaryDirectory() as temporary, tempfile.TemporaryDirectory() as outside_temporary:
            root = Path(temporary)
            (root / ".tmp").mkdir()
            (root / ".tmp" / "memo").symlink_to(Path(outside_temporary), target_is_directory=True)
            result = self.run_helper(root, "clean")
            fields = self.assert_toon_contract(result, 1)
            self.assertEqual(fields["code"], "unsafe-path")

    def test_toon_help_status_errors_and_exit_codes(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            for arguments in (("--help",), ("path", "--help"), ("clean", "-h")):
                with self.subTest(arguments=arguments):
                    fields = self.assert_toon_contract(self.run_helper(root, *arguments), 0)
                    self.assertEqual(fields["status"], "ok")
                    self.assertIn("usage", fields)

            unknown = self.run_helper(root, "unknown")
            unknown_fields = self.assert_toon_contract(unknown, 2)
            self.assertEqual(unknown_fields["code"], "invalid-usage")

            extra = self.run_helper(root, "path", "extra", marker=f"{MARKER_PREFIX}session")
            extra_fields = self.assert_toon_contract(extra, 2)
            self.assertEqual(extra_fields["code"], "invalid-usage")

            missing = self.run_helper(root, "path")
            missing_fields = self.assert_toon_contract(missing, 1)
            self.assertEqual(missing_fields["code"], "missing-session-id")

            status = self.run_helper(root, marker=f"{MARKER_PREFIX}session")
            status_fields = self.assert_toon_contract(status, 0)
            self.assertEqual(status_fields["command"], "status")
            self.assertFalse(status_fields["exists"])
            self.assertFalse(status_fields["git_repository"])

            clean = self.run_helper(root, "clean")
            clean_fields = self.assert_toon_contract(clean, 0)
            self.assertEqual(clean_fields["removed"], 0)

    def test_toon_uses_spec_control_character_escaping(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary) / "control-\b"
            root.mkdir()
            result = self.run_helper(root, "path", marker=f"{MARKER_PREFIX}session")
            fields = self.assert_toon_contract(result, 0)
            self.assertIn("\\u0008", result.stdout)
            self.assertEqual(fields["project_root"], str(root.resolve()))


if __name__ == "__main__":
    unittest.main()

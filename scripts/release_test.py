#!/usr/bin/env python3
"""Regression checks for deterministic archives and publication boundaries."""
import hashlib
import importlib.util
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import zipfile


def load(name, filename):
    spec = importlib.util.spec_from_file_location(name, Path(__file__).with_name(filename))
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


release = load('release', 'release.py')
publish = load('publish', 'publish-release.py')


class Archives(unittest.TestCase):
    def test_repeat_build_identity(self):
        def fake_go(args, **kwargs):
            Path(args[args.index('-o') + 1]).write_bytes(b'identical compiled binary')
        with tempfile.TemporaryDirectory() as tmp, patch.object(release, 'run', return_value='a' * 40), patch.object(release.subprocess, 'run', side_effect=fake_go):
            first, second = Path(tmp) / 'first', Path(tmp) / 'second'
            release.build(first)
            release.build(second)
            release.verify(first, 'a' * 40)
            self.assertEqual({p.name: p.read_bytes() for p in first.iterdir()}, {p.name: p.read_bytes() for p in second.iterdir()})
            for target in release.TARGETS:
                archive = first / release.asset(target)
                if target['archive'] == 'zip':
                    with zipfile.ZipFile(archive) as zipped:
                        self.assertTrue(all(i.date_time == (1980, 1, 1, 0, 0, 0) for i in zipped.infolist()))
                else:
                    self.assertEqual(archive.read_bytes()[4:8], b'\0' * 4)
            (first / 'SHA256SUMS').write_text('')
            with self.assertRaisesRegex(ValueError, 'checksum file set'):
                release.verify(first)

    def test_tag_version_mismatch(self):
        with self.assertRaisesRegex(ValueError, 'does not match'):
            release.identity('sjskills-v99.0.0')


class Publication(unittest.TestCase):
    def exercise(self, *, immutable=True, draft=True, mismatched=False):
        calls = []
        with tempfile.TemporaryDirectory() as tmp:
            output = Path(tmp)
            (output / 'archive').write_bytes(b'verified release payload')
            digest = 'sha256:' + hashlib.sha256((output / 'archive').read_bytes()).hexdigest()
            def gh(*args):
                calls.append(args)
                if args[-1].endswith('/immutable-releases'):
                    return json.dumps({'enabled': immutable})
                if args[-1].endswith('/releases'):
                    return json.dumps([[{'tag_name': 'sjskills-v' + release.VERSION, 'draft': draft, 'assets': [{'name': 'archive', 'digest': 'wrong' if mismatched else digest}]}]])
                if '/releases/tags/' in args[-1]:
                    return json.dumps({'id': 12})
                if args[-1].endswith('/assets'):
                    return json.dumps([[{'name': 'archive', 'digest': digest}]])
                return ''
            with patch.object(publish, 'gh', side_effect=gh), patch.object(publish.release, 'identity', return_value='a' * 40), patch.object(publish.release, 'verify'), patch.dict('os.environ', {'GITHUB_REPOSITORY': 'owner/repo'}), patch('sys.argv', ['publish', '--tag', 'sjskills-v' + release.VERSION, '--output', str(output)]):
                if not immutable or not draft or mismatched:
                    with self.assertRaises(SystemExit):
                        publish.main()
                    self.assertFalse(any(c[0] == 'release' for c in calls))
                else:
                    publish.main()
                    self.assertEqual(calls[-1][:2], ('release', 'edit'))
                    self.assertFalse(any(c[:2] == ('release', 'upload') for c in calls))

    def test_disabled_immutability_cannot_publish(self):
        self.exercise(immutable=False)

    def test_public_release_cannot_be_modified(self):
        self.exercise(draft=False)

    def test_mismatched_draft_cannot_be_replaced(self):
        self.exercise(mismatched=True)

    def test_identical_partial_draft_can_resume(self):
        self.exercise()


if __name__ == '__main__':
    unittest.main()

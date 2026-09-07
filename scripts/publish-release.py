#!/usr/bin/env python3
"""Publish only verified assets to an immutable GitHub release; called by CI."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess

spec = importlib.util.spec_from_file_location('release', Path(__file__).with_name('release.py'))
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)


def gh(*args):
    return subprocess.check_output(['gh', *args], text=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tag', required=True)
    parser.add_argument('--output', type=Path, required=True)
    args = parser.parse_args()
    repository = os.environ['GITHUB_REPOSITORY']
    commit = release.identity(args.tag)
    release.verify(args.output, commit)
    enabled = json.loads(gh('api', f'repos/{repository}/immutable-releases'))
    if not enabled['enabled']:
        raise SystemExit('Enable repository release immutability before publishing')
    # A draft is the only mutable phase. Never repair or replace public assets.
    drafts = json.loads(gh('api', '--paginate', '--slurp', f'repos/{repository}/releases'))
    matching = [r for page in drafts for r in page if r['tag_name'] == args.tag]
    if matching:
        if len(matching) != 1 or not matching[0]['draft']:
            raise SystemExit('Refusing to modify an existing published release')
        existing = {a['name']: a for a in matching[0]['assets']}
        expected = {p.name: p for p in args.output.iterdir()}
        if set(existing) - set(expected):
            raise SystemExit('Draft contains unexpected assets; inspect it manually')
        for name, asset in existing.items():
            if asset.get('digest') != 'sha256:' + hashlib.sha256(expected[name].read_bytes()).hexdigest():
                raise SystemExit('Draft asset differs; inspect it manually instead of replacing it')
    else:
        gh('release', 'create', args.tag, '--repo', repository, '--verify-tag', '--draft', '--title', args.tag, '--generate-notes')
        existing = {}
    for path in sorted(args.output.iterdir()):
        if path.name not in existing:
            gh('release', 'upload', args.tag, str(path), '--repo', repository)
    # Verify the exact server-side asset identities before the irreversible step.
    view = json.loads(gh('api', f'repos/{repository}/releases/tags/{args.tag}'))
    assets = json.loads(gh('api', '--paginate', '--slurp', f"repos/{repository}/releases/{view['id']}/assets"))
    uploaded = {a['name']: a['digest'] for page in assets for a in page}
    expected = {p.name: 'sha256:' + hashlib.sha256(p.read_bytes()).hexdigest() for p in args.output.iterdir()}
    if uploaded != expected:
        raise SystemExit('Uploaded asset identity mismatch; draft remains unpublished')
    gh('release', 'edit', args.tag, '--repo', repository, '--draft=false', '--latest')


if __name__ == '__main__':
    main()

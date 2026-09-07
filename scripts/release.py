#!/usr/bin/env python3
"""Build and verify the exact sjskills distributables with standard-library tools."""
import argparse
import gzip
import hashlib
import io
import json
import os
from pathlib import Path
import re
import subprocess
import tarfile
import tempfile
import zipfile

ROOT = Path(__file__).resolve().parents[1]
TARGETS = json.loads((ROOT / 'packaging/targets.json').read_text())
VERSION = (ROOT / 'internal/sjskills/VERSION').read_text().strip()


def run(*args, **kwargs):
    return subprocess.check_output(args, text=True, **kwargs).strip()


def asset(target, version=VERSION):
    return f"sjskills_{version}_{target['os']}_{target['arch']}.{target['archive']}"


def installers():
    shell = '\n'.join('  %s) asset="%s" ;;' % (t['unix'], asset(t, '${version}')) for t in TARGETS if 'unix' in t)
    powershell = '\n'.join('    \'%s\' { $asset = "%s" }' % (t['windows'], asset(t, '${Version}')) for t in TARGETS if 'windows' in t)
    for name, marker, replacement in [('install.sh', '@UNIX_TARGETS@', shell), ('install.ps1', '@WINDOWS_TARGETS@', powershell)]:
        yield name, (ROOT / 'packaging/templates' / name).read_text().replace(marker, replacement)


def identity(tag):
    if not re.fullmatch(r'(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)', VERSION):
        raise ValueError('VERSION must be a stable semantic version')
    if tag != f'sjskills-v{VERSION}':
        raise ValueError(f'tag {tag!r} does not match VERSION {VERSION}')
    commit = run('git', 'rev-parse', 'HEAD', cwd=ROOT)
    if run('git', 'rev-parse', f'{tag}^{{commit}}', cwd=ROOT) != commit:
        raise ValueError('release tag does not point at checkout HEAD')
    if run('git', 'status', '--porcelain', '--untracked-files=no', cwd=ROOT):
        raise ValueError('tracked release source has local modifications')
    return commit


def build(output, tag=None):
    commit = identity(tag) if tag else run('git', 'rev-parse', 'HEAD', cwd=ROOT)
    output.mkdir(parents=True, exist_ok=True)
    if list(output.iterdir()):
        raise ValueError('output directory must be empty')
    for name, data in installers():
        (output / name).write_bytes(data.encode())
    with tempfile.TemporaryDirectory() as directory:
        for target in TARGETS:
            executable = 'sjskills.exe' if target['os'] == 'windows' else 'sjskills'
            binary = Path(directory) / executable
            env = dict(os.environ, GOOS=target['os'], GOARCH=target['arch'], CGO_ENABLED='0')
            subprocess.run(['go', 'build', '-trimpath', '-buildvcs=true', '-o', str(binary), './cmd/sjskills'], cwd=ROOT, env=env, check=True)
            metadata = json.dumps({'version': VERSION, 'commit': commit, 'target': f"{target['os']}/{target['arch']}"}, sort_keys=True).encode() + b'\n'
            contents = {executable: binary.read_bytes(), 'release.json': metadata}
            path = output / asset(target)
            if target['archive'] == 'zip':
                with zipfile.ZipFile(path, 'w', compression=zipfile.ZIP_DEFLATED) as archive:
                    for name, data in contents.items():
                        info = zipfile.ZipInfo(name, date_time=(1980, 1, 1, 0, 0, 0))
                        info.compress_type = zipfile.ZIP_DEFLATED
                        info.create_system = 3
                        info.external_attr = (0o100755 if name == executable else 0o100644) << 16
                        archive.writestr(info, data)
            else:
                with path.open('wb') as raw, gzip.GzipFile(filename='', fileobj=raw, mode='wb', mtime=0) as compressed, tarfile.open(fileobj=compressed, mode='w') as archive:
                    for name, data in contents.items():
                        info = tarfile.TarInfo(name)
                        info.size = len(data)
                        info.mode = 0o755 if name == executable else 0o644
                        archive.addfile(info, io.BytesIO(data))
    manifest = ''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.name}\n' for p in sorted(output.iterdir()))
    (output / 'SHA256SUMS').write_bytes(manifest.encode())


def verify(output, commit=None):
    expected = {asset(t) for t in TARGETS} | {'install.sh', 'install.ps1'}
    if {p.name for p in output.iterdir()} != expected | {'SHA256SUMS'}:
        raise ValueError('release file set differs from target definition')
    manifest = {}
    for line in (output / 'SHA256SUMS').read_text().splitlines():
        match = re.fullmatch(r'([a-f0-9]{64})  ([a-zA-Z0-9_.-]+)', line)
        if not match or match[2] in manifest:
            raise ValueError('invalid or duplicate checksum entry')
        manifest[match[2]] = match[1]
    if set(manifest) != expected:
        raise ValueError('checksum file set differs from target definition')
    for name, digest in manifest.items():
        if hashlib.sha256((output / name).read_bytes()).hexdigest() != digest:
            raise ValueError(f'checksum mismatch: {name}')
    for name, data in installers():
        if (output / name).read_bytes() != data.encode():
            raise ValueError(f'installer differs from source: {name}')
    for target in TARGETS:
        path = output / asset(target)
        executable = 'sjskills.exe' if target['os'] == 'windows' else 'sjskills'
        if target['archive'] == 'zip':
            with zipfile.ZipFile(path) as archive:
                names = archive.namelist()
                metadata = archive.read('release.json')
        else:
            with tarfile.open(path) as archive:
                names = archive.getnames()
                if any(not member.isfile() for member in archive.getmembers()):
                    raise ValueError('archive has non-regular files')
                metadata = archive.extractfile('release.json').read()
        if sorted(names) != sorted([executable, 'release.json']):
            raise ValueError('unexpected archive contents')
        info = json.loads(metadata)
        if info['version'] != VERSION or info['target'] != f"{target['os']}/{target['arch']}" or (commit and info['commit'] != commit):
            raise ValueError('archive identity mismatch')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['matrix', 'build', 'verify'])
    parser.add_argument('--output', type=Path, default=ROOT / '.tmp/release')
    parser.add_argument('--tag', help='Require clean source and exact release tag; omit only for local development checks')
    args = parser.parse_args()
    if args.command == 'matrix':
        print(json.dumps({'include': TARGETS}, separators=(',', ':')))
    elif args.command == 'build':
        build(args.output.resolve(), args.tag)
    else:
        verify(args.output.resolve(), identity(args.tag) if args.tag else None)


if __name__ == '__main__':
    main()

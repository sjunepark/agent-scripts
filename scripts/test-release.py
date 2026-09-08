#!/usr/bin/env python3
"""Exercise built archives through the supported installers, including failures."""
import argparse
import importlib.util
import json
from datetime import datetime, timezone
import os
from pathlib import Path
import platform
import shutil
import subprocess
import tempfile

spec = importlib.util.spec_from_file_location('release', Path(__file__).with_name('release.py'))
release = importlib.util.module_from_spec(spec)
spec.loader.exec_module(release)


def invoke(installer, assets, destination, version=release.VERSION):
    if os.name == 'nt':
        command = ['powershell', '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', str(installer), '-Version', version,
                   '-InstallDir', str(destination), '-ArchiveDir', str(assets)]
    else:
        command = ['/bin/sh', str(installer), version, str(destination), str(assets)]
    return subprocess.run(command, capture_output=True, text=True)


def check_status(cli, consumer, configured):
    # Opt-out must not discover scopes or touch even the disposable cache.
    before = {str(p.relative_to(consumer)): (p.stat().st_mtime_ns, p.read_bytes() if p.is_file() else None)
              for p in consumer.rglob('*')}
    for arguments in [('--no-status-check',), ('status', '--no-status-check'),
                      ('--json', '--no-status-check'), ('--no-status-check', 'status', '--json')]:
        result = cli(*arguments)
        assert result.returncode == 0 and not result.stderr, result
        if '--json' in arguments:
            value = json.loads(result.stdout)
            assert value['operation'] == 'status' and value['result'] == 'success', value
            assert value['status'] == {'projectConfiguration': 'skipped'}, value
            assert not value.get('advisories') and 'cliAdvisory' not in value and 'plan' not in value, value
        else:
            assert result.stdout == 'Status checks disabled (--no-status-check).\n', result
    after = {str(p.relative_to(consumer)): (p.stat().st_mtime_ns, p.read_bytes() if p.is_file() else None)
             for p in consumer.rglob('*')}
    assert after == before, 'disabled status changed consumer files or cache'
    for arguments in [(), ('status',), ('--json',), ('--json', 'status'), ('status', '--json')]:
        result = cli(*arguments)
        assert result.returncode == 0 and not result.stderr, result
        if '--json' in arguments:
            value = json.loads(result.stdout)
            assert value['operation'] == 'status' and value['result'] == 'success', value
            assert 'plan' not in value, value
            assert value['cliAdvisory']['comparison'] == 'update', value
            assert value['cliAdvisory']['availableVersion'] == '999999.0.0', value
            assert value['cliAdvisory']['cached'], value
            assert value['status']['projectConfiguration'] == ('configured' if configured else 'not-configured'), value
            if configured:
                assert Path(value['status']['projectRoot']).resolve() == consumer.resolve(), value
            else:
                assert 'projectRoot' not in value['status'], value
            advisories = value['advisories']
            assert [v['scope'] for v in advisories] == (['project', 'global'] if configured else ['global']), value
            # No Bun and no existing evidence must produce honest unavailability,
            # while the advisory command itself still succeeds.
            assert all(v['freshness'] == 'unavailable' and v['error'] for v in advisories), value
        else:
            assert 'Global:' in result.stdout, result
            if configured:
                quoted_root = json.dumps(str(consumer.resolve()), ensure_ascii=True)
                assert f'Project ({quoted_root}):' in result.stdout, result
            else:
                assert 'Project: not configured' in result.stdout and 'sjskills init' in result.stdout, result
        if not configured:
            assert not (consumer / 'sjskills.toml').exists(), 'status initialized a project'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output', type=Path, default=release.ROOT / '.tmp/release')
    args = parser.parse_args()
    output = args.output.resolve()
    release.verify(output)
    system = {'Darwin': 'darwin', 'Windows': 'windows'}.get(platform.system())
    architecture = {'arm64': 'arm64', 'aarch64': 'arm64', 'x86_64': 'amd64', 'amd64': 'amd64'}.get(platform.machine().lower())
    target = next((t for t in release.TARGETS if t['os'] == system and t['arch'] == architecture), None)
    if not target:
        raise SystemExit('Consumer tests require a supported native release target')
    archive = release.asset(target)
    installer = output / ('install.ps1' if os.name == 'nt' else 'install.sh')
    executable = 'sjskills.exe' if os.name == 'nt' else 'sjskills'
    with tempfile.TemporaryDirectory(prefix='sjskills-consumer-') as temporary:
        root = Path(temporary)
        destination = root / 'install with spaces'
        result = invoke(installer, output, destination)
        assert result.returncode == 0, result.stderr + result.stdout
        binary = destination / executable
        original = binary.read_bytes()
        # An unrelated cwd and empty PATH establish that source, Go, gh, Git,
        # and Bun are unnecessary for the pure command surface. Status reports
        # unavailable upstream evidence successfully when Bun is absent.
        consumer = root / 'consumer'
        consumer.mkdir()
        staging = consumer / 'temporary'
        staging.mkdir()
        env = dict(os.environ, PATH='', HOME=str(consumer), USERPROFILE=str(consumer),
                   LOCALAPPDATA=str(consumer / 'cache'), XDG_CACHE_HOME=str(consumer / 'cache'),
                   TMPDIR=str(staging), TEMP=str(staging), TMP=str(staging),
                   HTTPS_PROXY='http://127.0.0.1:1', NO_PROXY='')
        # Exercise release comparison offline with the packaged executable and
        # empty PATH. Metadata is disposable and does not require Bun or gh.
        cache_base = consumer / 'Library' / 'Caches' if system == 'darwin' else consumer / 'cache'
        release_cache = cache_base / 'sjskills' / 'cli-status'
        release_cache.mkdir(parents=True)
        (release_cache / 'release.json').write_text(json.dumps({
            'schema': 1, 'source': 'sjunepark/agent-scripts',
            'platform': f"{target['os']}/{target['arch']}",
            'version': '999999.0.0', 'installable': True,
            'observedAt': datetime.now(timezone.utc).isoformat(),
            'retryAt': '0001-01-01T00:00:00Z',
        }))
        def cli(*arguments):
            return subprocess.run([str(binary), *arguments], cwd=consumer, env=env,
                                  capture_output=True, text=True, timeout=45)
        result = cli('--version')
        assert result.returncode == 0 and result.stdout == f'sjskills {release.VERSION}\n', result
        result = cli('--help')
        assert result.returncode == 0 and 'sjskills <command>' in result.stdout and 'status' in result.stdout, result
        check_status(cli, consumer, configured=False)
        result = cli('--json', 'profiles')
        assert result.returncode == 0 and json.loads(result.stdout)['result'] == 'success', result
        result = cli('init', 'dev')
        assert result.returncode == 0, result
        manifest = consumer / 'sjskills.toml'
        before = manifest.read_bytes()
        check_status(cli, consumer, configured=True)
        assert manifest.read_bytes() == before, 'status changed the project manifest'
        result = cli('init', 'go')
        assert result.returncode != 0 and manifest.read_bytes() == before, result
        result = cli('--json', 'plan')
        assert result.returncode != 0 and 'bunx' in result.stdout, result
        # Reinstallation uses the same supported path.
        result = invoke(installer, output, destination)
        assert result.returncode == 0, result.stderr + result.stdout
        assert binary.read_bytes() == original
        fixtures = root / 'fixtures'
        fixtures.mkdir()
        shutil.copy2(output / 'SHA256SUMS', fixtures / 'SHA256SUMS')
        # Missing asset and corrupt asset must both preserve the prior binary.
        assert invoke(installer, fixtures, destination).returncode != 0
        (fixtures / archive).write_bytes(b'corrupt archive')
        assert invoke(installer, fixtures, destination).returncode != 0
        assert binary.read_bytes() == original
        shutil.copy2(output / archive, fixtures / archive)
        sums = (output / 'SHA256SUMS').read_text()
        archive_line = next(line for line in sums.splitlines() if line.endswith('  ' + archive))
        for bad_manifest in ['', sums + archive_line + '\n']:
            (fixtures / 'SHA256SUMS').write_text(bad_manifest)
            assert invoke(installer, fixtures, destination).returncode != 0
            assert binary.read_bytes() == original
        assert invoke(installer, output, destination, '../1.0.0').returncode != 0
        assert binary.read_bytes() == original
        # A checksummed archive with a wrong binary identity cannot replace it.
        mismatch = sums.replace(archive, release.asset(target, '99.0.0'))
        (fixtures / 'SHA256SUMS').write_text(mismatch)
        shutil.copy2(output / archive, fixtures / release.asset(target, '99.0.0'))
        assert invoke(installer, fixtures, destination, '99.0.0').returncode != 0
        assert binary.read_bytes() == original
        # Refuse existing directories and links without following them.
        occupied = root / 'occupied'
        (occupied / executable).mkdir(parents=True)
        assert invoke(installer, output, occupied).returncode != 0
        if os.name != 'nt':
            linked = root / 'linked'
            linked.mkdir()
            (linked / executable).symlink_to(binary)
            assert invoke(installer, output, linked).returncode != 0
            assert (linked / executable).is_symlink()
        assert binary.read_bytes() == original
        assert not list(destination.glob('.sjskills-install.*')), 'installer left staging files'
        # Check the HTTPS downloader contract with controlled responses. This
        # does not substitute for native hosted CI or a live published release.
        download_env = dict(os.environ, RELEASE_FIXTURES=str(output), INSTALLER=str(installer),
                            INSTALL_DEST=str(destination), INSTALL_VERSION=release.VERSION)
        if os.name == 'nt':
            driver = root / 'download.ps1'
            driver.write_text("""
function Invoke-WebRequest {
    param($Uri, $OutFile)
    if ($env:FAIL_DOWNLOAD) { throw 'Simulated download failure' }
    $prefix = 'https://github.com/sjunepark/agent-scripts/releases/download/sjskills-v' + $env:INSTALL_VERSION + '/'
    if (-not $Uri.StartsWith($prefix)) { throw 'Unexpected release URL' }
    Copy-Item -LiteralPath (Join-Path $env:RELEASE_FIXTURES $Uri.Substring($prefix.Length)) -Destination $OutFile
}
& $env:INSTALLER -Version $env:INSTALL_VERSION -InstallDir $env:INSTALL_DEST
""")
            download_command = ['powershell', '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', str(driver)]
        else:
            commands = root / 'commands'
            commands.mkdir()
            curl = commands / 'curl'
            curl.write_text("""#!/bin/sh
set -eu
[ -z "${FAIL_DOWNLOAD:-}" ] || exit 22
url=''
output=''
while [ "$#" -gt 0 ]; do
  case "$1" in
    https://*) url=$1 ;;
    -o) shift; output=$1 ;;
  esac
  shift
done
prefix="https://github.com/sjunepark/agent-scripts/releases/download/sjskills-v$INSTALL_VERSION/"
case "$url" in "$prefix"*) ;; *) exit 23 ;; esac
cp "$RELEASE_FIXTURES/${url#"$prefix"}" "$output"
""")
            curl.chmod(0o755)
            download_env['PATH'] = str(commands) + os.pathsep + os.environ['PATH']
            download_command = ['/bin/sh', str(installer), release.VERSION, str(destination)]
        result = subprocess.run(download_command, env=download_env, capture_output=True, text=True)
        assert result.returncode == 0, result.stderr + result.stdout
        download_env['FAIL_DOWNLOAD'] = '1'
        result = subprocess.run(download_command, env=download_env, capture_output=True, text=True)
        assert result.returncode != 0
        assert binary.read_bytes() == original
        assert not list(destination.glob('.sjskills-install.*'))
    print(f"Release consumer and preservation checks passed: {system}/{architecture}")


if __name__ == '__main__':
    main()

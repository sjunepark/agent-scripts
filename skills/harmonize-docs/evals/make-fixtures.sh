#!/bin/sh
# Builds the harmonize-docs evaluation fixtures under $1 (must not exist yet).
set -eu
OUT="$1"
mkdir "$OUT"
cd "$OUT"
commit() { git init -q && git add -A && git -c user.name=fx -c user.email=fx@example.invalid commit -qm initial; }

# ---------- A: accurate docs, mixed audiences, unsurfaced references, instruction routing, ADR, roadmap
mkdir -p A/src A/docs/decisions A/tests A/scripts
cd A
cat > src/tally.py <<'EOF'
"""tally: count words in files using a TOML config."""
import sys, tomllib, pathlib

EXIT_OK, EXIT_USAGE, EXIT_CONFIG, EXIT_IO = 0, 1, 2, 3  # documented in docs/exit-codes.md

DEFAULT_CONFIG = "tally.toml"

def load_config(path):
    try:
        with open(path, "rb") as f:
            cfg = tomllib.load(f)
    except FileNotFoundError:
        return {"ignore_case": False, "min_length": 1, "stopwords": []}
    except tomllib.TOMLDecodeError:
        sys.exit(EXIT_CONFIG)
    return {
        "ignore_case": cfg.get("ignore_case", False),
        "min_length": cfg.get("min_length", 1),
        "stopwords": cfg.get("stopwords", []),
    }

def main(argv):
    if len(argv) < 2:
        print("usage: tally FILE [--config PATH]", file=sys.stderr)
        return EXIT_USAGE
    cfg_path = DEFAULT_CONFIG
    if "--config" in argv:
        cfg_path = argv[argv.index("--config") + 1]
    cfg = load_config(cfg_path)
    try:
        text = pathlib.Path(argv[1]).read_text()
    except OSError:
        return EXIT_IO
    words = text.split()
    if cfg["ignore_case"]:
        words = [w.lower() for w in words]
    words = [w for w in words if len(w) >= cfg["min_length"] and w not in cfg["stopwords"]]
    print(len(words))
    return EXIT_OK

if __name__ == "__main__":
    sys.exit(main(sys.argv))
EOF
printf 'test:\n\tpython -m pytest tests\nlint:\n\truff check src tests\nrelease:\n\tpython scripts/bump_version.py && git tag v$$(cat VERSION)\n' > Makefile
echo "def test_placeholder(): pass" > tests/test_tally.py
echo "0.4.0" > VERSION
echo "# bumps VERSION" > scripts/bump_version.py
cat > README.md <<'EOF'
# tally

tally counts the words in a text file, optionally filtering by a TOML config.

## Install

```
pip install tally-words
```

## Usage

```
tally notes.txt
tally notes.txt --config ./my-tally.toml
```

By default tally reads `tally.toml` from the current directory if present.

## Development

Install dev dependencies with `pip install -e .[dev]`. Run `make test` for the
test suite and `make lint` for ruff. Tests live in `tests/` and mirror the
module names in `src/`.

## Code layout

- `src/tally.py` holds the CLI entry point, config loading, and counting.
- `scripts/bump_version.py` bumps `VERSION` during releases.

## Release process

1. Make sure `make test` and `make lint` pass on `main`.
2. Run `make release`, which bumps `VERSION` and tags `v<version>`.
3. Push the tag; CI publishes to PyPI.
4. Announce the release in the changelog discussion.

## License

MIT
EOF
cat > AGENTS.md <<'EOF'
# Agent instructions

- Run `make test` and `make lint` before proposing changes.
- When asked to cut a release, follow the "Release process" section in README.md exactly.
- Do not edit files under docs/decisions/; they are accepted decision records.
EOF
cat > docs/config-format.md <<'EOF'
# Config format

tally reads an optional TOML file (`tally.toml` by default, or `--config PATH`).

| Key | Type | Default | Meaning |
| --- | --- | --- | --- |
| `ignore_case` | bool | `false` | Lowercase words before filtering. |
| `min_length` | int | `1` | Drop words shorter than this. |
| `stopwords` | list of strings | `[]` | Words to exclude. |

A missing file uses the defaults. A malformed file exits with code 2.
EOF
cat > docs/exit-codes.md <<'EOF'
# Exit codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | Usage error (no file given) |
| 2 | Config file could not be parsed |
| 3 | Input file could not be read |
EOF
cat > docs/roadmap.md <<'EOF'
# Roadmap

- **Planned, not implemented:** `--format json` output with per-word counts.
- **Under discussion:** reading from stdin when no file is given.
EOF
cat > docs/decisions/0001-use-toml.md <<'EOF'
# 0001: Use TOML for configuration

Status: Accepted (2025-03-02)

We chose TOML over YAML because Python 3.11 ships `tomllib`, avoiding a dependency.
EOF
commit
cd ..

# ---------- B: healthy small repository
mkdir -p B/src B/docs
cd B
cat > src/greet.py <<'EOF'
import os
def main():
    name = os.environ.get("GREET_NAME", "world")
    loud = os.environ.get("GREET_LOUD") == "1"
    msg = f"hello, {name}"
    print(msg.upper() if loud else msg)
if __name__ == "__main__":
    main()
EOF
cat > README.md <<'EOF'
# greet

Prints a greeting. Configure it with environment variables; see
[configuration](docs/config.md).

```
python src/greet.py
```

Contributors: see [CONTRIBUTING.md](CONTRIBUTING.md).
EOF
cat > CONTRIBUTING.md <<'EOF'
# Contributing

Run `python src/greet.py` to try changes. There is no test suite yet; describe
manual checks in your PR. Releases are tagged `vX.Y.Z` on `main`.
EOF
cat > docs/config.md <<'EOF'
# Configuration

| Variable | Default | Effect |
| --- | --- | --- |
| `GREET_NAME` | `world` | Name to greet. |
| `GREET_LOUD` | unset | Set to `1` to print in uppercase. |
EOF
commit
cd ..

# ---------- C: scoped change (flag rename) beside unrelated overlapping ops docs
mkdir -p C/src C/docs
cd C
cat > src/export.py <<'EOF'
import argparse
def main():
    p = argparse.ArgumentParser(prog="export")
    p.add_argument("source")
    p.add_argument("--out", default="export.csv", help="output file")
    p.add_argument("--delimiter", default=",")
    a = p.parse_args()
    print(f"exporting {a.source} to {a.out} with {a.delimiter!r}")
if __name__ == "__main__":
    main()
EOF
cat > README.md <<'EOF'
# export

Exports records from a source database file to CSV.

```
python src/export.py records.db --out records.csv
```

See [CLI reference](docs/cli.md). Operators: see docs/deployment.md, docs/ops-notes.md, and docs/runbook.md.
EOF
cat > docs/cli.md <<'EOF'
# CLI reference

`export SOURCE [--out FILE] [--delimiter CHAR]`

| Option | Default | Meaning |
| --- | --- | --- |
| `--out FILE` | `export.csv` | Output CSV path. |
| `--delimiter CHAR` | `,` | Field delimiter. |
EOF
cat > docs/deployment.md <<'EOF'
# Deployment

Run export nightly from cron on the reporting host. Logs go to /var/log/export.log.
If a run fails, check disk space on the reporting host first.
EOF
cat > docs/ops-notes.md <<'EOF'
# Ops notes

- Nightly cron on the reporting host runs export.
- Failed runs: check disk space, then the log at /var/log/export.log.
- Rotate the log weekly.
EOF
cat > docs/runbook.md <<'EOF'
# Runbook

When the nightly export fails: check disk space on the reporting host, read
/var/log/export.log, and rerun manually with `python src/export.py`.
EOF
commit
sed -i 's/"--out", default/"--output", "-o", default/; s/a\.out/a.output/' src/export.py
cd ..

# ---------- D (holdout): configuration reference scattered across README, ops, FAQ
mkdir -p D/src D/docs
cd D
cat > src/pinger.py <<'EOF'
import os, time, urllib.request
TARGET = os.environ["PINGER_TARGET"]
INTERVAL = int(os.environ.get("PINGER_INTERVAL", "30"))
TIMEOUT = float(os.environ.get("PINGER_TIMEOUT", "5"))
WEBHOOK = os.environ.get("PINGER_WEBHOOK")
FAILS_BEFORE_ALERT = int(os.environ.get("PINGER_FAILS", "3"))
def check():
    try:
        urllib.request.urlopen(TARGET, timeout=TIMEOUT); return True
    except Exception:
        return False
def main():
    fails = 0
    while True:
        fails = 0 if check() else fails + 1
        if fails == FAILS_BEFORE_ALERT and WEBHOOK:
            urllib.request.urlopen(urllib.request.Request(WEBHOOK, data=b'{"text":"down"}', method="POST"))
        time.sleep(INTERVAL)
if __name__ == "__main__":
    main()
EOF
cat > README.md <<'EOF'
# pinger

pinger polls a URL and posts to a webhook after repeated failures.

Set `PINGER_TARGET` to the URL to check and run `python src/pinger.py`.
For alerts, see docs/ops.md. Questions? See docs/faq.md.
EOF
cat > docs/ops.md <<'EOF'
# Operating pinger

Run it under systemd with `Restart=always`.

Set `PINGER_WEBHOOK` to a Slack-compatible incoming webhook URL to receive
alerts. pinger posts `{"text":"down"}` once per outage.

Polling defaults to every 30 seconds; change it with `PINGER_INTERVAL` (seconds).
EOF
cat > docs/faq.md <<'EOF'
# FAQ

**How many failures trigger an alert?** Three consecutive failures by default;
set `PINGER_FAILS` to change it.

**My target is slow and keeps alerting.** Raise `PINGER_TIMEOUT` (seconds,
default 5).

**Does pinger alert again while still down?** No, once per outage.
EOF
commit
cd ..

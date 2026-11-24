"""Objective checks for fixture A outputs (A3 content, A4 routing, A5 ADR, A7 links)."""
import json, pathlib, re, subprocess, sys

PRISTINE = pathlib.Path(sys.argv[1])
NEEDLES = ["pip install", "ignore_case", "min_length", "stopwords", "make test", "make lint",
           "make release", "Push the tag", "changelog", "--format json", "stdin", "MIT",
           "| 0 |", "| 1 |", "| 2 |", "| 3 |"]

def docs(root):
    return {p.relative_to(root).as_posix(): p.read_text(encoding="utf-8")
            for p in root.rglob("*.md") if ".git" not in p.parts}

out = {}
for run in sys.argv[2:]:
    root = pathlib.Path(run)
    d = docs(root)
    alltext = "\n".join(d.values())
    missing = [n for n in NEEDLES if n not in alltext]
    broken = []
    for name, text in d.items():
        for target in re.findall(r"\]\(([^)#\s]+)(?:#[^)]*)?\)", text):
            if target.startswith(("http:", "https:", "mailto:")):
                continue
            if not (root / name).parent.joinpath(target).resolve().exists():
                broken.append(f"{name} -> {target}")
    adr = "docs/decisions/0001-use-toml.md"
    readme = d.get("README.md", "")
    release_home = [n for n, t in d.items() if "Release process" in t and n != "AGENTS.md"]
    agents = d.get("AGENTS.md", "")
    out["/".join(pathlib.Path(run).parts[-2:])] = {
        "files": sorted(d),
        "A3_missing": missing,
        "A4_agents_exists": "AGENTS.md" in d,
        "A4_agents_release_target": re.findall(r"in ([A-Za-z./-]+\.md)", agents),
        "release_section_in": release_home,
        # A deleted ADR is a failed run, not a checker crash; a missing
        # pristine ADR still raises because the invocation is wrong.
        "A5_adr_identical": (root / adr).is_file()
        and (root / adr).read_bytes() == (PRISTINE / adr).read_bytes(),
        "A7_broken_links": broken,
        "readme_has_dev_or_release": any(h in readme for h in ("## Release process", "## Code layout", "## Development")),
        "links_config_and_exit": "config-format.md" in readme and "exit-codes.md" in readme,
    }
print(json.dumps(out, indent=1))

## Project selections

Profiles are named collections defined centrally in `agent-scripts`'s
`skill-registry.json`; `sjskills profiles` lists them. A project selects those
names, but cannot define new profiles in its manifest. It can independently add
skills from other repositories with `[[direct]]`, without changing the central
registry or publishing those skills in `agent-scripts`.

Inspect the access policy in `sjskills profiles` before selecting a profile.
`kicpa` is public; `kicpa-private` uses the enrolled private GitHub catalog.
Select both only when requested. Authenticated profiles and direct entries use
`access = "github-authenticated"`; omission means public. Use a CLI build that
supports access metadata; older executables can reject it. Declared source
names and URLs are public metadata, never a place for credentials.

Authenticated fetching requires Git, gh, and an existing login with repository
access. It accepts only GitHub.com shorthand or credential-free HTTPS sources.
The CLI uses controlled Git staging and the original gh configuration without
copying credentials. It refuses legacy/unsupported gh configuration before gh
can migrate it. On that error, report the requested `gh auth status` setup step;
do not start login, change accounts, or modify authentication configuration
without user authorization. Existing `GH_TOKEN`/`GITHUB_TOKEN` login also works.
A private fetch failure blocks its selected scope; do not omit that selection,
retry anonymously, or use status cache evidence to bypass it.

For example, using a placeholder source and skill name:

```toml
version = 1
profiles = ["dev"]

[[direct]]
name = "team-review"
source = "your-org/team-skills"
```

Use the source's actual skill name, not an alias. Repeat `[[direct]]` for more
skills, sorting entries by name and profile names alphabetically. Direct names
must be unique and cannot overlap selected profiles or the fixed global
baseline. Sources accept Git shorthand (`owner/repo[/path]`) or credential-free
HTTPS; local paths, embedded credentials, URL queries, npm specifiers, and other
schemes are unsupported. Set optional `full_depth = true` only when deeper
source discovery is needed. Direct skills use copy mode and the registry's
default targets, currently project `.agents/skills` and `.claude/skills`;
per-entry manager, mode, target, and workflow fields are unsupported.

For a direct-only project, omit `profiles` or use `profiles = []`, and include
at least one `[[direct]]` entry. Create that manifest directly when adoption is
requested: `sjskills init` requires a profile and cannot initialize this case.
For an existing manifest, edit the requested selection in place rather than
rerunning `init`. Keep the manifest as committed project configuration; commit
only when authorized. Review `sjskills plan` after configuration, and apply
only when installation or synchronization was requested.

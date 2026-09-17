# Status inspection

Use `sjskills` or `sjskills status` for CLI version and two-scope skill status.
The CLI comparison uses the latest stable published sjskills release, names
unavailable or absent release evidence, and links to updates without installing
them. Version equality does not establish checkout freshness. The report names the
resolved project root, offers setup guidance for a missing manifest, and reports
configuration failures without recommending replacement. It never initializes
or synchronizes. Status exits 0 even when inspection is unavailable; inspect the
findings and freshness rather than treating success as exact state. Cold checks
may take 30 seconds plus cleanup. Stale evidence remains explicitly labeled.

Explicit reports use stdout; other successful commands retain incidental stderr
notices; fresh skill scopes without findings and CLI comparisons without updates
or problems remain silent.
`--json` keeps full skill findings in `advisories` and version evidence in
`cliAdvisory`; status also includes setup metadata under `status`. None grants
apply authority or substitutes for a reviewed plan. Use the same compatible
executable for plan and apply; older loaders may reject the new advisory field.
Use `--no-status-check` to skip discovery, inspection, fetching, and cache writes;
status then reports checks disabled, while other commands retain their primary
verification.

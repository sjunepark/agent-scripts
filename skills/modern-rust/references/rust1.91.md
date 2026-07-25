# Rust 1.91

Released October 2025.

## Important changes

- C-style variadic function declarations became stable for the `sysv64`, `win64`, `efiapi`, and `aapcs` ABIs. Rust still cannot define variadic functions on stable.
- New diagnostics target unsafe assumptions: dangling raw pointers derived from locals and integer-to-pointer transmutes warn by default, while macro expansions that leave a trailing semicolon in expression position are denied by default.
- Cargo gained `build.build-dir` for separating build artifacts from the target directory and the `host-tuple` configuration value. The directory's internal layout is not a stable interface.
- Integer `strict_*` arithmetic became stable and always panics on overflow, independent of overflow-check settings or build profile.

## Compatibility and migration

- Using this syntax, configuration, or library surface raises the MSRV to 1.91. Update old expression-position macros that expand with a trailing semicolon.
- Apple targets may need explicit native-library search paths where builds relied on implicit `/usr/local/lib` discovery.
- Rust 1.91.1 fixed a severe cross-crate `wasm_import_module` regression that could cause link failures or call the wrong function, plus an Illumos target-directory locking issue.

Official details: [Rust 1.91 release notes](https://doc.rust-lang.org/stable/releases.html#version-1911-2025-11-10).

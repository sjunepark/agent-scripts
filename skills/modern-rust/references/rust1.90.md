# Rust 1.90

Released September 2025.

## Important changes

- `rust-lld` became the default linker for `x86_64-unknown-linux-gnu`. It is a compiler default, not Cargo configuration; do not add custom LLD setup unconditionally.
- Cargo stabilized dependency-ordered workspace publishing with `cargo publish --workspace`, including whole-workspace verification.
- Constants may end in references to mutable or external memory, although such constants cannot be patterns. Volatile access to non-Rust memory, including address zero, is accepted, but this does not make the address valid or the operation safe.
- Unsigned integers gained signed-subtraction methods; slice reversal and common float rounding became const-capable. `x86_64-apple-darwin` moved to Tier 2 with host tools.

## Compatibility and migration

- These changes require MSRV 1.90. LLD is not bug-for-bug compatible with GNU BFD; opt out with `-C linker-features=-lld` for linker-specific breakage.
- Workspace publishing is not atomic and can leave only some crates published after a failure. Also audit code relying on `SIGPIPE`: `UnixStream` writes now use `MSG_NOSIGNAL`.
- No Rust 1.90 patch release followed, so later patch fixes belong to other release lines.

Official details: [Rust 1.90 announcement](https://blog.rust-lang.org/2025/09/18/Rust-1.90.0/).

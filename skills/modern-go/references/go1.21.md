# Go 1.21

Released August 2023.

## Important changes

- **Language:** Added the `min`, `max`, and `clear` built-ins. Generic type
  inference became more capable for generic function values, interface method
  sets, and mixed untyped constants. Package initialization order is now
  specified precisely.
- **Versions and toolchains:** The `go` directive became a strict minimum
  requirement. Automatic toolchain selection and the `toolchain` directive
  separate the module's minimum Go version from its preferred toolchain.
  Compatibility defaults selected through `GODEBUG` follow the main module or
  workspace `go` version.
- **Compiler and runtime:** Profile-guided optimization became generally
  available. A `default.pgo` beside a main package is used automatically.
  `runtime.Pinner` supports explicitly pinned Go memory at cgo boundaries.
- **Library:** Added `log/slog`, `testing/slogtest`, and the generic `slices`,
  `maps`, and `cmp` packages. Added `context.AfterFunc`, `WithoutCancel`,
  deadline/timeout causes, the `sync.OnceFunc` family, and
  `errors.ErrUnsupported`.

## Migration notes

- `panic(nil)` now produces `*runtime.PanicNilError`, so a directly deferred
  `recover` during a panic returns non-nil. Modules targeting Go 1.20 or older
  retain the prior behavior; `GODEBUG=panicnil=1` is the temporary escape hatch.
- The first release in a family is now numbered `go1.N.0`, while `Go 1.N`
  names the language and release family.
- Per-iteration loop variables are only a preview in 1.21 under
  `GOEXPERIMENT=loopvar`; they become version-gated language behavior in 1.22.

Official details: [Go 1.21 release notes](https://go.dev/doc/go1.21) and
[Go toolchains](https://go.dev/doc/toolchain).

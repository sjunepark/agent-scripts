# Go 1.23

Released August 2024.

## Important changes

- **Language:** `for range` accepts iterator functions matching zero-, one-, or
  two-value yield signatures. The new `iter.Seq` and `iter.Seq2` types express
  those conventions. Generic type aliases remain a same-package preview under
  `GOEXPERIMENT=aliastypeparams`.
- **Tools:** `go vet` adds `stdversion`, which catches standard-library symbols
  newer than the effective module or file language version. `go.mod` and
  `go.work` accept `godebug` directives. `go mod tidy -diff` reports required
  changes without editing files. Telemetry stays local unless explicitly
  enabled with `go telemetry on`.
- **Compiler and runtime:** PGO build overhead drops substantially. The linker
  restricts new `//go:linkname` and assembly references to unmarked
  standard-library internals.
- **Library:** Added `iter`, `unique`, and `structs`. `slices` and `maps` gain
  iterator producers and consumers such as `All`, `Values`, and `Collect`.
  Other additions include `os.CopyFS`, `filepath.Localize`,
  `runtime/debug.SetCrashOutput`, `sync.Map.Clear`, and atomic `And`/`Or`.

## Migration notes

- For main modules targeting Go 1.23+, unreachable timers and tickers can be
  collected, timer channels are synchronous, and `Stop` or `Reset` prevents
  later stale values. Replace channel-length polling with a nonblocking
  receive. `GODEBUG=asynctimerchan=1` temporarily restores old semantics.
- `go/types` produces `Alias` nodes by default; analysis tools should use
  alias-aware APIs such as `types.Unalias`.
- TLS removes 3DES from defaults and enables an experimental post-quantum
  hybrid exchange by default. The minimum supported macOS version is 11.

Official details: [Go 1.23 release notes](https://go.dev/doc/go1.23) and
[range over function types](https://go.dev/blog/range-functions).

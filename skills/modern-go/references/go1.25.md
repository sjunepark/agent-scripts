# Go 1.25

Released August 2025. It makes no language changes.

## Important changes

- **Tools and modules:** `go.mod` gains `ignore` directives for excluding
  directory trees from package patterns, and the `work` pattern selects all
  packages in active workspace modules. `go vet` adds `waitgroup` and
  IPv6-safe `hostport` checks. Added `go version -m -json` and `go doc -http`.
- **Runtime:** On Linux, the default `GOMAXPROCS` respects cgroup CPU bandwidth
  limits and updates when limits or CPU availability change. Explicit
  configuration disables automatic updates. Added stable
  `runtime/trace.FlightRecorder`. DWARF 5 becomes the default.
  `GOEXPERIMENT=greenteagc` is an opt-in preview in this release.
- **Library and testing:** `testing/synctest` is stable. `sync.WaitGroup.Go`
  launches and tracks a function; its function must not panic. Added
  `net/http.CrossOriginProtection`, `reflect.TypeAssert[T]`, and testing
  attributes and output helpers. `os.Root` gains more root-scoped operations.
- **Preview:** `GOEXPERIMENT=jsonv2` exposes `encoding/json/v2` and
  `encoding/json/jsontext` and swaps the v1 package onto the new
  implementation. APIs and some observable details may still change.

## Migration notes

- Container workloads may receive a lower and dynamically changing
  `GOMAXPROCS`. Use `runtime.SetDefaultGOMAXPROCS` to reapply the automatic
  default after an override.
- More slice backing arrays can move to the stack, exposing invalid
  `unsafe.Pointer` assumptions and changing allocation-sensitive benchmarks.
- TLS 1.2 rejects SHA-1 signatures by default, and ASN.1/X.509 parsing rejects
  more malformed inputs.
- Go 1.25 requires macOS 12 or newer and is the final release for
  `windows/arm`.

Official details: [Go 1.25 release notes](https://go.dev/doc/go1.25).

# Go 1.22

Released February 2024.

## Important changes

- **Language:** Loop variables declared by `for` in packages targeting Go 1.22
  are fresh each iteration, fixing accidental closure and goroutine capture.
  Integer ranges are stable: `for i := range n` yields `0` through `n-1`.
  Range-over-function iterators remain a preview under
  `GOEXPERIMENT=rangefunc`.
- **Tools:** Workspaces gain vendoring through `go work vendor`. `go vet`
  understands new loop semantics and diagnoses empty `append`, prematurely
  evaluated deferred `time.Since`, and malformed `log/slog` calls.
- **Compiler and runtime:** Execution traces are streamable and record fuller
  scheduler and syscall detail. PGO devirtualization and GC metadata improve
  runtime performance. Call-site-aware inlining is still experimental.
- **Library:** Added `math/rand/v2`, with new generators and intentionally
  different streams; keep using `crypto/rand` for security. `net/http.ServeMux`
  patterns gain methods and wildcards, with captures exposed by
  `Request.PathValue`. Added `database/sql.Null[T]`, `go/version`,
  `reflect.TypeFor[T]`, and `slices.Concat`.

## Migration notes

- Raising a module or file's language version to 1.22 activates fresh loop
  variables. Code intentionally sharing one variable must model that sharing
  explicitly.
- New `ServeMux` parsing and precedence can change brace-containing or escaped
  patterns. `GODEBUG=httpmuxgo121=1` temporarily restores old routing.
- Slice deletion helpers now zero removed elements. TLS servers default to a
  TLS 1.2 minimum and disable RSA key-exchange suites by default.
- Go 1.22 is the last release supporting macOS 10.15.

Official details: [Go 1.22 release notes](https://go.dev/doc/go1.22).

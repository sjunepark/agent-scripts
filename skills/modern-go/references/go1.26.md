# Go 1.26

Released February 2026.

## Important changes

- **Language:** `new(expr)` allocates a variable initialized to an expression,
  making pointers to optional scalar values concise. Generic types may refer to
  themselves in type-parameter constraints.
- **Tools and modules:** `go fix` is rebuilt on `go/analysis`, with modernizers
  for current idioms and `//go:fix inline` support for source-level API
  migrations. Stable Go 1.N now makes `go mod init` default to `go 1.(N-1).0`;
  Go 1.26 therefore creates modules targeting Go 1.25. Use `go get go@version`
  to raise the target. `go tool doc` is removed; use `go doc`.
- **Compiler and runtime:** Green Tea GC is now the default rather than a
  preview. Cgo call overhead falls, 64-bit heap bases are randomized, and more
  slice backing stores can stay on the stack. The `goroutineleak` profile
  remains experimental.
- **Library and testing:** Added stable `crypto/hpke`, generic
  `errors.AsType`, reflection iterators, `log/slog.NewMultiHandler`, and
  `bytes.Buffer.Peek`. Tests gain managed artifact directories through
  `T.ArtifactDir`, `B.ArtifactDir`, and `F.ArtifactDir`, plus
  `go test -artifacts`. SIMD and `runtime/secret` packages remain experimental.

## Migration notes

- Several crypto APIs ignore caller-supplied randomness and use a secure global
  source. Use `testing/cryptotest.SetGlobalRandom` for deterministic tests;
  `GODEBUG=cryptocustomrand=1` temporarily restores older behavior.
- `net/http.ServeMux` trailing-slash redirects use status 307 instead of 301.
  `httputil.ReverseProxy.Director` is deprecated in favor of `Rewrite`.
  `net/url.Parse` rejects malformed unbracketed colon-containing hosts.
- The JPEG implementation changed, so encoded bytes and golden outputs can
  differ. The `windows/arm` port is removed.
- Building Go 1.26 from source requires Go 1.24.6 or later.

Official details: [Go 1.26 release notes](https://go.dev/doc/go1.26) and the
[Go release history](https://go.dev/doc/devel/release).

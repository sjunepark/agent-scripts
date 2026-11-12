# Go 1.27

Released August 2026.

## Important changes

- **Language:** Method declarations may declare their own type parameters.
  Interface methods cannot, and a generic method cannot satisfy an interface
  method. Struct literal keys may be any valid field selector, and function
  type inference applies wherever a generic function is assigned or converted
  to a matching function type.
- **Tools and modules:** `go test` runs the `stdversion` vet check by default,
  and `go vet` printf flags `%w` operands of type `*E` where `E` implements
  `error`. `go fix` adds the `atomictypes`, `embedlit`, `slicesbackward`, and
  `unsafefuncs` modernizers; `fmtappendf` is removed and `waitgroup` is renamed
  `waitgroupgo`. For `go 1.27` modules, `go mod tidy` merges `require` blocks
  into at most one direct and one indirect block. `go doc` accepts
  `package@version`. Bazaar support is removed.
- **Compiler and runtime:** Small allocations are size-specialized, speeding
  allocation-heavy code at a cost of roughly 60 KB of binary size. The
  `goroutineleak` profile is generally available through `runtime/pprof` and
  `/debug/pprof/goroutineleak`. Tracebacks in `go 1.27` modules include
  goroutine labels; `GODEBUG=tracebacklabels=0` opts out.
- **Library and testing:** `encoding/json/v2` and `encoding/json/jsontext` are
  stable, and `encoding/json` now runs on the v2 implementation. Added
  `crypto/mldsa`, the `uuid` package, `strings.CutLast`, `bytes.CutLast`,
  generic `math/rand/v2.Rand.N`, `net/url.URL.Clone`,
  `database/sql.ConvertAssign`, `testing/synctest.Sleep`, and
  `net/http/httptest.NewTestServer` for in-memory networking under `synctest`.
  Unicode moves to version 17. The `simd` package and `simd/archsimd` remain
  experimental.

## Migration notes

- `encoding/json` error text can change, and v2 rejects invalid UTF-8 and
  duplicate object names by default. Use the v2 options that select v1
  semantics while migrating. `GOEXPERIMENT=nojsonv2` temporarily restores the
  v1 implementation, but only builds when neither the module nor any
  dependency imports `encoding/json/v2` or `encoding/json/jsontext`.
- Removed `GODEBUG` settings, including `asynctimerchan`, `gotypesalias`,
  `x509keypairleaf`, `tlsrsakex`, `tls3des`, `tls10server`, and
  `tlsunsafeekm`, fail the build in `go.mod` `godebug` lines or `//go:debug`
  directives unless set to the final default. Timer channels are always
  unbuffered.
- `compress/flate` output can differ, affecting golden bytes from `gzip`,
  `zlib`, `zip`, and `png`. gofmt alignment no longer varies by architecture,
  so previously formatted files may see small whitespace changes. Closure
  symbol names change.
- `go tool trace -http=:port` listens only on localhost. `crypto/tls`
  deprecates `Config.Rand`.
- On Windows and Darwin, `x509.SystemCertPool` honors `SSL_CERT_FILE` and
  `SSL_CERT_DIR`: when set, Go loads roots from disk and uses its own verifier
  instead of the platform verifier, so CAs present only in the platform store
  stop verifying. `GODEBUG=x509sslcertoverrideplatform=0` disables this.
- Darwin requires macOS 13 or newer. Big-endian `linux/ppc64` switches to the
  ELFv2 ABI and requires Linux 3.13 or newer, or RHEL 7's 3.10 kernel, which
  backports ELFv2 support.

Official details: [Go 1.27 release notes](https://go.dev/doc/go1.27) and the
[Go release history](https://go.dev/doc/devel/release).

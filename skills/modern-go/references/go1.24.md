# Go 1.24

Released February 2025.

## Important changes

- **Language:** Generic type aliases are stable, including declarations such as
  `type Set[T comparable] = map[T]struct{}`.
- **Tools and modules:** `go.mod` gains `tool` directives for executable
  dependencies, replacing the blank-import `tools.go` convention. Manage them
  with `go get -tool` and run them through `go tool`. Build, install, and test
  gain richer JSON event output; `GOAUTH` supports private-module
  authentication.
- **Compiler and runtime:** Swiss Tables become the default map implementation.
  `sync.Map` is reimplemented for modification-heavy and disjoint-key
  workloads. Added stable `runtime.AddCleanup` and weak pointers in package
  `weak`.
- **Library and testing:** Added traversal-resistant, directory-scoped file
  access through `os.OpenRoot` and `os.Root`. Prefer `for b.Loop()` for new
  benchmarks. Tests gain `T.Context`, `B.Context`, `T.Chdir`, and `B.Chdir`.
  `encoding/json` gains `omitzero`; `bytes` and `strings` gain iterator APIs.
  New crypto packages include `crypto/mlkem`, `crypto/hkdf`,
  `crypto/pbkdf2`, and `crypto/sha3`.

## Migration notes

- `testing/synctest` is experimental in 1.24 and requires
  `GOEXPERIMENT=synctest`; its stable API arrives in 1.25.
- Top-level `math/rand.Seed` is now a no-op. Use a local generator;
  `GODEBUG=randseednop=0` temporarily restores the global seeding behavior.
- `go vet` flags nonconstant `fmt.Printf(s)` calls without formatting
  arguments for packages targeting Go 1.24+.
- TLS enables an ML-KEM hybrid exchange by default. X.509 and RSA parsing and
  validation are stricter, and SHA-1 certificate signatures are rejected.

Official details: [Go 1.24 release notes](https://go.dev/doc/go1.24).

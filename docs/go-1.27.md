# Go 1.27 implementation note

The module declares `go 1.27` and the pull-request workflow installs
`1.27.x` with `actions/setup-go`. This repository uses ordinary Go 1 APIs and
does not make its bounded semantics depend on a new language feature; the
version pin makes the verifier and generated binary target explicit.

The toolchain choice is grounded in the primary Go documentation:

- [Go 1.27 release notes](https://go.dev/doc/go1.27) — Go 1.27 release and
  command/toolchain behavior.
- [Go language specification](https://go.dev/ref/spec) — the current
  specification identifies language version go1.27.
- [Go release history](https://go.dev/doc/devel/release) — go1.27.0 was
  released on 2026-08-19.

The CI job is the authority for the version actually used by a verification
run. Local toolchain output is not release evidence.

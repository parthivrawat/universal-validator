# Contributing

This repository contains four parallel implementations of the same library:
Go (`go/`), Python (`python/`), Rust (`rust/`), and TypeScript
(`typescript/`).

## Golden rule: parity

The canonical behavior is defined by **`testdata/SPEC.md`**, and the shared
test vectors in **`testdata/vectors/cases.json`** are executed by all four
test suites. Any behavior change MUST:

1. Update `testdata/SPEC.md` (semantics, codes, exact message wording).
2. Add/adjust vectors in `testdata/vectors/cases.json` where applicable.
3. Be implemented in **all four** languages in the same change.

## Running the tests

| Language | Command |
|---|---|
| Go | `cd go && go test ./...` |
| Python | `cd python && python -m pytest -q` |
| Rust | `cd rust && cargo test` |
| TypeScript | `cd typescript && npm install && npx vitest run` |

Lint/typecheck: `gofmt -l go/` + `go vet ./...`, `ruff check python/`,
`cargo clippy` + `cargo fmt --check`, `npx tsc --noEmit`.

## Versioning & releases

All four implementations share **one version**, recorded in the root
`VERSION` file. When releasing:

1. Bump `VERSION`.
2. Set the same version in `python/pyproject.toml` (`project.version`),
   `rust/Cargo.toml` (`package.version`), and
   `typescript/package.json` (`version`). Go has no manifest version —
   tag the repo `vX.Y.Z` (the Go module is `.../go`, so also tag
   `go/vX.Y.Z` for Go module resolution).
3. Update `CHANGELOG.md`.
4. CI verifies the manifests match `VERSION`.

## Style notes

- Follow each implementation's existing conventions; keep public APIs
  documented in the per-language READMEs.
- Validators are immutable after construction and safe to share across
  threads/goroutines.
- No coercion/transform layer — validation never mutates input.

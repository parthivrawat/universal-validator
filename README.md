# Universal Validator

A comprehensive data validation library for API, database, and form contexts, implemented consistently across Python, TypeScript, Go, and Rust.

## Overview

This library provides a consistent set of validation primitives for checking values and schemas across different runtimes. It is designed for use in APIs, database models, and form handling without pulling in heavy dependencies.

## Languages

| Language | Package | README |
|---|---|---|
| **Go** | `go get github.com/parthivrawat/universal-validator/go/v2` | [Go README](go/README.md) |
| **Python** | `pip install universal-validator` | [Python README](python/README.md) |
| **Rust** | `cargo add universal-validator` | [Rust README](rust/README.md) |
| **TypeScript** | `npm install universal-validator-lib` | [TypeScript README](typescript/README.md) |

## Repository Layout

```
universal-validator/
├── go/            # Go module
├── python/        # Python package (universal_validator/)
├── rust/          # Rust crate
├── typescript/    # TypeScript/npm package
├── testdata/      # Shared conformance spec + test vectors (SPEC.md, vectors/)
├── .github/       # CI workflow
├── VERSION        # Single source of truth for the shared version
├── CHANGELOG.md
├── CONTRIBUTING.md
└── README.md      # This file
```

All four implementations follow the canonical behavior in
[`testdata/SPEC.md`](testdata/SPEC.md) and are verified against the shared
vectors in `testdata/vectors/cases.json`.

## License

MIT License — see [LICENSE](LICENSE) (also present in each language directory).

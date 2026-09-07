# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and all four
language implementations share a single version (see `VERSION`).

## [Unreleased]

## [2.0.0] — 2026-09-07

### Added
- Machine-readable `code` on every validation error (`required`,
  `unknown_field`, `type`, `min_length`, `max_length`, `min_value`,
  `max_value`, `pattern`, `choices`, `email`, `url`, `custom`) — asserted by
  the shared conformance corpus.
- Strict mode (unknown-key detection) in all four languages.
- Fail-fast mode (stop at first error) in all four languages.
- Shared conformance corpus: `testdata/SPEC.md` + `testdata/vectors/cases.json`
  consumed by all four test suites.
- New validators in all four languages: `uuid`, `date`, `datetime`, `numeric`
  (numeric strings), `not_empty`, `enum`, `one_of` (union), `regex` passthrough.
- `sensitive` option redacting offending values from error messages and
  `error.value`.
- `MAX_PATTERN_INPUT_LENGTH` (10,000 chars) cap applied before every pattern
  match (ReDoS bound for backtracking regex engines).
- Error aggregates: Go `ValidationErrors` + `Schema.ValidateOrError`, Python
  `ValidationErrors` (raised by `validate_or_raise`), Rust `ValidationErrors` +
  `Schema::try_validate`, TypeScript `ValidationErrors` (thrown by
  `validateOrThrow`).

### Changed (breaking)
- **Go**: options-struct constructors replaced by functional options
  (`String(MinLength(3), Optional())`); `Validator.Validate` now takes
  `context.Context`; `CustomValidator` signature is
  `func(ctx context.Context, value interface{}) error`; `ValidateOrPanic`
  removed.
- **Python**: `universal_validator.py` single module is now a package
  (`universal_validator/core.py`, `validators.py`); the `class validators`
  namespace is replaced by module-level factory functions
  (`uv.validators.int(...)` call sites unchanged); `Validator.custom()` returns
  a clone instead of mutating.
- **Rust**: `Validator::validate` operates on `serde_json::Value`.
- **TypeScript**: generic `Validator<T>` base, options-object constructors,
  `code: ErrorCode` literal union, `readonly` error arrays, `.custom()` clones.

### Fixed
- Error-swallowing in email/URL validators (TS/Python) — now code-based.
- Go options zero-value clobbering `Required` (per-field `Optional` defaults).
- Go typed-nil handling (`(*Foo)(nil)` counts as null).
- Rust regexes compiled once via `OnceLock`; `try_pattern` replaces panicking
  constructor; Python patterns precompiled and full-match everywhere.
- Choices membership is O(1) via set lookups.

### Security
- See `testdata/SPEC.md` "Security": `sensitive` redaction, pattern input cap,
  trusted-pattern requirement, email/URL permissiveness caveats.

## [1.0.0] — 2026-08-26

Initial release: string/integer/number/boolean/email/url/array/object
validators, nested schemas, custom validators, nullable/required.

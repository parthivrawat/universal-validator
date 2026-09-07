# Universal Validator — Conformance Spec

All four implementations (Go, Python, Rust, TypeScript) MUST produce identical
results for the vectors in `vectors/cases.json`. Each test suite loads the
corpus, builds the described schema via a small `buildSchema(spec)` test helper,
validates the input, and compares `(field, message)` pairs exactly and in order
of first occurrence (order within a single field's errors must match; ordering
across different fields is unspecified and tests should compare as a set per
field or sorted by field).

## Validator spec DSL

A schema spec is a JSON object:

```json
{
  "fields": { "<name>": <validator-spec>, ... },
  "strict": false
}
```

A validator-spec is a JSON object:

| Key | Applies to | Meaning |
|---|---|---|
| `type` | all | `"string" \| "integer" \| "number" \| "boolean" \| "email" \| "url" \| "array" \| "object" \| "uuid" \| "date" \| "datetime" \| "numeric" \| "not_empty" \| "enum" \| "one_of"` (required) |
| `required` | all | default `true` |
| `nullable` | all | default `false`; `null`/`nil`/`None`/`Value::Null` passes when true |
| `minLength`, `maxLength` | string, array | character count (Unicode scalar values) / element count |
| `minValue`, `maxValue` | integer, number | inclusive bounds |
| `pattern` | string | regex; MUST match the ENTIRE string (full-match semantics) |
| `choices` | string | list of allowed strings |
| `items` | array | validator-spec applied to each element |
| `fields` | object | nested field → validator-spec map |
| `strict` | object | default `false`; unknown keys → error |
| `values` | enum | list of allowed arbitrary JSON values (deep equality) |
| `of` | one_of | list of validator-specs; value must satisfy at least one |
| `sensitive` | all | default `false`; redacts the offending value in errors (see Security) |

## Canonical kind names

When reporting the actual type of a value, implementations map their native
runtime type to these JSON-kind names:

`null`, `boolean`, `integer`, `number` (non-integral numbers), `string`,
`array`, `object`. Unknown/other → `unknown`.

## Error codes

Every `ValidationError` MUST carry a machine-readable `code` in addition to
`field`/`message`/`value`. Canonical codes:

`required`, `unknown_field`, `type`, `min_length`, `max_length`, `min_value`,
`max_value`, `pattern`, `choices`, `email`, `url`, `custom`, `uuid`, `date`,
`datetime`, `numeric`, `not_empty`, `enum`, `one_of`.

Implementations use the code (not message text) internally — e.g. email/URL
validators identify the pattern error by `code == "pattern"` before rewriting
it to `code == "email"`/`"url"`.

## Presence & null semantics

A key that is PRESENT with a null-ish value (`null`, `None`, `nil`,
`undefined`, `Value::Null`) counts as present and is validated as a null value
— it is not "missing". Implementations must not confuse a present-null field
with an absent field; both produce `Field is required` when
`required && !nullable`, but through the validate path, not the missing-key
path. Typed nils (e.g. a Go `(*Foo)(nil)` stored in `interface{}`) count as
null for this purpose.

## Canonical messages (exact strings, `{}` = substitution)

| Condition | Code | Message |
|---|---|---|
| Required field absent, or value is null and `required && !nullable` | `required` | `Field is required` |
| Strict mode, key not in schema | `unknown_field` | `Unknown field` |
| Wrong type | `type` | `Expected {expected}, got {kind}` where `expected` ∈ `string, integer, number, boolean, array, object` |
| `integer` given non-integral number | `type` | `Expected integer, got number` |
| String too short | `min_length` | `String length must be at least {min}, got {n}` |
| String too long | `max_length` | `String length must be at most {max}, got {n}` |
| Array too short | `min_length` | `Array length must be at least {min}, got {n}` |
| Array too long | `max_length` | `Array length must be at most {max}, got {n}` |
| Below min | `min_value` | `Value must be at least {min}, got {v}` |
| Above max | `max_value` | `Value must be at most {max}, got {v}` |
| Pattern mismatch | `pattern` | `String does not match pattern {pattern}` (the raw pattern source, not `/…/ ` or compiled repr) |
| Choice mismatch | `choices` | `Value must be one of {a, b, c}, got '{v}'` (choices joined with `, `) |
| Email pattern mismatch | `email` | `Invalid email address: {v}` |
| URL pattern mismatch | `url` | `Invalid URL: {v}` |
| UUID format mismatch | `uuid` | `Invalid UUID: {v}` |
| Date format mismatch | `date` | `Invalid date: {v}` |
| Datetime format mismatch | `datetime` | `Invalid datetime: {v}` |
| String is not numeric | `numeric` | `Expected numeric string, got {v}` |
| Value is empty | `not_empty` | `Value must not be empty` |
| Not in enum `values` | `enum` | `Value must be one of {v1, v2, …}, got {v}` (each rendered as compact JSON — strings quoted, e.g. `"a", 1, true`) |
| No `one_of` branch matches | `one_of` | `Value does not match any allowed schema` |
| `not_empty` given other kind | `type` | `Expected string, array, or object, got {kind}` |
| Custom validator rejection | `custom` | (user-supplied message) |

Number formatting: shortest representation, no trailing zeros (`1.5` not
`1.500000`). Compact JSON rendering (used by `enum`): `json.Marshal` /
`json.dumps` / `serde_json::to_string` / `JSON.stringify` of the single value —
`"a"`, `1`, `1.5`, `true`, `null`.

## Semantics

- `null` input on `required && !nullable` → `Field is required`.
- `null` input otherwise (nullable, or not required) → no errors.
- `required=false` and key absent → no errors; key present with `null` →
  treated as null per the rule above.
- `number` validators accept integers; `integer` validators reject
  non-integral numbers with the type error above.
- Nested paths: `parent.child`, array items `field[0]`.
- `email`/`url` validators behave like a `string` validator with a fixed
  pattern; non-pattern errors (e.g. length) pass through unchanged, only the
  pattern-mismatch error is rewritten to the friendly message.
- `strict` applies per object level where set.
- `uuid` behaves like a `string` validator with the fixed pattern
  `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`;
  non-strings → `Expected string, got {kind}`; pattern failure is reported with
  code `uuid` and `Invalid UUID: {v}`.
- `date` behaves like a `string` validator with the fixed pattern
  `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$` (calendar semantics beyond
  month/day ranges are NOT checked); failures → code `date`,
  `Invalid date: {v}`.
- `datetime` behaves like a `string` validator with the fixed pattern
  `^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$`;
  failures → code `datetime`, `Invalid datetime: {v}`.
- `numeric` requires a string matching
  `^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`; non-strings →
  `Expected string, got {kind}`; failures → code `numeric`.
- `not_empty` accepts strings, arrays, and objects; any other kind →
  `Expected string, array, or object, got {kind}` (code `type`). An empty
  string/array/object → code `not_empty`, `Value must not be empty`.
- `enum` performs deep-equality membership against `values` (any JSON values);
  failure → code `enum` with the message above.
- `one_of` runs each spec in `of` against the value at the same field path;
  if any branch produces zero errors the value is valid, otherwise a single
  `one_of` error is produced at that field (branch errors are discarded).
- **Error path format** (JSONPath-ish, fixed for all languages): field names
  are joined with `.` (`user.address.city`); array indices are appended in
  brackets (`tags[0]`, `users[1].email`). A type error on the top-level
  (non-object) schema input uses the field name `root`.
- **Fail-fast mode** (opt-in, runtime option — not part of the DSL): when
  enabled on a `Schema`/`MapValidator`/nested-object validation, validation
  stops at the first error produced and returns a result containing exactly
  that one error. Names per language: Go `WithFailFast()` SchemaOption /
  `MapOptions.FailFast`; Python `Schema(..., fail_fast=True)` /
  `DictValidator(fail_fast=True)`; Rust `.fail_fast(true)`; TypeScript
  `{ failFast: true }`. Error ordering across sibling fields is unspecified,
  so WHICH error is "first" may differ across runtimes — only the count (1)
  is guaranteed. Default off.
- **No coercion/transform layer** (by design, out of scope): validators never
  mutate or normalize input. Trimming, case-folding, and parsing strings into
  numbers/dates are the caller's responsibility before validation.

## Security

- **`sensitive` option** (per validator, NOT inherited by nested validators):
  when set, the offending value is redacted — every `{v}` interpolation in
  error messages renders as the literal `***`, and `error.value` is set to
  null-ish (`nil` / `None` / `Value::Null` / `undefined`). Use for fields that
  may contain secrets (passwords, tokens) so errors can be logged safely.
- **Pattern input length cap**: before running any `pattern` check (including
  the fixed email/url/uuid/date/datetime/numeric patterns), implementations
  MUST skip the match and report the normal pattern error when the input
  string exceeds **10,000** characters. This bounds worst-case match cost on
  the backtracking regex engines (Python `re`, JS `RegExp`); Go (RE2) and Rust
  (`regex`) are linear-time but apply the same cap for parity.
- **Untrusted patterns**: `pattern` is executed with the platform regex engine;
  in Python and TypeScript a hostile pattern can still cause exponential
  backtracking on short inputs. Schemas MUST come from trusted sources.
- **Email/URL are sanity checks only**: the email pattern does not enforce RFC
  5321 length limits and cannot prove deliverability; the URL pattern performs
  no IDN/port validation and accepts `http://x.`.

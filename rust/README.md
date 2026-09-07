# Universal Data Validator (Rust)

A comprehensive data validation library for Rust that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, integer, float, bool, email, URL, UUID, date, datetime, numeric, not-empty, enum, one-of, array (`Vec`) and object (`Map`)
- ✅ **Schema-Based Validation**: Define complex, nested data structures
- ✅ **Custom Validators**: Add your own validation logic
- ✅ **Nullable & Optional**: Fine-grained control over missing and `null` values
- ✅ **Strict Mode**: Reject unknown fields
- ✅ **Fail-fast Mode**: Stop at the first validation error
- ✅ **Sensitive Fields**: Redact secrets from error messages and error values
- ✅ **Clear Error Messages**: Detailed error reporting with full field paths
- ✅ **Production Ready**: Comprehensive test coverage

## Installation

Add this to your `Cargo.toml`:

```toml
[dependencies]
universal-validator = "1.0"
serde_json = "1.0"
```

## Quick Start

```rust
use serde_json::json;
use universal_validator::{Schema, validators};

fn main() {
    let schema = Schema::new()
        .field("email", validators::email())
        .field("age", validators::int().min(0).max(120))
        .field("username", validators::string().min(3).max(20));

    let data = json!({
        "email": "user@example.com",
        "age": 25,
        "username": "john_doe"
    });

    let result = schema.validate(&data);

    if result.is_valid() {
        println!("✅ Data is valid!");
    } else {
        println!("❌ Validation errors:");
        for error in result.errors() {
            println!("  {}: {}", error.field, error.message);
        }
    }
}
```

## Usage Examples

### String Validation

```rust
use universal_validator::validators;
use regex::Regex;

// Basic string
let validator = validators::string();

// String with length constraints
let username_validator = validators::string().min(3).max(20);

// String with a pre-compiled regex pattern
let phone_validator = validators::string().pattern(Regex::new(r"^\d{3}-\d{4}$").unwrap());

// String with a fallible pattern compile
let pattern_validator = validators::string().try_pattern(r"^[a-z]+$").unwrap();

// String with choices
let theme_validator = validators::string().choices(vec![
    "light".to_string(),
    "dark".to_string(),
]);

// Optional and nullable string
let bio_validator = validators::string().optional().nullable();
```

### Integer Validation

```rust
use universal_validator::validators;

// Integer with range
let age_validator = validators::int().min(0).max(120);

// Optional integer
let score_validator = validators::int().optional();

// Reject floats like 1.5
let result = age_validator.validate(&serde_json::json!(1.5), "age");
assert!(!result.is_valid());
```

### Float Validation

```rust
use universal_validator::validators;

let score_validator = validators::float().min(0.0).max(100.0);
```

### Boolean Validation

```rust
use universal_validator::validators;

let active_validator = validators::bool();
```

### Email and URL Validation

```rust
use universal_validator::validators;

let email_validator = validators::email();
let url_validator = validators::url();
```

### Format, Membership, and Union Validation

```rust
use serde_json::json;
use universal_validator::validators;
use regex::Regex;

// Fixed-format string validators
let id_validator = validators::uuid();
let date_validator = validators::date();         // YYYY-MM-DD
let ts_validator = validators::datetime();       // ISO-8601 style
let numeric_validator = validators::numeric();   // string containing a number

// Non-empty string, array, or object
let required_body = validators::not_empty();

// Enum membership over arbitrary JSON values (deep equality).
// Named `enum_values` because `enum` is a Rust keyword; `validators::r#enum`
// is available as an alias.
let enum_validator = validators::enum_values(vec![json!("a"), json!(1), json!(true)]);

// Union: valid when at least one branch produces zero errors
let union_validator = validators::one_of(vec![
    Box::new(validators::string().min(3)),
    Box::new(validators::int().min(10)),
]);

// String validator anchored to a caller-supplied regex
let pattern_validator = validators::regex(Regex::new(r"^[a-z]+$").unwrap());
```

### Array and Object Validation

```rust
use universal_validator::validators;
use serde_json::json;

// Array of integers
let tags_validator = validators::vec()
    .item(validators::int().min(1))
    .min(1)
    .max(10);

// Nested object
let user_validator = validators::map()
    .field("name", validators::string().min(1))
    .field("age", validators::int().min(0));

let data = json!({"name": "Ada", "age": 30});
assert!(user_validator.validate(&data, "user").is_valid());
```

### Schema Validation

```rust
use serde_json::json;
use universal_validator::{Schema, validators};

let schema = Schema::new()
    .field("username", validators::string().min(3).max(20))
    .field("email", validators::email())
    .field("age", validators::int().min(13).optional())
    .field("bio", validators::string().max(500).optional().nullable())
    .field("website", validators::url().optional().nullable())
    .field("scores", validators::vec().item(validators::int().min(0)).optional())
    .strict(true);

let data = json!({
    "username": "john_doe",
    "email": "john@example.com",
    "age": 25,
    "scores": [80, 95]
});

let result = schema.validate(&data);

if result.is_valid() {
    println!("✅ Data is valid!");
} else {
    println!("❌ Validation errors:");
    for error in result.errors() {
        println!("  {}: {}", error.field, error.message);
    }
}
```

## Strict Mode

When strict mode is enabled, keys present in the input but not in the schema produce an `Unknown field` error.

```rust
use serde_json::json;
use universal_validator::{Schema, validators};

let schema = Schema::new()
    .field("name", validators::string())
    .strict(true);

let result = schema.validate(&json!({"name": "Ada", "extra": 1}));
assert!(!result.is_valid());
assert_eq!(result.errors()[0].message, "Unknown field");
assert_eq!(result.errors()[0].field, "extra");
```

## Fail-fast Mode

When fail-fast mode is enabled, validation stops at the first error and returns a `ValidationResult` containing exactly one error. Set `.fail_fast(true)` on a `Schema`, `MapValidator`, `VecValidator`, or any leaf validator. The flag is propagated to nested validators, so enabling it on a parent also applies to children.

```rust
use serde_json::json;
use universal_validator::{Schema, validators};

let schema = Schema::new()
    .fail_fast(true)
    .field("email", validators::email())
    .field("age", validators::int().min(0));

let result = schema.validate(&json!({"email": "not-an-email", "age": -1}));
assert!(!result.is_valid());
assert_eq!(result.errors().len(), 1);
```

Because `HashMap` iteration order is not deterministic, *which* error is returned first may vary; only the count (`1`) is guaranteed.

## Sensitive Fields

Mark fields that may contain secrets (passwords, tokens) with
`.sensitive(true)`. Every error a sensitive validator produces redacts the
offending value: message interpolations render as `***` and
`ValidationError::value` is `None`, so errors can be logged safely. Messages
that never embed the input (length/bounds, `Field is required`,
`Unknown field`) are unchanged.

The flag is **per-validator** and is *not* inherited by nested validators —
mark each sensitive field explicitly.

```rust
use serde_json::json;
use universal_validator::validators;

let validator = validators::email().sensitive(true);
let result = validator.validate(&json!("p@ssw0rd"), "secret");
assert_eq!(result.errors()[0].message, "Invalid email address: ***");
assert!(result.errors()[0].value.is_none());
```

## Pattern Security Notes

- **Input length cap**: before running any `pattern` check (including the
  fixed email/URL/UUID/date/datetime/numeric patterns), inputs longer than
  `MAX_PATTERN_INPUT_LENGTH` (**10,000** characters) skip the match and
  produce the normal pattern error. Rust's `regex` crate guarantees
  linear-time matching, so this is not strictly needed here — the cap exists
  for cross-language parity with backtracking engines (Python `re`,
  JS `RegExp`).
- **Trusted patterns only**: `pattern` is executed with the platform regex
  engine; in other runtimes a hostile pattern can cause exponential
  backtracking. Schemas should come from trusted sources.
- **Email/URL are sanity checks only**: the email pattern does not enforce
  RFC 5321 length limits and cannot prove deliverability; the URL pattern
  performs no IDN/port validation and accepts `http://x.`.

## Custom Validators

Add ad-hoc validation with `.custom()`:

```rust
use serde_json::json;
use universal_validator::validators;

let even_validator = validators::int().custom(|value| {
    if let Some(n) = value.as_i64() {
        if n % 2 == 0 {
            return None;
        }
    }
    Some("value must be even".to_string())
});

assert!(even_validator.validate(&json!(4), "x").is_valid());
assert!(!even_validator.validate(&json!(3), "x").is_valid());
```

## Error Handling

```rust
use serde_json::json;
use universal_validator::{Schema, validators};

let schema = Schema::new().field("email", validators::email());

let result = schema.validate(&json!({"email": "invalid"}));

if !result.is_valid() {
    for error in result.errors() {
        eprintln!("{}: {}", error.field, error.message);
        if let Some(value) = &error.value {
            eprintln!("  Invalid value: {}", value);
        }
    }
}
```

## API Reference

All validators and `Schema` support `.fail_fast(true)`: validation stops at the first error and the flag is propagated to nested children.

All validators also support `.sensitive(bool)`: redact the offending value from error messages (`***`) and leave `ValidationError::value` as `None`. Per-validator; not inherited by nested validators.

### Validators

- `validators::string()` - String validator
  - `.min(n)` - Minimum Unicode length
  - `.max(n)` - Maximum Unicode length
  - `.pattern(Regex)` - Pre-compiled regex pattern (infallible)
  - `.try_pattern(&str)` - Compile a regex, returns `Result<Self, regex::Error>`
  - `.choices(Vec<String>)` - Allowed values
  - `.optional()` - Make field optional
  - `.nullable()` - Allow explicit `null`
  - `.custom(...)` - Add custom validator

- `validators::int()` - Integer validator
  - `.min(n)` - Minimum value
  - `.max(n)` - Maximum value
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::float()` - Number validator (accepts any JSON number)
  - `.min(n)`, `.max(n)`, `.optional()`, `.nullable()`, `.custom(...)`

- `validators::bool()` - Boolean validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::email()` - Email validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::url()` - URL validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::uuid()` - UUID string validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::date()` - `YYYY-MM-DD` date string validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::datetime()` - ISO-8601 style datetime string validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::numeric()` - Numeric string validator
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::not_empty()` - Requires a non-empty string, array, or object
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::enum_values(Vec<Value>)` / `validators::r#enum(...)` - Membership in a set of JSON values
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::one_of(Vec<Box<dyn Validator>>)` - Accepts the value if any branch validates cleanly
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::regex(Regex)` - String validator anchored to a caller-supplied pattern
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::vec()` / `validators::list()` - Array validator
  - `.item(validator)` - Validator for each item
  - `.min(n)` / `.max(n)` - Array length bounds
  - `.fail_fast(bool)` - Stop at the first item error
  - `.optional()`, `.nullable()`, `.custom(...)`

- `validators::map()` / `validators::dict()` - Object validator
  - `.field(name, validator)` - Add a child field
  - `.schema(HashMap<...>)` - Replace the whole schema
  - `.strict(bool)` - Reject unknown keys
  - `.fail_fast(bool)` - Stop at the first field error
  - `.optional()`, `.nullable()`, `.custom(...)`

### Schema

- `Schema::new()` - Create a new schema
- `.field(name, validator)` - Add a top-level field validator
- `.strict(bool)` - Enable strict mode
- `.fail_fast(bool)` - Stop at the first validation error
- `.validate(data)` - Validate a `serde_json::Value` object
- `.validate_or_raise(data)` - Returns `Ok(())` or `Err(ValidationResult)`
- `.try_validate(data)` - Returns `Ok(())` or `Err(ValidationErrors)`, an error
  aggregate implementing `std::error::Error` (one `field: message` per line)

### ValidationResult

- `.is_valid()` - Check if validation passed
- `.errors()` - Get a slice of validation errors
- `.into_result()` - Convert to `Ok(())` or `Err(ValidationErrors)`

### ValidationError

- `.field` - Field path that failed validation
- `.code` - Machine-readable error code (e.g. `required`, `type`, `email`)
- `.message` - Error message
- `.value` - Optional `serde_json::Value` that failed validation

## Error Paths

Error `field` values follow a fixed JSONPath-ish format: object fields are
joined with `.` (`user.address.city`) and array indices are appended in
brackets (`tags[0]`, `users[1].email`). A type error on a top-level
(non-object) schema input uses the field name `root`.

## No Coercion / Transform

By design, validators never mutate or normalize input. Trimming,
case-folding, and parsing strings into numbers or dates are the caller's
responsibility before validation.

## Testing

```bash
# Run tests
cargo test

# Run tests with output
cargo test -- --nocapture

# Build
cargo build
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Changelog

### 1.0.0 (2024-01-15)
- Initial release
- Support for string, int, email, URL validators
- Schema-based validation
- Comprehensive test coverage

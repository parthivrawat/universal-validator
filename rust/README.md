# Universal Data Validator (Rust)

A comprehensive data validation library for Rust that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, Int, Email, URL
- ✅ **Schema-Based Validation**: Define complex data structures
- ✅ **Custom Validators**: Add your own validation logic
- ✅ **Clear Error Messages**: Detailed error reporting with field paths
- ✅ **Minimal Dependencies**: Only regex for pattern matching
- ✅ **Production Ready**: Comprehensive test coverage

## Installation

Add this to your `Cargo.toml`:

```toml
[dependencies]
universal-validator = "1.0"
```

## Quick Start

```rust
use universal_validator::{Schema, validators};
use std::collections::HashMap;

fn main() {
    let schema = Schema::new()
        .field("email", validators::email())
        .field("age", validators::int().min(0).max(120))
        .field("username", validators::string().min(3).max(20));

    let mut data = HashMap::new();
    data.insert("email".to_string(), "user@example.com".to_string());
    data.insert("age".to_string(), "25".to_string());
    data.insert("username".to_string(), "john_doe".to_string());

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

// Basic string
let validator = validators::string();

// String with length constraints
let username_validator = validators::string().min(3).max(20);

// String with pattern
let phone_validator = validators::string().pattern(r"^\d{3}-\d{4}$");

// String with choices
let theme_validator = validators::string().choices(vec![
    "light".to_string(),
    "dark".to_string(),
]);

// Optional string
let bio_validator = validators::string().optional();
```

### Integer Validation

```rust
// Integer with range
let age_validator = validators::int().min(0).max(120);

// Optional integer
let score_validator = validators::int().optional();
```

### Email and URL Validation

```rust
let email_validator = validators::email();
let url_validator = validators::url();
```

### Schema Validation

```rust
use universal_validator::{Schema, validators};
use std::collections::HashMap;

let schema = Schema::new()
    .field("username", validators::string().min(3).max(20))
    .field("email", validators::email())
    .field("age", validators::int().min(13).optional())
    .field("bio", validators::string().max(500).optional())
    .field("website", validators::url().optional());

let mut data = HashMap::new();
data.insert("username".to_string(), "john_doe".to_string());
data.insert("email".to_string(), "john@example.com".to_string());
data.insert("age".to_string(), "25".to_string());

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

## Real-World Examples

### User Registration

```rust
use universal_validator::{Schema, validators};

let registration_schema = Schema::new()
    .field("username", validators::string().min(3).max(20))
    .field("email", validators::email())
    .field("password", validators::string().min(8))
    .field("age", validators::int().min(13).optional());
```

### API Request Validation

```rust
let api_request_schema = Schema::new()
    .field("method", validators::string().choices(vec![
        "GET".to_string(),
        "POST".to_string(),
        "PUT".to_string(),
        "DELETE".to_string(),
    ]))
    .field("url", validators::url())
    .field("timeout", validators::int().min(0).optional());
```

## Error Handling

```rust
use universal_validator::{Schema, validators};
use std::collections::HashMap;

let schema = Schema::new()
    .field("email", validators::email());

let mut data = HashMap::new();
data.insert("email".to_string(), "invalid".to_string());

let result = schema.validate(&data);

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

### Validators

- `validators::string()` - String validator
  - `.min(n)` - Minimum length
  - `.max(n)` - Maximum length
  - `.pattern(regex)` - Regex pattern
  - `.choices(vec)` - Allowed values
  - `.optional()` - Make field optional

- `validators::int()` - Integer validator
  - `.min(n)` - Minimum value
  - `.max(n)` - Maximum value
  - `.optional()` - Make field optional

- `validators::email()` - Email validator
  - `.optional()` - Make field optional

- `validators::url()` - URL validator
  - `.optional()` - Make field optional

### Schema

- `Schema::new()` - Create a new schema
- `.field(name, validator)` - Add a field validator
- `.validate(data)` - Validate data and return ValidationResult

### ValidationResult

- `.is_valid()` - Check if validation passed
- `.errors()` - Get list of validation errors

### ValidationError

- `.field` - Field name that failed validation
- `.message` - Error message
- `.value` - Optional value that failed validation

## Testing

```bash
# Run tests
cargo test

# Run tests with output
cargo test -- --nocapture

# Run tests with coverage
cargo tarpaulin --out Html
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

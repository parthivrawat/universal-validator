# Universal Data Validator

A comprehensive data validation library for Python that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, int, float, bool, email, URL, UUID, date, datetime, numeric, not_empty, enum, one_of, list, dict
- ✅ **Schema-Based Validation**: Define complex data structures
- ✅ **Strict Mode**: Reject unknown keys in schemas and dict validators
- ✅ **Fail-Fast Mode**: Stop validation at the first error
- ✅ **Custom Validators**: Add your own validation logic
- ✅ **Nested Validation**: Validate nested objects and lists
- ✅ **Clear Error Messages**: Detailed error reporting with field paths
- ✅ **Zero Dependencies**: No external dependencies required
- ✅ **Production Ready**: Comprehensive test coverage

## Installation

```bash
pip install universal-validator
```

## Quick Start

```python
from universal_validator import Schema, validators

schema = Schema({
    'email': validators.email(),
    'age': validators.int(min_value=0, max_value=120),
    'username': validators.string(min_length=3, max_length=20),
    'tags': validators.list(validators.string())
})

result = schema.validate(data)
if not result.valid:
    for error in result.errors:
        print(f"{error.field}: {error.message}")
```

## Package Layout

The library is packaged as `universal_validator` with two submodules:

- `universal_validator.core` — validator classes, `Schema`, `ValidationError`, `ValidationErrors`, `ValidationResult`, and `Code`.
- `universal_validator.validators` — factory functions such as `validators.string`, `validators.int`, `validators.uuid`, etc.

All public names are re-exported from `universal_validator` and listed in `__all__`.

## Usage Examples

### String Validation

```python
from universal_validator import validators

# Basic string
validator = validators.string()

# String with length constraints
validator = validators.string(min_length=3, max_length=20)

# String with pattern
validator = validators.string(pattern=r'^\d{3}-\d{4}$')

# String with choices
validator = validators.string(choices=['red', 'green', 'blue'])

# Optional string
validator = validators.string(required=False)

# Nullable string
validator = validators.string(nullable=True)
```

### Integer and Float Validation

```python
# Integer with range
age_validator = validators.int(min_value=0, max_value=120)

# Float with range
price_validator = validators.float(min_value=0.0, max_value=9999.99)

# Optional integer
count_validator = validators.int(required=False)
```

### Boolean Validation

```python
# Boolean
terms_validator = validators.bool()

# Optional boolean
newsletter_validator = validators.bool(required=False)
```

### Email and URL Validation

```python
# Email validation
email_validator = validators.email()

# URL validation
url_validator = validators.url()
```

### List Validation

```python
# Simple list
tags_validator = validators.list()

# List with item validation
numbers_validator = validators.list(validators.int())

# List with length constraints
items_validator = validators.list(
    validators.string(),
    min_length=1,
    max_length=10
)
```

### Dictionary Validation

```python
# Dictionary with schema
address_validator = validators.dict({
    'street': validators.string(),
    'city': validators.string(),
    'zip': validators.string(pattern=r'^\d{5}$')
})

# Nested dictionary
user_validator = validators.dict({
    'name': validators.string(),
    'email': validators.email(),
    'address': validators.dict({
        'street': validators.string(),
        'city': validators.string()
    })
})
```

### Strict Mode

Enable `strict=True` on `Schema` or `validators.dict(...)` to report unknown keys as validation errors.

```python
schema = Schema({
    'name': validators.string()
}, strict=True)

result = schema.validate({'name': 'John', 'extra': 1})
# Error: extra: Unknown field

nested = validators.dict({
    'a': validators.dict({'b': validators.string()}, strict=True)
})

result = nested.validate({'a': {'b': 'ok', 'c': 2}}, 'root')
# Error: root.a.c: Unknown field
```

### Fail-Fast Mode

By default, validators collect all errors. Pass `fail_fast=True` to `Schema` or `validators.dict(...)` (and `validators.list(...)`) to stop at the first error. The result contains exactly one error; which sibling field is reported first follows the validator's iteration order and is not guaranteed across implementations.

```python
schema = Schema({
    'username': validators.string(min_length=3),
    'email': validators.email(),
    'age': validators.int(min_value=0)
}, fail_fast=True)

result = schema.validate({'username': 'ab', 'email': 'bad', 'age': -1})
# Only the first error encountered is reported
```

```python
profile = validators.dict({
    'name': validators.string(min_length=2),
    'age': validators.int(min_value=0)
}, fail_fast=True)

result = profile.validate({'name': 'x', 'age': -1}, 'profile')
# Only the first error encountered is reported
```

### Schema Validation

```python
from universal_validator import Schema, validators

# Define schema
user_schema = Schema({
    'username': validators.string(min_length=3, max_length=20),
    'email': validators.email(),
    'age': validators.int(min_value=13, required=False),
    'bio': validators.string(max_length=500, required=False),
    'tags': validators.list(validators.string()),
    'settings': validators.dict({
        'theme': validators.string(choices=['light', 'dark']),
        'notifications': validators.bool()
    })
})

# Validate data
data = {
    'username': 'john_doe',
    'email': 'john@example.com',
    'age': 25,
    'tags': ['python', 'javascript'],
    'settings': {
        'theme': 'dark',
        'notifications': True
    }
}

result = user_schema.validate(data)

if result.valid:
    print("✅ Data is valid!")
else:
    print("❌ Validation errors:")
    for error in result.errors:
        print(f"  {error.field}: {error.message}")
```

### Custom Validators

```python
from universal_validator import validators

def is_even(value):
    """Custom validator to check if number is even."""
    if value % 2 != 0:
        return "Value must be even"
    return None

# Add custom validator
validator = validators.int().custom(is_even)

result = validator.validate(4, 'number')
print(result.valid)  # True

result = validator.validate(3, 'number')
print(result.valid)  # False
print(result.errors[0].message)  # "Value must be even"
```

`Validator.custom(fn)` returns a **clone** of the validator with the new custom function appended; the original validator is unchanged.

### Multiple Custom Validators

```python
def is_positive(value):
    return "Must be positive" if value <= 0 else None

def is_even(value):
    return "Must be even" if value % 2 != 0 else None

validator = validators.int().custom(is_positive).custom(is_even)

result = validator.validate(4, 'field')
print(result.valid)  # True

result = validator.validate(-2, 'field')
print(result.valid)  # False
```

## Real-World Examples

### User Registration

```python
from universal_validator import Schema, validators

registration_schema = Schema({
    'username': validators.string(min_length=3, max_length=20),
    'email': validators.email(),
    'password': validators.string(min_length=8),
    'confirm_password': validators.string(min_length=8),
    'age': validators.int(min_value=13, required=False),
    'terms_accepted': validators.bool()
})

data = {
    'username': 'john_doe',
    'email': 'john@example.com',
    'password': 'secure_password_123',
    'confirm_password': 'secure_password_123',
    'age': 25,
    'terms_accepted': True
}

result = registration_schema.validate(data)
```

### API Request Validation

```python
api_request_schema = Schema({
    'method': validators.string(choices=['GET', 'POST', 'PUT', 'DELETE']),
    'url': validators.url(),
    'headers': validators.dict(required=False),
    'body': validators.dict(required=False),
    'timeout': validators.float(min_value=0, required=False)
})

request_data = {
    'method': 'POST',
    'url': 'https://api.example.com/users',
    'headers': {'Content-Type': 'application/json'},
    'body': {'name': 'John'},
    'timeout': 30.0
}

result = api_request_schema.validate(request_data)
```

### Configuration Validation

```python
config_schema = Schema({
    'database': validators.dict({
        'host': validators.string(),
        'port': validators.int(min_value=1, max_value=65535),
        'username': validators.string(),
        'password': validators.string(),
        'ssl': validators.bool()
    }),
    'cache': validators.dict({
        'enabled': validators.bool(),
        'ttl': validators.int(min_value=0),
        'max_size': validators.int(min_value=1)
    }),
    'features': validators.list(validators.string())
})

config = {
    'database': {
        'host': 'localhost',
        'port': 5432,
        'username': 'admin',
        'password': 'secret',
        'ssl': True
    },
    'cache': {
        'enabled': True,
        'ttl': 3600,
        'max_size': 1000
    },
    'features': ['feature1', 'feature2']
}

result = config_schema.validate(config)
```

### Sensitive Fields (Redaction)

Pass `sensitive=True` to any validator to redact the offending value from its
errors — useful for fields that may contain secrets (passwords, tokens) so
errors can be logged safely.

```python
password = validators.string(min_length=8, sensitive=True)

result = password.validate('hunter2', 'password')
print(result.errors[0].message)  # "String length must be at least 8, got 7"
print(result.errors[0].value)    # None

secret_email = validators.email(sensitive=True)
result = secret_email.validate('p@ssw0rd', 'email')
print(result.errors[0].message)  # "Invalid email address: ***"
print(result.errors[0].value)    # None
```

When `sensitive=True`:
- `error.value` is always `None`.
- Every error message that would interpolate the input value renders `***`
  instead (e.g. `Invalid email address: ***`, `Value must be one of a, b, got '***'`).
- Messages that do not embed the input value (e.g. `Field is required`,
  `Unknown field`, length bounds) are unchanged.

`sensitive` is per-validator and is **not** inherited by nested validators —
mark each sensitive field individually.

## Security

- **Trusted schemas only.** `pattern` (and `validators.regex`) compiles and runs
  with Python's `re`, a backtracking engine. A hostile pattern can cause
  exponential backtracking even on short inputs, so schemas and patterns MUST
  come from trusted sources.
- **Pattern input length cap.** Before running any pattern check — including the
  fixed email, URL, UUID, date, datetime, and numeric patterns — inputs longer
  than `MAX_PATTERN_INPUT_LENGTH` (10,000 characters, exported from
  `universal_validator`) are rejected with the normal pattern error without
  running the regex. This bounds worst-case match cost on the backtracking
  `re` engine.
- **Redact secrets.** Use `sensitive=True` on validators for fields that may
  contain secrets so `error.value` and messages never leak the input into logs.
- **Email/URL are sanity checks only.** The email pattern is permissive: it does
  not enforce RFC 5321 length limits and cannot prove deliverability. The URL
  pattern performs no IDN or port validation and accepts inputs such as
  `http://x.`. Do not rely on them for security decisions.

## Error Handling

```python
from universal_validator import Schema, validators, ValidationErrors

schema = Schema({
    'email': validators.email()
})

# Option 1: Check result
result = schema.validate({'email': 'invalid'})
if not result.valid:
    for error in result.errors:
        print(f"{error.field}: {error.message}")

# Option 2: Raise exception
try:
    schema.validate_or_raise({'email': 'invalid'})
except ValidationErrors as e:
    print(f"Validation failed: {e}")
```

`validate_or_raise` raises a `ValidationErrors` aggregate exception whose `.errors` list contains every `ValidationError`. Error paths follow the format `parent.child` for objects and `field[0]` for array items; a top-level type error on a `Schema` input uses the field name `root`.

## API Reference

### Validators

- `validators.string(min_length, max_length, pattern, choices, required, nullable, fail_fast, sensitive)` — String validator
- `validators.int(min_value, max_value, required, nullable, fail_fast, sensitive)` — Integer validator
- `validators.float(min_value, max_value, required, nullable, fail_fast, sensitive)` — Number validator
- `validators.bool(required, nullable, fail_fast, sensitive)` — Boolean validator
- `validators.email(required, nullable, fail_fast, sensitive)` — Email validator
- `validators.url(required, nullable, fail_fast, sensitive)` — URL validator
- `validators.list(item_validator, min_length, max_length, required, nullable, fail_fast, sensitive)` — List validator
- `validators.dict(schema, required, nullable, strict, fail_fast, sensitive)` — Dictionary validator
- `validators.uuid(required, nullable, fail_fast, sensitive)` — UUID string validator
- `validators.date(required, nullable, fail_fast, sensitive)` — `YYYY-MM-DD` string validator
- `validators.datetime(required, nullable, fail_fast, sensitive)` — ISO-8601 datetime string validator
- `validators.numeric(required, nullable, fail_fast, sensitive)` — Numeric string validator
- `validators.not_empty(required, nullable, fail_fast, sensitive)` — Non-empty string/array/object validator
- `validators.enum(values, required, nullable, fail_fast, sensitive)` — Arbitrary value enum validator
- `validators.one_of(*validators, required, nullable, fail_fast, sensitive)` — Union/disjunction validator
- `validators.regex(pattern, required, nullable, fail_fast, sensitive)` — String validator anchored to a pattern

All validators accept `sensitive=False` by default; see "Sensitive Fields (Redaction)" above.

### Schema

- `Schema(schema_dict, strict=False, fail_fast=False)` — Create a schema
- `schema.validate(data)` — Validate data and return ValidationResult
- `schema.validate_or_raise(data)` — Validate data and raise `ValidationErrors` if invalid

### ValidationResult

- `result.valid` - Boolean indicating if validation passed
- `result.errors` - List of ValidationError objects

### ValidationError

- `error.field` - Field name that failed validation
- `error.message` - Error message
- `error.value` - Value that failed validation
- `error.code` - Machine-readable canonical error code (e.g. `required`, `type`, `min_length`)

Canonical error codes are available via the `Code` class:

```python
from universal_validator import Code

assert result.errors[0].code == Code.REQUIRED  # or "required"
```

Codes include: `required`, `unknown_field`, `type`, `min_length`, `max_length`, `min_value`, `max_value`, `pattern`, `choices`, `email`, `url`, `custom`, `uuid`, `date`, `datetime`, `numeric`, `not_empty`, `enum`, `one_of`.

## Testing

```bash
# Install dev dependencies
pip install -e ".[dev]"

# Run tests
pytest test_universal_validator.py -v

# Run with coverage
pytest test_universal_validator.py --cov=universal_validator --cov-report=html
```

## Comparison with Other Libraries

| Feature | universal-validator | pydantic | marshmallow | cerberus |
|---------|-------------------|----------|-------------|----------|
| Schema-based | ✅ | ✅ | ✅ | ✅ |
| Custom validators | ✅ | ✅ | ✅ | ✅ |
| Zero dependencies | ✅ | ❌ | ❌ | ✅ |
| Nested validation | ✅ | ✅ | ✅ | ✅ |
| Clear errors | ✅ | ✅ | ✅ | ✅ |
| Simple API | ✅ | ❌ | ❌ | ✅ |
| Type hints | ✅ | ✅ | ❌ | ❌ |

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Notes

- Coercion, trimming, case-folding, and parsing strings into numbers/dates are intentionally out of scope; validate inputs after normalizing them.
- The public API is listed in `universal_validator.__all__`.

## Changelog

### 1.0.0 (2024-01-15)
- Initial release
- Support for string, int, float, bool, email, URL, list, dict validators
- Schema-based validation
- Custom validators
- Nested validation
- Comprehensive test coverage

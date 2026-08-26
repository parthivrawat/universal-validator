# Universal Data Validator

A comprehensive data validation library for Python that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, int, float, bool, email, URL, list, dict
- ✅ **Schema-Based Validation**: Define complex data structures
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

## Error Handling

```python
from universal_validator import Schema, validators, ValidationError

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
except ValidationError as e:
    print(f"Validation failed: {e}")
```

## API Reference

### Validators

- `validators.string(min_length, max_length, pattern, choices, required, nullable)` - String validator
- `validators.int(min_value, max_value, required, nullable)` - Integer validator
- `validators.float(min_value, max_value, required, nullable)` - Float validator
- `validators.bool(required, nullable)` - Boolean validator
- `validators.email(required, nullable)` - Email validator
- `validators.url(required, nullable)` - URL validator
- `validators.list(item_validator, min_length, max_length, required, nullable)` - List validator
- `validators.dict(schema, required, nullable)` - Dictionary validator

### Schema

- `Schema(schema_dict)` - Create a schema
- `schema.validate(data)` - Validate data and return ValidationResult
- `schema.validate_or_raise(data)` - Validate data and raise ValidationError if invalid

### ValidationResult

- `result.valid` - Boolean indicating if validation passed
- `result.errors` - List of ValidationError objects

### ValidationError

- `error.field` - Field name that failed validation
- `error.message` - Error message
- `error.value` - Value that failed validation

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

## Changelog

### 1.0.0 (2024-01-15)
- Initial release
- Support for string, int, float, bool, email, URL, list, dict validators
- Schema-based validation
- Custom validators
- Nested validation
- Comprehensive test coverage

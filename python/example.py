"""
Example usage of universal-validator library

This example demonstrates how to use the universal-validator library
for various validation scenarios.
"""

from universal_validator import Schema, ValidationErrors, validators


def example_basic_validation():
    """Example: Basic field validation."""
    print("\n=== Example 1: Basic Validation ===\n")
    
    # String validation
    name_validator = validators.string(min_length=2, max_length=50)
    result = name_validator.validate('John Doe', 'name')
    print(f"Name validation: {result.valid}")
    
    # Integer validation
    age_validator = validators.int(min_value=0, max_value=120)
    result = age_validator.validate(25, 'age')
    print(f"Age validation: {result.valid}")
    
    # Email validation
    email_validator = validators.email()
    result = email_validator.validate('john@example.com', 'email')
    print(f"Email validation: {result.valid}")


def example_user_registration():
    """Example: User registration form validation."""
    print("\n=== Example 2: User Registration ===\n")
    
    registration_schema = Schema({
        'username': validators.string(min_length=3, max_length=20),
        'email': validators.email(),
        'password': validators.string(min_length=8),
        'age': validators.int(min_value=13, required=False),
        'bio': validators.string(max_length=500, required=False),
        'terms_accepted': validators.bool()
    })
    
    # Valid data
    valid_data = {
        'username': 'john_doe',
        'email': 'john@example.com',
        'password': 'secure_password_123',
        'age': 25,
        'bio': 'Software developer',
        'terms_accepted': True
    }
    
    result = registration_schema.validate(valid_data)
    print(f"Valid registration: {result.valid}")
    
    # Invalid data
    invalid_data = {
        'username': 'ab',  # Too short
        'email': 'invalid-email',  # Invalid format
        'password': 'short',  # Too short
        'age': 10,  # Too young
        'terms_accepted': False
    }
    
    result = registration_schema.validate(invalid_data)
    print(f"\nInvalid registration: {result.valid}")
    if not result.valid:
        print("Errors:")
        for error in result.errors:
            print(f"  - {error.field}: {error.message}")


def example_nested_validation():
    """Example: Nested object validation."""
    print("\n=== Example 3: Nested Validation ===\n")
    
    user_schema = Schema({
        'name': validators.string(),
        'email': validators.email(),
        'address': validators.dict({
            'street': validators.string(),
            'city': validators.string(),
            'zip': validators.string(pattern=r'^\d{5}$'),
            'country': validators.string()
        }),
        'settings': validators.dict({
            'theme': validators.string(choices=['light', 'dark']),
            'notifications': validators.bool(),
            'language': validators.string(choices=['en', 'es', 'fr'])
        })
    })
    
    data = {
        'name': 'John Doe',
        'email': 'john@example.com',
        'address': {
            'street': '123 Main St',
            'city': 'New York',
            'zip': '10001',
            'country': 'USA'
        },
        'settings': {
            'theme': 'dark',
            'notifications': True,
            'language': 'en'
        }
    }
    
    result = user_schema.validate(data)
    print(f"Nested validation: {result.valid}")


def example_list_validation():
    """Example: List validation."""
    print("\n=== Example 4: List Validation ===\n")
    
    # Simple list
    tags_validator = validators.list(validators.string())
    result = tags_validator.validate(['python', 'javascript', 'go'], 'tags')
    print(f"Tags validation: {result.valid}")
    
    # List of objects
    users_schema = Schema({
        'users': validators.list(
            validators.dict({
                'name': validators.string(),
                'email': validators.email(),
                'age': validators.int(min_value=0)
            }),
            min_length=1
        )
    })
    
    data = {
        'users': [
            {'name': 'John', 'email': 'john@example.com', 'age': 30},
            {'name': 'Jane', 'email': 'jane@example.com', 'age': 25}
        ]
    }
    
    result = users_schema.validate(data)
    print(f"Users list validation: {result.valid}")


def example_custom_validators():
    """Example: Custom validation logic."""
    print("\n=== Example 5: Custom Validators ===\n")
    
    def is_even(value):
        """Check if number is even."""
        if value % 2 != 0:
            return "Value must be even"
        return None
    
    def is_positive(value):
        """Check if number is positive."""
        if value <= 0:
            return "Value must be positive"
        return None
    
    # Single custom validator
    even_validator = validators.int().custom(is_even)
    
    result = even_validator.validate(4, 'number')
    print(f"Even number (4): {result.valid}")
    
    result = even_validator.validate(3, 'number')
    print(f"Even number (3): {result.valid}")
    if not result.valid:
        print(f"  Error: {result.errors[0].message}")
    
    # Multiple custom validators
    positive_even_validator = validators.int().custom(is_positive).custom(is_even)
    
    result = positive_even_validator.validate(4, 'number')
    print(f"\nPositive even (4): {result.valid}")
    
    result = positive_even_validator.validate(-2, 'number')
    print(f"Positive even (-2): {result.valid}")
    if not result.valid:
        print(f"  Error: {result.errors[0].message}")


def example_api_validation():
    """Example: API request validation."""
    print("\n=== Example 6: API Request Validation ===\n")
    
    api_request_schema = Schema({
        'method': validators.string(choices=['GET', 'POST', 'PUT', 'DELETE', 'PATCH']),
        'url': validators.url(),
        'headers': validators.dict(required=False),
        'body': validators.dict(required=False),
        'timeout': validators.float(min_value=0, required=False),
        'retry': validators.dict({
            'max_attempts': validators.int(min_value=1, max_value=10),
            'backoff': validators.float(min_value=0)
        }, required=False)
    })
    
    request = {
        'method': 'POST',
        'url': 'https://api.example.com/users',
        'headers': {
            'Content-Type': 'application/json',
            'Authorization': 'Bearer token123'
        },
        'body': {
            'name': 'John Doe',
            'email': 'john@example.com'
        },
        'timeout': 30.0,
        'retry': {
            'max_attempts': 3,
            'backoff': 1.5
        }
    }
    
    result = api_request_schema.validate(request)
    print(f"API request validation: {result.valid}")


def example_config_validation():
    """Example: Application configuration validation."""
    print("\n=== Example 7: Configuration Validation ===\n")
    
    config_schema = Schema({
        'app': validators.dict({
            'name': validators.string(),
            'version': validators.string(pattern=r'^\d+\.\d+\.\d+$'),
            'debug': validators.bool()
        }),
        'database': validators.dict({
            'host': validators.string(),
            'port': validators.int(min_value=1, max_value=65535),
            'name': validators.string(),
            'username': validators.string(),
            'password': validators.string(),
            'pool_size': validators.int(min_value=1, max_value=100),
            'ssl': validators.bool()
        }),
        'cache': validators.dict({
            'enabled': validators.bool(),
            'backend': validators.string(choices=['redis', 'memcached', 'memory']),
            'ttl': validators.int(min_value=0),
            'max_size': validators.int(min_value=1)
        }),
        'logging': validators.dict({
            'level': validators.string(choices=['DEBUG', 'INFO', 'WARNING', 'ERROR']),
            'format': validators.string(choices=['json', 'text']),
            'output': validators.string(choices=['stdout', 'file'])
        }),
        'features': validators.list(validators.string())
    })
    
    config = {
        'app': {
            'name': 'MyApp',
            'version': '1.0.0',
            'debug': False
        },
        'database': {
            'host': 'localhost',
            'port': 5432,
            'name': 'myapp',
            'username': 'admin',
            'password': 'secret',
            'pool_size': 10,
            'ssl': True
        },
        'cache': {
            'enabled': True,
            'backend': 'redis',
            'ttl': 3600,
            'max_size': 1000
        },
        'logging': {
            'level': 'INFO',
            'format': 'json',
            'output': 'stdout'
        },
        'features': ['feature1', 'feature2', 'feature3']
    }
    
    result = config_schema.validate(config)
    print(f"Configuration validation: {result.valid}")


def example_error_handling():
    """Example: Error handling."""
    print("\n=== Example 8: Error Handling ===\n")
    
    schema = Schema({
        'email': validators.email(),
        'age': validators.int(min_value=0)
    })
    
    # Option 1: Check result
    print("Option 1: Check ValidationResult")
    result = schema.validate({'email': 'invalid', 'age': -1})
    if not result.valid:
        print("Validation failed:")
        for error in result.errors:
            print(f"  - {error.field}: {error.message}")
    
    # Option 2: Raise exception
    print("\nOption 2: Raise ValidationError")
    try:
        schema.validate_or_raise({'email': 'invalid', 'age': -1})
    except ValidationErrors as e:
        print(f"Caught exception: {e}")


def example_optional_fields():
    """Example: Optional and nullable fields."""
    print("\n=== Example 9: Optional and Nullable Fields ===\n")
    
    schema = Schema({
        'name': validators.string(),  # Required
        'email': validators.email(),  # Required
        'phone': validators.string(required=False),  # Optional
        'bio': validators.string(nullable=True),  # Can be None
        'age': validators.int(required=False, nullable=True)  # Optional and nullable
    })
    
    # Valid with optional fields missing
    data1 = {
        'name': 'John',
        'email': 'john@example.com',
        'bio': None
    }
    result = schema.validate(data1)
    print(f"With optional fields missing: {result.valid}")
    
    # Valid with all fields
    data2 = {
        'name': 'John',
        'email': 'john@example.com',
        'phone': '123-456-7890',
        'bio': 'Software developer',
        'age': 30
    }
    result = schema.validate(data2)
    print(f"With all fields: {result.valid}")


def example_real_world_ecommerce():
    """Example: E-commerce order validation."""
    print("\n=== Example 10: E-commerce Order ===\n")
    
    order_schema = Schema({
        'order_id': validators.string(pattern=r'^ORD-\d{8}$'),
        'customer': validators.dict({
            'id': validators.string(),
            'name': validators.string(),
            'email': validators.email(),
            'phone': validators.string(required=False)
        }),
        'items': validators.list(
            validators.dict({
                'product_id': validators.string(),
                'name': validators.string(),
                'quantity': validators.int(min_value=1),
                'price': validators.float(min_value=0),
                'discount': validators.float(min_value=0, max_value=100, required=False)
            }),
            min_length=1
        ),
        'shipping': validators.dict({
            'address': validators.dict({
                'street': validators.string(),
                'city': validators.string(),
                'state': validators.string(),
                'zip': validators.string(pattern=r'^\d{5}$'),
                'country': validators.string()
            }),
            'method': validators.string(choices=['standard', 'express', 'overnight']),
            'cost': validators.float(min_value=0)
        }),
        'payment': validators.dict({
            'method': validators.string(choices=['credit_card', 'paypal', 'bank_transfer']),
            'status': validators.string(choices=['pending', 'completed', 'failed']),
            'amount': validators.float(min_value=0)
        }),
        'notes': validators.string(max_length=500, required=False)
    })
    
    order = {
        'order_id': 'ORD-12345678',
        'customer': {
            'id': 'CUST-001',
            'name': 'John Doe',
            'email': 'john@example.com',
            'phone': '555-1234'
        },
        'items': [
            {
                'product_id': 'PROD-001',
                'name': 'Widget',
                'quantity': 2,
                'price': 29.99,
                'discount': 10.0
            },
            {
                'product_id': 'PROD-002',
                'name': 'Gadget',
                'quantity': 1,
                'price': 49.99
            }
        ],
        'shipping': {
            'address': {
                'street': '123 Main St',
                'city': 'New York',
                'state': 'NY',
                'zip': '10001',
                'country': 'USA'
            },
            'method': 'express',
            'cost': 15.00
        },
        'payment': {
            'method': 'credit_card',
            'status': 'completed',
            'amount': 94.98
        },
        'notes': 'Please deliver before 5 PM'
    }
    
    result = order_schema.validate(order)
    print(f"E-commerce order validation: {result.valid}")
    if result.valid:
        print("✅ Order is valid and ready for processing!")


def main():
    """Run all examples."""
    print("=" * 60)
    print("Universal Data Validator - Examples")
    print("=" * 60)
    
    example_basic_validation()
    example_user_registration()
    example_nested_validation()
    example_list_validation()
    example_custom_validators()
    example_api_validation()
    example_config_validation()
    example_error_handling()
    example_optional_fields()
    example_real_world_ecommerce()
    
    print("\n" + "=" * 60)
    print("Examples completed!")
    print("=" * 60)


if __name__ == '__main__':
    main()

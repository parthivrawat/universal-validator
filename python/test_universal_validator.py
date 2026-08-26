"""
Tests for Universal Data Validator Library
"""

import pytest
from universal_validator import (
    Schema, validators, ValidationError, ValidationResult,
    StringValidator, IntValidator, FloatValidator, BoolValidator,
    EmailValidator, UrlValidator, ListValidator, DictValidator
)


class TestStringValidator:
    def test_valid_string(self):
        validator = validators.string()
        result = validator.validate('hello', 'field')
        assert result.valid
        assert len(result.errors) == 0
    
    def test_invalid_type(self):
        validator = validators.string()
        result = validator.validate(123, 'field')
        assert not result.valid
        assert len(result.errors) == 1
        assert 'Expected string' in result.errors[0].message
    
    def test_min_length(self):
        validator = validators.string(min_length=5)
        
        result = validator.validate('hello', 'field')
        assert result.valid
        
        result = validator.validate('hi', 'field')
        assert not result.valid
        assert 'at least 5' in result.errors[0].message
    
    def test_max_length(self):
        validator = validators.string(max_length=5)
        
        result = validator.validate('hello', 'field')
        assert result.valid
        
        result = validator.validate('hello world', 'field')
        assert not result.valid
        assert 'at most 5' in result.errors[0].message
    
    def test_pattern(self):
        validator = validators.string(pattern=r'^\d{3}-\d{4}$')
        
        result = validator.validate('123-4567', 'field')
        assert result.valid
        
        result = validator.validate('invalid', 'field')
        assert not result.valid
        assert 'does not match pattern' in result.errors[0].message
    
    def test_choices(self):
        validator = validators.string(choices=['red', 'green', 'blue'])
        
        result = validator.validate('red', 'field')
        assert result.valid
        
        result = validator.validate('yellow', 'field')
        assert not result.valid
        assert 'must be one of' in result.errors[0].message
    
    def test_required(self):
        validator = validators.string(required=True)
        result = validator.validate(None, 'field')
        assert not result.valid
        assert 'required' in result.errors[0].message.lower()
    
    def test_optional(self):
        validator = validators.string(required=False)
        result = validator.validate(None, 'field')
        assert result.valid
    
    def test_nullable(self):
        validator = validators.string(nullable=True)
        result = validator.validate(None, 'field')
        assert result.valid


class TestIntValidator:
    def test_valid_int(self):
        validator = validators.int()
        result = validator.validate(42, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.int()
        result = validator.validate('not an int', 'field')
        assert not result.valid
        assert 'Expected int' in result.errors[0].message
    
    def test_bool_rejected(self):
        validator = validators.int()
        result = validator.validate(True, 'field')
        assert not result.valid
    
    def test_min_value(self):
        validator = validators.int(min_value=0)
        
        result = validator.validate(5, 'field')
        assert result.valid
        
        result = validator.validate(-1, 'field')
        assert not result.valid
        assert 'at least 0' in result.errors[0].message
    
    def test_max_value(self):
        validator = validators.int(max_value=100)
        
        result = validator.validate(50, 'field')
        assert result.valid
        
        result = validator.validate(101, 'field')
        assert not result.valid
        assert 'at most 100' in result.errors[0].message


class TestFloatValidator:
    def test_valid_float(self):
        validator = validators.float()
        result = validator.validate(3.14, 'field')
        assert result.valid
    
    def test_int_accepted(self):
        validator = validators.float()
        result = validator.validate(42, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.float()
        result = validator.validate('not a float', 'field')
        assert not result.valid
    
    def test_min_value(self):
        validator = validators.float(min_value=0.0)
        
        result = validator.validate(5.5, 'field')
        assert result.valid
        
        result = validator.validate(-0.1, 'field')
        assert not result.valid
    
    def test_max_value(self):
        validator = validators.float(max_value=100.0)
        
        result = validator.validate(50.5, 'field')
        assert result.valid
        
        result = validator.validate(100.1, 'field')
        assert not result.valid


class TestBoolValidator:
    def test_valid_bool(self):
        validator = validators.bool()
        
        result = validator.validate(True, 'field')
        assert result.valid
        
        result = validator.validate(False, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.bool()
        result = validator.validate('not a bool', 'field')
        assert not result.valid
        assert 'Expected bool' in result.errors[0].message


class TestEmailValidator:
    def test_valid_email(self):
        validator = validators.email()
        
        valid_emails = [
            'user@example.com',
            'test.user@example.com',
            'user+tag@example.co.uk',
            'user123@test-domain.com'
        ]
        
        for email in valid_emails:
            result = validator.validate(email, 'field')
            assert result.valid, f"Failed for: {email}"
    
    def test_invalid_email(self):
        validator = validators.email()
        
        invalid_emails = [
            'not-an-email',
            '@example.com',
            'user@',
            'user @example.com',
            'user@example',
        ]
        
        for email in invalid_emails:
            result = validator.validate(email, 'field')
            assert not result.valid, f"Should fail for: {email}"
            assert 'Invalid email' in result.errors[0].message


class TestUrlValidator:
    def test_valid_url(self):
        validator = validators.url()
        
        valid_urls = [
            'http://example.com',
            'https://example.com',
            'https://example.com/path',
            'https://example.com/path?query=value',
            'http://subdomain.example.com',
        ]
        
        for url in valid_urls:
            result = validator.validate(url, 'field')
            assert result.valid, f"Failed for: {url}"
    
    def test_invalid_url(self):
        validator = validators.url()
        
        invalid_urls = [
            'not-a-url',
            'ftp://example.com',
            'example.com',
            'http://',
        ]
        
        for url in invalid_urls:
            result = validator.validate(url, 'field')
            assert not result.valid, f"Should fail for: {url}"


class TestListValidator:
    def test_valid_list(self):
        validator = validators.list()
        result = validator.validate(['a', 'b', 'c'], 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.list()
        result = validator.validate('not a list', 'field')
        assert not result.valid
        assert 'Expected list' in result.errors[0].message
    
    def test_min_length(self):
        validator = validators.list(min_length=2)
        
        result = validator.validate(['a', 'b'], 'field')
        assert result.valid
        
        result = validator.validate(['a'], 'field')
        assert not result.valid
        assert 'at least 2' in result.errors[0].message
    
    def test_max_length(self):
        validator = validators.list(max_length=3)
        
        result = validator.validate(['a', 'b'], 'field')
        assert result.valid
        
        result = validator.validate(['a', 'b', 'c', 'd'], 'field')
        assert not result.valid
        assert 'at most 3' in result.errors[0].message
    
    def test_item_validator(self):
        validator = validators.list(item_validator=validators.int())
        
        result = validator.validate([1, 2, 3], 'field')
        assert result.valid
        
        result = validator.validate([1, 'two', 3], 'field')
        assert not result.valid
        assert 'field[1]' in result.errors[0].field


class TestDictValidator:
    def test_valid_dict(self):
        validator = validators.dict()
        result = validator.validate({'key': 'value'}, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.dict()
        result = validator.validate('not a dict', 'field')
        assert not result.valid
        assert 'Expected dict' in result.errors[0].message
    
    def test_schema_validation(self):
        validator = validators.dict({
            'name': validators.string(),
            'age': validators.int(min_value=0)
        })
        
        result = validator.validate({'name': 'John', 'age': 30}, 'field')
        assert result.valid
        
        result = validator.validate({'name': 'John', 'age': -1}, 'field')
        assert not result.valid
        assert 'field.age' in result.errors[0].field
    
    def test_missing_required_field(self):
        validator = validators.dict({
            'name': validators.string(required=True)
        })
        
        result = validator.validate({}, 'field')
        assert not result.valid
        assert 'required' in result.errors[0].message.lower()


class TestSchema:
    def test_simple_schema(self):
        schema = Schema({
            'username': validators.string(min_length=3),
            'email': validators.email(),
            'age': validators.int(min_value=0, max_value=120)
        })
        
        data = {
            'username': 'john_doe',
            'email': 'john@example.com',
            'age': 30
        }
        
        result = schema.validate(data)
        assert result.valid
    
    def test_schema_with_errors(self):
        schema = Schema({
            'username': validators.string(min_length=3),
            'email': validators.email(),
            'age': validators.int(min_value=0)
        })
        
        data = {
            'username': 'ab',
            'email': 'invalid-email',
            'age': -1
        }
        
        result = schema.validate(data)
        assert not result.valid
        assert len(result.errors) == 3
    
    def test_nested_schema(self):
        schema = Schema({
            'user': validators.dict({
                'name': validators.string(),
                'email': validators.email()
            }),
            'settings': validators.dict({
                'theme': validators.string(choices=['light', 'dark']),
                'notifications': validators.bool()
            })
        })
        
        data = {
            'user': {
                'name': 'John',
                'email': 'john@example.com'
            },
            'settings': {
                'theme': 'dark',
                'notifications': True
            }
        }
        
        result = schema.validate(data)
        assert result.valid
    
    def test_list_of_objects(self):
        schema = Schema({
            'users': validators.list(
                validators.dict({
                    'name': validators.string(),
                    'age': validators.int(min_value=0)
                })
            )
        })
        
        data = {
            'users': [
                {'name': 'John', 'age': 30},
                {'name': 'Jane', 'age': 25}
            ]
        }
        
        result = schema.validate(data)
        assert result.valid
    
    def test_validate_or_raise(self):
        schema = Schema({
            'email': validators.email()
        })
        
        with pytest.raises(ValidationError):
            schema.validate_or_raise({'email': 'invalid'})


class TestCustomValidators:
    def test_custom_validator(self):
        def is_even(value):
            if value % 2 != 0:
                return "Value must be even"
            return None
        
        validator = validators.int().custom(is_even)
        
        result = validator.validate(4, 'field')
        assert result.valid
        
        result = validator.validate(3, 'field')
        assert not result.valid
        assert 'must be even' in result.errors[0].message
    
    def test_multiple_custom_validators(self):
        def is_positive(value):
            return "Must be positive" if value <= 0 else None
        
        def is_even(value):
            return "Must be even" if value % 2 != 0 else None
        
        validator = validators.int().custom(is_positive).custom(is_even)
        
        result = validator.validate(4, 'field')
        assert result.valid
        
        result = validator.validate(-2, 'field')
        assert not result.valid
        assert any('positive' in e.message for e in result.errors)


class TestRealWorldScenarios:
    def test_user_registration(self):
        schema = Schema({
            'username': validators.string(min_length=3, max_length=20),
            'email': validators.email(),
            'password': validators.string(min_length=8),
            'age': validators.int(min_value=13, required=False),
            'terms_accepted': validators.bool()
        })
        
        valid_data = {
            'username': 'john_doe',
            'email': 'john@example.com',
            'password': 'secure_password_123',
            'age': 25,
            'terms_accepted': True
        }
        
        result = schema.validate(valid_data)
        assert result.valid
        
        invalid_data = {
            'username': 'ab',
            'email': 'invalid',
            'password': 'short',
            'terms_accepted': False
        }
        
        result = schema.validate(invalid_data)
        assert not result.valid
        assert len(result.errors) >= 3
    
    def test_api_request(self):
        schema = Schema({
            'method': validators.string(choices=['GET', 'POST', 'PUT', 'DELETE']),
            'url': validators.url(),
            'headers': validators.dict(required=False),
            'body': validators.dict(required=False),
            'timeout': validators.float(min_value=0, required=False)
        })
        
        data = {
            'method': 'POST',
            'url': 'https://api.example.com/users',
            'headers': {'Content-Type': 'application/json'},
            'body': {'name': 'John'},
            'timeout': 30.0
        }
        
        result = schema.validate(data)
        assert result.valid
    
    def test_configuration(self):
        schema = Schema({
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
        
        data = {
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
        
        result = schema.validate(data)
        assert result.valid


if __name__ == '__main__':
    pytest.main([__file__, '-v'])

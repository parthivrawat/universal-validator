"""
Tests for Universal Data Validator Library
"""

import json
from pathlib import Path

import pytest

from universal_validator import (
    BoolValidator,
    Code,
    DateTimeValidator,
    DateValidator,
    DictValidator,
    EmailValidator,
    EnumValidator,
    FloatValidator,
    IntValidator,
    ListValidator,
    NotEmptyValidator,
    NumericValidator,
    OneOfValidator,
    Schema,
    StringValidator,
    UrlValidator,
    UuidValidator,
    ValidationErrors,
    Validator,
    validators,
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
        assert result.errors[0].code == Code.TYPE
    
    def test_min_length(self):
        validator = validators.string(min_length=5)
        
        result = validator.validate('hello', 'field')
        assert result.valid
        
        result = validator.validate('hi', 'field')
        assert not result.valid
        assert 'at least 5' in result.errors[0].message
        assert result.errors[0].code == Code.MIN_LENGTH
    
    def test_max_length(self):
        validator = validators.string(max_length=5)
        
        result = validator.validate('hello', 'field')
        assert result.valid
        
        result = validator.validate('hello world', 'field')
        assert not result.valid
        assert 'at most 5' in result.errors[0].message
        assert result.errors[0].code == Code.MAX_LENGTH
    
    def test_pattern(self):
        validator = validators.string(pattern=r'^\d{3}-\d{4}$')
        
        result = validator.validate('123-4567', 'field')
        assert result.valid
        
        result = validator.validate('invalid', 'field')
        assert not result.valid
        assert 'does not match pattern' in result.errors[0].message
        assert result.errors[0].code == Code.PATTERN
    
    def test_choices(self):
        validator = validators.string(choices=['red', 'green', 'blue'])
        
        result = validator.validate('red', 'field')
        assert result.valid
        
        result = validator.validate('yellow', 'field')
        assert not result.valid
        assert 'must be one of' in result.errors[0].message
        assert result.errors[0].code == Code.CHOICES
    
    def test_choices_message_preserves_order(self):
        validator = validators.string(choices=['one', 'two', 'three'])
        result = validator.validate('four', 'field')
        assert not result.valid
        assert result.errors[0].message == "Value must be one of one, two, three, got 'four'"
    
    def test_required(self):
        validator = validators.string(required=True)
        result = validator.validate(None, 'field')
        assert not result.valid
        assert 'required' in result.errors[0].message.lower()
        assert result.errors[0].code == Code.REQUIRED
    
    def test_optional(self):
        validator = validators.string(required=False)
        result = validator.validate(None, 'field')
        assert result.valid
    
    def test_nullable(self):
        validator = validators.string(nullable=True)
        result = validator.validate(None, 'field')
        assert result.valid

    def test_pattern_uses_fullmatch(self):
        validator = validators.string(pattern=r'\d+')

        result = validator.validate('12', 'field')
        assert result.valid

        result = validator.validate('12a', 'field')
        assert not result.valid
        assert 'does not match pattern' in result.errors[0].message
        assert result.errors[0].code == Code.PATTERN


class TestIntValidator:
    def test_valid_int(self):
        validator = validators.int()
        result = validator.validate(42, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.int()
        result = validator.validate('not an int', 'field')
        assert not result.valid
        assert 'Expected integer' in result.errors[0].message
        assert result.errors[0].code == Code.TYPE
    
    def test_bool_rejected(self):
        validator = validators.int()
        result = validator.validate(True, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.TYPE
    
    def test_min_value(self):
        validator = validators.int(min_value=0)
        
        result = validator.validate(5, 'field')
        assert result.valid
        
        result = validator.validate(-1, 'field')
        assert not result.valid
        assert 'at least 0' in result.errors[0].message
        assert result.errors[0].code == Code.MIN_VALUE
    
    def test_max_value(self):
        validator = validators.int(max_value=100)
        
        result = validator.validate(50, 'field')
        assert result.valid
        
        result = validator.validate(101, 'field')
        assert not result.valid
        assert 'at most 100' in result.errors[0].message
        assert result.errors[0].code == Code.MAX_VALUE


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
        assert result.errors[0].code == Code.TYPE
    
    def test_min_value(self):
        validator = validators.float(min_value=0.0)
        
        result = validator.validate(5.5, 'field')
        assert result.valid
        
        result = validator.validate(-0.1, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.MIN_VALUE
    
    def test_max_value(self):
        validator = validators.float(max_value=100.0)
        
        result = validator.validate(50.5, 'field')
        assert result.valid
        
        result = validator.validate(100.1, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.MAX_VALUE


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
        assert 'Expected boolean' in result.errors[0].message
        assert result.errors[0].code == Code.TYPE


class TestEmailValidator:
    def test_length_errors_preserved_with_friendly_message(self):
        validator = EmailValidator(min_length=10, max_length=30)

        # Too short and invalid format
        result = validator.validate('bad', 'field')
        assert not result.valid
        assert any('at least 10' in e.message for e in result.errors)
        assert any(Code.MIN_LENGTH == e.code and 'at least 10' in e.message for e in result.errors)
        assert any('Invalid email' in e.message for e in result.errors)
        assert any(Code.EMAIL == e.code for e in result.errors)

        # Too long and invalid format
        result = validator.validate('x' * 35, 'field')
        assert not result.valid
        assert any('at most 30' in e.message for e in result.errors)
        assert any(Code.MAX_LENGTH == e.code and 'at most 30' in e.message for e in result.errors)
        assert any('Invalid email' in e.message for e in result.errors)
        assert any(Code.EMAIL == e.code for e in result.errors)

        # Length OK but invalid format
        result = validator.validate('not-an-email', 'field')
        assert not result.valid
        assert all('at least' not in e.message and 'at most' not in e.message for e in result.errors)
        assert any('Invalid email' in e.message for e in result.errors)
        assert all(e.code in (Code.EMAIL,) for e in result.errors)

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
            assert result.errors[0].code == Code.EMAIL


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
            assert 'Invalid URL' in result.errors[0].message
            assert result.errors[0].code == Code.URL


class TestListValidator:
    def test_valid_list(self):
        validator = validators.list()
        result = validator.validate(['a', 'b', 'c'], 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.list()
        result = validator.validate('not a list', 'field')
        assert not result.valid
        assert 'Expected array' in result.errors[0].message
        assert result.errors[0].code == Code.TYPE
    
    def test_min_length(self):
        validator = validators.list(min_length=2)
        
        result = validator.validate(['a', 'b'], 'field')
        assert result.valid
        
        result = validator.validate(['a'], 'field')
        assert not result.valid
        assert 'at least 2' in result.errors[0].message
        assert result.errors[0].code == Code.MIN_LENGTH
    
    def test_max_length(self):
        validator = validators.list(max_length=3)
        
        result = validator.validate(['a', 'b'], 'field')
        assert result.valid
        
        result = validator.validate(['a', 'b', 'c', 'd'], 'field')
        assert not result.valid
        assert 'at most 3' in result.errors[0].message
        assert result.errors[0].code == Code.MAX_LENGTH
    
    def test_item_validator(self):
        validator = validators.list(item_validator=validators.int())
        
        result = validator.validate([1, 2, 3], 'field')
        assert result.valid
        
        result = validator.validate([1, 'two', 3], 'field')
        assert not result.valid
        assert 'field[1]' in result.errors[0].field
        assert result.errors[0].code == Code.TYPE


class TestDictValidator:
    def test_valid_dict(self):
        validator = validators.dict()
        result = validator.validate({'key': 'value'}, 'field')
        assert result.valid
    
    def test_invalid_type(self):
        validator = validators.dict()
        result = validator.validate('not a dict', 'field')
        assert not result.valid
        assert 'Expected object' in result.errors[0].message
        assert result.errors[0].code == Code.TYPE
    
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
        assert result.errors[0].code == Code.MIN_VALUE
    
    def test_missing_required_field(self):
        validator = validators.dict({
            'name': validators.string(required=True)
        })
        
        result = validator.validate({}, 'field')
        assert not result.valid
        assert 'required' in result.errors[0].message.lower()
        assert result.errors[0].code == Code.REQUIRED

    def test_present_null_required(self):
        """A present key with None must go through validate and report required."""
        validator = validators.dict({
            'name': validators.string(required=True)
        })

        result = validator.validate({'name': None}, 'field')
        assert not result.valid
        assert result.errors[0].field == 'field.name'
        assert 'required' in result.errors[0].message.lower()
        assert result.errors[0].code == Code.REQUIRED

    def test_strict_unknown_keys(self):
        validator = validators.dict({
            'name': validators.string()
        }, strict=True)

        result = validator.validate({'name': 'John', 'extra': 1}, 'field')
        assert not result.valid
        assert any(e.field == 'field.extra' and e.message == 'Unknown field' for e in result.errors)
        assert all(e.code == Code.UNKNOWN_FIELD for e in result.errors if 'Unknown field' in e.message)

    def test_strict_off_by_default(self):
        validator = validators.dict({
            'name': validators.string()
        })

        result = validator.validate({'name': 'John', 'extra': 1}, 'field')
        assert result.valid


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

        with pytest.raises(ValidationErrors):
            schema.validate_or_raise({'email': 'invalid'})

    def test_strict_unknown_keys(self):
        schema = Schema({'name': validators.string()}, strict=True)

        result = schema.validate({'name': 'John', 'extra': 1})
        assert not result.valid
        assert any(e.field == 'extra' and e.message == 'Unknown field' for e in result.errors)
        assert all(e.code == Code.UNKNOWN_FIELD for e in result.errors if e.message == 'Unknown field')

    def test_strict_off_by_default(self):
        schema = Schema({'name': validators.string()})

        result = schema.validate({'name': 'John', 'extra': 1})
        assert result.valid

    def test_strict_nested_dict_path(self):
        schema = Schema({
            'a': validators.dict({'b': validators.string()}, strict=True)
        })

        result = schema.validate({'a': {'b': 'ok', 'c': 1}})
        assert not result.valid
        assert any(e.field == 'a.c' and e.message == 'Unknown field' for e in result.errors)
        assert all(e.code == Code.UNKNOWN_FIELD for e in result.errors if e.message == 'Unknown field')


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
        assert result.errors[0].code == Code.CUSTOM
    
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
        assert all(e.code == Code.CUSTOM for e in result.errors)


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


class TestFailFast:
    def test_schema_default_collects_all_errors(self):
        schema = Schema({
            'username': validators.string(min_length=3),
            'email': validators.email(),
            'age': validators.int(min_value=0)
        })

        result = schema.validate({
            'username': 'ab',
            'email': 'invalid-email',
            'age': -1
        })
        assert not result.valid
        assert len(result.errors) == 3

    def test_schema_fail_fast_returns_one_error(self):
        schema = Schema({
            'username': validators.string(min_length=3),
            'email': validators.email(),
            'age': validators.int(min_value=0)
        }, fail_fast=True)

        result = schema.validate({
            'username': 'ab',
            'email': 'invalid-email',
            'age': -1
        })
        assert not result.valid
        assert len(result.errors) == 1

    def test_dict_fail_fast_returns_one_error(self):
        validator = validators.dict({
            'a': validators.int(min_value=0),
            'b': validators.int(min_value=0)
        }, fail_fast=True)

        result = validator.validate({'a': -1, 'b': -2}, 'root')
        assert not result.valid
        assert len(result.errors) == 1
        assert result.errors[0].field == 'root.a'

    def test_list_fail_fast_returns_one_error(self):
        validator = validators.list(validators.int(min_value=0), fail_fast=True)

        result = validator.validate([-1, -2], 'items')
        assert not result.valid
        assert len(result.errors) == 1
        assert result.errors[0].field == 'items[0]'

    def test_nested_dict_fail_fast(self):
        schema = Schema({
            'user': validators.dict({
                'name': validators.string(min_length=3),
                'age': validators.int(min_value=0)
            })
        }, fail_fast=True)

        result = schema.validate({'user': {'name': 'ab', 'age': -1}})
        assert not result.valid
        assert len(result.errors) == 1

    def test_nested_list_fail_fast(self):
        item = validators.dict({
            'a': validators.int(min_value=0),
            'b': validators.int(min_value=0)
        })
        validator = validators.list(item, fail_fast=True)

        result = validator.validate([{'a': -1, 'b': -2}], 'items')
        assert not result.valid
        assert len(result.errors) == 1
        assert result.errors[0].field == 'items[0].a'


class TestSensitive:
    def test_email_redacts_message_and_value(self):
        validator = validators.email(sensitive=True)
        result = validator.validate('p@ssw0rd', 'secret')
        assert not result.valid
        assert result.errors[0].message == "Invalid email address: ***"
        assert result.errors[0].value is None
        assert result.errors[0].code == Code.EMAIL

    def test_choices_redacts_got_interpolation(self):
        validator = validators.string(choices=['a', 'b'], sensitive=True)
        result = validator.validate('hunter2', 'token')
        assert not result.valid
        assert result.errors[0].message == "Value must be one of a, b, got '***'"
        assert result.errors[0].value is None

    def test_non_value_messages_unchanged(self):
        validator = validators.string(min_length=5, sensitive=True)
        result = validator.validate('abc', 'secret')
        assert not result.valid
        assert result.errors[0].message == "String length must be at least 5, got 3"
        assert result.errors[0].value is None

    def test_int_bounds_redacted(self):
        validator = validators.int(min_value=0, sensitive=True)
        result = validator.validate(-1, 'pin')
        assert not result.valid
        assert result.errors[0].message == "Value must be at least 0, got ***"
        assert result.errors[0].value is None

    def test_type_error_value_redacted(self):
        validator = validators.int(sensitive=True)
        result = validator.validate('hunter2', 'pin')
        assert not result.valid
        assert 'got string' in result.errors[0].message
        assert result.errors[0].value is None

    def test_not_inherited_by_nested_validators(self):
        validator = validators.dict(
            {'inner': validators.string()},
            sensitive=True,
        )
        result = validator.validate({'inner': 123}, 'field')
        assert not result.valid
        # Nested validator is not sensitive: raw value still reported.
        assert result.errors[0].value == 123

    def test_default_not_sensitive(self):
        validator = validators.email()
        result = validator.validate('p@ssw0rd', 'field')
        assert not result.valid
        assert result.errors[0].message == "Invalid email address: p@ssw0rd"
        assert result.errors[0].value == 'p@ssw0rd'


class TestPatternInputLengthCap:
    def test_over_length_input_reports_pattern_error(self):
        validator = validators.string(pattern=r'^[a-z]+$')
        result = validator.validate('a' * 10_001, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.PATTERN
        assert 'does not match pattern' in result.errors[0].message

    def test_cap_applies_to_fixed_patterns(self):
        result = validators.email().validate('a' * 10_001, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.EMAIL
        assert result.errors[0].message == f"Invalid email address: {'a' * 10_001}"

    def test_at_cap_still_matches(self):
        validator = validators.string(pattern=r'^[a-z]+$')
        result = validator.validate('a' * 10_000, 'field')
        assert result.valid

    def test_cap_with_sensitive(self):
        validator = validators.string(pattern=r'^[a-z]+$', sensitive=True)
        result = validator.validate('a' * 10_001, 'field')
        assert not result.valid
        assert result.errors[0].code == Code.PATTERN
        assert result.errors[0].value is None


CASES_PATH = Path(__file__).parent.parent / "testdata" / "vectors" / "cases.json"
with open(CASES_PATH, encoding="utf-8") as _f:
    CASES = json.load(_f)


def build_validator(spec: dict) -> Validator:
    """Build a Validator from a SPEC.md validator-spec dict."""
    vtype = spec["type"]
    common = {
        "required": spec.get("required", True),
        "nullable": spec.get("nullable", False),
        "sensitive": spec.get("sensitive", False),
    }
    if vtype == "string":
        return StringValidator(
            min_length=spec.get("minLength"),
            max_length=spec.get("maxLength"),
            pattern=spec.get("pattern"),
            choices=spec.get("choices"),
            **common,
        )
    if vtype == "integer":
        return IntValidator(
            min_value=spec.get("minValue"),
            max_value=spec.get("maxValue"),
            **common,
        )
    if vtype == "number":
        return FloatValidator(
            min_value=spec.get("minValue"),
            max_value=spec.get("maxValue"),
            **common,
        )
    if vtype == "boolean":
        return BoolValidator(**common)
    if vtype == "email":
        return EmailValidator(**common)
    if vtype == "url":
        return UrlValidator(**common)
    if vtype == "array":
        return ListValidator(
            item_validator=build_validator(spec["items"]) if "items" in spec else None,
            min_length=spec.get("minLength"),
            max_length=spec.get("maxLength"),
            **common,
        )
    if vtype == "object":
        fields = spec.get("fields", {})
        return DictValidator(
            schema={name: build_validator(sub) for name, sub in fields.items()},
            strict=spec.get("strict", False),
            **common,
        )
    if vtype == "uuid":
        return UuidValidator(**common)
    if vtype == "date":
        return DateValidator(**common)
    if vtype == "datetime":
        return DateTimeValidator(**common)
    if vtype == "numeric":
        return NumericValidator(**common)
    if vtype == "not_empty":
        return NotEmptyValidator(**common)
    if vtype == "enum":
        return EnumValidator(values=spec["values"], **common)
    if vtype == "one_of":
        return OneOfValidator(
            validators=[build_validator(s) for s in spec["of"]],
            **common,
        )
    raise ValueError(f"Unknown validator type: {vtype}")


def build_schema(spec: dict) -> Schema:
    """Build a Schema from a SPEC.md schema-spec dict."""
    fields = spec.get("fields", {})
    return Schema(
        {name: build_validator(sub) for name, sub in fields.items()},
        strict=spec.get("strict", False),
    )


class TestConformance:
    """Run the shared conformance vectors from testdata/vectors/cases.json."""

    def test_cases(self):
        for case in CASES:
            schema = build_schema(case["schema"])
            result = schema.validate(case["input"])
            actual = sorted((e.field, e.code, e.message) for e in result.errors)
            expected = sorted((e["field"], e["code"], e["message"]) for e in case["expected"])
            assert actual == expected, (
                f"Case '{case['name']}' failed:\n"
                f"  expected: {expected}\n"
                f"  actual:   {actual}"
            )


if __name__ == '__main__':
    pytest.main([__file__, '-v'])

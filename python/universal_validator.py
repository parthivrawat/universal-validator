"""
Universal Data Validator Library

A comprehensive data validation library that works across API, database, and form contexts.

Author: Parthiv Rawat
License: MIT
"""

from typing import Any, Callable, Dict, List, Optional, Union, Type, TypeVar, Generic
from dataclasses import dataclass
from enum import Enum
import re
from datetime import datetime


T = TypeVar('T')


class ValidationError(Exception):
    """Base exception for validation errors."""
    
    def __init__(self, field: str, message: str, value: Any = None):
        self.field = field
        self.message = message
        self.value = value
        super().__init__(f"{field}: {message}")


class ValidationResult:
    """Result of a validation operation."""
    
    def __init__(self, valid: bool = True, errors: Optional[List[ValidationError]] = None):
        self.valid = valid
        self.errors = errors or []
    
    def add_error(self, error: ValidationError):
        """Add a validation error."""
        self.valid = False
        self.errors.append(error)
    
    def __bool__(self):
        return self.valid
    
    def __repr__(self):
        if self.valid:
            return "ValidationResult(valid=True)"
        return f"ValidationResult(valid=False, errors={len(self.errors)})"


class Validator(Generic[T]):
    """Base validator class."""
    
    def __init__(self, required: bool = True, nullable: bool = False):
        self.required = required
        self.nullable = nullable
        self.custom_validators: List[Callable[[Any], Optional[str]]] = []
    
    def validate(self, value: Any, field: str = "field") -> ValidationResult:
        """Validate a value."""
        result = ValidationResult()
        
        if value is None:
            if self.required and not self.nullable:
                result.add_error(ValidationError(field, "Field is required"))
            return result
        
        type_error = self._validate_type(value, field)
        if type_error:
            result.add_error(type_error)
            return result
        
        constraint_errors = self._validate_constraints(value, field)
        for error in constraint_errors:
            result.add_error(error)
        
        for custom_validator in self.custom_validators:
            error_msg = custom_validator(value)
            if error_msg:
                result.add_error(ValidationError(field, error_msg, value))
        
        return result
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        """Validate the type of the value."""
        return None
    
    def _validate_constraints(self, value: Any, field: str) -> List[ValidationError]:
        """Validate constraints on the value."""
        return []
    
    def custom(self, validator: Callable[[Any], Optional[str]]) -> 'Validator[T]':
        """Add a custom validator function."""
        self.custom_validators.append(validator)
        return self


class StringValidator(Validator[str]):
    """Validator for string values."""
    
    def __init__(self, 
                 min_length: Optional[int] = None,
                 max_length: Optional[int] = None,
                 pattern: Optional[str] = None,
                 choices: Optional[List[str]] = None,
                 **kwargs):
        super().__init__(**kwargs)
        self.min_length = min_length
        self.max_length = max_length
        self.pattern = pattern
        self.choices = choices
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, str):
            return ValidationError(field, f"Expected string, got {type(value).__name__}", value)
        return None
    
    def _validate_constraints(self, value: str, field: str) -> List[ValidationError]:
        errors = []
        
        if self.min_length is not None and len(value) < self.min_length:
            errors.append(ValidationError(
                field, 
                f"String length must be at least {self.min_length}, got {len(value)}",
                value
            ))
        
        if self.max_length is not None and len(value) > self.max_length:
            errors.append(ValidationError(
                field,
                f"String length must be at most {self.max_length}, got {len(value)}",
                value
            ))
        
        if self.pattern is not None and not re.match(self.pattern, value):
            errors.append(ValidationError(
                field,
                f"String does not match pattern {self.pattern}",
                value
            ))
        
        if self.choices is not None and value not in self.choices:
            errors.append(ValidationError(
                field,
                f"Value must be one of {self.choices}, got '{value}'",
                value
            ))
        
        return errors


class IntValidator(Validator[int]):
    """Validator for integer values."""
    
    def __init__(self,
                 min_value: Optional[int] = None,
                 max_value: Optional[int] = None,
                 **kwargs):
        super().__init__(**kwargs)
        self.min_value = min_value
        self.max_value = max_value
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, int) or isinstance(value, bool):
            return ValidationError(field, f"Expected int, got {type(value).__name__}", value)
        return None
    
    def _validate_constraints(self, value: int, field: str) -> List[ValidationError]:
        errors = []
        
        if self.min_value is not None and value < self.min_value:
            errors.append(ValidationError(
                field,
                f"Value must be at least {self.min_value}, got {value}",
                value
            ))
        
        if self.max_value is not None and value > self.max_value:
            errors.append(ValidationError(
                field,
                f"Value must be at most {self.max_value}, got {value}",
                value
            ))
        
        return errors


class FloatValidator(Validator[float]):
    """Validator for float values."""
    
    def __init__(self,
                 min_value: Optional[float] = None,
                 max_value: Optional[float] = None,
                 **kwargs):
        super().__init__(**kwargs)
        self.min_value = min_value
        self.max_value = max_value
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, (int, float)) or isinstance(value, bool):
            return ValidationError(field, f"Expected float, got {type(value).__name__}", value)
        return None
    
    def _validate_constraints(self, value: float, field: str) -> List[ValidationError]:
        errors = []
        
        if self.min_value is not None and value < self.min_value:
            errors.append(ValidationError(
                field,
                f"Value must be at least {self.min_value}, got {value}",
                value
            ))
        
        if self.max_value is not None and value > self.max_value:
            errors.append(ValidationError(
                field,
                f"Value must be at most {self.max_value}, got {value}",
                value
            ))
        
        return errors


class BoolValidator(Validator[bool]):
    """Validator for boolean values."""
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, bool):
            return ValidationError(field, f"Expected bool, got {type(value).__name__}", value)
        return None


class EmailValidator(StringValidator):
    """Validator for email addresses."""
    
    def __init__(self, **kwargs):
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        super().__init__(pattern=pattern, **kwargs)
    
    def _validate_constraints(self, value: str, field: str) -> List[ValidationError]:
        errors = super()._validate_constraints(value, field)
        
        if not re.match(self.pattern, value):
            errors = [ValidationError(field, f"Invalid email address: {value}", value)]
        
        return errors


class UrlValidator(StringValidator):
    """Validator for URLs."""
    
    def __init__(self, **kwargs):
        pattern = r'^https?://[^\s/$.?#].[^\s]*$'
        super().__init__(pattern=pattern, **kwargs)
    
    def _validate_constraints(self, value: str, field: str) -> List[ValidationError]:
        errors = super()._validate_constraints(value, field)
        
        if not re.match(self.pattern, value):
            errors = [ValidationError(field, f"Invalid URL: {value}", value)]
        
        return errors


class ListValidator(Validator[List]):
    """Validator for list values."""
    
    def __init__(self,
                 item_validator: Optional[Validator] = None,
                 min_length: Optional[int] = None,
                 max_length: Optional[int] = None,
                 **kwargs):
        super().__init__(**kwargs)
        self.item_validator = item_validator
        self.min_length = min_length
        self.max_length = max_length
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, list):
            return ValidationError(field, f"Expected list, got {type(value).__name__}", value)
        return None
    
    def _validate_constraints(self, value: List, field: str) -> List[ValidationError]:
        errors = []
        
        if self.min_length is not None and len(value) < self.min_length:
            errors.append(ValidationError(
                field,
                f"List length must be at least {self.min_length}, got {len(value)}",
                value
            ))
        
        if self.max_length is not None and len(value) > self.max_length:
            errors.append(ValidationError(
                field,
                f"List length must be at most {self.max_length}, got {len(value)}",
                value
            ))
        
        if self.item_validator:
            for i, item in enumerate(value):
                item_result = self.item_validator.validate(item, f"{field}[{i}]")
                errors.extend(item_result.errors)
        
        return errors


class DictValidator(Validator[Dict]):
    """Validator for dictionary values."""
    
    def __init__(self,
                 schema: Optional[Dict[str, Validator]] = None,
                 **kwargs):
        super().__init__(**kwargs)
        self.schema = schema or {}
    
    def _validate_type(self, value: Any, field: str) -> Optional[ValidationError]:
        if not isinstance(value, dict):
            return ValidationError(field, f"Expected dict, got {type(value).__name__}", value)
        return None
    
    def _validate_constraints(self, value: Dict, field: str) -> List[ValidationError]:
        errors = []
        
        for key, validator in self.schema.items():
            if key in value:
                result = validator.validate(value[key], f"{field}.{key}")
                errors.extend(result.errors)
            elif validator.required:
                errors.append(ValidationError(f"{field}.{key}", "Field is required"))
        
        return errors


class Schema:
    """Schema for validating complex data structures."""
    
    def __init__(self, schema: Dict[str, Validator]):
        self.schema = schema
    
    def validate(self, data: Dict[str, Any]) -> ValidationResult:
        """Validate data against the schema."""
        result = ValidationResult()
        
        if not isinstance(data, dict):
            result.add_error(ValidationError("root", f"Expected dict, got {type(data).__name__}"))
            return result
        
        for field, validator in self.schema.items():
            if field in data:
                field_result = validator.validate(data[field], field)
                for error in field_result.errors:
                    result.add_error(error)
            elif validator.required:
                result.add_error(ValidationError(field, "Field is required"))
        
        return result
    
    def validate_or_raise(self, data: Dict[str, Any]) -> None:
        """Validate data and raise exception if invalid."""
        result = self.validate(data)
        if not result.valid:
            error_messages = [f"{e.field}: {e.message}" for e in result.errors]
            raise ValidationError("validation", "\n".join(error_messages))


class validators:
    """Namespace for validator factory functions."""
    
    @staticmethod
    def string(min_length: Optional[int] = None,
               max_length: Optional[int] = None,
               pattern: Optional[str] = None,
               choices: Optional[List[str]] = None,
               required: bool = True,
               nullable: bool = False) -> StringValidator:
        """Create a string validator."""
        return StringValidator(
            min_length=min_length,
            max_length=max_length,
            pattern=pattern,
            choices=choices,
            required=required,
            nullable=nullable
        )
    
    @staticmethod
    def int(min_value: Optional[int] = None,
            max_value: Optional[int] = None,
            required: bool = True,
            nullable: bool = False) -> IntValidator:
        """Create an integer validator."""
        return IntValidator(
            min_value=min_value,
            max_value=max_value,
            required=required,
            nullable=nullable
        )
    
    @staticmethod
    def float(min_value: Optional[float] = None,
              max_value: Optional[float] = None,
              required: bool = True,
              nullable: bool = False) -> FloatValidator:
        """Create a float validator."""
        return FloatValidator(
            min_value=min_value,
            max_value=max_value,
            required=required,
            nullable=nullable
        )
    
    @staticmethod
    def bool(required: bool = True, nullable: bool = False) -> BoolValidator:
        """Create a boolean validator."""
        return BoolValidator(required=required, nullable=nullable)
    
    @staticmethod
    def email(required: bool = True, nullable: bool = False) -> EmailValidator:
        """Create an email validator."""
        return EmailValidator(required=required, nullable=nullable)
    
    @staticmethod
    def url(required: bool = True, nullable: bool = False) -> UrlValidator:
        """Create a URL validator."""
        return UrlValidator(required=required, nullable=nullable)
    
    @staticmethod
    def list(item_validator: Optional[Validator] = None,
             min_length: Optional[int] = None,
             max_length: Optional[int] = None,
             required: bool = True,
             nullable: bool = False) -> ListValidator:
        """Create a list validator."""
        return ListValidator(
            item_validator=item_validator,
            min_length=min_length,
            max_length=max_length,
            required=required,
            nullable=nullable
        )
    
    @staticmethod
    def dict(schema: Optional[Dict[str, Validator]] = None,
             required: bool = True,
             nullable: bool = False) -> DictValidator:
        """Create a dictionary validator."""
        return DictValidator(schema=schema, required=required, nullable=nullable)

"""Core validation classes and validators for universal-validator."""

from __future__ import annotations

import copy
import json
import re
from typing import Any, Callable, Dict, Generic, List, TypeVar

__all__ = [
    "MAX_PATTERN_INPUT_LENGTH",
    "BoolValidator",
    "Code",
    "DateTimeValidator",
    "DateValidator",
    "DictValidator",
    "EmailValidator",
    "EnumValidator",
    "FloatValidator",
    "IntValidator",
    "ListValidator",
    "NotEmptyValidator",
    "NumericValidator",
    "OneOfValidator",
    "Schema",
    "StringValidator",
    "UrlValidator",
    "UuidValidator",
    "ValidationError",
    "ValidationErrors",
    "ValidationResult",
    "Validator",
]

T = TypeVar("T")

# Maximum input length for any pattern check (user-supplied `pattern` and the
# fixed email/url/uuid/date/datetime/numeric patterns). Python's `re` is a
# backtracking engine, so this bounds worst-case match cost; inputs longer
# than this are reported as pattern failures without running the regex.
MAX_PATTERN_INPUT_LENGTH = 10_000


def _kind_name(value: Any) -> str:
    """Map a Python value to its canonical JSON-kind name."""
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "integer"
    if isinstance(value, float):
        return "number"
    if isinstance(value, str):
        return "string"
    if isinstance(value, (list, tuple)):
        return "array"
    if isinstance(value, dict):
        return "object"
    return "unknown"


def _compact_json(value: Any) -> str:
    """Render a single JSON value compactly without extra whitespace."""
    return json.dumps(value, separators=(",", ":"), ensure_ascii=False)


class Code:
    """Canonical machine-readable validation error codes."""

    REQUIRED = "required"
    UNKNOWN_FIELD = "unknown_field"
    TYPE = "type"
    MIN_LENGTH = "min_length"
    MAX_LENGTH = "max_length"
    MIN_VALUE = "min_value"
    MAX_VALUE = "max_value"
    PATTERN = "pattern"
    CHOICES = "choices"
    EMAIL = "email"
    URL = "url"
    CUSTOM = "custom"
    UUID = "uuid"
    DATE = "date"
    DATETIME = "datetime"
    NUMERIC = "numeric"
    NOT_EMPTY = "not_empty"
    ENUM = "enum"
    ONE_OF = "one_of"


class ValidationError(Exception):
    """Single validation error."""

    def __init__(self, field: str, message: str, value: Any = None, code: str = ""):
        self.field = field
        self.message = message
        self.value = value
        self.code = code
        super().__init__(f"{field}: {message}")


class ValidationErrors(Exception):
    """Aggregate exception collecting multiple ValidationError instances."""

    def __init__(self, errors: list[ValidationError]):
        self.errors = list(errors)
        message = "\n".join(f"{e.field}: {e.message}" for e in self.errors)
        super().__init__(message)


class ValidationResult:
    """Result of a validation operation."""

    def __init__(self, valid: bool = True, errors: list[ValidationError] | None = None):
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

    def __init__(
        self,
        required: bool = True,
        nullable: bool = False,
        fail_fast: bool = False,
        sensitive: bool = False,
    ):
        self.required = required
        self.nullable = nullable
        self.fail_fast = fail_fast
        self.sensitive = sensitive
        self.custom_validators: list[Callable[[Any], str | None]] = []

    def _display(self, value: Any) -> Any:
        """Return the representation of a value to embed in an error message."""
        return "***" if self.sensitive else value

    def _err_value(self, value: Any) -> Any:
        """Return the value to store on a ValidationError (None if sensitive)."""
        return None if self.sensitive else value

    def validate(self, value: Any, field: str = "field") -> ValidationResult:
        """Validate a value."""
        result = ValidationResult()

        if value is None:
            if self.required and not self.nullable:
                result.add_error(ValidationError(field, "Field is required", value, Code.REQUIRED))
            return result

        type_error = self._validate_type(value, field)
        if type_error:
            result.add_error(type_error)
            return result

        constraint_errors = self._validate_constraints(value, field)
        for error in constraint_errors:
            result.add_error(error)
            if self.fail_fast:
                return result

        for custom_validator in self.custom_validators:
            error_msg = custom_validator(value)
            if error_msg:
                result.add_error(
                    ValidationError(field, error_msg, self._err_value(value), Code.CUSTOM)
                )
                if self.fail_fast:
                    return result

        return result

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        """Validate the type of the value."""
        return None

    def _validate_constraints(self, value: Any, field: str) -> list[ValidationError]:
        """Validate constraints on the value."""
        return []

    def custom(self, validator: Callable[[Any], str | None]) -> Validator[T]:
        """Return a cloned validator with an additional custom validator.

        The original validator is left unchanged; the returned clone has the
        new custom validator appended to its own custom validator list.
        """
        clone = copy.copy(self)
        clone.custom_validators = list(self.custom_validators)
        clone.custom_validators.append(validator)
        return clone


class StringValidator(Validator[str]):
    """Validator for string values."""

    def __init__(
        self,
        min_length: int | None = None,
        max_length: int | None = None,
        pattern: str | None = None,
        choices: list[str] | None = None,
        **kwargs,
    ):
        super().__init__(**kwargs)
        self.min_length = min_length
        self.max_length = max_length
        self.pattern = pattern
        self._compiled_pattern = re.compile(pattern) if pattern is not None else None
        self.choices = choices
        self._choices_set = set(choices) if choices is not None else None

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, str):
            return ValidationError(
                field,
                f"Expected string, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = []

        if self.min_length is not None and len(value) < self.min_length:
            errors.append(
                ValidationError(
                    field,
                    f"String length must be at least {self.min_length}, got {len(value)}",
                    self._err_value(value),
                    Code.MIN_LENGTH,
                )
            )
            if self.fail_fast:
                return errors

        if self.max_length is not None and len(value) > self.max_length:
            errors.append(
                ValidationError(
                    field,
                    f"String length must be at most {self.max_length}, got {len(value)}",
                    self._err_value(value),
                    Code.MAX_LENGTH,
                )
            )
            if self.fail_fast:
                return errors

        if self._compiled_pattern is not None:
            # Cap input length before matching: `re` is a backtracking engine,
            # so unbounded inputs are a ReDoS risk. Over-length inputs report
            # the normal pattern error without running the regex.
            matches = (
                len(value) <= MAX_PATTERN_INPUT_LENGTH
                and self._compiled_pattern.fullmatch(value) is not None
            )
            if not matches:
                errors.append(
                    ValidationError(
                        field,
                        f"String does not match pattern {self.pattern}",
                        self._err_value(value),
                        Code.PATTERN,
                    )
                )
                if self.fail_fast:
                    return errors

        if self._choices_set is not None and value not in self._choices_set:
            errors.append(
                ValidationError(
                    field,
                    f"Value must be one of {', '.join(self.choices)}, got '{self._display(value)}'",
                    self._err_value(value),
                    Code.CHOICES,
                )
            )
            if self.fail_fast:
                return errors

        return errors


class IntValidator(Validator[int]):
    """Validator for integer values."""

    def __init__(
        self,
        min_value: int | None = None,
        max_value: int | None = None,
        **kwargs,
    ):
        super().__init__(**kwargs)
        self.min_value = min_value
        self.max_value = max_value

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, int) or isinstance(value, bool):
            return ValidationError(
                field,
                f"Expected integer, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: int, field: str) -> list[ValidationError]:
        errors = []

        if self.min_value is not None and value < self.min_value:
            errors.append(
                ValidationError(
                    field,
                    f"Value must be at least {self.min_value}, got {self._display(value)}",
                    self._err_value(value),
                    Code.MIN_VALUE,
                )
            )
            if self.fail_fast:
                return errors

        if self.max_value is not None and value > self.max_value:
            errors.append(
                ValidationError(
                    field,
                    f"Value must be at most {self.max_value}, got {self._display(value)}",
                    self._err_value(value),
                    Code.MAX_VALUE,
                )
            )
            if self.fail_fast:
                return errors

        return errors


class FloatValidator(Validator[float]):
    """Validator for number (float/integer) values."""

    def __init__(
        self,
        min_value: float | None = None,
        max_value: float | None = None,
        **kwargs,
    ):
        super().__init__(**kwargs)
        self.min_value = min_value
        self.max_value = max_value

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, (int, float)) or isinstance(value, bool):
            return ValidationError(
                field,
                f"Expected number, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: float, field: str) -> list[ValidationError]:
        errors = []

        if self.min_value is not None and value < self.min_value:
            errors.append(
                ValidationError(
                    field,
                    f"Value must be at least {self.min_value}, got {self._display(value)}",
                    self._err_value(value),
                    Code.MIN_VALUE,
                )
            )
            if self.fail_fast:
                return errors

        if self.max_value is not None and value > self.max_value:
            errors.append(
                ValidationError(
                    field,
                    f"Value must be at most {self.max_value}, got {self._display(value)}",
                    self._err_value(value),
                    Code.MAX_VALUE,
                )
            )
            if self.fail_fast:
                return errors

        return errors


class BoolValidator(Validator[bool]):
    """Validator for boolean values."""

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, bool):
            return ValidationError(
                field,
                f"Expected boolean, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None


class EmailValidator(StringValidator):
    """Validator for email addresses."""

    def __init__(self, **kwargs):
        pattern = r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
        super().__init__(pattern=pattern, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Invalid email address: {self._display(value)}",
                    self._err_value(value),
                    Code.EMAIL,
                )
                break

        return errors


class UrlValidator(StringValidator):
    """Validator for URLs."""

    def __init__(self, **kwargs):
        pattern = r"^https?://[^\s/$.?#].[^\s]*$"
        super().__init__(pattern=pattern, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Invalid URL: {self._display(value)}",
                    self._err_value(value),
                    Code.URL,
                )
                break

        return errors


class UuidValidator(StringValidator):
    """Validator for UUID strings."""

    _PATTERN = r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"

    def __init__(self, **kwargs):
        super().__init__(pattern=self._PATTERN, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Invalid UUID: {self._display(value)}",
                    self._err_value(value),
                    Code.UUID,
                )
                break

        return errors


class DateValidator(StringValidator):
    """Validator for ISO date strings (YYYY-MM-DD)."""

    _PATTERN = r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$"

    def __init__(self, **kwargs):
        super().__init__(pattern=self._PATTERN, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Invalid date: {self._display(value)}",
                    self._err_value(value),
                    Code.DATE,
                )
                break

        return errors


class DateTimeValidator(StringValidator):
    """Validator for ISO datetime strings."""

    _PATTERN = r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$"

    def __init__(self, **kwargs):
        super().__init__(pattern=self._PATTERN, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Invalid datetime: {self._display(value)}",
                    self._err_value(value),
                    Code.DATETIME,
                )
                break

        return errors


class NumericValidator(StringValidator):
    """Validator for numeric strings."""

    _PATTERN = r"^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$"

    def __init__(self, **kwargs):
        super().__init__(pattern=self._PATTERN, **kwargs)

    def _validate_constraints(self, value: str, field: str) -> list[ValidationError]:
        errors = super()._validate_constraints(value, field)

        for i, error in enumerate(errors):
            if error.code == Code.PATTERN:
                errors[i] = ValidationError(
                    field,
                    f"Expected numeric string, got {self._display(value)}",
                    self._err_value(value),
                    Code.NUMERIC,
                )
                break

        return errors


class NotEmptyValidator(Validator[Any]):
    """Validator that rejects empty strings, arrays, or objects."""

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, (str, list, tuple, dict)):
            return ValidationError(
                field,
                f"Expected string, array, or object, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: Any, field: str) -> list[ValidationError]:
        if len(value) == 0:
            return [
                ValidationError(
                    field,
                    "Value must not be empty",
                    self._err_value(value),
                    Code.NOT_EMPTY,
                )
            ]
        return []


class EnumValidator(Validator[Any]):
    """Validator that requires a value to be one of an arbitrary list."""

    def __init__(self, values: list[Any], **kwargs):
        super().__init__(**kwargs)
        self.values = values

    def _validate_constraints(self, value: Any, field: str) -> list[ValidationError]:
        for allowed in self.values:
            if value == allowed:
                return []

        choices = ", ".join(_compact_json(v) for v in self.values)
        return [
            ValidationError(
                field,
                f"Value must be one of {choices}, got {self._display(_compact_json(value))}",
                self._err_value(value),
                Code.ENUM,
            )
        ]


class OneOfValidator(Validator[Any]):
    """Validator that accepts a value if any branch validator accepts it."""

    def __init__(self, validators: list[Validator], **kwargs):
        super().__init__(**kwargs)
        self.validators = validators

    def _validate_constraints(self, value: Any, field: str) -> list[ValidationError]:
        for validator in self.validators:
            branch_result = validator.validate(value, field)
            if branch_result.valid:
                return []

        return [
            ValidationError(
                field,
                "Value does not match any allowed schema",
                self._err_value(value),
                Code.ONE_OF,
            )
        ]


class ListValidator(Validator[List]):
    """Validator for list values."""

    def __init__(
        self,
        item_validator: Validator | None = None,
        min_length: int | None = None,
        max_length: int | None = None,
        fail_fast: bool = False,
        **kwargs,
    ):
        super().__init__(**kwargs, fail_fast=fail_fast)
        self.item_validator = item_validator
        self.min_length = min_length
        self.max_length = max_length

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, (list, tuple)):
            return ValidationError(
                field,
                f"Expected array, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: list, field: str) -> list[ValidationError]:
        errors = []

        if self.min_length is not None and len(value) < self.min_length:
            errors.append(
                ValidationError(
                    field,
                    f"Array length must be at least {self.min_length}, got {len(value)}",
                    self._err_value(value),
                    Code.MIN_LENGTH,
                )
            )
            if self.fail_fast:
                return errors

        if self.max_length is not None and len(value) > self.max_length:
            errors.append(
                ValidationError(
                    field,
                    f"Array length must be at most {self.max_length}, got {len(value)}",
                    self._err_value(value),
                    Code.MAX_LENGTH,
                )
            )
            if self.fail_fast:
                return errors

        if self.item_validator:
            for i, item in enumerate(value):
                old_fail_fast = self.item_validator.fail_fast
                self.item_validator.fail_fast = old_fail_fast or self.fail_fast
                try:
                    item_result = self.item_validator.validate(item, f"{field}[{i}]")
                finally:
                    self.item_validator.fail_fast = old_fail_fast
                for error in item_result.errors:
                    errors.append(error)
                    if self.fail_fast:
                        return errors

        return errors


class DictValidator(Validator[Dict]):
    """Validator for dictionary values."""

    def __init__(
        self,
        schema: dict[str, Validator] | None = None,
        strict: bool = False,
        fail_fast: bool = False,
        **kwargs,
    ):
        super().__init__(**kwargs, fail_fast=fail_fast)
        self.schema = schema or {}
        self.strict = strict

    def _validate_type(self, value: Any, field: str) -> ValidationError | None:
        if not isinstance(value, dict):
            return ValidationError(
                field,
                f"Expected object, got {_kind_name(value)}",
                self._err_value(value),
                Code.TYPE,
            )
        return None

    def _validate_constraints(self, value: dict, field: str) -> list[ValidationError]:
        errors = []

        if self.strict:
            for key, item in value.items():
                if key not in self.schema:
                    errors.append(
                        ValidationError(
                            f"{field}.{key}",
                            "Unknown field",
                            self._err_value(item),
                            Code.UNKNOWN_FIELD,
                        )
                    )
                    if self.fail_fast:
                        return errors

        for key, validator in self.schema.items():
            if key in value:
                old_fail_fast = validator.fail_fast
                validator.fail_fast = old_fail_fast or self.fail_fast
                try:
                    result = validator.validate(value[key], f"{field}.{key}")
                finally:
                    validator.fail_fast = old_fail_fast
                for error in result.errors:
                    errors.append(error)
                    if self.fail_fast:
                        return errors
            elif validator.required:
                errors.append(ValidationError(f"{field}.{key}", "Field is required", value=None, code=Code.REQUIRED))
                if self.fail_fast:
                    return errors

        return errors


class Schema:
    """Schema for validating complex data structures."""

    def __init__(self, schema: dict[str, Validator], strict: bool = False, fail_fast: bool = False):
        self.schema = schema
        self.strict = strict
        self.fail_fast = fail_fast

    def validate(self, data: dict[str, Any]) -> ValidationResult:
        """Validate data against the schema."""
        result = ValidationResult()

        if not isinstance(data, dict):
            result.add_error(ValidationError("root", f"Expected object, got {_kind_name(data)}", data, Code.TYPE))
            return result

        if self.strict:
            for field, item in data.items():
                if field not in self.schema:
                    result.add_error(ValidationError(field, "Unknown field", item, Code.UNKNOWN_FIELD))
                    if self.fail_fast:
                        return result

        for field, validator in self.schema.items():
            if field in data:
                old_fail_fast = validator.fail_fast
                validator.fail_fast = old_fail_fast or self.fail_fast
                try:
                    field_result = validator.validate(data[field], field)
                finally:
                    validator.fail_fast = old_fail_fast
                for error in field_result.errors:
                    result.add_error(error)
                    if self.fail_fast:
                        return result
            elif validator.required:
                result.add_error(ValidationError(field, "Field is required", code=Code.REQUIRED))
                if self.fail_fast:
                    return result

        return result

    def validate_or_raise(self, data: dict[str, Any]) -> None:
        """Validate data and raise ValidationErrors if invalid."""
        result = self.validate(data)
        if not result.valid:
            raise ValidationErrors(result.errors)

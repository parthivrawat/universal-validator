"""Factory functions for creating validator instances."""
from __future__ import annotations

from builtins import list as _list
from typing import Any, Dict

from .core import (
    BoolValidator,
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
    Validator,
)

__all__ = [
    "bool",
    "date",
    "datetime",
    "dict",
    "email",
    "enum",
    "float",
    "int",
    "list",
    "not_empty",
    "numeric",
    "one_of",
    "regex",
    "schema",
    "string",
    "url",
    "uuid",
]


def string(
    min_length: int | None = None,
    max_length: int | None = None,
    pattern: str | None = None,
    choices: _list[str] | None = None,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> StringValidator:
    """Create a string validator."""
    return StringValidator(
        min_length=min_length,
        max_length=max_length,
        pattern=pattern,
        choices=choices,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def regex(
    pattern: str,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> StringValidator:
    """Create a string validator anchored to the supplied pattern."""
    return StringValidator(
        pattern=pattern,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def int(
    min_value: int | None = None,
    max_value: int | None = None,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> IntValidator:
    """Create an integer validator."""
    return IntValidator(
        min_value=min_value,
        max_value=max_value,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def float(
    min_value: float | None = None,
    max_value: float | None = None,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> FloatValidator:
    """Create a number validator."""
    return FloatValidator(
        min_value=min_value,
        max_value=max_value,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def bool(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> BoolValidator:
    """Create a boolean validator."""
    return BoolValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def email(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> EmailValidator:
    """Create an email validator."""
    return EmailValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def url(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> UrlValidator:
    """Create a URL validator."""
    return UrlValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def list(
    item_validator: Validator | None = None,
    min_length: int | None = None,
    max_length: int | None = None,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> ListValidator:
    """Create a list (array) validator."""
    return ListValidator(
        item_validator=item_validator,
        min_length=min_length,
        max_length=max_length,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def dict(
    schema: Dict[str, Validator] | None = None,
    required: bool = True,
    nullable: bool = False,
    strict: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> DictValidator:
    """Create a dictionary (object) validator."""
    return DictValidator(
        schema=schema,
        required=required,
        nullable=nullable,
        strict=strict,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def uuid(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> UuidValidator:
    """Create a UUID validator."""
    return UuidValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def date(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> DateValidator:
    """Create a date validator."""
    return DateValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def datetime(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> DateTimeValidator:
    """Create a datetime validator."""
    return DateTimeValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def numeric(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> NumericValidator:
    """Create a numeric-string validator."""
    return NumericValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def not_empty(
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> NotEmptyValidator:
    """Create a not-empty validator."""
    return NotEmptyValidator(required=required, nullable=nullable, fail_fast=fail_fast, sensitive=sensitive)


def enum(
    values: _list[Any],
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> EnumValidator:
    """Create an enum validator."""
    return EnumValidator(
        values=values,
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def one_of(
    *validators: Validator,
    required: bool = True,
    nullable: bool = False,
    fail_fast: bool = False,
    sensitive: bool = False,
) -> OneOfValidator:
    """Create a one-of validator."""
    return OneOfValidator(
        validators=_list(validators),
        required=required,
        nullable=nullable,
        fail_fast=fail_fast,
        sensitive=sensitive,
    )


def schema(
    fields: Dict[str, Validator],
    strict: bool = False,
    fail_fast: bool = False,
) -> Schema:
    """Create a top-level Schema."""
    return Schema(fields, strict=strict, fail_fast=fail_fast)

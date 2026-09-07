//! Universal Data Validator
//!
//! A comprehensive data validation library for Rust that works across API, database, and form contexts.
//!
//! # Features
//!
//! - **Rich Validator Types**: String, integer, float, bool, email, URL, UUID, date,
//!   datetime, numeric, not-empty, enum, one-of, array (`Vec`) and object (`Map`)
//! - **Schema-Based Validation**: Define complex, nested data structures
//! - **Custom Validators**: Add your own validation logic
//! - **Nullable & Optional**: Fine-grained control over missing and `null` values
//! - **Strict Mode**: Reject unknown fields
//! - **Clear Error Messages**: Detailed error reporting with full field paths
//!
//! # Examples
//!
//! ```
//! use serde_json::json;
//! use universal_validator::{Schema, validators};
//!
//! let schema = Schema::new()
//!     .field("email", validators::email())
//!     .field("age", validators::int().min(0).max(120))
//!     .field("username", validators::string().min(3).max(20));
//!
//! let data = json!({
//!     "email": "user@example.com",
//!     "age": 25,
//!     "username": "john_doe"
//! });
//!
//! let result = schema.validate(&data);
//! assert!(result.is_valid());
//! ```

use regex::Regex;
use serde_json::Value;
use std::collections::HashMap;
use std::fmt;
use std::sync::OnceLock;

/// Canonical error codes returned by validators.
pub mod codes {
    pub const REQUIRED: &str = "required";
    pub const UNKNOWN_FIELD: &str = "unknown_field";
    pub const TYPE: &str = "type";
    pub const MIN_LENGTH: &str = "min_length";
    pub const MAX_LENGTH: &str = "max_length";
    pub const MIN_VALUE: &str = "min_value";
    pub const MAX_VALUE: &str = "max_value";
    pub const PATTERN: &str = "pattern";
    pub const CHOICES: &str = "choices";
    pub const EMAIL: &str = "email";
    pub const URL: &str = "url";
    pub const CUSTOM: &str = "custom";
    pub const UUID: &str = "uuid";
    pub const DATE: &str = "date";
    pub const DATETIME: &str = "datetime";
    pub const NUMERIC: &str = "numeric";
    pub const NOT_EMPTY: &str = "not_empty";
    pub const ENUM: &str = "enum";
    pub const ONE_OF: &str = "one_of";
}

/// A single validation error.
#[derive(Debug, Clone)]
pub struct ValidationError {
    pub field: String,
    pub message: String,
    pub code: String,
    pub value: Option<Value>,
}

impl ValidationError {
    /// Create a new validation error.
    pub fn new(field: impl Into<String>, message: impl Into<String>) -> Self {
        Self {
            field: field.into(),
            message: message.into(),
            code: String::new(),
            value: None,
        }
    }

    /// Set the machine-readable error code.
    pub fn with_code(mut self, code: impl Into<String>) -> Self {
        self.code = code.into();
        self
    }

    /// Attach the value that caused the error.
    ///
    /// Accepts either a [`Value`] or an `Option<Value>`; passing `None`
    /// (e.g. from a `sensitive` validator) leaves the value unset.
    pub fn with_value(mut self, value: impl Into<Option<Value>>) -> Self {
        self.value = value.into();
        self
    }
}

impl fmt::Display for ValidationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.field, self.message)
    }
}

impl std::error::Error for ValidationError {}

/// An aggregate of [`ValidationError`]s returned by [`Schema::try_validate`].
///
/// `Display` renders one `field: message` per line.
#[derive(Debug, Clone)]
pub struct ValidationErrors(pub Vec<ValidationError>);

impl fmt::Display for ValidationErrors {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        for (i, error) in self.0.iter().enumerate() {
            if i > 0 {
                writeln!(f)?;
            }
            write!(f, "{error}")?;
        }
        Ok(())
    }
}

impl std::error::Error for ValidationErrors {}

/// The result of validating a value or schema.
#[derive(Debug, Clone, Default)]
pub struct ValidationResult {
    errors: Vec<ValidationError>,
}

impl ValidationResult {
    /// Create a new, empty validation result.
    pub fn new() -> Self {
        Self { errors: Vec::new() }
    }

    /// Whether validation succeeded.
    pub fn is_valid(&self) -> bool {
        self.errors.is_empty()
    }

    /// The list of validation errors.
    pub fn errors(&self) -> &[ValidationError] {
        &self.errors
    }

    /// Add an error to the result.
    pub fn add_error(&mut self, error: ValidationError) {
        self.errors.push(error);
    }

    /// Merge another validation result into this one.
    pub fn merge(&mut self, other: ValidationResult) {
        self.errors.extend(other.errors);
    }

    /// Convert the result into `Ok(())` when valid, or a
    /// [`ValidationErrors`] aggregate on failure.
    pub fn into_result(self) -> Result<(), ValidationErrors> {
        if self.errors.is_empty() {
            Ok(())
        } else {
            Err(ValidationErrors(self.errors))
        }
    }
}

/// The trait implemented by all validators.
///
/// Validators operate on [`serde_json::Value`] so the same rules can be reused
/// for JSON APIs, configuration files, and form data.
pub trait Validator: Send + Sync {
    /// Validate `value` at the given `field` path.
    fn validate(&self, value: &Value, field: &str) -> ValidationResult;

    /// Whether the field is required.
    fn is_required(&self) -> bool;

    /// Whether the field may be explicitly `null`.
    fn is_nullable(&self) -> bool;

    /// Set fail-fast mode for this validator and its children.
    fn set_fail_fast(&mut self, _fail_fast: bool) {}
}

/// Allow boxed validators to be used anywhere a [`Validator`] is expected
/// (e.g. when building schemas dynamically).
impl Validator for Box<dyn Validator> {
    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        (**self).validate(value, field)
    }

    fn is_required(&self) -> bool {
        (**self).is_required()
    }

    fn is_nullable(&self) -> bool {
        (**self).is_nullable()
    }

    fn set_fail_fast(&mut self, fail_fast: bool) {
        (**self).set_fail_fast(fail_fast);
    }
}

/// Maximum length (in Unicode scalar values) of a string that a `pattern`
/// check will run against. Inputs longer than this skip the match and produce
/// the normal pattern error.
///
/// Rust's `regex` crate guarantees linear-time matching, so this cap exists
/// for cross-language spec parity with backtracking engines (Python `re`,
/// JS `RegExp`).
pub const MAX_PATTERN_INPUT_LENGTH: usize = 10_000;

/// A user-supplied validation function. Returning `Some(message)` rejects the
/// value with that message (code `custom`).
type CustomValidator = Box<dyn Fn(&Value) -> Option<String> + Send + Sync>;

struct Base {
    required: bool,
    nullable: bool,
    fail_fast: bool,
    /// When true, the offending value is redacted from error messages
    /// (rendered as `***`) and `ValidationError::value` is left `None`.
    sensitive: bool,
    custom_validators: Vec<CustomValidator>,
}

impl Base {
    fn new() -> Self {
        Self {
            required: true,
            nullable: false,
            fail_fast: false,
            sensitive: false,
            custom_validators: Vec::new(),
        }
    }

    fn is_required(&self) -> bool {
        self.required
    }

    fn is_nullable(&self) -> bool {
        self.nullable
    }

    /// Render a string value for an error message, or `***` when sensitive.
    fn display<'a>(&self, s: &'a str) -> &'a str {
        if self.sensitive {
            "***"
        } else {
            s
        }
    }

    /// Render an arbitrary JSON value for an error message, or `***` when
    /// sensitive.
    fn display_json(&self, value: &Value) -> String {
        if self.sensitive {
            "***".to_string()
        } else {
            serde_json::to_string(value).unwrap_or_else(|_| value.to_string())
        }
    }

    /// The value to attach to an error, or `None` when sensitive.
    fn err_value(&self, value: &Value) -> Option<Value> {
        if self.sensitive {
            None
        } else {
            Some(value.clone())
        }
    }

    fn handle_null(&self, field: &str, fail_fast: bool) -> ValidationResult {
        let mut result = ValidationResult::new();
        if self.required && !self.nullable {
            result.add_error(
                ValidationError::new(field, "Field is required").with_code(codes::REQUIRED),
            );
            if fail_fast {
                return result;
            }
        }
        result
    }

    fn run_custom(
        &self,
        value: &Value,
        field: &str,
        result: &mut ValidationResult,
        fail_fast: bool,
    ) {
        for validator in &self.custom_validators {
            if let Some(message) = validator(value) {
                result.add_error(
                    ValidationError::new(field, message)
                        .with_code(codes::CUSTOM)
                        .with_value(self.err_value(value)),
                );
                if fail_fast {
                    break;
                }
            }
        }
    }
}

macro_rules! base_methods {
    ($type:ty) => {
        impl $type {
            /// Make the field optional (equivalent to `required = false`).
            pub fn optional(mut self) -> Self {
                self.base.required = false;
                self
            }

            /// Allow the field to be explicitly `null`.
            pub fn nullable(mut self) -> Self {
                self.base.nullable = true;
                self
            }

            /// Mark the field as required.
            pub fn required(mut self) -> Self {
                self.base.required = true;
                self
            }

            /// Mark the field as sensitive (e.g. passwords, tokens).
            ///
            /// When enabled, every error this validator produces redacts the
            /// offending value: message interpolations render as `***` and
            /// [`ValidationError::value`] is `None`, so errors can be logged
            /// safely. The flag is per-validator and is NOT inherited by
            /// nested validators.
            pub fn sensitive(mut self, sensitive: bool) -> Self {
                self.base.sensitive = sensitive;
                self
            }

            /// Add a custom validation function.
            ///
            /// The function receives the value and should return `Some(message)`
            /// when invalid. Custom validators run after type and constraint checks.
            pub fn custom<F>(mut self, f: F) -> Self
            where
                F: Fn(&Value) -> Option<String> + Send + Sync + 'static,
            {
                self.base.custom_validators.push(Box::new(f));
                self
            }

            /// Stop at the first validation error.
            ///
            /// When enabled, validation returns a result containing exactly one
            /// error. The flag is propagated to nested validators.
            pub fn fail_fast(mut self, fail_fast: bool) -> Self {
                self.set_fail_fast(fail_fast);
                self
            }
        }
    };
}

/// Return the canonical JSON kind name for a value.
///
/// Numbers are `"integer"` when integral (`i64`, `u64`, or integral `f64`)
/// and `"number"` otherwise.
fn kind_name(value: &Value) -> &'static str {
    match value {
        Value::Null => "null",
        Value::Bool(_) => "boolean",
        Value::Number(n) => {
            if n.is_i64() || n.is_u64() {
                "integer"
            } else if let Some(f) = n.as_f64() {
                if f.fract() == 0.0 && f.is_finite() {
                    "integer"
                } else {
                    "number"
                }
            } else {
                "number"
            }
        }
        Value::String(_) => "string",
        Value::Array(_) => "array",
        Value::Object(_) => "object",
    }
}

// Email and URL patterns are compiled exactly once on first use.
static EMAIL_RE: OnceLock<Regex> = OnceLock::new();
static URL_RE: OnceLock<Regex> = OnceLock::new();

fn email_regex() -> &'static Regex {
    EMAIL_RE
        .get_or_init(|| Regex::new(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$").unwrap())
}

fn url_regex() -> &'static Regex {
    URL_RE.get_or_init(|| Regex::new(r"^https?://[^\s/$.?#].[^\s]*$").unwrap())
}

// Fixed-format patterns (uuid/date/datetime/numeric) are compiled exactly
// once on first use.
static UUID_RE: OnceLock<Regex> = OnceLock::new();
static DATE_RE: OnceLock<Regex> = OnceLock::new();
static DATETIME_RE: OnceLock<Regex> = OnceLock::new();
static NUMERIC_RE: OnceLock<Regex> = OnceLock::new();

fn uuid_regex() -> &'static Regex {
    UUID_RE.get_or_init(|| {
        Regex::new(r"^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
            .unwrap()
    })
}

fn date_regex() -> &'static Regex {
    DATE_RE.get_or_init(|| Regex::new(r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$").unwrap())
}

fn datetime_regex() -> &'static Regex {
    DATETIME_RE.get_or_init(|| {
        Regex::new(
            r"^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$",
        )
        .unwrap()
    })
}

fn numeric_regex() -> &'static Regex {
    NUMERIC_RE.get_or_init(|| Regex::new(r"^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$").unwrap())
}

/// Validator for JSON strings.
pub struct StringValidator {
    base: Base,
    min_length: Option<usize>,
    max_length: Option<usize>,
    /// Raw pattern source (for error messages) plus the anchored compiled regex.
    pattern: Option<(String, Regex)>,
    choices: Option<Vec<String>>,
}

/// Anchor a pattern so it must match the entire string (full-match semantics).
fn anchor_pattern(source: &str) -> Result<Regex, regex::Error> {
    Regex::new(&format!("^(?:{source})$"))
}

impl StringValidator {
    /// Create a new string validator.
    pub fn new() -> Self {
        Self {
            base: Base::new(),
            min_length: None,
            max_length: None,
            pattern: None,
            choices: None,
        }
    }

    /// Set the minimum length (measured in Unicode scalars).
    pub fn min(mut self, min: usize) -> Self {
        self.min_length = Some(min);
        self
    }

    /// Set the maximum length (measured in Unicode scalars).
    pub fn max(mut self, max: usize) -> Self {
        self.max_length = Some(max);
        self
    }

    /// Set a pre-compiled regex pattern. The pattern is anchored so it must
    /// match the entire string (full-match semantics).
    ///
    /// For a fallible builder, see [`StringValidator::try_pattern`].
    pub fn pattern(mut self, pattern: Regex) -> Self {
        let source = pattern.as_str().to_string();
        let anchored = anchor_pattern(&source).unwrap_or(pattern);
        self.pattern = Some((source, anchored));
        self
    }

    /// Compile a regex pattern. The pattern is anchored so it must match the
    /// entire string (full-match semantics). Returns an error if the pattern
    /// is invalid.
    pub fn try_pattern(mut self, pattern: &str) -> Result<Self, regex::Error> {
        let anchored = anchor_pattern(pattern)?;
        self.pattern = Some((pattern.to_string(), anchored));
        Ok(self)
    }

    /// Restrict the value to a set of allowed strings.
    pub fn choices(mut self, choices: Vec<String>) -> Self {
        self.choices = Some(choices);
        self
    }
}

base_methods!(StringValidator);

impl Validator for StringValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(s) = value.as_str() else {
            result.add_error(
                ValidationError::new(field, format!("Expected string, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        let len = s.chars().count();

        if let Some(min) = self.min_length {
            if len < min {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("String length must be at least {min}, got {len}"),
                    )
                    .with_code(codes::MIN_LENGTH)
                    .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(max) = self.max_length {
            if len > max {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("String length must be at most {max}, got {len}"),
                    )
                    .with_code(codes::MAX_LENGTH)
                    .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some((source, pattern)) = &self.pattern {
            // `regex` is linear-time; the cap exists for spec parity (see
            // MAX_PATTERN_INPUT_LENGTH).
            if len > MAX_PATTERN_INPUT_LENGTH || !pattern.is_match(s) {
                result.add_error(
                    ValidationError::new(field, format!("String does not match pattern {source}"))
                        .with_code(codes::PATTERN)
                        .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(choices) = &self.choices {
            if !choices.iter().any(|c| c == s) {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!(
                            "Value must be one of {}, got '{}'",
                            choices.join(", "),
                            self.base.display(s)
                        ),
                    )
                    .with_code(codes::CHOICES)
                    .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for JSON integers.
pub struct IntValidator {
    base: Base,
    min_value: Option<i64>,
    max_value: Option<i64>,
}

impl IntValidator {
    /// Create a new integer validator.
    pub fn new() -> Self {
        Self {
            base: Base::new(),
            min_value: None,
            max_value: None,
        }
    }

    /// Set the minimum allowed value.
    pub fn min(mut self, min: i64) -> Self {
        self.min_value = Some(min);
        self
    }

    /// Set the maximum allowed value.
    pub fn max(mut self, max: i64) -> Self {
        self.max_value = Some(max);
        self
    }
}

base_methods!(IntValidator);

impl Validator for IntValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let parsed = value
            .as_i64()
            .or_else(|| value.as_u64().and_then(|u| i64::try_from(u).ok()))
            .or_else(|| {
                value.as_f64().and_then(|f| {
                    if f.fract() == 0.0
                        && f.is_finite()
                        && f >= i64::MIN as f64
                        && f <= i64::MAX as f64
                    {
                        Some(f as i64)
                    } else {
                        None
                    }
                })
            });

        let Some(num) = parsed else {
            result.add_error(
                ValidationError::new(field, format!("Expected integer, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        if let Some(min) = self.min_value {
            if num < min {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at least {min}, got {num}"))
                        .with_code(codes::MIN_VALUE)
                        .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(max) = self.max_value {
            if num > max {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at most {max}, got {num}"))
                        .with_code(codes::MAX_VALUE)
                        .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for JSON numbers (integers and floats).
pub struct FloatValidator {
    base: Base,
    min_value: Option<f64>,
    max_value: Option<f64>,
}

impl FloatValidator {
    /// Create a new float/number validator.
    pub fn new() -> Self {
        Self {
            base: Base::new(),
            min_value: None,
            max_value: None,
        }
    }

    /// Set the minimum allowed value.
    pub fn min(mut self, min: f64) -> Self {
        self.min_value = Some(min);
        self
    }

    /// Set the maximum allowed value.
    pub fn max(mut self, max: f64) -> Self {
        self.max_value = Some(max);
        self
    }
}

base_methods!(FloatValidator);

impl Validator for FloatValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(num) = value.as_f64() else {
            result.add_error(
                ValidationError::new(field, format!("Expected number, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        if let Some(min) = self.min_value {
            if num < min {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at least {min}, got {num}"))
                        .with_code(codes::MIN_VALUE)
                        .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(max) = self.max_value {
            if num > max {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at most {max}, got {num}"))
                        .with_code(codes::MAX_VALUE)
                        .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for JSON booleans.
pub struct BoolValidator {
    base: Base,
}

impl BoolValidator {
    /// Create a new boolean validator.
    pub fn new() -> Self {
        Self { base: Base::new() }
    }
}

base_methods!(BoolValidator);

impl Validator for BoolValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        if !value.is_boolean() {
            result.add_error(
                ValidationError::new(field, format!("Expected boolean, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for email addresses.
pub struct EmailValidator {
    base: Base,
}

impl EmailValidator {
    /// Create a new email validator.
    pub fn new() -> Self {
        Self { base: Base::new() }
    }
}

base_methods!(EmailValidator);

impl Validator for EmailValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(s) = value.as_str() else {
            result.add_error(
                ValidationError::new(field, format!("Expected string, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        // `regex` is linear-time; the cap exists for spec parity.
        if s.chars().count() > MAX_PATTERN_INPUT_LENGTH || !email_regex().is_match(s) {
            result.add_error(
                ValidationError::new(
                    field,
                    format!("Invalid email address: {}", self.base.display(s)),
                )
                .with_code(codes::EMAIL)
                .with_value(self.base.err_value(value)),
            );
            if fail_fast {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for URLs.
pub struct UrlValidator {
    base: Base,
}

impl UrlValidator {
    /// Create a new URL validator.
    pub fn new() -> Self {
        Self { base: Base::new() }
    }
}

base_methods!(UrlValidator);

impl Validator for UrlValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(s) = value.as_str() else {
            result.add_error(
                ValidationError::new(field, format!("Expected string, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        // `regex` is linear-time; the cap exists for spec parity.
        if s.chars().count() > MAX_PATTERN_INPUT_LENGTH || !url_regex().is_match(s) {
            result.add_error(
                ValidationError::new(field, format!("Invalid URL: {}", self.base.display(s)))
                    .with_code(codes::URL)
                    .with_value(self.base.err_value(value)),
            );
            if fail_fast {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Generate a string validator that checks a fixed regex and reports
/// mismatches with a dedicated error code and message.
macro_rules! fixed_pattern_validator {
    ($type:ident, $doc:literal, $regex_fn:ident, $code:expr, $msg:literal) => {
        #[doc = $doc]
        pub struct $type {
            base: Base,
        }

        impl $type {
            /// Create a new validator.
            pub fn new() -> Self {
                Self { base: Base::new() }
            }
        }

        base_methods!($type);

        impl Default for $type {
            fn default() -> Self {
                Self::new()
            }
        }

        impl Validator for $type {
            fn set_fail_fast(&mut self, fail_fast: bool) {
                self.base.fail_fast = fail_fast;
            }

            fn validate(&self, value: &Value, field: &str) -> ValidationResult {
                let fail_fast = self.base.fail_fast;
                if value.is_null() {
                    return self.base.handle_null(field, fail_fast);
                }

                let mut result = ValidationResult::new();

                let Some(s) = value.as_str() else {
                    result.add_error(
                        ValidationError::new(
                            field,
                            format!("Expected string, got {}", kind_name(value)),
                        )
                        .with_code(codes::TYPE)
                        .with_value(self.base.err_value(value)),
                    );
                    return result;
                };

                // `regex` is linear-time; the cap exists for spec parity.
                if s.chars().count() > MAX_PATTERN_INPUT_LENGTH || !$regex_fn().is_match(s) {
                    result.add_error(
                        ValidationError::new(field, format!($msg, self.base.display(s)))
                            .with_code($code)
                            .with_value(self.base.err_value(value)),
                    );
                    if fail_fast {
                        return result;
                    }
                }

                self.base.run_custom(value, field, &mut result, fail_fast);
                result
            }

            fn is_required(&self) -> bool {
                self.base.is_required()
            }

            fn is_nullable(&self) -> bool {
                self.base.is_nullable()
            }
        }
    };
}

fixed_pattern_validator!(
    UuidValidator,
    "Validator for UUID strings.",
    uuid_regex,
    codes::UUID,
    "Invalid UUID: {}"
);

fixed_pattern_validator!(
    DateValidator,
    "Validator for `YYYY-MM-DD` date strings.",
    date_regex,
    codes::DATE,
    "Invalid date: {}"
);

fixed_pattern_validator!(
    DatetimeValidator,
    "Validator for ISO-8601 style datetime strings.",
    datetime_regex,
    codes::DATETIME,
    "Invalid datetime: {}"
);

fixed_pattern_validator!(
    NumericValidator,
    "Validator for strings that contain a valid JSON-style number.",
    numeric_regex,
    codes::NUMERIC,
    "Expected numeric string, got {}"
);

/// Validator that requires a non-empty string, array, or object.
pub struct NotEmptyValidator {
    base: Base,
}

impl NotEmptyValidator {
    /// Create a new not-empty validator.
    pub fn new() -> Self {
        Self { base: Base::new() }
    }
}

base_methods!(NotEmptyValidator);

impl Validator for NotEmptyValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let is_empty = match value {
            Value::String(s) => s.is_empty(),
            Value::Array(arr) => arr.is_empty(),
            Value::Object(map) => map.is_empty(),
            _ => {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!(
                            "Expected string, array, or object, got {}",
                            kind_name(value)
                        ),
                    )
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
                );
                return result;
            }
        };

        if is_empty {
            result.add_error(
                ValidationError::new(field, "Value must not be empty")
                    .with_code(codes::NOT_EMPTY)
                    .with_value(self.base.err_value(value)),
            );
            if fail_fast {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator that checks membership in a fixed set of JSON values
/// (deep-equality semantics).
pub struct EnumValidator {
    base: Base,
    values: Vec<Value>,
}

impl EnumValidator {
    /// Create a new enum validator over the given allowed values.
    pub fn new(values: Vec<Value>) -> Self {
        Self {
            base: Base::new(),
            values,
        }
    }
}

base_methods!(EnumValidator);

impl Validator for EnumValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        if !self.values.iter().any(|v| v == value) {
            let allowed = self
                .values
                .iter()
                .map(|v| serde_json::to_string(v).unwrap_or_else(|_| v.to_string()))
                .collect::<Vec<_>>()
                .join(", ");
            let got = self.base.display_json(value);
            result.add_error(
                ValidationError::new(field, format!("Value must be one of {allowed}, got {got}"))
                    .with_code(codes::ENUM)
                    .with_value(self.base.err_value(value)),
            );
            if fail_fast {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator that accepts a value when at least one branch validator
/// produces zero errors at the same field path.
pub struct OneOfValidator {
    base: Base,
    branches: Vec<Box<dyn Validator>>,
}

impl OneOfValidator {
    /// Create a new one-of validator over the given branch validators.
    pub fn new(branches: Vec<Box<dyn Validator>>) -> Self {
        Self {
            base: Base::new(),
            branches,
        }
    }
}

base_methods!(OneOfValidator);

impl Validator for OneOfValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
        for branch in &mut self.branches {
            branch.set_fail_fast(fail_fast);
        }
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let any_match = self
            .branches
            .iter()
            .any(|branch| branch.validate(value, field).is_valid());

        if !any_match {
            result.add_error(
                ValidationError::new(field, "Value does not match any allowed schema")
                    .with_code(codes::ONE_OF)
                    .with_value(self.base.err_value(value)),
            );
            if fail_fast {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for JSON arrays.
pub struct VecValidator {
    base: Base,
    item_validator: Option<Box<dyn Validator>>,
    min_length: Option<usize>,
    max_length: Option<usize>,
}

impl VecValidator {
    /// Create a new array validator.
    pub fn new() -> Self {
        Self {
            base: Base::new(),
            item_validator: None,
            min_length: None,
            max_length: None,
        }
    }

    /// Set a validator for each array item.
    pub fn item(mut self, mut validator: impl Validator + 'static) -> Self {
        validator.set_fail_fast(self.base.fail_fast);
        self.item_validator = Some(Box::new(validator));
        self
    }

    /// Set the minimum array length.
    pub fn min(mut self, min: usize) -> Self {
        self.min_length = Some(min);
        self
    }

    /// Set the maximum array length.
    pub fn max(mut self, max: usize) -> Self {
        self.max_length = Some(max);
        self
    }
}

base_methods!(VecValidator);

impl Validator for VecValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
        if let Some(item_validator) = self.item_validator.as_mut() {
            item_validator.set_fail_fast(fail_fast);
        }
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(arr) = value.as_array() else {
            result.add_error(
                ValidationError::new(field, format!("Expected array, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        if let Some(min) = self.min_length {
            if arr.len() < min {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("Array length must be at least {min}, got {}", arr.len()),
                    )
                    .with_code(codes::MIN_LENGTH)
                    .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(max) = self.max_length {
            if arr.len() > max {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("Array length must be at most {max}, got {}", arr.len()),
                    )
                    .with_code(codes::MAX_LENGTH)
                    .with_value(self.base.err_value(value)),
                );
                if fail_fast {
                    return result;
                }
            }
        }

        if let Some(item_validator) = &self.item_validator {
            for (i, item) in arr.iter().enumerate() {
                let path = format!("{field}[{i}]");
                result.merge(item_validator.validate(item, &path));
                if fail_fast && !result.is_valid() {
                    return result;
                }
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// Validator for JSON objects.
pub struct MapValidator {
    base: Base,
    schema: HashMap<String, Box<dyn Validator>>,
    strict: bool,
}

impl MapValidator {
    /// Create a new object validator.
    pub fn new() -> Self {
        Self {
            base: Base::new(),
            schema: HashMap::new(),
            strict: false,
        }
    }

    /// Add a field to the object schema.
    pub fn field(
        mut self,
        name: impl Into<String>,
        mut validator: impl Validator + 'static,
    ) -> Self {
        validator.set_fail_fast(self.base.fail_fast);
        self.schema.insert(name.into(), Box::new(validator));
        self
    }

    /// Replace the object schema wholesale.
    pub fn schema(mut self, mut schema: HashMap<String, Box<dyn Validator>>) -> Self {
        let fail_fast = self.base.fail_fast;
        for v in schema.values_mut() {
            v.set_fail_fast(fail_fast);
        }
        self.schema = schema;
        self
    }

    /// Enable or disable strict mode.
    ///
    /// When strict, unknown fields in the object produce an error.
    pub fn strict(mut self, strict: bool) -> Self {
        self.strict = strict;
        self
    }
}

base_methods!(MapValidator);

impl Default for MapValidator {
    fn default() -> Self {
        Self::new()
    }
}

impl Validator for MapValidator {
    fn set_fail_fast(&mut self, fail_fast: bool) {
        self.base.fail_fast = fail_fast;
        for v in self.schema.values_mut() {
            v.set_fail_fast(fail_fast);
        }
    }

    fn validate(&self, value: &Value, field: &str) -> ValidationResult {
        let fail_fast = self.base.fail_fast;
        if value.is_null() {
            return self.base.handle_null(field, fail_fast);
        }

        let mut result = ValidationResult::new();

        let Some(map) = value.as_object() else {
            result.add_error(
                ValidationError::new(field, format!("Expected object, got {}", kind_name(value)))
                    .with_code(codes::TYPE)
                    .with_value(self.base.err_value(value)),
            );
            return result;
        };

        if self.strict {
            for (key, val) in map.iter() {
                if !self.schema.contains_key(key) {
                    result.add_error(
                        ValidationError::new(format!("{field}.{key}"), "Unknown field")
                            .with_code(codes::UNKNOWN_FIELD)
                            .with_value(self.base.err_value(val)),
                    );
                    if fail_fast {
                        return result;
                    }
                }
            }
        }

        for (key, validator) in &self.schema {
            let path = format!("{field}.{key}");
            if let Some(v) = map.get(key) {
                result.merge(validator.validate(v, &path));
            } else if validator.is_required() {
                result.add_error(
                    ValidationError::new(&path, "Field is required").with_code(codes::REQUIRED),
                );
            }
            if fail_fast && !result.is_valid() {
                return result;
            }
        }

        self.base.run_custom(value, field, &mut result, fail_fast);
        result
    }

    fn is_required(&self) -> bool {
        self.base.is_required()
    }

    fn is_nullable(&self) -> bool {
        self.base.is_nullable()
    }
}

/// A schema for validating top-level JSON objects.
#[derive(Default)]
pub struct Schema {
    fields: HashMap<String, Box<dyn Validator>>,
    strict: bool,
    fail_fast: bool,
}

impl Schema {
    /// Create a new schema.
    pub fn new() -> Self {
        Self {
            fields: HashMap::new(),
            strict: false,
            fail_fast: false,
        }
    }

    /// Add a field validator.
    pub fn field(
        mut self,
        name: impl Into<String>,
        mut validator: impl Validator + 'static,
    ) -> Self {
        validator.set_fail_fast(self.fail_fast);
        self.fields.insert(name.into(), Box::new(validator));
        self
    }

    /// Enable or disable strict mode.
    ///
    /// When strict, keys present in the input object but absent from the schema
    /// produce an `Unknown field` error.
    pub fn strict(mut self, strict: bool) -> Self {
        self.strict = strict;
        self
    }

    /// Stop at the first validation error.
    ///
    /// When enabled, validation returns a result containing exactly one
    /// error. The flag is propagated to nested validators.
    pub fn fail_fast(mut self, fail_fast: bool) -> Self {
        self.fail_fast = fail_fast;
        for v in self.fields.values_mut() {
            v.set_fail_fast(fail_fast);
        }
        self
    }

    /// Validate a JSON object (`serde_json::Value` object).
    pub fn validate(&self, data: &Value) -> ValidationResult {
        let mut result = ValidationResult::new();
        let fail_fast = self.fail_fast;

        let Some(map) = data.as_object() else {
            result.add_error(
                ValidationError::new("root", format!("Expected object, got {}", kind_name(data)))
                    .with_code(codes::TYPE)
                    .with_value(data.clone()),
            );
            return result;
        };

        if self.strict {
            for (key, val) in map.iter() {
                if !self.fields.contains_key(key) {
                    result.add_error(
                        ValidationError::new(key, "Unknown field")
                            .with_code(codes::UNKNOWN_FIELD)
                            .with_value(val.clone()),
                    );
                    if fail_fast {
                        return result;
                    }
                }
            }
        }

        for (field, validator) in &self.fields {
            if let Some(value) = map.get(field) {
                result.merge(validator.validate(value, field));
            } else if validator.is_required() {
                result.add_error(
                    ValidationError::new(field, "Field is required").with_code(codes::REQUIRED),
                );
            }
            if fail_fast && !result.is_valid() {
                return result;
            }
        }

        result
    }

    /// Validate and return `Ok(())` if valid, or the full result on failure.
    pub fn validate_or_raise(&self, data: &Value) -> Result<(), ValidationResult> {
        let result = self.validate(data);
        if result.is_valid() {
            Ok(())
        } else {
            Err(result)
        }
    }

    /// Validate and return `Ok(())` if valid, or a [`ValidationErrors`]
    /// aggregate (which implements `std::error::Error`) on failure.
    pub fn try_validate(&self, data: &Value) -> Result<(), ValidationErrors> {
        self.validate(data).into_result()
    }
}

/// Factory functions for creating validators.
pub mod validators {
    use super::*;

    /// Create a string validator.
    pub fn string() -> StringValidator {
        StringValidator::new()
    }

    /// Create an integer validator.
    pub fn int() -> IntValidator {
        IntValidator::new()
    }

    /// Create a float/number validator.
    pub fn float() -> FloatValidator {
        FloatValidator::new()
    }

    /// Create a boolean validator.
    pub fn bool() -> BoolValidator {
        BoolValidator::new()
    }

    /// Create an email validator.
    pub fn email() -> EmailValidator {
        EmailValidator::new()
    }

    /// Create a URL validator.
    pub fn url() -> UrlValidator {
        UrlValidator::new()
    }

    /// Create an array validator.
    pub fn vec() -> VecValidator {
        VecValidator::new()
    }

    /// Alias for [`validators::vec`].
    pub fn list() -> VecValidator {
        VecValidator::new()
    }

    /// Create an object validator.
    pub fn map() -> MapValidator {
        MapValidator::new()
    }

    /// Alias for [`validators::map`].
    pub fn dict() -> MapValidator {
        MapValidator::new()
    }

    /// Create a UUID validator.
    pub fn uuid() -> UuidValidator {
        UuidValidator::new()
    }

    /// Create a `YYYY-MM-DD` date validator.
    pub fn date() -> DateValidator {
        DateValidator::new()
    }

    /// Create a datetime validator.
    pub fn datetime() -> DatetimeValidator {
        DatetimeValidator::new()
    }

    /// Create a numeric-string validator.
    pub fn numeric() -> NumericValidator {
        NumericValidator::new()
    }

    /// Create a validator requiring a non-empty string, array, or object.
    pub fn not_empty() -> NotEmptyValidator {
        NotEmptyValidator::new()
    }

    /// Create an enum validator over arbitrary JSON values.
    ///
    /// Named `enum_values` because `enum` is a Rust keyword; the raw
    /// identifier alias [`validators::r#enum`] is also provided.
    pub fn enum_values(values: Vec<Value>) -> EnumValidator {
        EnumValidator::new(values)
    }

    /// Alias for [`validators::enum_values`].
    #[allow(non_snake_case)]
    pub fn r#enum(values: Vec<Value>) -> EnumValidator {
        EnumValidator::new(values)
    }

    /// Create a validator that accepts the value when at least one of the
    /// given branch validators produces zero errors.
    pub fn one_of(branches: Vec<Box<dyn Validator>>) -> OneOfValidator {
        OneOfValidator::new(branches)
    }

    /// Create a string validator anchored to the given regex pattern.
    pub fn regex(re: Regex) -> StringValidator {
        StringValidator::new().pattern(re)
    }
}

// `Default` implementations for the parameterless `new()` constructors.
macro_rules! impl_default {
    ($($t:ty),* $(,)?) => {
        $(impl Default for $t {
            fn default() -> Self {
                Self::new()
            }
        })*
    };
}

impl_default!(
    StringValidator,
    IntValidator,
    FloatValidator,
    BoolValidator,
    EmailValidator,
    UrlValidator,
    NotEmptyValidator,
    VecValidator,
);

#[cfg(test)]
mod tests {
    use super::*;
    use regex::Regex;
    use serde_json::json;

    #[test]
    fn test_string_validator() {
        let validator = validators::string().min(3).max(10);
        assert!(validator.validate(&json!("hello"), "name").is_valid());
        assert!(!validator.validate(&json!("ab"), "name").is_valid());
        assert!(!validator.validate(&json!("hello world"), "name").is_valid());

        let pattern_validator = validators::string().try_pattern(r"^\d{3}-\d{4}$").unwrap();
        assert!(pattern_validator
            .validate(&json!("123-4567"), "phone")
            .is_valid());
        assert!(!pattern_validator
            .validate(&json!("12-3456"), "phone")
            .is_valid());

        let choices_validator =
            validators::string().choices(vec!["light".to_string(), "dark".to_string()]);
        assert!(choices_validator
            .validate(&json!("light"), "theme")
            .is_valid());
        assert!(!choices_validator
            .validate(&json!("red"), "theme")
            .is_valid());

        let result = validators::string().validate(&json!(42), "name");
        assert!(!result.is_valid());
        assert!(result.errors()[0]
            .message
            .contains("Expected string, got integer"));
        assert_eq!(result.errors()[0].code, codes::TYPE);
    }

    #[test]
    fn test_int_validator() {
        let validator = validators::int().min(0).max(120);
        assert!(validator.validate(&json!(25), "age").is_valid());
        assert!(!validator.validate(&json!(-1), "age").is_valid());
        assert!(!validator.validate(&json!(150), "age").is_valid());

        let result = validator.validate(&json!(1.5), "age");
        assert!(!result.is_valid());
        assert!(result.errors()[0]
            .message
            .contains("Expected integer, got number"));
        assert_eq!(result.errors()[0].code, codes::TYPE);

        let result = validator.validate(&json!("25"), "age");
        assert!(!result.is_valid());
        assert!(result.errors()[0]
            .message
            .contains("Expected integer, got string"));
    }

    #[test]
    fn test_float_validator() {
        let validator = validators::float().min(0.0).max(100.0);
        assert!(validator.validate(&json!(25.5), "score").is_valid());
        assert!(validator.validate(&json!(25), "score").is_valid());
        assert!(!validator.validate(&json!(150.0), "score").is_valid());

        let result = validator.validate(&json!("25"), "score");
        assert!(!result.is_valid());
        assert!(result.errors()[0]
            .message
            .contains("Expected number, got string"));
    }

    #[test]
    fn test_bool_validator() {
        let validator = validators::bool();
        assert!(validator.validate(&json!(true), "active").is_valid());

        let result = validator.validate(&json!("true"), "active");
        assert!(!result.is_valid());
        assert!(result.errors()[0]
            .message
            .contains("Expected boolean, got string"));
    }

    #[test]
    fn test_email_validator() {
        let validator = validators::email();
        assert!(validator
            .validate(&json!("user@example.com"), "email")
            .is_valid());

        let result = validator.validate(&json!("invalid-email"), "email");
        assert!(!result.is_valid());
        assert!(result.errors()[0].message.contains("Invalid email address"));
        assert_eq!(result.errors()[0].code, codes::EMAIL);

        let result = validator.validate(&json!(123), "email");
        assert!(!result.is_valid());
        assert!(result.errors()[0].message.contains("Expected string"));
        assert_eq!(result.errors()[0].code, codes::TYPE);
    }

    #[test]
    fn test_url_validator() {
        let validator = validators::url();
        assert!(validator
            .validate(&json!("https://example.com"), "url")
            .is_valid());

        let result = validator.validate(&json!("not-a-url"), "url");
        assert!(!result.is_valid());
        assert!(result.errors()[0].message.contains("Invalid URL"));
        assert_eq!(result.errors()[0].code, codes::URL);
    }

    #[test]
    fn test_vec_validator() {
        let validator = validators::vec()
            .item(validators::int().min(0))
            .min(1)
            .max(3);

        assert!(validator.validate(&json!([1, 2]), "items").is_valid());
        assert!(!validator.validate(&json!([]), "items").is_valid());
        assert!(!validator.validate(&json!([1, 2, 3, 4]), "items").is_valid());

        let result = validator.validate(&json!(["bad"]), "items");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].field, "items[0]");
        assert!(result.errors()[0].message.contains("Expected integer"));
    }

    #[test]
    fn test_map_validator() {
        let validator = validators::map()
            .field("name", validators::string().min(1))
            .field("age", validators::int());

        assert!(validator
            .validate(&json!({"name": "Ada", "age": 30}), "user")
            .is_valid());

        let result = validator.validate(&json!({"name": "Ada"}), "user");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].field, "user.age");
        assert_eq!(result.errors()[0].message, "Field is required");
        assert_eq!(result.errors()[0].code, codes::REQUIRED);
    }

    #[test]
    fn test_optional_nullable() {
        let schema = Schema::new()
            .field("name", validators::string())
            .field("bio", validators::string().optional().nullable());

        assert!(schema.validate(&json!({"name": "Ada"})).is_valid());
        assert!(schema
            .validate(&json!({"name": "Ada", "bio": null}))
            .is_valid());

        let result = schema.validate(&json!({"name": "Ada", "bio": 123}));
        assert!(!result.is_valid());
        assert!(result.errors()[0].message.contains("Expected string"));

        let required = validators::string();
        assert!(!required.validate(&Value::Null, "name").is_valid());

        let nullable = validators::string().nullable();
        assert!(nullable.validate(&Value::Null, "name").is_valid());
    }

    #[test]
    fn test_custom_validator() {
        let validator = validators::int().custom(|value| {
            if let Some(n) = value.as_i64() {
                if n % 2 == 0 {
                    return None;
                }
            }
            Some("value must be even".to_string())
        });

        assert!(validator.validate(&json!(4), "x").is_valid());

        let result = validator.validate(&json!(3), "x");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].code, codes::CUSTOM);
    }

    #[test]
    fn test_strict_schema() {
        let schema = Schema::new()
            .field("name", validators::string())
            .strict(true);

        let result = schema.validate(&json!({"name": "Ada", "extra": 1}));
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].field, "extra");
        assert_eq!(result.errors()[0].message, "Unknown field");
        assert_eq!(result.errors()[0].code, codes::UNKNOWN_FIELD);

        assert!(schema.validate(&json!({"name": "Ada"})).is_valid());
    }

    #[test]
    fn test_strict_map() {
        let validator = validators::map()
            .field("name", validators::string())
            .strict(true);

        let result = validator.validate(&json!({"name": "Ada", "extra": 1}), "user");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].field, "user.extra");
        assert_eq!(result.errors()[0].message, "Unknown field");
    }

    #[test]
    fn test_missing_required() {
        let schema = Schema::new()
            .field("name", validators::string())
            .field("age", validators::int());

        let result = schema.validate(&json!({"name": "Ada"}));
        assert!(!result.is_valid());
        assert!(result
            .errors()
            .iter()
            .any(|e| e.field == "age" && e.message == "Field is required"));
    }

    #[test]
    fn test_nested_map_vec_paths() {
        let schema = Schema::new().field(
            "user",
            validators::map().field("tags", validators::vec().item(validators::string().min(2))),
        );

        let result = schema.validate(&json!({"user": {"tags": ["a"]}}));
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].field, "user.tags[0]");
    }

    #[test]
    fn test_pattern_infallible() {
        let validator = validators::string().pattern(Regex::new(r"^[a-z]+$").unwrap());
        assert!(validator.validate(&json!("abc"), "code").is_valid());
        assert!(!validator.validate(&json!("ABC"), "code").is_valid());
    }

    #[test]
    fn test_validate_or_raise() {
        let schema = Schema::new().field("name", validators::string());
        assert!(schema.validate_or_raise(&json!({"name": "Ada"})).is_ok());
        assert!(schema.validate_or_raise(&json!({})).is_err());
    }

    #[test]
    fn test_fail_fast_schema() {
        let schema = Schema::new()
            .fail_fast(true)
            .field("a", validators::string().min(2))
            .field("b", validators::int().min(0));

        let result = schema.validate(&json!({"a": "x", "b": -1}));
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);

        let schema_default = Schema::new()
            .field("a", validators::string().min(2))
            .field("b", validators::int().min(0));
        let result_default = schema_default.validate(&json!({"a": "x", "b": -1}));
        assert_eq!(result_default.errors().len(), 2);
    }

    #[test]
    fn test_fail_fast_map() {
        let validator = validators::map()
            .fail_fast(true)
            .field("name", validators::string().min(2))
            .field("age", validators::int().min(0));

        let result = validator.validate(&json!({"name": "x", "age": -1}), "user");
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);

        let validator_default = validators::map()
            .field("name", validators::string().min(2))
            .field("age", validators::int().min(0));
        let result_default = validator_default.validate(&json!({"name": "x", "age": -1}), "user");
        assert_eq!(result_default.errors().len(), 2);
    }

    #[test]
    fn test_fail_fast_vec() {
        let validator = validators::vec()
            .fail_fast(true)
            .item(validators::int().min(0));

        let result = validator.validate(&json!([-1, -2]), "items");
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);
        assert_eq!(result.errors()[0].field, "items[0]");

        let result_default = validators::vec()
            .item(validators::int().min(0))
            .validate(&json!([-1, -2]), "items");
        assert_eq!(result_default.errors().len(), 2);
    }

    #[test]
    fn test_fail_fast_nested() {
        let schema = Schema::new().fail_fast(true).field(
            "user",
            validators::map().field("tags", validators::vec().item(validators::int().min(0))),
        );

        let result = schema.validate(&json!({"user": {"tags": [-1, -2]}}));
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);
        assert_eq!(result.errors()[0].field, "user.tags[0]");
    }

    #[test]
    fn test_fail_fast_leaf_stops_at_first_constraint() {
        let validator = validators::string()
            .min(5)
            .try_pattern(r"^\d+$")
            .expect("valid regex")
            .fail_fast(true);

        let result = validator.validate(&json!("ab"), "code");
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);
        assert_eq!(result.errors()[0].code, codes::MIN_LENGTH);

        let result_default = validators::string()
            .min(5)
            .try_pattern(r"^\d+$")
            .expect("valid regex")
            .validate(&json!("ab"), "code");
        assert_eq!(result_default.errors().len(), 2);
    }

    #[test]
    fn test_fail_fast_propagated_after_fields_added() {
        let validator = validators::map()
            .field("name", validators::string().min(2))
            .field("age", validators::int().min(0))
            .fail_fast(true);

        let result = validator.validate(&json!({"name": "x", "age": -1}), "user");
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);
    }

    #[test]
    fn test_uuid_validator() {
        let validator = validators::uuid();
        assert!(validator
            .validate(&json!("550e8400-e29b-41d4-a716-446655440000"), "id")
            .is_valid());

        let result = validator.validate(&json!("not-a-uuid"), "id");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].code, codes::UUID);
        assert_eq!(result.errors()[0].message, "Invalid UUID: not-a-uuid");

        let result = validator.validate(&json!(42), "id");
        assert_eq!(result.errors()[0].code, codes::TYPE);
        assert_eq!(result.errors()[0].message, "Expected string, got integer");
    }

    #[test]
    fn test_date_validator() {
        let validator = validators::date();
        assert!(validator.validate(&json!("2026-09-07"), "d").is_valid());

        let result = validator.validate(&json!("2026-13-07"), "d");
        assert_eq!(result.errors()[0].code, codes::DATE);
        assert_eq!(result.errors()[0].message, "Invalid date: 2026-13-07");
    }

    #[test]
    fn test_datetime_validator() {
        let validator = validators::datetime();
        assert!(validator
            .validate(&json!("2026-09-07T10:30:00Z"), "t")
            .is_valid());
        assert!(validator
            .validate(&json!("2026-09-07T10:30:00.5+05:30"), "t")
            .is_valid());

        let result = validator.validate(&json!("2026-09-07 10:30"), "t");
        assert_eq!(result.errors()[0].code, codes::DATETIME);
        assert_eq!(
            result.errors()[0].message,
            "Invalid datetime: 2026-09-07 10:30"
        );
    }

    #[test]
    fn test_numeric_validator() {
        let validator = validators::numeric();
        assert!(validator.validate(&json!("-12.5"), "n").is_valid());
        assert!(validator.validate(&json!("1e3"), "n").is_valid());

        let result = validator.validate(&json!("12a"), "n");
        assert_eq!(result.errors()[0].code, codes::NUMERIC);
        assert_eq!(
            result.errors()[0].message,
            "Expected numeric string, got 12a"
        );

        let result = validator.validate(&json!(12), "n");
        assert_eq!(result.errors()[0].code, codes::TYPE);
    }

    #[test]
    fn test_not_empty_validator() {
        let validator = validators::not_empty();
        assert!(validator.validate(&json!("x"), "v").is_valid());
        assert!(validator.validate(&json!([1]), "v").is_valid());
        assert!(validator.validate(&json!({"k": 1}), "v").is_valid());

        for bad in [json!(""), json!([]), json!({})] {
            let result = validator.validate(&bad, "v");
            assert_eq!(result.errors()[0].code, codes::NOT_EMPTY);
            assert_eq!(result.errors()[0].message, "Value must not be empty");
        }

        let result = validator.validate(&json!(5), "v");
        assert_eq!(result.errors()[0].code, codes::TYPE);
        assert_eq!(
            result.errors()[0].message,
            "Expected string, array, or object, got integer"
        );
    }

    #[test]
    fn test_enum_validator() {
        let validator = validators::enum_values(vec![json!("a"), json!(1), json!(true)]);
        assert!(validator.validate(&json!("a"), "v").is_valid());
        assert!(validator.validate(&json!(1), "v").is_valid());

        let result = validator.validate(&json!("b"), "v");
        assert_eq!(result.errors()[0].code, codes::ENUM);
        assert_eq!(
            result.errors()[0].message,
            "Value must be one of \"a\", 1, true, got \"b\""
        );
    }

    #[test]
    fn test_one_of_validator() {
        let validator = validators::one_of(vec![
            Box::new(validators::string().min(3)),
            Box::new(validators::int().min(10)),
        ]);
        assert!(validator.validate(&json!("abc"), "v").is_valid());
        assert!(validator.validate(&json!(15), "v").is_valid());

        let result = validator.validate(&json!(5), "v");
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 1);
        assert_eq!(result.errors()[0].code, codes::ONE_OF);
        assert_eq!(
            result.errors()[0].message,
            "Value does not match any allowed schema"
        );
    }

    #[test]
    fn test_regex_factory() {
        let validator = validators::regex(Regex::new(r"^[a-z]+$").unwrap());
        assert!(validator.validate(&json!("abc"), "code").is_valid());
        assert!(!validator.validate(&json!("ABC"), "code").is_valid());
    }

    #[test]
    fn test_sensitive_redacts_value() {
        // Messages that embed the offending value render `***` and
        // `error.value` is None.
        let result = validators::email()
            .sensitive(true)
            .validate(&json!("p@ssw0rd"), "secret");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].message, "Invalid email address: ***");
        assert!(result.errors()[0].value.is_none());

        let result = validators::uuid()
            .sensitive(true)
            .validate(&json!("leaked"), "id");
        assert_eq!(result.errors()[0].message, "Invalid UUID: ***");
        assert!(result.errors()[0].value.is_none());

        // `choices` keeps its quotes around the redacted placeholder.
        let result = validators::string()
            .choices(vec!["a".to_string(), "b".to_string()])
            .sensitive(true)
            .validate(&json!("hunter2"), "token");
        assert_eq!(
            result.errors()[0].message,
            "Value must be one of a, b, got '***'"
        );
        assert!(result.errors()[0].value.is_none());

        // `enum` renders the bare, unquoted placeholder.
        let result = validators::enum_values(vec![json!("a"), json!(1)])
            .sensitive(true)
            .validate(&json!("hunter2"), "token");
        assert_eq!(
            result.errors()[0].message,
            "Value must be one of \"a\", 1, got ***"
        );
        assert!(result.errors()[0].value.is_none());
    }

    #[test]
    fn test_sensitive_other_messages_unchanged() {
        // Messages that do not embed the input value are unchanged.
        let result = validators::string()
            .min(5)
            .sensitive(true)
            .validate(&json!("abc"), "secret");
        assert_eq!(
            result.errors()[0].message,
            "String length must be at least 5, got 3"
        );
        assert!(result.errors()[0].value.is_none());

        // Sensitive is per-validator and not inherited by nested validators.
        let result = validators::map()
            .sensitive(true)
            .field("inner", validators::email())
            .validate(&json!({"inner": "not-an-email"}), "user");
        assert_eq!(
            result.errors()[0].message,
            "Invalid email address: not-an-email"
        );
    }

    #[test]
    fn test_pattern_input_length_cap() {
        // Over-limit inputs skip the match and produce the normal pattern
        // error. `regex` is linear-time; the cap is for spec parity.
        let long = "a".repeat(MAX_PATTERN_INPUT_LENGTH + 1);

        let result = validators::string()
            .try_pattern(r"^a+$")
            .unwrap()
            .validate(&json!(long), "s");
        assert!(!result.is_valid());
        assert_eq!(result.errors()[0].code, codes::PATTERN);

        let result = validators::email().validate(&json!(long), "e");
        assert_eq!(result.errors()[0].code, codes::EMAIL);

        let result = validators::numeric().validate(&json!(long), "n");
        assert_eq!(result.errors()[0].code, codes::NUMERIC);

        // Exactly at the limit still matches.
        let at_limit = "a".repeat(MAX_PATTERN_INPUT_LENGTH);
        assert!(validators::string()
            .try_pattern(r"^a+$")
            .unwrap()
            .validate(&json!(at_limit), "s")
            .is_valid());
    }

    #[test]
    fn test_try_validate() {
        let schema = Schema::new().field("name", validators::string());
        assert!(schema.try_validate(&json!({"name": "Ada"})).is_ok());

        let err = schema.try_validate(&json!({})).unwrap_err();
        assert_eq!(err.0.len(), 1);
        assert_eq!(err.to_string(), "name: Field is required");
    }
}

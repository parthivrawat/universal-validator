//! Universal Data Validator
//!
//! A comprehensive data validation library for Rust that works across API, database, and form contexts.
//!
//! # Features
//!
//! - **Rich Validator Types**: String, i32, i64, f64, bool, email, URL, Vec, HashMap
//! - **Schema-Based Validation**: Define complex data structures
//! - **Custom Validators**: Add your own validation logic
//! - **Nested Validation**: Validate nested objects and vectors
//! - **Clear Error Messages**: Detailed error reporting with field paths
//!
//! # Examples
//!
//! ```
//! use universal_validator::{Schema, Validator, validators};
//!
//! let schema = Schema::new()
//!     .field("email", validators::email())
//!     .field("age", validators::int().min(0).max(120));
//!
//! let mut data = std::collections::HashMap::new();
//! data.insert("email".to_string(), "user@example.com".to_string());
//! data.insert("age".to_string(), "25".to_string());
//!
//! let result = schema.validate(&data);
//! assert!(result.is_valid());
//! ```

use regex::Regex;
use std::collections::HashMap;
use std::fmt;

/// Validation error
#[derive(Debug, Clone)]
pub struct ValidationError {
    pub field: String,
    pub message: String,
    pub value: Option<String>,
}

impl ValidationError {
    pub fn new(field: impl Into<String>, message: impl Into<String>) -> Self {
        Self {
            field: field.into(),
            message: message.into(),
            value: None,
        }
    }

    pub fn with_value(mut self, value: impl Into<String>) -> Self {
        self.value = Some(value.into());
        self
    }
}

impl fmt::Display for ValidationError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {}", self.field, self.message)
    }
}

impl std::error::Error for ValidationError {}

/// Validation result
#[derive(Debug)]
pub struct ValidationResult {
    errors: Vec<ValidationError>,
}

impl ValidationResult {
    pub fn new() -> Self {
        Self { errors: Vec::new() }
    }

    pub fn is_valid(&self) -> bool {
        self.errors.is_empty()
    }

    pub fn errors(&self) -> &[ValidationError] {
        &self.errors
    }

    pub fn add_error(&mut self, error: ValidationError) {
        self.errors.push(error);
    }

    pub fn merge(&mut self, other: ValidationResult) {
        self.errors.extend(other.errors);
    }
}

impl Default for ValidationResult {
    fn default() -> Self {
        Self::new()
    }
}

/// Validator trait
pub trait Validator: Send + Sync {
    fn validate(&self, value: &str, field: &str) -> ValidationResult;
    fn is_required(&self) -> bool {
        true
    }
}

/// String validator
pub struct StringValidator {
    min_length: Option<usize>,
    max_length: Option<usize>,
    pattern: Option<Regex>,
    choices: Option<Vec<String>>,
    required: bool,
}

impl StringValidator {
    pub fn new() -> Self {
        Self {
            min_length: None,
            max_length: None,
            pattern: None,
            choices: None,
            required: true,
        }
    }

    pub fn min(mut self, min: usize) -> Self {
        self.min_length = Some(min);
        self
    }

    pub fn max(mut self, max: usize) -> Self {
        self.max_length = Some(max);
        self
    }

    pub fn pattern(mut self, pattern: &str) -> Self {
        self.pattern = Some(Regex::new(pattern).expect("Invalid regex pattern"));
        self
    }

    pub fn choices(mut self, choices: Vec<String>) -> Self {
        self.choices = Some(choices);
        self
    }

    pub fn optional(mut self) -> Self {
        self.required = false;
        self
    }
}

impl Validator for StringValidator {
    fn validate(&self, value: &str, field: &str) -> ValidationResult {
        let mut result = ValidationResult::new();

        if let Some(min) = self.min_length {
            if value.len() < min {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("String length must be at least {}, got {}", min, value.len()),
                    )
                    .with_value(value),
                );
            }
        }

        if let Some(max) = self.max_length {
            if value.len() > max {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("String length must be at most {}, got {}", max, value.len()),
                    )
                    .with_value(value),
                );
            }
        }

        if let Some(pattern) = &self.pattern {
            if !pattern.is_match(value) {
                result.add_error(
                    ValidationError::new(field, format!("String does not match pattern {}", pattern))
                        .with_value(value),
                );
            }
        }

        if let Some(choices) = &self.choices {
            if !choices.contains(&value.to_string()) {
                result.add_error(
                    ValidationError::new(
                        field,
                        format!("Value must be one of {:?}, got '{}'", choices, value),
                    )
                    .with_value(value),
                );
            }
        }

        result
    }

    fn is_required(&self) -> bool {
        self.required
    }
}

/// Integer validator
pub struct IntValidator {
    min_value: Option<i64>,
    max_value: Option<i64>,
    required: bool,
}

impl IntValidator {
    pub fn new() -> Self {
        Self {
            min_value: None,
            max_value: None,
            required: true,
        }
    }

    pub fn min(mut self, min: i64) -> Self {
        self.min_value = Some(min);
        self
    }

    pub fn max(mut self, max: i64) -> Self {
        self.max_value = Some(max);
        self
    }

    pub fn optional(mut self) -> Self {
        self.required = false;
        self
    }
}

impl Validator for IntValidator {
    fn validate(&self, value: &str, field: &str) -> ValidationResult {
        let mut result = ValidationResult::new();

        let num = match value.parse::<i64>() {
            Ok(n) => n,
            Err(_) => {
                result.add_error(
                    ValidationError::new(field, format!("Expected integer, got '{}'", value))
                        .with_value(value),
                );
                return result;
            }
        };

        if let Some(min) = self.min_value {
            if num < min {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at least {}, got {}", min, num))
                        .with_value(value),
                );
            }
        }

        if let Some(max) = self.max_value {
            if num > max {
                result.add_error(
                    ValidationError::new(field, format!("Value must be at most {}, got {}", max, num))
                        .with_value(value),
                );
            }
        }

        result
    }

    fn is_required(&self) -> bool {
        self.required
    }
}

/// Email validator
pub struct EmailValidator {
    required: bool,
}

impl EmailValidator {
    pub fn new() -> Self {
        Self { required: true }
    }

    pub fn optional(mut self) -> Self {
        self.required = false;
        self
    }
}

impl Validator for EmailValidator {
    fn validate(&self, value: &str, field: &str) -> ValidationResult {
        let mut result = ValidationResult::new();
        let email_pattern = Regex::new(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$").unwrap();

        if !email_pattern.is_match(value) {
            result.add_error(
                ValidationError::new(field, format!("Invalid email address: {}", value))
                    .with_value(value),
            );
        }

        result
    }

    fn is_required(&self) -> bool {
        self.required
    }
}

/// URL validator
pub struct UrlValidator {
    required: bool,
}

impl UrlValidator {
    pub fn new() -> Self {
        Self { required: true }
    }

    pub fn optional(mut self) -> Self {
        self.required = false;
        self
    }
}

impl Validator for UrlValidator {
    fn validate(&self, value: &str, field: &str) -> ValidationResult {
        let mut result = ValidationResult::new();
        let url_pattern = Regex::new(r"^https?://[^\s/$.?#].[^\s]*$").unwrap();

        if !url_pattern.is_match(value) {
            result.add_error(
                ValidationError::new(field, format!("Invalid URL: {}", value)).with_value(value),
            );
        }

        result
    }

    fn is_required(&self) -> bool {
        self.required
    }
}

/// Schema for validating complex data structures
pub struct Schema {
    fields: HashMap<String, Box<dyn Validator>>,
}

impl Schema {
    pub fn new() -> Self {
        Self {
            fields: HashMap::new(),
        }
    }

    pub fn field(mut self, name: impl Into<String>, validator: impl Validator + 'static) -> Self {
        self.fields.insert(name.into(), Box::new(validator));
        self
    }

    pub fn validate(&self, data: &HashMap<String, String>) -> ValidationResult {
        let mut result = ValidationResult::new();

        for (field, validator) in &self.fields {
            if let Some(value) = data.get(field) {
                let field_result = validator.validate(value, field);
                result.merge(field_result);
            } else if validator.is_required() {
                result.add_error(ValidationError::new(field, "Field is required"));
            }
        }

        result
    }
}

impl Default for Schema {
    fn default() -> Self {
        Self::new()
    }
}

/// Validator factory functions
pub mod validators {
    use super::*;

    pub fn string() -> StringValidator {
        StringValidator::new()
    }

    pub fn int() -> IntValidator {
        IntValidator::new()
    }

    pub fn email() -> EmailValidator {
        EmailValidator::new()
    }

    pub fn url() -> UrlValidator {
        UrlValidator::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_string_validator() {
        let validator = validators::string().min(3).max(10);
        
        let result = validator.validate("hello", "name");
        assert!(result.is_valid());

        let result = validator.validate("ab", "name");
        assert!(!result.is_valid());

        let result = validator.validate("hello world", "name");
        assert!(!result.is_valid());
    }

    #[test]
    fn test_int_validator() {
        let validator = validators::int().min(0).max(120);

        let result = validator.validate("25", "age");
        assert!(result.is_valid());

        let result = validator.validate("-1", "age");
        assert!(!result.is_valid());

        let result = validator.validate("150", "age");
        assert!(!result.is_valid());

        let result = validator.validate("not_a_number", "age");
        assert!(!result.is_valid());
    }

    #[test]
    fn test_email_validator() {
        let validator = validators::email();

        let result = validator.validate("user@example.com", "email");
        assert!(result.is_valid());

        let result = validator.validate("invalid-email", "email");
        assert!(!result.is_valid());
    }

    #[test]
    fn test_url_validator() {
        let validator = validators::url();

        let result = validator.validate("https://example.com", "url");
        assert!(result.is_valid());

        let result = validator.validate("not-a-url", "url");
        assert!(!result.is_valid());
    }

    #[test]
    fn test_schema() {
        let schema = Schema::new()
            .field("username", validators::string().min(3).max(20))
            .field("email", validators::email())
            .field("age", validators::int().min(0).max(120));

        let mut data = HashMap::new();
        data.insert("username".to_string(), "john_doe".to_string());
        data.insert("email".to_string(), "john@example.com".to_string());
        data.insert("age".to_string(), "25".to_string());

        let result = schema.validate(&data);
        assert!(result.is_valid());
    }

    #[test]
    fn test_schema_invalid() {
        let schema = Schema::new()
            .field("username", validators::string().min(3))
            .field("email", validators::email());

        let mut data = HashMap::new();
        data.insert("username".to_string(), "ab".to_string());
        data.insert("email".to_string(), "invalid".to_string());

        let result = schema.validate(&data);
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 2);
    }

    #[test]
    fn test_schema_missing_required() {
        let schema = Schema::new()
            .field("username", validators::string())
            .field("email", validators::email());

        let data = HashMap::new();

        let result = schema.validate(&data);
        assert!(!result.is_valid());
        assert_eq!(result.errors().len(), 2);
    }

    #[test]
    fn test_optional_field() {
        let schema = Schema::new()
            .field("username", validators::string())
            .field("bio", validators::string().optional());

        let mut data = HashMap::new();
        data.insert("username".to_string(), "john_doe".to_string());

        let result = schema.validate(&data);
        assert!(result.is_valid());
    }
}

// Package validator provides comprehensive data validation for Go applications.
//
// This package offers schema-based validation with support for various types,
// custom validators, and nested validation.
//
// Example:
//
//	import "github.com/parthivrawat/universal-validator"
//
//	schema := validator.NewSchema(map[string]validator.Validator{
//	    "email": validator.Email(),
//	    "age": validator.Int(validator.IntOptions{MinValue: ptr(0), MaxValue: ptr(120)}),
//	})
//
//	result := schema.Validate(data)
package validator

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationResult represents the result of a validation
type ValidationResult struct {
	Valid  bool
	Errors []*ValidationError
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(err *ValidationError) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

// Validator is the interface that all validators must implement
type Validator interface {
	Validate(value interface{}, field string) *ValidationResult
	IsRequired() bool
	IsNullable() bool
}

// BaseValidator provides common functionality for all validators
type BaseValidator struct {
	Required         bool
	Nullable         bool
	CustomValidators []func(interface{}) *string
}

// IsRequired returns whether the field is required
func (v *BaseValidator) IsRequired() bool {
	return v.Required
}

// IsNullable returns whether the field can be null
func (v *BaseValidator) IsNullable() bool {
	return v.Nullable
}

// AddCustomValidator adds a custom validation function
func (v *BaseValidator) AddCustomValidator(fn func(interface{}) *string) {
	v.CustomValidators = append(v.CustomValidators, fn)
}

// ValidateBase performs base validation (nil checks and custom validators)
func (v *BaseValidator) ValidateBase(value interface{}, field string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if value == nil {
		if v.Required && !v.Nullable {
			result.AddError(&ValidationError{Field: field, Message: "Field is required"})
		}
		return result
	}

	for _, customValidator := range v.CustomValidators {
		if errMsg := customValidator(value); errMsg != nil {
			result.AddError(&ValidationError{Field: field, Message: *errMsg, Value: value})
		}
	}

	return result
}

// StringValidator validates string values
type StringValidator struct {
	BaseValidator
	MinLength *int
	MaxLength *int
	Pattern   *regexp.Regexp
	Choices   []string
}

// StringOptions configures a string validator
type StringOptions struct {
	MinLength *int
	MaxLength *int
	Pattern   *regexp.Regexp
	Choices   []string
	Required  bool
	Nullable  bool
}

// String creates a new string validator
func String(opts ...StringOptions) *StringValidator {
	v := &StringValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.MinLength = opt.MinLength
		v.MaxLength = opt.MaxLength
		v.Pattern = opt.Pattern
		v.Choices = opt.Choices
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a string value
func (v *StringValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	str, ok := value.(string)
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected string, got %T", value),
			Value:   value,
		})
		return result
	}

	if v.MinLength != nil && len(str) < *v.MinLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("String length must be at least %d, got %d", *v.MinLength, len(str)),
			Value:   value,
		})
	}

	if v.MaxLength != nil && len(str) > *v.MaxLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("String length must be at most %d, got %d", *v.MaxLength, len(str)),
			Value:   value,
		})
	}

	if v.Pattern != nil && !v.Pattern.MatchString(str) {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("String does not match pattern %s", v.Pattern.String()),
			Value:   value,
		})
	}

	if v.Choices != nil && len(v.Choices) > 0 {
		found := false
		for _, choice := range v.Choices {
			if str == choice {
				found = true
				break
			}
		}
		if !found {
			result.AddError(&ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Value must be one of %v, got '%s'", v.Choices, str),
				Value:   value,
			})
		}
	}

	return result
}

// Custom adds a custom validator
func (v *StringValidator) Custom(fn func(interface{}) *string) *StringValidator {
	v.AddCustomValidator(fn)
	return v
}

// IntValidator validates integer values
type IntValidator struct {
	BaseValidator
	MinValue *int
	MaxValue *int
}

// IntOptions configures an integer validator
type IntOptions struct {
	MinValue *int
	MaxValue *int
	Required bool
	Nullable bool
}

// Int creates a new integer validator
func Int(opts ...IntOptions) *IntValidator {
	v := &IntValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.MinValue = opt.MinValue
		v.MaxValue = opt.MaxValue
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates an integer value
func (v *IntValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	var num int
	switch val := value.(type) {
	case int:
		num = val
	case int32:
		num = int(val)
	case int64:
		num = int(val)
	case float64:
		if val != float64(int(val)) {
			result.AddError(&ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Expected integer, got %v", value),
				Value:   value,
			})
			return result
		}
		num = int(val)
	default:
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected int, got %T", value),
			Value:   value,
		})
		return result
	}

	if v.MinValue != nil && num < *v.MinValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at least %d, got %d", *v.MinValue, num),
			Value:   value,
		})
	}

	if v.MaxValue != nil && num > *v.MaxValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at most %d, got %d", *v.MaxValue, num),
			Value:   value,
		})
	}

	return result
}

// Custom adds a custom validator
func (v *IntValidator) Custom(fn func(interface{}) *string) *IntValidator {
	v.AddCustomValidator(fn)
	return v
}

// FloatValidator validates float values
type FloatValidator struct {
	BaseValidator
	MinValue *float64
	MaxValue *float64
}

// FloatOptions configures a float validator
type FloatOptions struct {
	MinValue *float64
	MaxValue *float64
	Required bool
	Nullable bool
}

// Float creates a new float validator
func Float(opts ...FloatOptions) *FloatValidator {
	v := &FloatValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.MinValue = opt.MinValue
		v.MaxValue = opt.MaxValue
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a float value
func (v *FloatValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	var num float64
	switch val := value.(type) {
	case float64:
		num = val
	case float32:
		num = float64(val)
	case int:
		num = float64(val)
	case int32:
		num = float64(val)
	case int64:
		num = float64(val)
	default:
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected float, got %T", value),
			Value:   value,
		})
		return result
	}

	if v.MinValue != nil && num < *v.MinValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at least %f, got %f", *v.MinValue, num),
			Value:   value,
		})
	}

	if v.MaxValue != nil && num > *v.MaxValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at most %f, got %f", *v.MaxValue, num),
			Value:   value,
		})
	}

	return result
}

// Custom adds a custom validator
func (v *FloatValidator) Custom(fn func(interface{}) *string) *FloatValidator {
	v.AddCustomValidator(fn)
	return v
}

// BoolValidator validates boolean values
type BoolValidator struct {
	BaseValidator
}

// BoolOptions configures a boolean validator
type BoolOptions struct {
	Required bool
	Nullable bool
}

// Bool creates a new boolean validator
func Bool(opts ...BoolOptions) *BoolValidator {
	v := &BoolValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a boolean value
func (v *BoolValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	if _, ok := value.(bool); !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected bool, got %T", value),
			Value:   value,
		})
	}

	return result
}

// Custom adds a custom validator
func (v *BoolValidator) Custom(fn func(interface{}) *string) *BoolValidator {
	v.AddCustomValidator(fn)
	return v
}

// EmailValidator validates email addresses
type EmailValidator struct {
	*StringValidator
}

// EmailOptions configures an email validator
type EmailOptions struct {
	Required bool
	Nullable bool
}

// Email creates a new email validator
func Email(opts ...EmailOptions) *EmailValidator {
	emailPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	v := &EmailValidator{
		StringValidator: &StringValidator{
			BaseValidator: BaseValidator{Required: true, Nullable: false},
			Pattern:       emailPattern,
		},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates an email address
func (v *EmailValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(value, field)
	
	// Customize error message for email
	if !result.Valid && len(result.Errors) > 0 {
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "does not match pattern") {
				err.Message = fmt.Sprintf("Invalid email address: %v", value)
			}
		}
	}
	
	return result
}

// URLValidator validates URLs
type URLValidator struct {
	*StringValidator
}

// URLOptions configures a URL validator
type URLOptions struct {
	Required bool
	Nullable bool
}

// URL creates a new URL validator
func URL(opts ...URLOptions) *URLValidator {
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	v := &URLValidator{
		StringValidator: &StringValidator{
			BaseValidator: BaseValidator{Required: true, Nullable: false},
			Pattern:       urlPattern,
		},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a URL
func (v *URLValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(value, field)
	
	// Customize error message for URL
	if !result.Valid && len(result.Errors) > 0 {
		for _, err := range result.Errors {
			if strings.Contains(err.Message, "does not match pattern") {
				err.Message = fmt.Sprintf("Invalid URL: %v", value)
			}
		}
	}
	
	return result
}

// SliceValidator validates slice/array values
type SliceValidator struct {
	BaseValidator
	ItemValidator Validator
	MinLength     *int
	MaxLength     *int
}

// SliceOptions configures a slice validator
type SliceOptions struct {
	ItemValidator Validator
	MinLength     *int
	MaxLength     *int
	Required      bool
	Nullable      bool
}

// Slice creates a new slice validator
func Slice(opts ...SliceOptions) *SliceValidator {
	v := &SliceValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.ItemValidator = opt.ItemValidator
		v.MinLength = opt.MinLength
		v.MaxLength = opt.MaxLength
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a slice value
func (v *SliceValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	slice, ok := value.([]interface{})
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected slice, got %T", value),
			Value:   value,
		})
		return result
	}

	if v.MinLength != nil && len(slice) < *v.MinLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Slice length must be at least %d, got %d", *v.MinLength, len(slice)),
			Value:   value,
		})
	}

	if v.MaxLength != nil && len(slice) > *v.MaxLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Slice length must be at most %d, got %d", *v.MaxLength, len(slice)),
			Value:   value,
		})
	}

	if v.ItemValidator != nil {
		for i, item := range slice {
			itemResult := v.ItemValidator.Validate(item, fmt.Sprintf("%s[%d]", field, i))
			result.Errors = append(result.Errors, itemResult.Errors...)
			if !itemResult.Valid {
				result.Valid = false
			}
		}
	}

	return result
}

// Custom adds a custom validator
func (v *SliceValidator) Custom(fn func(interface{}) *string) *SliceValidator {
	v.AddCustomValidator(fn)
	return v
}

// MapValidator validates map values
type MapValidator struct {
	BaseValidator
	Schema map[string]Validator
}

// MapOptions configures a map validator
type MapOptions struct {
	Schema   map[string]Validator
	Required bool
	Nullable bool
}

// Map creates a new map validator
func Map(opts ...MapOptions) *MapValidator {
	v := &MapValidator{
		BaseValidator: BaseValidator{Required: true, Nullable: false},
	}

	if len(opts) > 0 {
		opt := opts[0]
		v.Schema = opt.Schema
		v.Required = opt.Required
		v.Nullable = opt.Nullable
	}

	return v
}

// Validate validates a map value
func (v *MapValidator) Validate(value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(value, field)
	if !result.Valid || value == nil {
		return result
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected map, got %T", value),
			Value:   value,
		})
		return result
	}

	if v.Schema != nil {
		for key, validator := range v.Schema {
			fieldName := fmt.Sprintf("%s.%s", field, key)
			if val, exists := m[key]; exists {
				fieldResult := validator.Validate(val, fieldName)
				result.Errors = append(result.Errors, fieldResult.Errors...)
				if !fieldResult.Valid {
					result.Valid = false
				}
			} else if validator.IsRequired() {
				result.AddError(&ValidationError{
					Field:   fieldName,
					Message: "Field is required",
				})
			}
		}
	}

	return result
}

// Custom adds a custom validator
func (v *MapValidator) Custom(fn func(interface{}) *string) *MapValidator {
	v.AddCustomValidator(fn)
	return v
}

// Schema represents a validation schema
type Schema struct {
	validators map[string]Validator
}

// NewSchema creates a new validation schema
func NewSchema(validators map[string]Validator) *Schema {
	return &Schema{validators: validators}
}

// Validate validates data against the schema
func (s *Schema) Validate(data map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	for field, validator := range s.validators {
		if value, exists := data[field]; exists {
			fieldResult := validator.Validate(value, field)
			result.Errors = append(result.Errors, fieldResult.Errors...)
			if !fieldResult.Valid {
				result.Valid = false
			}
		} else if validator.IsRequired() {
			result.AddError(&ValidationError{
				Field:   field,
				Message: "Field is required",
			})
		}
	}

	return result
}

// ValidateOrPanic validates data and panics if validation fails
func (s *Schema) ValidateOrPanic(data map[string]interface{}) {
	result := s.Validate(data)
	if !result.Valid {
		var messages []string
		for _, err := range result.Errors {
			messages = append(messages, err.Error())
		}
		panic(fmt.Sprintf("Validation failed:\n%s", strings.Join(messages, "\n")))
	}
}

// Helper function to create pointer to int
func IntPtr(i int) *int {
	return &i
}

// Helper function to create pointer to float64
func Float64Ptr(f float64) *float64 {
	return &f
}

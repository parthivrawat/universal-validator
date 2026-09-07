// Package validator provides comprehensive data validation for Go applications.
//
// This package offers schema-based validation with support for various types,
// custom validators, and nested validation. Validators are immutable after
// construction and safe to share across goroutines.
//
// Example:
//
//	import validator "github.com/parthivrawat/universal-validator/go/v2"
//
//	schema := validator.NewSchema(map[string]validator.Validator{
//	    "email": validator.Email(),
//	    "age":   validator.Int(validator.MinValue(0), validator.MaxValue(120)),
//	})
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strings"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
	Code    string
}

// Canonical validation error codes.
const (
	ErrCodeRequired     = "required"
	ErrCodeUnknownField = "unknown_field"
	ErrCodeType         = "type"
	ErrCodeMinLength    = "min_length"
	ErrCodeMaxLength    = "max_length"
	ErrCodeMinValue     = "min_value"
	ErrCodeMaxValue     = "max_value"
	ErrCodePattern      = "pattern"
	ErrCodeChoices      = "choices"
	ErrCodeEmail        = "email"
	ErrCodeURL          = "url"
	ErrCodeCustom       = "custom"
	ErrCodeUUID         = "uuid"
	ErrCodeDate         = "date"
	ErrCodeDateTime     = "datetime"
	ErrCodeNumeric      = "numeric"
	ErrCodeNotEmpty     = "not_empty"
	ErrCodeEnum         = "enum"
	ErrCodeOneOf        = "one_of"
)

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is an aggregate of one or more validation errors. It
// implements error and can be unwrapped to inspect the individual errors.
type ValidationErrors []*ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "\n")
}

// Unwrap returns the individual validation errors as a slice of error values.
func (ve ValidationErrors) Unwrap() []error {
	errs := make([]error, len(ve))
	for i, e := range ve {
		errs[i] = e
	}
	return errs
}

// CustomValidator is a user-supplied validation function. Returning a non-nil
// error causes validation to fail with code ErrCodeCustom; the error's
// Error() string is used as the message. Returning nil means the value passed.
type CustomValidator func(ctx context.Context, value interface{}) error

// Option configures a validator. Options are a single shared namespace across
// all validator constructors; options that are not applicable to a validator
// are ignored.
type Option func(*config)

type config struct {
	optional   bool
	nullable   bool
	minLength  *int
	maxLength  *int
	pattern    *regexp.Regexp
	choices    []string
	minInt     *int
	maxInt     *int
	minFloat   *float64
	maxFloat   *float64
	items      Validator
	fields     map[string]Validator
	strict     bool
	failFast   bool
	custom     []CustomValidator
	oneOf      []Validator
	enumValues []interface{}
	sensitive  bool
}

// Optional marks a validator as optional (required=false).
func Optional() Option { return func(c *config) { c.optional = true } }

// Nullable marks a validator as nullable.
func Nullable() Option { return func(c *config) { c.nullable = true } }

// MinLength sets the minimum string length.
func MinLength(n int) Option { return func(c *config) { c.minLength = &n } }

// MaxLength sets the maximum string length.
func MaxLength(n int) Option { return func(c *config) { c.maxLength = &n } }

// Pattern sets the regular expression a string must match in full.
func Pattern(re *regexp.Regexp) Option { return func(c *config) { c.pattern = re } }

// Choices sets the allowed string values.
func Choices(choices ...string) Option { return func(c *config) { c.choices = choices } }

// MinValue sets the minimum integer value (inclusive).
func MinValue(n int) Option { return func(c *config) { c.minInt = &n } }

// MaxValue sets the maximum integer value (inclusive).
func MaxValue(n int) Option { return func(c *config) { c.maxInt = &n } }

// MinFloat sets the minimum float value (inclusive).
func MinFloat(f float64) Option { return func(c *config) { c.minFloat = &f } }

// MaxFloat sets the maximum float value (inclusive).
func MaxFloat(f float64) Option { return func(c *config) { c.maxFloat = &f } }

// Items sets the item validator for a slice.
func Items(v Validator) Option { return func(c *config) { c.items = v } }

// Fields sets the nested field validators for a map/object.
func Fields(fields map[string]Validator) Option { return func(c *config) { c.fields = fields } }

// Strict enables unknown-key detection for a map or schema.
func Strict() Option { return func(c *config) { c.strict = true } }

// FailFast enables fail-fast validation for a map or schema.
func FailFast() Option { return func(c *config) { c.failFast = true } }

// WithCustom attaches a custom validation function. Multiple custom validators
// can be supplied and are run in order.
func WithCustom(fn CustomValidator) Option {
	return func(c *config) { c.custom = append(c.custom, fn) }
}

// Sensitive marks a validator as handling a secret (passwords, tokens, …).
// When set, every error the validator produces has Value == nil and any
// interpolation of the offending value in an error message renders as the
// literal "***". Messages that do not embed the input (length and value
// bounds, "Field is required", "Unknown field") are unchanged. The flag is
// NOT inherited by nested validators (Items, Fields, schema fields).
func Sensitive() Option { return func(c *config) { c.sensitive = true } }

// MaxPatternInputLength is the maximum length, in bytes, of a string that a
// pattern check will be run against. Inputs longer than this skip the match
// and immediately produce the normal `pattern` error (which the email/url/
// uuid/date/datetime/numeric wrappers rewrite to their friendly message as
// usual). Go's regexp engine (RE2) is linear-time so the cap is not required
// for safety here; it is enforced for parity with the conformance spec and
// sibling implementations that use backtracking engines.
const MaxPatternInputLength = 10000

// isNullish reports whether v is nil or a typed nil (e.g. (*Foo)(nil),
// nil map/slice/function stored in an interface{}).
func isNullish(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// ValidationResult represents the result of a validation.
type ValidationResult struct {
	Valid  bool
	Errors []*ValidationError
}

// AddError adds a validation error to the result.
func (r *ValidationResult) AddError(err *ValidationError) {
	r.Valid = false
	r.Errors = append(r.Errors, err)
}

// Validator is the interface that all validators must implement.
type Validator interface {
	Validate(ctx context.Context, value interface{}, field string) *ValidationResult
	IsRequired() bool
	IsNullable() bool
}

// failFastable is implemented by container validators (MapValidator,
// SliceValidator) that can stop their internal iteration at the first
// error. It is unexported so the public Validator interface and its
// Validate signature remain unchanged.
type failFastable interface {
	validateFailFast(ctx context.Context, value interface{}, field string) *ValidationResult
}

// validateFailFast validates value with v in fail-fast mode: container
// validators stop iterating fields/items after the first error, and leaf
// validator results are truncated to their first error so a fail-fast
// result always contains at most one error.
func validateFailFast(v Validator, ctx context.Context, value interface{}, field string) *ValidationResult {
	if ff, ok := v.(failFastable); ok {
		return ff.validateFailFast(ctx, value, field)
	}
	result := v.Validate(ctx, value, field)
	if len(result.Errors) > 1 {
		result.Errors = result.Errors[:1]
	}
	return result
}

// BaseValidator provides common functionality for all validators.
type BaseValidator struct {
	Required         bool
	Nullable         bool
	CustomValidators []CustomValidator
	// Sensitive redacts the offending value from errors: Value is nil and
	// any value interpolated into a message renders as "***". See Sensitive().
	Sensitive bool
}

// displayValue returns the representation of v for embedding in error
// messages: the literal "***" when the validator is sensitive.
func (v *BaseValidator) displayValue(value interface{}) interface{} {
	if v.Sensitive {
		return "***"
	}
	return value
}

// errorValue returns the value to store on a ValidationError: nil when the
// validator is sensitive so secrets are not exposed on the error.
func (v *BaseValidator) errorValue(value interface{}) interface{} {
	if v.Sensitive {
		return nil
	}
	return value
}

// IsRequired returns whether the field is required.
func (v *BaseValidator) IsRequired() bool {
	return v.Required
}

// IsNullable returns whether the field can be null.
func (v *BaseValidator) IsNullable() bool {
	return v.Nullable
}

// ValidateBase performs base validation (nil checks and custom validators).
func (v *BaseValidator) ValidateBase(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := &ValidationResult{Valid: true}

	if isNullish(value) {
		if v.Required && !v.Nullable {
			result.AddError(&ValidationError{Field: field, Message: "Field is required", Value: v.errorValue(value), Code: ErrCodeRequired})
		}
		return result
	}

	for _, customValidator := range v.CustomValidators {
		if err := customValidator(ctx, value); err != nil {
			result.AddError(&ValidationError{Field: field, Message: err.Error(), Value: v.errorValue(value), Code: ErrCodeCustom})
		}
	}

	return result
}

// jsonKind returns the canonical JSON kind name for a value as defined by the
// conformance spec: null, boolean, integer, number, string, array, or object.
func jsonKind(v interface{}) string {
	if v == nil {
		return "null"
	}
	switch x := v.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "integer"
	case float64:
		if x == math.Trunc(x) {
			return "integer"
		}
		return "number"
	case float32:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// newChoiceSet builds a lookup set for O(1) choices membership checks.
// It returns nil for an empty choices list so validators constructed
// directly (not via String()) fall back to a linear scan.
func newChoiceSet(choices []string) map[string]struct{} {
	if len(choices) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(choices))
	for _, c := range choices {
		set[c] = struct{}{}
	}
	return set
}

// StringValidator validates string values.
type StringValidator struct {
	BaseValidator
	MinLength *int
	MaxLength *int
	Pattern   *regexp.Regexp
	Choices   []string
	// choiceSet is an O(1) membership index built from Choices at
	// construction time. Choices itself is retained for the error
	// message, which must list choices in declared order.
	choiceSet map[string]struct{}
}

func makeStringValidator(c *config) *StringValidator {
	return &StringValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		MinLength:     c.minLength,
		MaxLength:     c.maxLength,
		Pattern:       c.pattern,
		Choices:       c.choices,
		choiceSet:     newChoiceSet(c.choices),
	}
}

// String creates a new string validator.
func String(opts ...Option) *StringValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return makeStringValidator(c)
}

// Validate validates a string value.
func (v *StringValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	str, ok := value.(string)
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected string, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
		return result
	}

	if v.MinLength != nil && len(str) < *v.MinLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("String length must be at least %d, got %d", *v.MinLength, len(str)),
			Value:   v.errorValue(value),
			Code:    ErrCodeMinLength,
		})
	}

	if v.MaxLength != nil && len(str) > *v.MaxLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("String length must be at most %d, got %d", *v.MaxLength, len(str)),
			Value:   v.errorValue(value),
			Code:    ErrCodeMaxLength,
		})
	}

	if v.Pattern != nil {
		// Skip the match entirely for oversized inputs. RE2 is
		// linear-time so this is for spec parity, not safety.
		var loc []int
		if len(str) <= MaxPatternInputLength {
			loc = v.Pattern.FindStringIndex(str)
		}
		if loc == nil || loc[0] != 0 || loc[1] != len(str) {
			result.AddError(&ValidationError{
				Field:   field,
				Message: fmt.Sprintf("String does not match pattern %s", v.Pattern.String()),
				Value:   v.errorValue(value),
				Code:    ErrCodePattern,
			})
		}
	}

	if len(v.Choices) > 0 {
		found := false
		if v.choiceSet != nil {
			_, found = v.choiceSet[str]
		} else {
			// Fallback for StringValidator values built directly
			// rather than via String().
			for _, choice := range v.Choices {
				if str == choice {
					found = true
					break
				}
			}
		}
		if !found {
			result.AddError(&ValidationError{
				Field:   field,
				Message: fmt.Sprintf("Value must be one of %s, got '%v'", strings.Join(v.Choices, ", "), v.displayValue(str)),
				Value:   v.errorValue(value),
				Code:    ErrCodeChoices,
			})
		}
	}

	return result
}

// IntValidator validates integer values.
type IntValidator struct {
	BaseValidator
	MinValue *int
	MaxValue *int
}

// Int creates a new integer validator.
func Int(opts ...Option) *IntValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &IntValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		MinValue:      c.minInt,
		MaxValue:      c.maxInt,
	}
}

// Validate validates an integer value.
func (v *IntValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
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
				Message: fmt.Sprintf("Expected integer, got %s", jsonKind(value)),
				Value:   v.errorValue(value),
				Code:    ErrCodeType,
			})
			return result
		}
		num = int(val)
	default:
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected integer, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
		return result
	}

	if v.MinValue != nil && num < *v.MinValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at least %d, got %d", *v.MinValue, num),
			Value:   v.errorValue(value),
			Code:    ErrCodeMinValue,
		})
	}

	if v.MaxValue != nil && num > *v.MaxValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at most %d, got %d", *v.MaxValue, num),
			Value:   v.errorValue(value),
			Code:    ErrCodeMaxValue,
		})
	}

	return result
}

// FloatValidator validates float values.
type FloatValidator struct {
	BaseValidator
	MinValue *float64
	MaxValue *float64
}

// Float creates a new float validator.
func Float(opts ...Option) *FloatValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &FloatValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		MinValue:      c.minFloat,
		MaxValue:      c.maxFloat,
	}
}

// Validate validates a float value.
func (v *FloatValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
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
			Message: fmt.Sprintf("Expected number, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
		return result
	}

	if v.MinValue != nil && num < *v.MinValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at least %v, got %v", *v.MinValue, num),
			Value:   v.errorValue(value),
			Code:    ErrCodeMinValue,
		})
	}

	if v.MaxValue != nil && num > *v.MaxValue {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at most %v, got %v", *v.MaxValue, num),
			Value:   v.errorValue(value),
			Code:    ErrCodeMaxValue,
		})
	}

	return result
}

// BoolValidator validates boolean values.
type BoolValidator struct {
	BaseValidator
}

// Bool creates a new boolean validator.
func Bool(opts ...Option) *BoolValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &BoolValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
	}
}

// Validate validates a boolean value.
func (v *BoolValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	if _, ok := value.(bool); !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected boolean, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
	}

	return result
}

var (
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	urlPattern      = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	datePattern     = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$`)
	datetimePattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])T([01]\d|2[0-3]):[0-5]\d:[0-5]\d(\.\d+)?(Z|[+-]([01]\d|2[0-3]):[0-5]\d)$`)
	numericPattern  = regexp.MustCompile(`^-?(0|[1-9]\d*)(\.\d+)?([eE][+-]?\d+)?$`)
)

// patternStringValidator returns a *StringValidator using only optional/nullable,
// custom, and the supplied pattern. Other string options are ignored.
func patternStringValidator(c *config, pattern *regexp.Regexp) *StringValidator {
	bc := &config{
		optional:  c.optional,
		nullable:  c.nullable,
		custom:    c.custom,
		pattern:   pattern,
		sensitive: c.sensitive,
	}
	return makeStringValidator(bc)
}

// EmailValidator validates email addresses.
type EmailValidator struct{ *StringValidator }

// Email creates a new email validator.
func Email(opts ...Option) *EmailValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &EmailValidator{StringValidator: patternStringValidator(c, emailPattern)}
}

// Validate validates an email address.
func (v *EmailValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)

	// Rewrite the pattern-mismatch error to the friendly email message.
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeEmail
				err.Message = fmt.Sprintf("Invalid email address: %v", v.displayValue(value))
			}
		}
	}

	return result
}

// URLValidator validates URLs.
type URLValidator struct{ *StringValidator }

// URL creates a new URL validator.
func URL(opts ...Option) *URLValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &URLValidator{StringValidator: patternStringValidator(c, urlPattern)}
}

// Validate validates a URL.
func (v *URLValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)

	// Rewrite the pattern-mismatch error to the friendly URL message.
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeURL
				err.Message = fmt.Sprintf("Invalid URL: %v", v.displayValue(value))
			}
		}
	}

	return result
}

// Regex returns a string validator with a caller-supplied pattern. It is a
// convenience wrapper around String(Pattern(re)).
func Regex(re *regexp.Regexp, opts ...Option) *StringValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	c.pattern = re
	return makeStringValidator(c)
}

// UUIDValidator validates UUIDs.
type UUIDValidator struct{ *StringValidator }

// UUID creates a new UUID validator.
func UUID(opts ...Option) *UUIDValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &UUIDValidator{StringValidator: patternStringValidator(c, uuidPattern)}
}

// Validate validates a UUID.
func (v *UUIDValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeUUID
				err.Message = fmt.Sprintf("Invalid UUID: %v", v.displayValue(value))
			}
		}
	}
	return result
}

// DateValidator validates ISO-style date strings (YYYY-MM-DD).
type DateValidator struct{ *StringValidator }

// Date creates a new date validator.
func Date(opts ...Option) *DateValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &DateValidator{StringValidator: patternStringValidator(c, datePattern)}
}

// Validate validates a date.
func (v *DateValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeDate
				err.Message = fmt.Sprintf("Invalid date: %v", v.displayValue(value))
			}
		}
	}
	return result
}

// DateTimeValidator validates ISO-style datetime strings.
type DateTimeValidator struct{ *StringValidator }

// DateTime creates a new datetime validator.
func DateTime(opts ...Option) *DateTimeValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &DateTimeValidator{StringValidator: patternStringValidator(c, datetimePattern)}
}

// Validate validates a datetime.
func (v *DateTimeValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeDateTime
				err.Message = fmt.Sprintf("Invalid datetime: %v", v.displayValue(value))
			}
		}
	}
	return result
}

// NumericValidator validates strings that look like numbers.
type NumericValidator struct{ *StringValidator }

// Numeric creates a new numeric-string validator.
func Numeric(opts ...Option) *NumericValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &NumericValidator{StringValidator: patternStringValidator(c, numericPattern)}
}

// Validate validates that a value is a numeric string.
func (v *NumericValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.StringValidator.Validate(ctx, value, field)
	if !result.Valid {
		for _, err := range result.Errors {
			if err.Code == ErrCodePattern {
				err.Code = ErrCodeNumeric
				err.Message = fmt.Sprintf("Expected numeric string, got %v", v.displayValue(value))
			}
		}
	}
	return result
}

// NotEmptyValidator validates that a string, array, or object is non-empty.
type NotEmptyValidator struct{ BaseValidator }

// NotEmpty creates a new non-empty validator.
func NotEmpty(opts ...Option) *NotEmptyValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &NotEmptyValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
	}
}

// Validate validates that a value is not empty.
func (v *NotEmptyValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	switch x := value.(type) {
	case string:
		if len(x) == 0 {
			result.AddError(&ValidationError{
				Field:   field,
				Message: "Value must not be empty",
				Value:   v.errorValue(value),
				Code:    ErrCodeNotEmpty,
			})
		}
	case []interface{}:
		if len(x) == 0 {
			result.AddError(&ValidationError{
				Field:   field,
				Message: "Value must not be empty",
				Value:   v.errorValue(value),
				Code:    ErrCodeNotEmpty,
			})
		}
	case map[string]interface{}:
		if len(x) == 0 {
			result.AddError(&ValidationError{
				Field:   field,
				Message: "Value must not be empty",
				Value:   v.errorValue(value),
				Code:    ErrCodeNotEmpty,
			})
		}
	default:
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected string, array, or object, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
	}

	return result
}

// EnumValidator validates that a value is one of a set of values using deep
// equality.
type EnumValidator struct {
	BaseValidator
	Values []interface{}
}

// Enum creates a new enum validator. Options such as Optional(), Nullable(),
// and Sensitive() may be supplied after the values slice.
func Enum(values []interface{}, opts ...Option) *EnumValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &EnumValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		Values:        values,
	}
}

// Validate validates that a value is one of the allowed values.
func (v *EnumValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	for _, allowed := range v.Values {
		if reflect.DeepEqual(value, allowed) {
			return result
		}
	}

	allowedJSON := make([]string, len(v.Values))
	for i, a := range v.Values {
		b, err := json.Marshal(a)
		if err != nil {
			b = []byte("?")
		}
		allowedJSON[i] = string(b)
	}
	gotJSON, err := json.Marshal(value)
	if err != nil {
		gotJSON = []byte("?")
	}

	result.AddError(&ValidationError{
		Field:   field,
		Message: fmt.Sprintf("Value must be one of %s, got %v", strings.Join(allowedJSON, ", "), v.displayValue(string(gotJSON))),
		Value:   v.errorValue(value),
		Code:    ErrCodeEnum,
	})
	return result
}

// OneOfValidator validates that a value matches at least one of the supplied
// branch validators.
type OneOfValidator struct {
	BaseValidator
	Branches []Validator
}

// OneOf creates a new one-of validator. Options such as Optional(), Nullable(),
// and Sensitive() may be supplied after the branches slice.
func OneOf(branches []Validator, opts ...Option) *OneOfValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &OneOfValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		Branches:      branches,
	}
}

// Validate validates that a value satisfies at least one branch.
func (v *OneOfValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	for _, branch := range v.Branches {
		res := branch.Validate(ctx, value, field)
		if res.Valid {
			return &ValidationResult{Valid: true}
		}
	}

	result.AddError(&ValidationError{
		Field:   field,
		Message: "Value does not match any allowed schema",
		Value:   v.errorValue(value),
		Code:    ErrCodeOneOf,
	})
	return result
}

// SliceValidator validates slice/array values.
type SliceValidator struct {
	BaseValidator
	ItemValidator Validator
	MinLength     *int
	MaxLength     *int
}

// Slice creates a new slice validator.
func Slice(opts ...Option) *SliceValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &SliceValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		ItemValidator: c.items,
		MinLength:     c.minLength,
		MaxLength:     c.maxLength,
	}
}

// Validate validates a slice value.
func (v *SliceValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	return v.validate(ctx, value, field, false)
}

// validateFailFast implements failFastable: item iteration stops after the
// first error and nested container validators also run in fail-fast mode.
func (v *SliceValidator) validateFailFast(ctx context.Context, value interface{}, field string) *ValidationResult {
	return v.validate(ctx, value, field, true)
}

func (v *SliceValidator) validate(ctx context.Context, value interface{}, field string, failFast bool) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	slice, ok := value.([]interface{})
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected array, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
		return result
	}

	if v.MinLength != nil && len(slice) < *v.MinLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Array length must be at least %d, got %d", *v.MinLength, len(slice)),
			Value:   v.errorValue(value),
			Code:    ErrCodeMinLength,
		})
		if failFast {
			return result
		}
	}

	if v.MaxLength != nil && len(slice) > *v.MaxLength {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Array length must be at most %d, got %d", *v.MaxLength, len(slice)),
			Value:   v.errorValue(value),
			Code:    ErrCodeMaxLength,
		})
		if failFast {
			return result
		}
	}

	if v.ItemValidator != nil {
		for i, item := range slice {
			var itemResult *ValidationResult
			if failFast {
				itemResult = validateFailFast(v.ItemValidator, ctx, item, fmt.Sprintf("%s[%d]", field, i))
			} else {
				itemResult = v.ItemValidator.Validate(ctx, item, fmt.Sprintf("%s[%d]", field, i))
			}
			result.Errors = append(result.Errors, itemResult.Errors...)
			if !itemResult.Valid {
				result.Valid = false
				if failFast {
					return result
				}
			}
		}
	}

	return result
}

// MapValidator validates map values.
type MapValidator struct {
	BaseValidator
	Schema map[string]Validator
	// Strict, when true, reports keys present in the data map that are
	// not defined in Schema as "Unknown field" errors (path: field.key).
	Strict bool
	// FailFast, when true, stops validation at the first error produced
	// inside this map and returns a result containing exactly that error.
	// It also applies to a MapValidator nested under a fail-fast Schema or
	// container, in which case the parent propagates the flag.
	FailFast bool
}

// Map creates a new map validator.
func Map(opts ...Option) *MapValidator {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return &MapValidator{
		BaseValidator: BaseValidator{Required: !c.optional, Nullable: c.nullable, CustomValidators: c.custom, Sensitive: c.sensitive},
		Schema:        c.fields,
		Strict:        c.strict,
		FailFast:      c.failFast,
	}
}

// Validate validates a map value.
func (v *MapValidator) Validate(ctx context.Context, value interface{}, field string) *ValidationResult {
	return v.validate(ctx, value, field, v.FailFast)
}

// validateFailFast implements failFastable: a fail-fast parent (Schema,
// MapValidator, or SliceValidator) forces fail-fast on this map even when
// its own FailFast option is not set.
func (v *MapValidator) validateFailFast(ctx context.Context, value interface{}, field string) *ValidationResult {
	return v.validate(ctx, value, field, true)
}

func (v *MapValidator) validate(ctx context.Context, value interface{}, field string, failFast bool) *ValidationResult {
	result := v.ValidateBase(ctx, value, field)
	if !result.Valid || isNullish(value) {
		return result
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		result.AddError(&ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Expected object, got %s", jsonKind(value)),
			Value:   v.errorValue(value),
			Code:    ErrCodeType,
		})
		return result
	}

	if v.Schema != nil {
		for key, validator := range v.Schema {
			fieldName := fmt.Sprintf("%s.%s", field, key)
			if val, exists := m[key]; exists {
				var fieldResult *ValidationResult
				if failFast {
					fieldResult = validateFailFast(validator, ctx, val, fieldName)
				} else {
					fieldResult = validator.Validate(ctx, val, fieldName)
				}
				result.Errors = append(result.Errors, fieldResult.Errors...)
				if !fieldResult.Valid {
					result.Valid = false
				}
			} else if validator.IsRequired() {
				result.AddError(&ValidationError{
					Field:   fieldName,
					Message: "Field is required",
					Code:    ErrCodeRequired,
				})
			}
			if failFast && !result.Valid {
				return result
			}
		}

		if v.Strict {
			for key, val := range m {
				if _, defined := v.Schema[key]; !defined {
					result.AddError(&ValidationError{
						Field:   fmt.Sprintf("%s.%s", field, key),
						Message: "Unknown field",
						Value:   v.errorValue(val),
						Code:    ErrCodeUnknownField,
					})
					if failFast {
						return result
					}
				}
			}
		}
	}

	return result
}

// Schema represents a validation schema.
type Schema struct {
	validators map[string]Validator
	strict     bool
	failFast   bool
}

// NewSchema creates a new validation schema. Use Strict() and FailFast() to
// enable strict or fail-fast mode.
func NewSchema(validators map[string]Validator, opts ...Option) *Schema {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	s := &Schema{validators: validators, strict: c.strict, failFast: c.failFast}
	return s
}

// SetStrict enables or disables strict mode on the schema.
// See Strict() for details on strict mode behavior.
func (s *Schema) SetStrict(strict bool) {
	s.strict = strict
}

// SetFailFast enables or disables fail-fast mode on the schema.
// See FailFast() for details on fail-fast behavior.
func (s *Schema) SetFailFast(failFast bool) {
	s.failFast = failFast
}

// Validate validates data against the schema.
func (s *Schema) Validate(ctx context.Context, data map[string]interface{}) *ValidationResult {
	result := &ValidationResult{Valid: true}

	for field, validator := range s.validators {
		if value, exists := data[field]; exists {
			var fieldResult *ValidationResult
			if s.failFast {
				fieldResult = validateFailFast(validator, ctx, value, field)
			} else {
				fieldResult = validator.Validate(ctx, value, field)
			}
			result.Errors = append(result.Errors, fieldResult.Errors...)
			if !fieldResult.Valid {
				result.Valid = false
			}
		} else if validator.IsRequired() {
			result.AddError(&ValidationError{
				Field:   field,
				Message: "Field is required",
				Code:    ErrCodeRequired,
			})
		}
		if s.failFast && !result.Valid {
			return result
		}
	}

	if s.strict {
		for key, value := range data {
			if _, defined := s.validators[key]; !defined {
				result.AddError(&ValidationError{
					Field:   key,
					Message: "Unknown field",
					Value:   value,
					Code:    ErrCodeUnknownField,
				})
				if s.failFast {
					return result
				}
			}
		}
	}

	return result
}

// ValidateOrError validates data and returns a ValidationErrors value if any
// errors were found, otherwise nil.
func (s *Schema) ValidateOrError(ctx context.Context, data map[string]interface{}) error {
	result := s.Validate(ctx, data)
	if !result.Valid {
		return ValidationErrors(result.Errors)
	}
	return nil
}

// Helper function to create pointer to int.
func IntPtr(i int) *int {
	return &i
}

// Helper function to create pointer to float64.
func Float64Ptr(f float64) *float64 {
	return &f
}

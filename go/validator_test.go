package validator

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func TestStringValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates valid string", func(t *testing.T) {
		v := String()
		result := v.Validate(ctx, "hello", "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := String()
		result := v.Validate(ctx, 123, "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min length", func(t *testing.T) {
		v := String(MinLength(5))

		if !v.Validate(ctx, "hello", "field").Valid {
			t.Error("expected valid for 'hello'")
		}
		if v.Validate(ctx, "hi", "field").Valid {
			t.Error("expected invalid for 'hi'")
		}
	})

	t.Run("validates max length", func(t *testing.T) {
		v := String(MaxLength(5))

		if !v.Validate(ctx, "hello", "field").Valid {
			t.Error("expected valid for 'hello'")
		}
		if v.Validate(ctx, "hello world", "field").Valid {
			t.Error("expected invalid for 'hello world'")
		}
	})

	t.Run("validates pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		v := String(Pattern(pattern))

		if !v.Validate(ctx, "123-4567", "field").Valid {
			t.Error("expected valid for '123-4567'")
		}
		if v.Validate(ctx, "invalid", "field").Valid {
			t.Error("expected invalid for 'invalid'")
		}
	})

	t.Run("validates choices", func(t *testing.T) {
		v := String(Choices("red", "green", "blue"))

		if !v.Validate(ctx, "red", "field").Valid {
			t.Error("expected valid for 'red'")
		}
		if v.Validate(ctx, "yellow", "field").Valid {
			t.Error("expected invalid for 'yellow'")
		}
	})

	t.Run("handles required field", func(t *testing.T) {
		v := String()
		result := v.Validate(ctx, nil, "field")
		if result.Valid {
			t.Error("expected invalid for nil")
		}
	})

	t.Run("handles optional field", func(t *testing.T) {
		v := String(Optional())
		result := v.Validate(ctx, nil, "field")
		if !result.Valid {
			t.Error("expected valid for nil")
		}
	})
}

func TestIntValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates valid int", func(t *testing.T) {
		v := Int()
		result := v.Validate(ctx, 42, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Int()
		result := v.Validate(ctx, "not an int", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min value", func(t *testing.T) {
		v := Int(MinValue(0))

		if !v.Validate(ctx, 5, "field").Valid {
			t.Error("expected valid for 5")
		}
		if v.Validate(ctx, -1, "field").Valid {
			t.Error("expected invalid for -1")
		}
	})

	t.Run("validates max value", func(t *testing.T) {
		v := Int(MaxValue(100))

		if !v.Validate(ctx, 50, "field").Valid {
			t.Error("expected valid for 50")
		}
		if v.Validate(ctx, 101, "field").Valid {
			t.Error("expected invalid for 101")
		}
	})

	t.Run("accepts float64 if integer", func(t *testing.T) {
		v := Int()
		if !v.Validate(ctx, float64(42), "field").Valid {
			t.Error("expected valid for float64(42)")
		}
		if v.Validate(ctx, 3.14, "field").Valid {
			t.Error("expected invalid for 3.14")
		}
	})
}

func TestFloatValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates valid float", func(t *testing.T) {
		v := Float()
		result := v.Validate(ctx, 3.14, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("accepts int as float", func(t *testing.T) {
		v := Float()
		result := v.Validate(ctx, 42, "field")
		if !result.Valid {
			t.Error("expected valid for int")
		}
	})

	t.Run("validates min value", func(t *testing.T) {
		v := Float(MinFloat(0.0))

		if !v.Validate(ctx, 5.5, "field").Valid {
			t.Error("expected valid for 5.5")
		}
		if v.Validate(ctx, -0.1, "field").Valid {
			t.Error("expected invalid for -0.1")
		}
	})

	t.Run("validates max value", func(t *testing.T) {
		v := Float(MaxFloat(100.0))

		if !v.Validate(ctx, 50.5, "field").Valid {
			t.Error("expected valid for 50.5")
		}
		if v.Validate(ctx, 100.1, "field").Valid {
			t.Error("expected invalid for 100.1")
		}
	})
}

func TestBoolValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates boolean", func(t *testing.T) {
		v := Bool()

		if !v.Validate(ctx, true, "field").Valid {
			t.Error("expected valid for true")
		}
		if !v.Validate(ctx, false, "field").Valid {
			t.Error("expected valid for false")
		}
		if v.Validate(ctx, "not a bool", "field").Valid {
			t.Error("expected invalid for string")
		}
	})
}

func TestEmailValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates valid emails", func(t *testing.T) {
		v := Email()

		validEmails := []string{
			"user@example.com",
			"test.user@example.com",
			"user+tag@example.co.uk",
			"user123@test-domain.com",
		}

		for _, email := range validEmails {
			if !v.Validate(ctx, email, "field").Valid {
				t.Errorf("expected valid for %s", email)
			}
		}
	})

	t.Run("rejects invalid emails", func(t *testing.T) {
		v := Email()

		invalidEmails := []string{
			"not-an-email",
			"@example.com",
			"user@",
			"user @example.com",
		}

		for _, email := range invalidEmails {
			if v.Validate(ctx, email, "field").Valid {
				t.Errorf("expected invalid for %s", email)
			}
		}
	})
}

func TestURLValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates valid URLs", func(t *testing.T) {
		v := URL()

		validURLs := []string{
			"http://example.com",
			"https://example.com",
			"https://example.com/path",
			"https://example.com/path?query=value",
		}

		for _, url := range validURLs {
			if !v.Validate(ctx, url, "field").Valid {
				t.Errorf("expected valid for %s", url)
			}
		}
	})

	t.Run("rejects invalid URLs", func(t *testing.T) {
		v := URL()

		invalidURLs := []string{
			"not-a-url",
			"example.com",
			"http://",
		}

		for _, url := range invalidURLs {
			if v.Validate(ctx, url, "field").Valid {
				t.Errorf("expected invalid for %s", url)
			}
		}
	})
}

func TestSliceValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates slice", func(t *testing.T) {
		v := Slice()
		result := v.Validate(ctx, []interface{}{"a", "b", "c"}, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Slice()
		result := v.Validate(ctx, "not a slice", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min length", func(t *testing.T) {
		v := Slice(MinLength(2))

		if !v.Validate(ctx, []interface{}{"a", "b"}, "field").Valid {
			t.Error("expected valid for length 2")
		}
		if v.Validate(ctx, []interface{}{"a"}, "field").Valid {
			t.Error("expected invalid for length 1")
		}
	})

	t.Run("validates max length", func(t *testing.T) {
		v := Slice(MaxLength(3))

		if !v.Validate(ctx, []interface{}{"a", "b"}, "field").Valid {
			t.Error("expected valid for length 2")
		}
		if v.Validate(ctx, []interface{}{"a", "b", "c", "d"}, "field").Valid {
			t.Error("expected invalid for length 4")
		}
	})

	t.Run("validates items", func(t *testing.T) {
		v := Slice(Items(Int()))

		if !v.Validate(ctx, []interface{}{1, 2, 3}, "field").Valid {
			t.Error("expected valid for [1, 2, 3]")
		}
		if v.Validate(ctx, []interface{}{1, "two", 3}, "field").Valid {
			t.Error("expected invalid for [1, 'two', 3]")
		}
	})
}

func TestMapValidator(t *testing.T) {
	ctx := context.Background()

	t.Run("validates map", func(t *testing.T) {
		v := Map()
		result := v.Validate(ctx, map[string]interface{}{"key": "value"}, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Map()
		result := v.Validate(ctx, "not a map", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates schema", func(t *testing.T) {
		v := Map(Fields(map[string]Validator{
			"name": String(),
			"age":  Int(MinValue(0)),
		}))

		validData := map[string]interface{}{
			"name": "John",
			"age":  30,
		}
		if !v.Validate(ctx, validData, "field").Valid {
			t.Error("expected valid for valid data")
		}

		invalidData := map[string]interface{}{
			"name": "John",
			"age":  -1,
		}
		if v.Validate(ctx, invalidData, "field").Valid {
			t.Error("expected invalid for age -1")
		}
	})

	t.Run("checks required fields", func(t *testing.T) {
		v := Map(Fields(map[string]Validator{
			"name": String(),
		}))

		result := v.Validate(ctx, map[string]interface{}{}, "field")
		if result.Valid {
			t.Error("expected invalid for missing required field")
		}
	})

	t.Run("strict mode reports unknown keys with field path", func(t *testing.T) {
		v := Map(
			Strict(),
			Fields(map[string]Validator{
				"name": String(),
			}),
		)

		result := v.Validate(ctx, map[string]interface{}{
			"name":    "John",
			"unknown": "value",
		}, "parent")

		if result.Valid {
			t.Fatal("expected invalid result in strict mode")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "parent.unknown" && err.Code == ErrCodeUnknownField && err.Message == "Unknown field" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected error {Field: parent.unknown, Message: Unknown field}, got %v", result.Errors)
		}
	})

	t.Run("non-strict mode ignores unknown keys by default", func(t *testing.T) {
		v := Map(Fields(map[string]Validator{
			"name": String(),
		}))

		result := v.Validate(ctx, map[string]interface{}{
			"name":    "John",
			"unknown": "value",
		}, "parent")

		if !result.Valid {
			t.Errorf("expected valid result in non-strict mode, got %v", result.Errors)
		}
	})
}

func TestSchema(t *testing.T) {
	ctx := context.Background()

	t.Run("validates simple schema", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"username": String(MinLength(3), MaxLength(20)),
			"email":    Email(),
			"age":      Int(MinValue(0), MaxValue(120)),
		})

		data := map[string]interface{}{
			"username": "john_doe",
			"email":    "john@example.com",
			"age":      30,
		}

		result := schema.Validate(ctx, data)
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("collects multiple errors", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"username": String(MinLength(3)),
			"email":    Email(),
			"age":      Int(MinValue(0)),
		})

		data := map[string]interface{}{
			"username": "ab",
			"email":    "invalid-email",
			"age":      -1,
		}

		result := schema.Validate(ctx, data)
		if result.Valid {
			t.Error("expected invalid result")
		}
		if len(result.Errors) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors))
		}
	})

	t.Run("validates nested schema", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"user": Map(Fields(map[string]Validator{
				"name":  String(),
				"email": Email(),
			})),
			"settings": Map(Fields(map[string]Validator{
				"theme":         String(Choices("light", "dark")),
				"notifications": Bool(),
			})),
		})

		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name":  "John",
				"email": "john@example.com",
			},
			"settings": map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
			},
		}

		result := schema.Validate(ctx, data)
		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}
	})

	t.Run("validates slice of maps", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"users": Slice(Items(Map(Fields(map[string]Validator{
				"name": String(),
				"age":  Int(MinValue(0)),
			})))),
		})

		data := map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"name": "John", "age": 30},
				map[string]interface{}{"name": "Jane", "age": 25},
			},
		}

		result := schema.Validate(ctx, data)
		if !result.Valid {
			t.Error("expected valid result")
		}
	})
}

func TestSchemaStrictMode(t *testing.T) {
	ctx := context.Background()

	t.Run("strict mode reports unknown top-level keys", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"name": String(),
		}, Strict())

		result := schema.Validate(ctx, map[string]interface{}{
			"name":  "John",
			"extra": "value",
		})

		if result.Valid {
			t.Fatal("expected invalid result in strict mode")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "extra" && err.Code == ErrCodeUnknownField && err.Message == "Unknown field" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected error {Field: extra, Message: Unknown field}, got %v", result.Errors)
		}
	})

	t.Run("strict mode off by default", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"name": String(),
		})

		result := schema.Validate(ctx, map[string]interface{}{
			"name":  "John",
			"extra": "value",
		})

		if !result.Valid {
			t.Errorf("expected valid result by default, got %v", result.Errors)
		}
	})

	t.Run("SetStrict toggles strict mode", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"name": String(),
		})
		schema.SetStrict(true)

		result := schema.Validate(ctx, map[string]interface{}{
			"name":  "John",
			"extra": 1,
		})

		if result.Valid {
			t.Error("expected invalid result after SetStrict(true)")
		}
	})

	t.Run("nested MapValidator strict reports parent.unknown path", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"parent": Map(
				Strict(),
				Fields(map[string]Validator{
					"known": String(),
				}),
			),
		})

		result := schema.Validate(ctx, map[string]interface{}{
			"parent": map[string]interface{}{
				"known":   "value",
				"unknown": "value",
			},
		})

		if result.Valid {
			t.Fatal("expected invalid result")
		}

		found := false
		for _, err := range result.Errors {
			if err.Field == "parent.unknown" && err.Code == ErrCodeUnknownField && err.Message == "Unknown field" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected error {Field: parent.unknown, Message: Unknown field}, got %v", result.Errors)
		}
	})

	t.Run("non-map input still type-errors under strict MapValidator", func(t *testing.T) {
		v := Map(
			Strict(),
			Fields(map[string]Validator{
				"name": String(),
			}),
		)

		result := v.Validate(ctx, "not a map", "field")
		if result.Valid {
			t.Fatal("expected invalid result for non-map input")
		}
		if len(result.Errors) == 0 || result.Errors[0].Field != "field" {
			t.Errorf("expected type error on 'field', got %v", result.Errors)
		}
	})
}

func TestCustomValidators(t *testing.T) {
	ctx := context.Background()

	t.Run("applies custom validator", func(t *testing.T) {
		isEven := func(ctx context.Context, value interface{}) error {
			if num, ok := value.(int); ok {
				if num%2 != 0 {
					return fmt.Errorf("Value must be even")
				}
			}
			return nil
		}

		v := Int(WithCustom(isEven))

		if !v.Validate(ctx, 4, "field").Valid {
			t.Error("expected valid for 4")
		}
		if v.Validate(ctx, 3, "field").Valid {
			t.Error("expected invalid for 3")
		}
	})

	t.Run("applies multiple custom validators", func(t *testing.T) {
		isPositive := func(ctx context.Context, value interface{}) error {
			if num, ok := value.(int); ok {
				if num <= 0 {
					return fmt.Errorf("Must be positive")
				}
			}
			return nil
		}

		isEven := func(ctx context.Context, value interface{}) error {
			if num, ok := value.(int); ok {
				if num%2 != 0 {
					return fmt.Errorf("Must be even")
				}
			}
			return nil
		}

		v := Int(WithCustom(isPositive), WithCustom(isEven))

		if !v.Validate(ctx, 4, "field").Valid {
			t.Error("expected valid for 4")
		}
		if v.Validate(ctx, -2, "field").Valid {
			t.Error("expected invalid for -2")
		}
	})
}

func TestRealWorldScenarios(t *testing.T) {
	ctx := context.Background()

	t.Run("validates user registration", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"username":       String(MinLength(3), MaxLength(20)),
			"email":          Email(),
			"password":       String(MinLength(8)),
			"age":            Int(MinValue(13), Optional()),
			"terms_accepted": Bool(),
		})

		validData := map[string]interface{}{
			"username":       "john_doe",
			"email":          "john@example.com",
			"password":       "secure_password_123",
			"age":            25,
			"terms_accepted": true,
		}

		if !schema.Validate(ctx, validData).Valid {
			t.Error("expected valid for valid data")
		}

		invalidData := map[string]interface{}{
			"username":       "ab",
			"email":          "invalid",
			"password":       "short",
			"terms_accepted": false,
		}

		result := schema.Validate(ctx, invalidData)
		if result.Valid {
			t.Error("expected invalid for invalid data")
		}
		if len(result.Errors) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors))
		}
	})

	t.Run("validates API request", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"method":  String(Choices("GET", "POST", "PUT", "DELETE")),
			"url":     URL(),
			"headers": Map(Optional()),
			"body":    Map(Optional()),
			"timeout": Float(MinFloat(0), Optional()),
		})

		data := map[string]interface{}{
			"method": "POST",
			"url":    "https://api.example.com/users",
			"headers": map[string]interface{}{
				"Content-Type": "application/json",
			},
			"body": map[string]interface{}{
				"name": "John",
			},
			"timeout": 30.0,
		}

		if !schema.Validate(ctx, data).Valid {
			t.Error("expected valid result")
		}
	})
}

func TestTypedNilHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("required Int rejects typed nil as missing", func(t *testing.T) {
		var p *int = nil
		result := Int().Validate(ctx, p, "count")
		if result.Valid {
			t.Fatal("expected invalid result for typed nil on required Int")
		}
		if len(result.Errors) != 1 || result.Errors[0].Code != ErrCodeRequired || result.Errors[0].Message != "Field is required" {
			t.Errorf("expected required error, got %v", result.Errors)
		}
	})

	t.Run("nullable Int accepts typed nil", func(t *testing.T) {
		var p *int = nil
		result := Int(Nullable()).Validate(ctx, p, "count")
		if !result.Valid {
			t.Errorf("expected valid result for typed nil on nullable Int, got %v", result.Errors)
		}
	})

	t.Run("optional String accepts typed nil", func(t *testing.T) {
		var s *string = nil
		result := String(Optional()).Validate(ctx, s, "bio")
		if !result.Valid {
			t.Errorf("expected valid result for typed nil on optional String, got %v", result.Errors)
		}
	})
}

func TestErrorCodes(t *testing.T) {
	ctx := context.Background()
	pattern := regexp.MustCompile(`^\d+$`)

	t.Run("string errors carry canonical codes", func(t *testing.T) {
		tests := []struct {
			value   interface{}
			v       Validator
			want    string
			message string
		}{
			{nil, String(), ErrCodeRequired, "Field is required"},
			{123, String(), ErrCodeType, "Expected string, got integer"},
			{"ab", String(MinLength(5)), ErrCodeMinLength, "String length must be at least 5, got 2"},
			{"abcd", String(MaxLength(3)), ErrCodeMaxLength, "String length must be at most 3, got 4"},
			{"abc", String(Pattern(pattern)), ErrCodePattern, "String does not match pattern ^\\d+$"},
			{"yellow", String(Choices("red", "green")), ErrCodeChoices, "Value must be one of red, green, got 'yellow'"},
		}

		for _, tc := range tests {
			result := tc.v.Validate(ctx, tc.value, "field")
			if result.Valid {
				t.Fatalf("expected invalid for %v", tc.value)
			}
			found := false
			for _, err := range result.Errors {
				if err.Code == tc.want && err.Message == tc.message {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("value %v: expected code %q message %q, got %v", tc.value, tc.want, tc.message, result.Errors)
			}
		}
	})

	t.Run("email and URL rewrite pattern codes", func(t *testing.T) {
		emailResult := Email().Validate(ctx, "not-an-email", "email")
		if emailResult.Valid || emailResult.Errors[0].Code != ErrCodeEmail || emailResult.Errors[0].Message != "Invalid email address: not-an-email" {
			t.Errorf("unexpected email error: %v", emailResult.Errors)
		}

		urlResult := URL().Validate(ctx, "not-a-url", "url")
		if urlResult.Valid || urlResult.Errors[0].Code != ErrCodeURL || urlResult.Errors[0].Message != "Invalid URL: not-a-url" {
			t.Errorf("unexpected URL error: %v", urlResult.Errors)
		}
	})
}

func TestChoicesLookup(t *testing.T) {
	ctx := context.Background()

	t.Run("large choices list validates members and preserves message order", func(t *testing.T) {
		choices := make([]string, 500)
		for i := range choices {
			choices[i] = fmt.Sprintf("choice-%03d", i)
		}
		v := String(Choices(choices...))

		// First, middle and last declared choices all pass.
		for _, c := range []string{choices[0], choices[250], choices[499]} {
			if !v.Validate(ctx, c, "field").Valid {
				t.Errorf("expected valid for %q", c)
			}
		}

		result := v.Validate(ctx, "nope", "field")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly one choices error, got %v", result.Errors)
		}
		want := fmt.Sprintf("Value must be one of %s, got 'nope'", strings.Join(choices, ", "))
		if result.Errors[0].Message != want {
			t.Error("choices error message does not list choices in declared order")
		}
		if result.Errors[0].Code != ErrCodeChoices {
			t.Errorf("expected code %q, got %q", ErrCodeChoices, result.Errors[0].Code)
		}
	})

	t.Run("directly constructed StringValidator still checks choices", func(t *testing.T) {
		v := &StringValidator{
			BaseValidator: BaseValidator{Required: true},
			Choices:       []string{"a", "b"},
		}
		if !v.Validate(ctx, "b", "field").Valid {
			t.Error("expected valid for 'b'")
		}
		if v.Validate(ctx, "z", "field").Valid {
			t.Error("expected invalid for 'z'")
		}
	})
}

func TestFailFast(t *testing.T) {
	ctx := context.Background()
	failingSchema := func() map[string]Validator {
		return map[string]Validator{
			"username": String(MinLength(5)),
			"email":    Email(),
			"age":      Int(MinValue(0)),
		}
	}
	failingData := map[string]interface{}{
		"username": "ab",
		"email":    "not-an-email",
		"age":      -1,
	}

	t.Run("off by default collects all errors", func(t *testing.T) {
		result := NewSchema(failingSchema()).Validate(ctx, failingData)
		if result.Valid {
			t.Fatal("expected invalid result")
		}
		if len(result.Errors) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors))
		}
	})

	t.Run("FailFast returns exactly one error", func(t *testing.T) {
		result := NewSchema(failingSchema(), FailFast()).Validate(ctx, failingData)
		if result.Valid {
			t.Fatal("expected invalid result")
		}
		if len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %d: %v", len(result.Errors), result.Errors)
		}
	})

	t.Run("SetFailFast toggles fail-fast mode", func(t *testing.T) {
		schema := NewSchema(failingSchema())
		schema.SetFailFast(true)
		if got := len(schema.Validate(ctx, failingData).Errors); got != 1 {
			t.Fatalf("expected 1 error after SetFailFast(true), got %d", got)
		}
		schema.SetFailFast(false)
		if got := len(schema.Validate(ctx, failingData).Errors); got < 3 {
			t.Fatalf("expected >=3 errors after SetFailFast(false), got %d", got)
		}
	})

	t.Run("fail-fast propagates into nested MapValidator fields", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"user": Map(Fields(map[string]Validator{
				"name": String(MinLength(5)),
				"age":  Int(MinValue(0)),
			})),
		}, FailFast())

		result := schema.Validate(ctx, map[string]interface{}{
			"user": map[string]interface{}{"name": "ab", "age": -1},
		})
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
	})

	t.Run("fail-fast propagates into SliceValidator items", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"nums": Slice(Items(Int())),
		}, FailFast())

		result := schema.Validate(ctx, map[string]interface{}{
			"nums": []interface{}{"a", "b", "c"},
		})
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		if result.Errors[0].Field != "nums[0]" {
			t.Errorf("expected first item error at nums[0], got %v", result.Errors)
		}
	})

	t.Run("single validator returns only its first error", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"s": String(MinLength(10), MaxLength(2)),
		}, FailFast())

		result := schema.Validate(ctx, map[string]interface{}{"s": "abc"})
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
	})

	t.Run("fail-fast stops at first missing required field", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"a": String(),
			"b": String(),
		}, FailFast())

		result := schema.Validate(ctx, map[string]interface{}{})
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		if result.Errors[0].Code != ErrCodeRequired {
			t.Errorf("expected required error, got %v", result.Errors[0])
		}
	})

	t.Run("fail-fast stops at first strict unknown field", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"a": String(),
		}, Strict(), FailFast())

		result := schema.Validate(ctx, map[string]interface{}{
			"a": "ok",
			"x": 1,
			"y": 2,
		})
		if result.Valid || len(result.Errors) != 1 || result.Errors[0].Code != ErrCodeUnknownField {
			t.Fatalf("expected exactly 1 unknown_field error, got %v", result.Errors)
		}
	})

	t.Run("valid data still passes in fail-fast mode", func(t *testing.T) {
		schema := NewSchema(failingSchema(), FailFast())
		result := schema.Validate(ctx, map[string]interface{}{
			"username": "john_doe",
			"email":    "john@example.com",
			"age":      30,
		})
		if !result.Valid {
			t.Errorf("expected valid result, got %v", result.Errors)
		}
	})
}

func TestMapValidatorFailFast(t *testing.T) {
	ctx := context.Background()

	t.Run("Map FailFast returns exactly one error", func(t *testing.T) {
		v := Map(
			FailFast(),
			Fields(map[string]Validator{
				"name": String(MinLength(5)),
				"age":  Int(MinValue(0)),
			}),
		)

		result := v.Validate(ctx, map[string]interface{}{"name": "ab", "age": -1}, "user")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
	})

	t.Run("Map FailFast propagates to nested maps", func(t *testing.T) {
		v := Map(
			FailFast(),
			Fields(map[string]Validator{
				"inner": Map(Fields(map[string]Validator{
					"x": Int(),
					"y": Int(),
				})),
			}),
		)

		result := v.Validate(ctx, map[string]interface{}{
			"inner": map[string]interface{}{"x": "bad", "y": "bad"},
		}, "outer")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
	})

	t.Run("Map FailFast off collects all errors", func(t *testing.T) {
		v := Map(Fields(map[string]Validator{
			"name": String(MinLength(5)),
			"age":  Int(MinValue(0)),
		}))

		result := v.Validate(ctx, map[string]interface{}{"name": "ab", "age": -1}, "user")
		if result.Valid {
			t.Fatal("expected invalid result")
		}
		if len(result.Errors) != 2 {
			t.Fatalf("expected 2 errors, got %v", result.Errors)
		}
	})
}

func TestSensitive(t *testing.T) {
	ctx := context.Background()

	t.Run("redacts value from message and Value field", func(t *testing.T) {
		v := Email(Sensitive())
		result := v.Validate(ctx, "p@ssw0rd", "secret")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		err := result.Errors[0]
		if err.Code != ErrCodeEmail {
			t.Errorf("expected code %q, got %q", ErrCodeEmail, err.Code)
		}
		if err.Message != "Invalid email address: ***" {
			t.Errorf("expected redacted message, got %q", err.Message)
		}
		if err.Value != nil {
			t.Errorf("expected nil Value, got %v", err.Value)
		}
	})

	t.Run("redacts choices interpolation", func(t *testing.T) {
		v := String(Choices("a", "b"), Sensitive())
		result := v.Validate(ctx, "hunter2", "token")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		err := result.Errors[0]
		if err.Message != "Value must be one of a, b, got '***'" {
			t.Errorf("expected redacted message, got %q", err.Message)
		}
		if err.Value != nil {
			t.Errorf("expected nil Value, got %v", err.Value)
		}
	})

	t.Run("does not redact length or bound values", func(t *testing.T) {
		v := String(MinLength(5), Sensitive())
		result := v.Validate(ctx, "abc", "secret")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		err := result.Errors[0]
		if err.Message != "String length must be at least 5, got 3" {
			t.Errorf("length must not be redacted, got %q", err.Message)
		}
		if err.Value != nil {
			t.Errorf("expected nil Value, got %v", err.Value)
		}
	})

	t.Run("is not inherited by nested validators", func(t *testing.T) {
		v := Slice(Sensitive(), Items(String(Choices("a", "b"))))
		result := v.Validate(ctx, []interface{}{"x"}, "field")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		if result.Errors[0].Message != "Value must be one of a, b, got 'x'" {
			t.Errorf("nested validator should not redact, got %q", result.Errors[0].Message)
		}
		if result.Errors[0].Value != "x" {
			t.Errorf("nested validator should expose value, got %v", result.Errors[0].Value)
		}
	})
}

func TestMaxPatternInputLength(t *testing.T) {
	ctx := context.Background()

	t.Run("oversized input produces pattern error without matching", func(t *testing.T) {
		big := strings.Repeat("a", MaxPatternInputLength+1)
		v := String(Pattern(regexp.MustCompile(`^a+$`)))
		result := v.Validate(ctx, big, "field")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		if result.Errors[0].Code != ErrCodePattern {
			t.Errorf("expected code %q, got %q", ErrCodePattern, result.Errors[0].Code)
		}
	})

	t.Run("cap applies to fixed pattern validators", func(t *testing.T) {
		big := strings.Repeat("a", MaxPatternInputLength+1)
		result := Email().Validate(ctx, big, "field")
		if result.Valid || len(result.Errors) != 1 {
			t.Fatalf("expected exactly 1 error, got %v", result.Errors)
		}
		if result.Errors[0].Code != ErrCodeEmail {
			t.Errorf("expected code %q, got %q", ErrCodeEmail, result.Errors[0].Code)
		}
	})

	t.Run("input at the cap is still matched", func(t *testing.T) {
		big := strings.Repeat("a", MaxPatternInputLength)
		v := String(Pattern(regexp.MustCompile(`^a+$`)))
		if !v.Validate(ctx, big, "field").Valid {
			t.Error("expected valid result at exactly the cap")
		}
	})
}

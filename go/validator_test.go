package validator

import (
	"regexp"
	"testing"
)

func TestStringValidator(t *testing.T) {
	t.Run("validates valid string", func(t *testing.T) {
		v := String()
		result := v.Validate("hello", "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := String()
		result := v.Validate(123, "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min length", func(t *testing.T) {
		minLen := 5
		v := String(StringOptions{MinLength: &minLen})

		if !v.Validate("hello", "field").Valid {
			t.Error("expected valid for 'hello'")
		}
		if v.Validate("hi", "field").Valid {
			t.Error("expected invalid for 'hi'")
		}
	})

	t.Run("validates max length", func(t *testing.T) {
		maxLen := 5
		v := String(StringOptions{MaxLength: &maxLen})

		if !v.Validate("hello", "field").Valid {
			t.Error("expected valid for 'hello'")
		}
		if v.Validate("hello world", "field").Valid {
			t.Error("expected invalid for 'hello world'")
		}
	})

	t.Run("validates pattern", func(t *testing.T) {
		pattern := regexp.MustCompile(`^\d{3}-\d{4}$`)
		v := String(StringOptions{Pattern: pattern})

		if !v.Validate("123-4567", "field").Valid {
			t.Error("expected valid for '123-4567'")
		}
		if v.Validate("invalid", "field").Valid {
			t.Error("expected invalid for 'invalid'")
		}
	})

	t.Run("validates choices", func(t *testing.T) {
		v := String(StringOptions{Choices: []string{"red", "green", "blue"}})

		if !v.Validate("red", "field").Valid {
			t.Error("expected valid for 'red'")
		}
		if v.Validate("yellow", "field").Valid {
			t.Error("expected invalid for 'yellow'")
		}
	})

	t.Run("handles required field", func(t *testing.T) {
		v := String(StringOptions{Required: true})
		result := v.Validate(nil, "field")
		if result.Valid {
			t.Error("expected invalid for nil")
		}
	})

	t.Run("handles optional field", func(t *testing.T) {
		v := String(StringOptions{Required: false})
		result := v.Validate(nil, "field")
		if !result.Valid {
			t.Error("expected valid for nil")
		}
	})
}

func TestIntValidator(t *testing.T) {
	t.Run("validates valid int", func(t *testing.T) {
		v := Int()
		result := v.Validate(42, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Int()
		result := v.Validate("not an int", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min value", func(t *testing.T) {
		minVal := 0
		v := Int(IntOptions{MinValue: &minVal})

		if !v.Validate(5, "field").Valid {
			t.Error("expected valid for 5")
		}
		if v.Validate(-1, "field").Valid {
			t.Error("expected invalid for -1")
		}
	})

	t.Run("validates max value", func(t *testing.T) {
		maxVal := 100
		v := Int(IntOptions{MaxValue: &maxVal})

		if !v.Validate(50, "field").Valid {
			t.Error("expected valid for 50")
		}
		if v.Validate(101, "field").Valid {
			t.Error("expected invalid for 101")
		}
	})

	t.Run("accepts float64 if integer", func(t *testing.T) {
		v := Int()
		if !v.Validate(float64(42), "field").Valid {
			t.Error("expected valid for float64(42)")
		}
		if v.Validate(3.14, "field").Valid {
			t.Error("expected invalid for 3.14")
		}
	})
}

func TestFloatValidator(t *testing.T) {
	t.Run("validates valid float", func(t *testing.T) {
		v := Float()
		result := v.Validate(3.14, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("accepts int as float", func(t *testing.T) {
		v := Float()
		result := v.Validate(42, "field")
		if !result.Valid {
			t.Error("expected valid for int")
		}
	})

	t.Run("validates min value", func(t *testing.T) {
		minVal := 0.0
		v := Float(FloatOptions{MinValue: &minVal})

		if !v.Validate(5.5, "field").Valid {
			t.Error("expected valid for 5.5")
		}
		if v.Validate(-0.1, "field").Valid {
			t.Error("expected invalid for -0.1")
		}
	})

	t.Run("validates max value", func(t *testing.T) {
		maxVal := 100.0
		v := Float(FloatOptions{MaxValue: &maxVal})

		if !v.Validate(50.5, "field").Valid {
			t.Error("expected valid for 50.5")
		}
		if v.Validate(100.1, "field").Valid {
			t.Error("expected invalid for 100.1")
		}
	})
}

func TestBoolValidator(t *testing.T) {
	t.Run("validates boolean", func(t *testing.T) {
		v := Bool()

		if !v.Validate(true, "field").Valid {
			t.Error("expected valid for true")
		}
		if !v.Validate(false, "field").Valid {
			t.Error("expected valid for false")
		}
		if v.Validate("not a bool", "field").Valid {
			t.Error("expected invalid for string")
		}
	})
}

func TestEmailValidator(t *testing.T) {
	t.Run("validates valid emails", func(t *testing.T) {
		v := Email()

		validEmails := []string{
			"user@example.com",
			"test.user@example.com",
			"user+tag@example.co.uk",
			"user123@test-domain.com",
		}

		for _, email := range validEmails {
			if !v.Validate(email, "field").Valid {
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
			if v.Validate(email, "field").Valid {
				t.Errorf("expected invalid for %s", email)
			}
		}
	})
}

func TestURLValidator(t *testing.T) {
	t.Run("validates valid URLs", func(t *testing.T) {
		v := URL()

		validURLs := []string{
			"http://example.com",
			"https://example.com",
			"https://example.com/path",
			"https://example.com/path?query=value",
		}

		for _, url := range validURLs {
			if !v.Validate(url, "field").Valid {
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
			if v.Validate(url, "field").Valid {
				t.Errorf("expected invalid for %s", url)
			}
		}
	})
}

func TestSliceValidator(t *testing.T) {
	t.Run("validates slice", func(t *testing.T) {
		v := Slice()
		result := v.Validate([]interface{}{"a", "b", "c"}, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Slice()
		result := v.Validate("not a slice", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates min length", func(t *testing.T) {
		minLen := 2
		v := Slice(SliceOptions{MinLength: &minLen})

		if !v.Validate([]interface{}{"a", "b"}, "field").Valid {
			t.Error("expected valid for length 2")
		}
		if v.Validate([]interface{}{"a"}, "field").Valid {
			t.Error("expected invalid for length 1")
		}
	})

	t.Run("validates max length", func(t *testing.T) {
		maxLen := 3
		v := Slice(SliceOptions{MaxLength: &maxLen})

		if !v.Validate([]interface{}{"a", "b"}, "field").Valid {
			t.Error("expected valid for length 2")
		}
		if v.Validate([]interface{}{"a", "b", "c", "d"}, "field").Valid {
			t.Error("expected invalid for length 4")
		}
	})

	t.Run("validates items", func(t *testing.T) {
		v := Slice(SliceOptions{ItemValidator: Int()})

		if !v.Validate([]interface{}{1, 2, 3}, "field").Valid {
			t.Error("expected valid for [1, 2, 3]")
		}
		if v.Validate([]interface{}{1, "two", 3}, "field").Valid {
			t.Error("expected invalid for [1, 'two', 3]")
		}
	})
}

func TestMapValidator(t *testing.T) {
	t.Run("validates map", func(t *testing.T) {
		v := Map()
		result := v.Validate(map[string]interface{}{"key": "value"}, "field")
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("rejects invalid type", func(t *testing.T) {
		v := Map()
		result := v.Validate("not a map", "field")
		if result.Valid {
			t.Error("expected invalid result")
		}
	})

	t.Run("validates schema", func(t *testing.T) {
		v := Map(MapOptions{
			Schema: map[string]Validator{
				"name": String(),
				"age":  Int(IntOptions{MinValue: IntPtr(0)}),
			},
		})

		validData := map[string]interface{}{
			"name": "John",
			"age":  30,
		}
		if !v.Validate(validData, "field").Valid {
			t.Error("expected valid for valid data")
		}

		invalidData := map[string]interface{}{
			"name": "John",
			"age":  -1,
		}
		if v.Validate(invalidData, "field").Valid {
			t.Error("expected invalid for age -1")
		}
	})

	t.Run("checks required fields", func(t *testing.T) {
		v := Map(MapOptions{
			Schema: map[string]Validator{
				"name": String(StringOptions{Required: true}),
			},
		})

		result := v.Validate(map[string]interface{}{}, "field")
		if result.Valid {
			t.Error("expected invalid for missing required field")
		}
	})
}

func TestSchema(t *testing.T) {
	t.Run("validates simple schema", func(t *testing.T) {
		minLen := 3
		minAge := 0
		maxAge := 120

		schema := NewSchema(map[string]Validator{
			"username": String(StringOptions{MinLength: &minLen}),
			"email":    Email(),
			"age":      Int(IntOptions{MinValue: &minAge, MaxValue: &maxAge}),
		})

		data := map[string]interface{}{
			"username": "john_doe",
			"email":    "john@example.com",
			"age":      30,
		}

		result := schema.Validate(data)
		if !result.Valid {
			t.Error("expected valid result")
		}
	})

	t.Run("collects multiple errors", func(t *testing.T) {
		minLen := 3
		minAge := 0

		schema := NewSchema(map[string]Validator{
			"username": String(StringOptions{MinLength: &minLen}),
			"email":    Email(),
			"age":      Int(IntOptions{MinValue: &minAge}),
		})

		data := map[string]interface{}{
			"username": "ab",
			"email":    "invalid-email",
			"age":      -1,
		}

		result := schema.Validate(data)
		if result.Valid {
			t.Error("expected invalid result")
		}
		if len(result.Errors) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors))
		}
	})

	t.Run("validates nested schema", func(t *testing.T) {
		schema := NewSchema(map[string]Validator{
			"user": Map(MapOptions{
				Schema: map[string]Validator{
					"name":  String(),
					"email": Email(),
				},
			}),
			"settings": Map(MapOptions{
				Schema: map[string]Validator{
					"theme":         String(StringOptions{Choices: []string{"light", "dark"}}),
					"notifications": Bool(),
				},
			}),
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

		result := schema.Validate(data)
		if !result.Valid {
			t.Errorf("expected valid result, got errors: %v", result.Errors)
		}
	})

	t.Run("validates slice of maps", func(t *testing.T) {
		minAge := 0

		schema := NewSchema(map[string]Validator{
			"users": Slice(SliceOptions{
				ItemValidator: Map(MapOptions{
					Schema: map[string]Validator{
						"name": String(),
						"age":  Int(IntOptions{MinValue: &minAge}),
					},
				}),
			}),
		})

		data := map[string]interface{}{
			"users": []interface{}{
				map[string]interface{}{"name": "John", "age": 30},
				map[string]interface{}{"name": "Jane", "age": 25},
			},
		}

		result := schema.Validate(data)
		if !result.Valid {
			t.Error("expected valid result")
		}
	})
}

func TestCustomValidators(t *testing.T) {
	t.Run("applies custom validator", func(t *testing.T) {
		isEven := func(value interface{}) *string {
			if num, ok := value.(int); ok {
				if num%2 != 0 {
					msg := "Value must be even"
					return &msg
				}
			}
			return nil
		}

		v := Int().Custom(isEven)

		if !v.Validate(4, "field").Valid {
			t.Error("expected valid for 4")
		}
		if v.Validate(3, "field").Valid {
			t.Error("expected invalid for 3")
		}
	})

	t.Run("applies multiple custom validators", func(t *testing.T) {
		isPositive := func(value interface{}) *string {
			if num, ok := value.(int); ok {
				if num <= 0 {
					msg := "Must be positive"
					return &msg
				}
			}
			return nil
		}

		isEven := func(value interface{}) *string {
			if num, ok := value.(int); ok {
				if num%2 != 0 {
					msg := "Must be even"
					return &msg
				}
			}
			return nil
		}

		v := Int().Custom(isPositive).Custom(isEven)

		if !v.Validate(4, "field").Valid {
			t.Error("expected valid for 4")
		}
		if v.Validate(-2, "field").Valid {
			t.Error("expected invalid for -2")
		}
	})
}

func TestRealWorldScenarios(t *testing.T) {
	t.Run("validates user registration", func(t *testing.T) {
		minUserLen := 3
		maxUserLen := 20
		minPassLen := 8
		minAge := 13

		schema := NewSchema(map[string]Validator{
			"username":       String(StringOptions{MinLength: &minUserLen, MaxLength: &maxUserLen}),
			"email":          Email(),
			"password":       String(StringOptions{MinLength: &minPassLen}),
			"age":            Int(IntOptions{MinValue: &minAge, Required: false}),
			"terms_accepted": Bool(),
		})

		validData := map[string]interface{}{
			"username":       "john_doe",
			"email":          "john@example.com",
			"password":       "secure_password_123",
			"age":            25,
			"terms_accepted": true,
		}

		if !schema.Validate(validData).Valid {
			t.Error("expected valid for valid data")
		}

		invalidData := map[string]interface{}{
			"username":       "ab",
			"email":          "invalid",
			"password":       "short",
			"terms_accepted": false,
		}

		result := schema.Validate(invalidData)
		if result.Valid {
			t.Error("expected invalid for invalid data")
		}
		if len(result.Errors) < 3 {
			t.Errorf("expected at least 3 errors, got %d", len(result.Errors))
		}
	})

	t.Run("validates API request", func(t *testing.T) {
		minTimeout := 0.0

		schema := NewSchema(map[string]Validator{
			"method":  String(StringOptions{Choices: []string{"GET", "POST", "PUT", "DELETE"}}),
			"url":     URL(),
			"headers": Map(MapOptions{Required: false}),
			"body":    Map(MapOptions{Required: false}),
			"timeout": Float(FloatOptions{MinValue: &minTimeout, Required: false}),
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

		if !schema.Validate(data).Valid {
			t.Error("expected valid result")
		}
	})
}

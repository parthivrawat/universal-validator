package main

import (
	"fmt"
	"regexp"
	"strings"

	v "github.com/parthivrawat/universal-validator/go"
)

func exampleBasicValidation() {
	fmt.Println("\n=== Example 1: Basic Validation ===")

	// String validation
	nameValidator := v.String(v.StringOptions{MinLength: v.IntPtr(2), MaxLength: v.IntPtr(50)})
	result := nameValidator.Validate("John Doe", "name")
	fmt.Printf("Name validation: %v\n", result.Valid)

	// Integer validation
	ageValidator := v.Int(v.IntOptions{MinValue: v.IntPtr(0), MaxValue: v.IntPtr(120)})
	result = ageValidator.Validate(25, "age")
	fmt.Printf("Age validation: %v\n", result.Valid)

	// Email validation
	emailValidator := v.Email()
	result = emailValidator.Validate("john@example.com", "email")
	fmt.Printf("Email validation: %v\n", result.Valid)
}

func exampleUserRegistration() {
	fmt.Println("\n=== Example 2: User Registration ===")

	schema := v.NewSchema(map[string]v.Validator{
		"username": v.String(v.StringOptions{MinLength: v.IntPtr(3), MaxLength: v.IntPtr(20)}),
		"email":    v.Email(),
		"password": v.String(v.StringOptions{MinLength: v.IntPtr(8)}),
		"age":      v.Int(v.IntOptions{MinValue: v.IntPtr(13), Required: false}),
		"bio":      v.String(v.StringOptions{MaxLength: v.IntPtr(500), Required: false}),
		"terms_accepted": v.Bool(),
	})

	// Valid data
	validData := map[string]interface{}{
		"username":       "john_doe",
		"email":          "john@example.com",
		"password":       "secure_password_123",
		"age":            25,
		"bio":            "Software developer",
		"terms_accepted": true,
	}

	result := schema.Validate(validData)
	fmt.Printf("Valid registration: %v\n", result.Valid)

	// Invalid data
	invalidData := map[string]interface{}{
		"username":       "ab",
		"email":          "invalid-email",
		"password":       "short",
		"age":            10,
		"terms_accepted": false,
	}

	result = schema.Validate(invalidData)
	fmt.Printf("\nInvalid registration: %v\n", result.Valid)
	if !result.Valid {
		fmt.Println("Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}
}

func exampleNestedValidation() {
	fmt.Println("\n=== Example 3: Nested Validation ===")

	schema := v.NewSchema(map[string]v.Validator{
		"name":  v.String(),
		"email": v.Email(),
		"address": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"street": v.String(),
				"city":   v.String(),
				"zip":    v.String(v.StringOptions{Pattern: regexp.MustCompile(`^\d{5}$`)}),
				"country": v.String(),
			},
		}),
		"settings": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"theme":         v.String(v.StringOptions{Choices: []string{"light", "dark"}}),
				"notifications": v.Bool(),
				"language":      v.String(v.StringOptions{Choices: []string{"en", "es", "fr"}}),
			},
		}),
	})

	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"address": map[string]interface{}{
			"street":  "123 Main St",
			"city":    "New York",
			"zip":     "10001",
			"country": "USA",
		},
		"settings": map[string]interface{}{
			"theme":         "dark",
			"notifications": true,
			"language":      "en",
		},
	}

	result := schema.Validate(data)
	fmt.Printf("Nested validation: %v\n", result.Valid)
}

func exampleSliceValidation() {
	fmt.Println("\n=== Example 4: Slice Validation ===")

	// Simple slice
	tagsValidator := v.Slice(v.SliceOptions{ItemValidator: v.String()})
	result := tagsValidator.Validate([]interface{}{"go", "rust", "python"}, "tags")
	fmt.Printf("Tags validation: %v\n", result.Valid)

	// Slice of maps
	usersSchema := v.NewSchema(map[string]v.Validator{
		"users": v.Slice(v.SliceOptions{
			ItemValidator: v.Map(v.MapOptions{
				Schema: map[string]v.Validator{
					"name":  v.String(),
					"email": v.Email(),
					"age":   v.Int(v.IntOptions{MinValue: v.IntPtr(0)}),
				},
			}),
			MinLength: v.IntPtr(1),
		}),
	})

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "John", "email": "john@example.com", "age": 30},
			map[string]interface{}{"name": "Jane", "email": "jane@example.com", "age": 25},
		},
	}

	result = usersSchema.Validate(data)
	fmt.Printf("Users slice validation: %v\n", result.Valid)
}

func exampleCustomValidators() {
	fmt.Println("\n=== Example 5: Custom Validators ===")

	isEven := func(value interface{}) *string {
		if num, ok := value.(int); ok {
			if num%2 != 0 {
				msg := "Value must be even"
				return &msg
			}
		}
		return nil
	}

	isPositive := func(value interface{}) *string {
		if num, ok := value.(int); ok {
			if num <= 0 {
				msg := "Value must be positive"
				return &msg
			}
		}
		return nil
	}

	// Single custom validator
	evenValidator := v.Int().Custom(isEven)

	result := evenValidator.Validate(4, "number")
	fmt.Printf("Even number (4): %v\n", result.Valid)

	result = evenValidator.Validate(3, "number")
	fmt.Printf("Even number (3): %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("  Error: %s\n", result.Errors[0].Message)
	}

	// Multiple custom validators
	positiveEvenValidator := v.Int().Custom(isPositive).Custom(isEven)

	result = positiveEvenValidator.Validate(4, "number")
	fmt.Printf("\nPositive even (4): %v\n", result.Valid)

	result = positiveEvenValidator.Validate(-2, "number")
	fmt.Printf("Positive even (-2): %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("  Error: %s\n", result.Errors[0].Message)
	}
}

func exampleAPIValidation() {
	fmt.Println("\n=== Example 6: API Request Validation ===")

	schema := v.NewSchema(map[string]v.Validator{
		"method": v.String(v.StringOptions{Choices: []string{"GET", "POST", "PUT", "DELETE", "PATCH"}}),
		"url":    v.URL(),
		"headers": v.Map(v.MapOptions{Required: false}),
		"body":    v.Map(v.MapOptions{Required: false}),
		"timeout": v.Float(v.FloatOptions{MinValue: v.Float64Ptr(0), Required: false}),
		"retry": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"max_attempts": v.Int(v.IntOptions{MinValue: v.IntPtr(1), MaxValue: v.IntPtr(10)}),
				"backoff":      v.Float(v.FloatOptions{MinValue: v.Float64Ptr(0)}),
			},
			Required: false,
		}),
	})

	request := map[string]interface{}{
		"method": "POST",
		"url":    "https://api.example.com/users",
		"headers": map[string]interface{}{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		"body": map[string]interface{}{
			"name":  "John Doe",
			"email": "john@example.com",
		},
		"timeout": 30.0,
		"retry": map[string]interface{}{
			"max_attempts": 3,
			"backoff":      1.5,
		},
	}

	result := schema.Validate(request)
	fmt.Printf("API request validation: %v\n", result.Valid)
}

func exampleConfigValidation() {
	fmt.Println("\n=== Example 7: Configuration Validation ===")

	schema := v.NewSchema(map[string]v.Validator{
		"app": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"name":    v.String(),
				"version": v.String(v.StringOptions{Pattern: regexp.MustCompile(`^\d+\.\d+\.\d+$`)}),
				"debug":   v.Bool(),
			},
		}),
		"database": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"host":      v.String(),
				"port":      v.Int(v.IntOptions{MinValue: v.IntPtr(1), MaxValue: v.IntPtr(65535)}),
				"name":      v.String(),
				"username":  v.String(),
				"password":  v.String(),
				"pool_size": v.Int(v.IntOptions{MinValue: v.IntPtr(1), MaxValue: v.IntPtr(100)}),
				"ssl":       v.Bool(),
			},
		}),
		"cache": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"enabled":  v.Bool(),
				"backend":  v.String(v.StringOptions{Choices: []string{"redis", "memcached", "memory"}}),
				"ttl":      v.Int(v.IntOptions{MinValue: v.IntPtr(0)}),
				"max_size": v.Int(v.IntOptions{MinValue: v.IntPtr(1)}),
			},
		}),
		"logging": v.Map(v.MapOptions{
			Schema: map[string]v.Validator{
				"level":  v.String(v.StringOptions{Choices: []string{"DEBUG", "INFO", "WARNING", "ERROR"}}),
				"format": v.String(v.StringOptions{Choices: []string{"json", "text"}}),
				"output": v.String(v.StringOptions{Choices: []string{"stdout", "file"}}),
			},
		}),
		"features": v.Slice(v.SliceOptions{ItemValidator: v.String()}),
	})

	config := map[string]interface{}{
		"app": map[string]interface{}{
			"name":    "MyApp",
			"version": "1.0.0",
			"debug":   false,
		},
		"database": map[string]interface{}{
			"host":      "localhost",
			"port":      5432,
			"name":      "myapp",
			"username":  "admin",
			"password":  "secret",
			"pool_size": 10,
			"ssl":       true,
		},
		"cache": map[string]interface{}{
			"enabled":  true,
			"backend":  "redis",
			"ttl":      3600,
			"max_size": 1000,
		},
		"logging": map[string]interface{}{
			"level":  "INFO",
			"format": "json",
			"output": "stdout",
		},
		"features": []interface{}{"feature1", "feature2", "feature3"},
	}

	result := schema.Validate(config)
	fmt.Printf("Configuration validation: %v\n", result.Valid)
}

func exampleErrorHandling() {
	fmt.Println("\n=== Example 8: Error Handling ===")

	schema := v.NewSchema(map[string]v.Validator{
		"email": v.Email(),
		"age":   v.Int(v.IntOptions{MinValue: v.IntPtr(0)}),
	})

	// Option 1: Check result
	fmt.Println("Option 1: Check ValidationResult")
	result := schema.Validate(map[string]interface{}{"email": "invalid", "age": -1})
	if !result.Valid {
		fmt.Println("Validation failed:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}

	// Option 2: Panic on error
	fmt.Println("\nOption 2: Panic on error")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Caught panic: %v\n", r)
		}
	}()
	schema.ValidateOrPanic(map[string]interface{}{"email": "invalid", "age": -1})
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Universal Data Validator - Examples")
	fmt.Println(strings.Repeat("=", 60))

	exampleBasicValidation()
	exampleUserRegistration()
	exampleNestedValidation()
	exampleSliceValidation()
	exampleCustomValidators()
	exampleAPIValidation()
	exampleConfigValidation()
	exampleErrorHandling()

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Examples completed!")
	fmt.Println(strings.Repeat("=", 60))
}

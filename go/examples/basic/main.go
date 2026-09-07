package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	v "github.com/parthivrawat/universal-validator/go/v2"
)

func exampleBasicValidation() {
	fmt.Println("\n=== Example 1: Basic Validation ===")

	// String validation
	nameValidator := v.String(v.MinLength(2), v.MaxLength(50))
	result := nameValidator.Validate(context.Background(), "John Doe", "name")
	fmt.Printf("Name validation: %v\n", result.Valid)

	// Integer validation
	ageValidator := v.Int(v.MinValue(0), v.MaxValue(120))
	result = ageValidator.Validate(context.Background(), 25, "age")
	fmt.Printf("Age validation: %v\n", result.Valid)

	// Email validation
	emailValidator := v.Email()
	result = emailValidator.Validate(context.Background(), "john@example.com", "email")
	fmt.Printf("Email validation: %v\n", result.Valid)
}

func exampleUserRegistration() {
	fmt.Println("\n=== Example 2: User Registration ===")

	schema := v.NewSchema(map[string]v.Validator{
		"username":       v.String(v.MinLength(3), v.MaxLength(20)),
		"email":          v.Email(),
		"password":       v.String(v.MinLength(8)),
		"age":            v.Int(v.MinValue(13), v.Optional()),
		"bio":            v.String(v.MaxLength(500), v.Optional()),
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

	result := schema.Validate(context.Background(), validData)
	fmt.Printf("Valid registration: %v\n", result.Valid)

	// Invalid data
	invalidData := map[string]interface{}{
		"username":       "ab",
		"email":          "invalid-email",
		"password":       "short",
		"age":            10,
		"terms_accepted": false,
	}

	result = schema.Validate(context.Background(), invalidData)
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
		"address": v.Map(v.Fields(map[string]v.Validator{
			"street":  v.String(),
			"city":    v.String(),
			"zip":     v.Regex(regexp.MustCompile(`^\d{5}$`)),
			"country": v.String(),
		})),
		"settings": v.Map(v.Fields(map[string]v.Validator{
			"theme":         v.String(v.Choices("light", "dark")),
			"notifications": v.Bool(),
			"language":      v.String(v.Choices("en", "es", "fr")),
		})),
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

	result := schema.Validate(context.Background(), data)
	fmt.Printf("Nested validation: %v\n", result.Valid)
}

func exampleSliceValidation() {
	fmt.Println("\n=== Example 4: Slice Validation ===")

	// Simple slice
	tagsValidator := v.Slice(v.Items(v.String()))
	result := tagsValidator.Validate(context.Background(), []interface{}{"go", "rust", "python"}, "tags")
	fmt.Printf("Tags validation: %v\n", result.Valid)

	// Slice of maps
	usersSchema := v.NewSchema(map[string]v.Validator{
		"users": v.Slice(
			v.Items(v.Map(v.Fields(map[string]v.Validator{
				"name":  v.String(),
				"email": v.Email(),
				"age":   v.Int(v.MinValue(0)),
			}))),
			v.MinLength(1),
		),
	})

	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"name": "John", "email": "john@example.com", "age": 30},
			map[string]interface{}{"name": "Jane", "email": "jane@example.com", "age": 25},
		},
	}

	result = usersSchema.Validate(context.Background(), data)
	fmt.Printf("Users slice validation: %v\n", result.Valid)
}

func exampleCustomValidators() {
	fmt.Println("\n=== Example 5: Custom Validators ===")

	isEven := func(ctx context.Context, value interface{}) error {
		if num, ok := value.(int); ok {
			if num%2 != 0 {
				return fmt.Errorf("Value must be even")
			}
		}
		return nil
	}

	isPositive := func(ctx context.Context, value interface{}) error {
		if num, ok := value.(int); ok {
			if num <= 0 {
				return fmt.Errorf("Value must be positive")
			}
		}
		return nil
	}

	// Single custom validator
	evenValidator := v.Int(v.WithCustom(isEven))

	result := evenValidator.Validate(context.Background(), 4, "number")
	fmt.Printf("Even number (4): %v\n", result.Valid)

	result = evenValidator.Validate(context.Background(), 3, "number")
	fmt.Printf("Even number (3): %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("  Error: %s\n", result.Errors[0].Message)
	}

	// Multiple custom validators
	positiveEvenValidator := v.Int(v.WithCustom(isPositive), v.WithCustom(isEven))

	result = positiveEvenValidator.Validate(context.Background(), 4, "number")
	fmt.Printf("\nPositive even (4): %v\n", result.Valid)

	result = positiveEvenValidator.Validate(context.Background(), -2, "number")
	fmt.Printf("Positive even (-2): %v\n", result.Valid)
	if !result.Valid {
		fmt.Printf("  Error: %s\n", result.Errors[0].Message)
	}
}

func exampleAPIValidation() {
	fmt.Println("\n=== Example 6: API Request Validation ===")

	schema := v.NewSchema(map[string]v.Validator{
		"method":  v.String(v.Choices("GET", "POST", "PUT", "DELETE", "PATCH")),
		"url":     v.URL(),
		"headers": v.Map(v.Optional()),
		"body":    v.Map(v.Optional()),
		"timeout": v.Float(v.MinFloat(0), v.Optional()),
		"retry": v.Map(
			v.Optional(),
			v.Fields(map[string]v.Validator{
				"max_attempts": v.Int(v.MinValue(1), v.MaxValue(10)),
				"backoff":      v.Float(v.MinFloat(0)),
			}),
		),
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

	result := schema.Validate(context.Background(), request)
	fmt.Printf("API request validation: %v\n", result.Valid)
}

func exampleConfigValidation() {
	fmt.Println("\n=== Example 7: Configuration Validation ===")

	schema := v.NewSchema(map[string]v.Validator{
		"app": v.Map(v.Fields(map[string]v.Validator{
			"name":    v.String(),
			"version": v.Regex(regexp.MustCompile(`^\d+\.\d+\.\d+$`)),
			"debug":   v.Bool(),
		})),
		"database": v.Map(v.Fields(map[string]v.Validator{
			"host":      v.String(),
			"port":      v.Int(v.MinValue(1), v.MaxValue(65535)),
			"name":      v.String(),
			"username":  v.String(),
			"password":  v.String(),
			"pool_size": v.Int(v.MinValue(1), v.MaxValue(100)),
			"ssl":       v.Bool(),
		})),
		"cache": v.Map(v.Fields(map[string]v.Validator{
			"enabled":  v.Bool(),
			"backend":  v.String(v.Choices("redis", "memcached", "memory")),
			"ttl":      v.Int(v.MinValue(0)),
			"max_size": v.Int(v.MinValue(1)),
		})),
		"logging": v.Map(v.Fields(map[string]v.Validator{
			"level":  v.String(v.Choices("DEBUG", "INFO", "WARNING", "ERROR")),
			"format": v.String(v.Choices("json", "text")),
			"output": v.String(v.Choices("stdout", "file")),
		})),
		"features": v.Slice(v.Items(v.String())),
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

	result := schema.Validate(context.Background(), config)
	fmt.Printf("Configuration validation: %v\n", result.Valid)
}

func exampleErrorHandling() {
	fmt.Println("\n=== Example 8: Error Handling ===")

	schema := v.NewSchema(map[string]v.Validator{
		"email": v.Email(),
		"age":   v.Int(v.MinValue(0)),
	})

	// Option 1: Check result
	fmt.Println("Option 1: Check ValidationResult")
	result := schema.Validate(context.Background(), map[string]interface{}{"email": "invalid", "age": -1})
	if !result.Valid {
		fmt.Println("Validation failed:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s: %s\n", err.Field, err.Message)
		}
	}

	// Option 2: Return error
	fmt.Println("\nOption 2: ValidateOrError")
	if err := schema.ValidateOrError(context.Background(), map[string]interface{}{"email": "invalid", "age": -1}); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	}
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

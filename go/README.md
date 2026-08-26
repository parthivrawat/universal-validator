# Universal Data Validator (Go)

A comprehensive data validation library for Go that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, Int, Float, Bool, Email, URL, Slice, Map
- ✅ **Schema-Based Validation**: Define complex data structures
- ✅ **Custom Validators**: Add your own validation logic
- ✅ **Nested Validation**: Validate nested objects and slices
- ✅ **Clear Error Messages**: Detailed error reporting with field paths
- ✅ **Zero Dependencies**: No external dependencies required
- ✅ **Production Ready**: Comprehensive test coverage

## Installation

```bash
go get github.com/parthivrawat/universal-validator/go
```

## Quick Start

```go
package main

import (
    "fmt"
    v "github.com/parthivrawat/universal-validator/go"
)

func main() {
    schema := v.NewSchema(map[string]v.Validator{
        "email": v.Email(),
        "age":   v.Int(v.IntOptions{MinValue: v.IntPtr(0), MaxValue: v.IntPtr(120)}),
        "username": v.String(v.StringOptions{MinLength: v.IntPtr(3), MaxLength: v.IntPtr(20)}),
    })

    data := map[string]interface{}{
        "email":    "user@example.com",
        "age":      25,
        "username": "john_doe",
    }

    result := schema.Validate(data)
    if !result.Valid {
        for _, err := range result.Errors {
            fmt.Printf("%s: %s\n", err.Field, err.Message)
        }
    }
}
```

## Usage Examples

### String Validation

```go
import v "github.com/parthivrawat/universal-validator/go"

// Basic string
validator := v.String()

// String with length constraints
usernameValidator := v.String(v.StringOptions{
    MinLength: v.IntPtr(3),
    MaxLength: v.IntPtr(20),
})

// String with pattern
phoneValidator := v.String(v.StringOptions{
    Pattern: regexp.MustCompile(`^\d{3}-\d{4}$`),
})

// String with choices
themeValidator := v.String(v.StringOptions{
    Choices: []string{"light", "dark"},
})

// Optional string
bioValidator := v.String(v.StringOptions{Required: false})
```

### Integer and Float Validation

```go
// Integer with range
ageValidator := v.Int(v.IntOptions{
    MinValue: v.IntPtr(0),
    MaxValue: v.IntPtr(120),
})

// Float with range
priceValidator := v.Float(v.FloatOptions{
    MinValue: v.Float64Ptr(0.0),
    MaxValue: v.Float64Ptr(9999.99),
})

// Optional integer
scoreValidator := v.Int(v.IntOptions{Required: false})
```

### Boolean Validation

```go
termsValidator := v.Bool()
newsletterValidator := v.Bool(v.BoolOptions{Required: false})
```

### Email and URL Validation

```go
emailValidator := v.Email()
urlValidator := v.URL()
```

### Slice Validation

```go
// Simple slice
tagsValidator := v.Slice()

// Slice with item validation
numbersValidator := v.Slice(v.SliceOptions{
    ItemValidator: v.Int(),
})

// Slice with length constraints
itemsValidator := v.Slice(v.SliceOptions{
    ItemValidator: v.String(),
    MinLength:     v.IntPtr(1),
    MaxLength:     v.IntPtr(10),
})
```

### Map Validation

```go
// Map with schema
addressValidator := v.Map(v.MapOptions{
    Schema: map[string]v.Validator{
        "street": v.String(),
        "city":   v.String(),
        "zip":    v.String(v.StringOptions{Pattern: regexp.MustCompile(`^\d{5}$`)}),
    },
})

// Nested map
userValidator := v.Map(v.MapOptions{
    Schema: map[string]v.Validator{
        "name":  v.String(),
        "email": v.Email(),
        "address": v.Map(v.MapOptions{
            Schema: map[string]v.Validator{
                "street": v.String(),
                "city":   v.String(),
            },
        }),
    },
})
```

### Schema Validation

```go
import v "github.com/parthivrawat/universal-validator/go"

schema := v.NewSchema(map[string]v.Validator{
    "username": v.String(v.StringOptions{MinLength: v.IntPtr(3), MaxLength: v.IntPtr(20)}),
    "email":    v.Email(),
    "age":      v.Int(v.IntOptions{MinValue: v.IntPtr(13), Required: false}),
    "bio":      v.String(v.StringOptions{MaxLength: v.IntPtr(500), Required: false}),
    "tags":     v.Slice(v.SliceOptions{ItemValidator: v.String()}),
    "settings": v.Map(v.MapOptions{
        Schema: map[string]v.Validator{
            "theme":         v.String(v.StringOptions{Choices: []string{"light", "dark"}}),
            "notifications": v.Bool(),
        },
    }),
})

data := map[string]interface{}{
    "username": "john_doe",
    "email":    "john@example.com",
    "age":      25,
    "tags":     []interface{}{"go", "programming"},
    "settings": map[string]interface{}{
        "theme":         "dark",
        "notifications": true,
    },
}

result := schema.Validate(data)

if result.Valid {
    fmt.Println("✅ Data is valid!")
} else {
    fmt.Println("❌ Validation errors:")
    for _, err := range result.Errors {
        fmt.Printf("  %s: %s\n", err.Field, err.Message)
    }
}
```

### Custom Validators

```go
isEven := func(value interface{}) *string {
    if num, ok := value.(int); ok {
        if num%2 != 0 {
            msg := "Value must be even"
            return &msg
        }
    }
    return nil
}

validator := v.Int().Custom(isEven)

result := validator.Validate(4, "number")
fmt.Println(result.Valid) // true

result = validator.Validate(3, "number")
fmt.Println(result.Valid) // false
```

## Real-World Examples

### User Registration

```go
schema := v.NewSchema(map[string]v.Validator{
    "username": v.String(v.StringOptions{MinLength: v.IntPtr(3), MaxLength: v.IntPtr(20)}),
    "email":    v.Email(),
    "password": v.String(v.StringOptions{MinLength: v.IntPtr(8)}),
    "age":      v.Int(v.IntOptions{MinValue: v.IntPtr(13), Required: false}),
    "terms_accepted": v.Bool(),
})
```

### API Request Validation

```go
schema := v.NewSchema(map[string]v.Validator{
    "method": v.String(v.StringOptions{Choices: []string{"GET", "POST", "PUT", "DELETE"}}),
    "url":    v.URL(),
    "headers": v.Map(v.MapOptions{Required: false}),
    "body":    v.Map(v.MapOptions{Required: false}),
    "timeout": v.Float(v.FloatOptions{MinValue: v.Float64Ptr(0), Required: false}),
})
```

### Configuration Validation

```go
schema := v.NewSchema(map[string]v.Validator{
    "database": v.Map(v.MapOptions{
        Schema: map[string]v.Validator{
            "host":     v.String(),
            "port":     v.Int(v.IntOptions{MinValue: v.IntPtr(1), MaxValue: v.IntPtr(65535)}),
            "username": v.String(),
            "password": v.String(),
            "ssl":      v.Bool(),
        },
    }),
    "cache": v.Map(v.MapOptions{
        Schema: map[string]v.Validator{
            "enabled":  v.Bool(),
            "ttl":      v.Int(v.IntOptions{MinValue: v.IntPtr(0)}),
            "max_size": v.Int(v.IntOptions{MinValue: v.IntPtr(1)}),
        },
    }),
    "features": v.Slice(v.SliceOptions{ItemValidator: v.String()}),
})
```

## Error Handling

```go
schema := v.NewSchema(map[string]v.Validator{
    "email": v.Email(),
})

// Option 1: Check result
result := schema.Validate(map[string]interface{}{"email": "invalid"})
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("%s: %s\n", err.Field, err.Message)
    }
}

// Option 2: Panic on error
defer func() {
    if r := recover(); r != nil {
        fmt.Printf("Validation failed: %v\n", r)
    }
}()
schema.ValidateOrPanic(map[string]interface{}{"email": "invalid"})
```

## API Reference

### Validators

- `String(opts...)` - String validator
- `Int(opts...)` - Integer validator
- `Float(opts...)` - Float validator
- `Bool(opts...)` - Boolean validator
- `Email(opts...)` - Email validator
- `URL(opts...)` - URL validator
- `Slice(opts...)` - Slice validator
- `Map(opts...)` - Map validator

### Schema

- `NewSchema(validators)` - Create a schema
- `schema.Validate(data)` - Validate data and return ValidationResult
- `schema.ValidateOrPanic(data)` - Validate data and panic if invalid

### Helper Functions

- `IntPtr(i)` - Create pointer to int
- `Float64Ptr(f)` - Create pointer to float64

## Testing

```bash
# Run tests
go test -v

# Run tests with coverage
go test -v -cover

# Generate coverage report
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

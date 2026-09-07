# Universal Data Validator (Go)

A comprehensive data validation library for Go that works across API, database, and form contexts.

## Features

- **Rich Validator Types**: String, Int, Float, Bool, Email, URL, Slice, Map, UUID, Date, DateTime, Numeric, NotEmpty, Enum, OneOf, Regex
- **Schema-Based Validation**: Define complex data structures
- **Functional Options**: One `Option` type shared across all validators
- **Context-Aware Custom Validators**: Receive `context.Context` and return `error`
- **Nested Validation**: Validate nested objects and slices
- **Clear Error Messages**: Detailed error reporting with field paths
- **Error Aggregation**: `ValidationErrors` implementing `error` and `Unwrap() []error`
- **Immutable Validators**: Safe to share across goroutines
- **Zero Dependencies**: No external dependencies required
- **Production Ready**: Comprehensive test coverage

## Installation

```bash
go get github.com/parthivrawat/universal-validator/go/v2
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    v "github.com/parthivrawat/universal-validator/go/v2"
)

func main() {
    schema := v.NewSchema(map[string]v.Validator{
        "email":    v.Email(),
        "age":      v.Int(v.MinValue(0), v.MaxValue(120)),
        "username": v.String(v.MinLength(3), v.MaxLength(20)),
    })

    data := map[string]interface{}{
        "email":    "user@example.com",
        "age":      25,
        "username": "john_doe",
    }

    result := schema.Validate(context.Background(), data)
    if !result.Valid {
        for _, err := range result.Errors {
            fmt.Printf("%s: %s\n", err.Field, err.Message)
        }
    }
}
```

## Functional Options

All validator constructors accept `...Option`. The same `Option` type is used for every validator; options that do not apply to a particular validator are ignored. Options are applied at construction time; validators are immutable and safe to share across goroutines.

```go
Optional()                  // required=false (required is the default)
Nullable()                  // allow null/nil values
MinLength(int)              // string min length
MaxLength(int)              // string max length
Pattern(*regexp.Regexp)     // full regex match for strings
Choices(...string)          // allowed string values
MinValue(int)               // integer min value
MaxValue(int)               // integer max value
MinFloat(float64)           // float min value
MaxFloat(float64)           // float max value
Items(Validator)            // item validator for slices
Fields(map[string]Validator) // field validators for maps/objects
Strict()                    // unknown-key detection for maps/schemata
FailFast()                  // stop at the first error for maps/schemata
WithCustom(CustomValidator) // attach a custom validator
Sensitive()                 // redact the offending value from errors
```

## Usage Examples

### String Validation

```go
import v "github.com/parthivrawat/universal-validator/go/v2"

// Basic string
validator := v.String()

// String with length constraints
usernameValidator := v.String(v.MinLength(3), v.MaxLength(20))

// String with pattern
phoneValidator := v.String(v.Pattern(regexp.MustCompile(`^\d{3}-\d{4}$`)))

// String with choices
themeValidator := v.String(v.Choices("light", "dark"))

// Optional string
bioValidator := v.String(v.Optional())
```

### Integer and Float Validation

```go
// Integer with range
ageValidator := v.Int(v.MinValue(0), v.MaxValue(120))

// Float with range
priceValidator := v.Float(v.MinFloat(0.0), v.MaxFloat(9999.99))

// Optional integer
scoreValidator := v.Int(v.Optional())
```

### Boolean, Email, URL

```go
termsValidator := v.Bool()
newsletterValidator := v.Bool(v.Optional())
emailValidator := v.Email()
urlValidator := v.URL()
```

### Slice and Map Validation

```go
// Slice with item validation
numbersValidator := v.Slice(v.Items(v.Int()))

// Slice with length constraints
itemsValidator := v.Slice(
    v.Items(v.String()),
    v.MinLength(1),
    v.MaxLength(10),
)

// Map with schema
addressValidator := v.Map(v.Fields(map[string]v.Validator{
    "street": v.String(),
    "city":   v.String(),
    "zip":    v.Regex(regexp.MustCompile(`^\d{5}$`)),
}))
```

### New Validators

```go
// UUID
code := v.UUID()

// Date and datetime (pattern only; no calendar normalization)
dateValidator := v.Date()
datetimeValidator := v.DateTime()

// Numeric string (e.g. "-12.5", "1e3")
numericValidator := v.Numeric()

// Non-empty string, array, or object
notEmpty := v.NotEmpty()

// Enum with deep-equality
enumValidator := v.Enum([]interface{}{"a", 1, true})

// OneOf: value must satisfy at least one branch
oneOf := v.OneOf([]v.Validator{v.String(v.MinLength(3)), v.Int(v.MinValue(10))})

// Regex: string validator with a caller-supplied pattern
regexValidator := v.Regex(regexp.MustCompile(`^[A-Z]{3}$`))
```

### Schema Validation

```go
schema := v.NewSchema(map[string]v.Validator{
    "username": v.String(v.MinLength(3), v.MaxLength(20)),
    "email":    v.Email(),
    "age":      v.Int(v.MinValue(13), v.Optional()),
    "bio":      v.String(v.MaxLength(500), v.Optional()),
    "tags":     v.Slice(v.Items(v.String())),
    "settings": v.Map(v.Fields(map[string]v.Validator{
        "theme":         v.String(v.Choices("light", "dark")),
        "notifications": v.Bool(),
    })),
})
```

### Strict Mode (Unknown-Key Detection)

```go
schema := v.NewSchema(map[string]v.Validator{
    "name": v.String(),
}, v.Strict())

// Or toggle after construction
schema.SetStrict(true)
```

For nested maps:

```go
userSchema := v.NewSchema(map[string]v.Validator{
    "address": v.Map(
        v.Strict(),
        v.Fields(map[string]v.Validator{
            "city": v.String(),
        }),
    ),
})
```

### Fail-Fast Mode

```go
schema := v.NewSchema(map[string]v.Validator{
    "name": v.String(),
    "age":  v.Int(v.MinValue(0)),
}, v.FailFast())

// Or toggle after construction
schema.SetFailFast(true)
```

Fail-fast also works on `Map` and `Slice`:

```go
userValidator := v.Map(
    v.FailFast(),
    v.Fields(map[string]v.Validator{
        "name": v.String(),
        "age":  v.Int(),
    }),
)
```

### Context-Aware Custom Validators

```go
isEven := func(ctx context.Context, value interface{}) error {
    if num, ok := value.(int); ok && num%2 != 0 {
        return fmt.Errorf("Value must be even")
    }
    return nil
}

validator := v.Int(v.WithCustom(isEven))
```

### Error Handling

`Schema.ValidateOrError` returns a `ValidationErrors` value (or `nil` on success). `ValidationErrors` joins `field: message` with newlines and supports `errors.Unwrap`:

```go
if err := schema.ValidateOrError(context.Background(), data); err != nil {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        for _, e := range ve {
            fmt.Printf("%s: %s\n", e.Field, e.Message)
        }
    }
}
```

### Error Codes

Each `ValidationError` carries a machine-readable `Code`:

- `validator.ErrCodeRequired`
- `validator.ErrCodeUnknownField`
- `validator.ErrCodeType`
- `validator.ErrCodeMinLength`
- `validator.ErrCodeMaxLength`
- `validator.ErrCodeMinValue`
- `validator.ErrCodeMaxValue`
- `validator.ErrCodePattern`
- `validator.ErrCodeChoices`
- `validator.ErrCodeEmail`
- `validator.ErrCodeURL`
- `validator.ErrCodeUUID`
- `validator.ErrCodeDate`
- `validator.ErrCodeDateTime`
- `validator.ErrCodeNumeric`
- `validator.ErrCodeNotEmpty`
- `validator.ErrCodeEnum`
- `validator.ErrCodeOneOf`
- `validator.ErrCodeCustom`

## Error Path Format

Nested field paths use `.` (e.g. `user.address.city`) and array indices use brackets (e.g. `tags[0]`, `users[1].email`). Top-level type errors on a schema use `root`.

## No Coercion / No Mutation

Validation is read-only. It never trims strings, folds case, parses strings into numbers, or otherwise mutates the input. Normalize data before validating if your domain requires it.

## Security

### Sensitive fields

Mark a validator with `Sensitive()` when it handles secrets (passwords, tokens, API keys). A sensitive validator redacts the offending value from every error it produces: `ValidationError.Value` is `nil` and any place the value would be interpolated into the message renders as the literal `***`:

```go
v.String(v.Sensitive(), v.Choices("a", "b")).Validate(ctx, "hunter2", "token")
// => token: Value must be one of a, b, got '***'   (Value == nil)

v.Email(v.Sensitive()).Validate(ctx, "p@ssw0rd", "secret")
// => secret: Invalid email address: ***
```

Messages that do not embed the input — length and value bounds (`got 3`), `Field is required`, `Unknown field` — are unchanged. The flag is **not** inherited by nested validators (`Items`, `Fields`, schema fields); set it on each validator that needs it.

### Pattern input length cap

Pattern checks are skipped for input strings longer than `MaxPatternInputLength` (10,000 bytes) and immediately produce the normal `pattern` error (rewritten to the friendly `email`/`url`/`uuid`/`date`/`datetime`/`numeric` messages where applicable). Go's regexp engine (RE2) is linear-time, so this cap exists for conformance parity with sibling implementations that use backtracking engines — it is not required for safety in Go.

### Trusted patterns only

Patterns supplied via `Pattern`/`Regex` are executed by the platform regex engine. While Go's RE2 cannot backtrack exponentially, the same schema can be catastrophic in runtimes that use backtracking engines — patterns must come from trusted sources, never from end users.

### Email/URL are sanity checks only

The built-in email pattern is a permissive sanity check: it does not enforce RFC 5321 length limits and cannot prove an address is deliverable. The URL pattern performs no IDN or port validation and accepts inputs such as `http://x.`. Do not rely on either for security decisions.

## Concurrency

Validators are immutable after construction and may be shared freely across goroutines. `SetStrict`/`SetFailFast` on `Schema` mutate the schema and are not safe for concurrent use with `Validate`.

## Testing

```bash
# Run tests
go test -v

# Run tests with coverage
go test -v -cover
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

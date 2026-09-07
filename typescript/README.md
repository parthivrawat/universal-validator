# Universal Data Validator (TypeScript)

A comprehensive data validation library for TypeScript that works across API, database, and form contexts.

## Features

- Rich Validator Types: String, number, int, boolean, email, URL, UUID, date, datetime, numeric, not-empty, enum, one-of, array, object, and regex
- Schema-Based Validation: Define complex data structures
- Custom Validators: Add your own validation logic
- Nested Validation: Validate nested objects and arrays
- Clear Error Messages: Detailed error reporting with field paths
- Type Safety: Full TypeScript support with generic validators
- Zero Dependencies: No external dependencies required
- Production Ready: Comprehensive test coverage

## Installation

```bash
npm install universal-validator
# or
yarn add universal-validator
# or
pnpm add universal-validator
```

## Quick Start

```typescript
import { Schema, validators } from 'universal-validator';

const schema = new Schema({
  email: validators.email(),
  age: validators.int({ minValue: 0, maxValue: 120 }),
  username: validators.string({ minLength: 3, maxLength: 20 }),
  tags: validators.array({ itemValidator: validators.string() }),
});

const result = schema.validate(data);
if (!result.valid) {
  for (const error of result.errors) {
    console.log(`${error.field}: ${error.message}`);
  }
}
```

## Options objects

All validator classes accept a single options object. Factory helpers keep the same friendly signatures but forward options objects internally.

```typescript
// Options object
const v1 = new StringValidator({ minLength: 2, maxLength: 10, required: true });

// Equivalent factory helper
const v2 = validators.string({ minLength: 2, maxLength: 10, required: true });
```

Shared base options (`required`, `nullable`, `failFast`, `sensitive`) default to `true`, `false`, `false`, `false` respectively.

## Usage Examples

### String Validation

```typescript
import { validators } from 'universal-validator';

// Basic string
const validator = validators.string();

// String with length constraints
const usernameValidator = validators.string({ minLength: 3, maxLength: 20 });

// String with pattern
const phoneValidator = validators.string({ pattern: /^[0-9]{3}-[0-9]{4}$/ });

// String with choices
const themeValidator = validators.string({ choices: ['light', 'dark'] });

// Optional string
const bioValidator = validators.string({ required: false });
```

### Number and Integer Validation

```typescript
// Number with range
const priceValidator = validators.number({ minValue: 0, maxValue: 9999.99 });

// Integer with range
const ageValidator = validators.int({ minValue: 0, maxValue: 120 });

// Optional number
const scoreValidator = validators.number({ required: false });
```

### Boolean Validation

```typescript
const termsValidator = validators.boolean();
const newsletterValidator = validators.boolean({ required: false });
```

### Email and URL Validation

```typescript
const emailValidator = validators.email();
const urlValidator = validators.url();
```

### UUID, Date, Datetime, and Numeric Validation

```typescript
const uuidValidator = validators.uuid();
const dateValidator = validators.date();
const datetimeValidator = validators.datetime();
const numericValidator = validators.numeric();
```

### Not-Empty Validation

Accepts strings, arrays, and plain objects. Rejects other kinds and any empty container.

```typescript
const notEmpty = validators.notEmpty();
notEmpty.validate('', 'field');      // invalid - empty string
notEmpty.validate([], 'field');      // invalid - empty array
notEmpty.validate({}, 'field');      // invalid - empty object
notEmpty.validate(5, 'field');       // invalid - not a container
```

### Enum Validation

Performs deep-equality membership against any JSON values.

```typescript
const enumValidator = validators.enum({ values: ['a', 1, true] });
enumValidator.validate('a', 'field');   // valid
enumValidator.validate('b', 'field');   // invalid
```

### One-Of Validation

Valid if any branch produces zero errors at the same field path.

```typescript
const oneOf = validators.oneOf(
  validators.string({ minLength: 3 }),
  validators.int({ minValue: 10 })
);
oneOf.validate('abc', 'field'); // valid
oneOf.validate(15, 'field');    // valid
oneOf.validate(5, 'field');     // invalid
```

You can also pass an options object:

```typescript
const oneOf = validators.oneOf({
  validators: [validators.string(), validators.int()],
});
```

### Regex Factory

Creates a string validator whose value must fully match the supplied pattern.

```typescript
const pin = validators.regex(/^[0-9]{4}$/);
const pin2 = validators.regex('^[0-9]{4}$');
```

### Array Validation

```typescript
// Simple array
const tagsValidator = validators.array();

// Array with item validation
const numbersValidator = validators.array({ itemValidator: validators.int() });

// Array with length constraints
const itemsValidator = validators.array({
  itemValidator: validators.string(),
  minLength: 1,
  maxLength: 10,
});
```

### Object Validation

```typescript
// Object with schema
const addressValidator = validators.object({
  schema: {
    street: validators.string(),
    city: validators.string(),
    zip: validators.string({ pattern: /^[0-9]{5}$/ }),
  },
});

// Nested object
const userValidator = validators.object({
  schema: {
    name: validators.string(),
    email: validators.email(),
    address: validators.object({
      schema: {
        street: validators.string(),
        city: validators.string(),
      },
    }),
  },
});
```

### Schema Validation

```typescript
import { Schema, validators } from 'universal-validator';

const userSchema = new Schema({
  username: validators.string({ minLength: 3, maxLength: 20 }),
  email: validators.email(),
  age: validators.int({ minValue: 13, required: false }),
  bio: validators.string({ maxLength: 500, required: false }),
  tags: validators.array({ itemValidator: validators.string() }),
  settings: validators.object({
    schema: {
      theme: validators.string({ choices: ['light', 'dark'] }),
      notifications: validators.boolean(),
    },
  }),
});

const data = {
  username: 'john_doe',
  email: 'john@example.com',
  age: 25,
  tags: ['typescript', 'javascript'],
  settings: {
    theme: 'dark',
    notifications: true,
  },
};

const result = userSchema.validate(data);

if (result.valid) {
  console.log('Data is valid!');
} else {
  console.log('Validation errors:');
  for (const error of result.errors) {
    console.log(`  ${error.field}: ${error.message}`);
  }
}
```

### Strict Mode (Unknown-Key Detection)

By default, keys present in the data but not in the schema are ignored. Enable strict mode to report them as `Unknown field` errors:

```typescript
// Top-level strict mode
const strictSchema = new Schema(
  { name: validators.string() },
  { strict: true }
);

strictSchema.validate({ name: 'John', extra: 1 });
// => error on field 'extra': 'Unknown field'

// Nested objects: strict per object validator
const nestedSchema = new Schema({
  user: validators.object({
    schema: { name: validators.string() },
    strict: true,
  }),
});

nestedSchema.validate({ user: { name: 'John', extra: 1 } });
// => error on field 'user.extra': 'Unknown field'
```

### Fail-Fast Mode

By default, validation collects all errors. Enable `failFast` to stop at the first error produced. The returned `ValidationResult` contains exactly one error.

```typescript
// Schema-level fail-fast
const fastSchema = new Schema(
  {
    username: validators.string({ minLength: 3 }),
    email: validators.email(),
  },
  { failFast: true }
);

const result = fastSchema.validate({ username: 'ab', email: 'bad' });
result.errors.length === 1; // only the first error is reported

// Per-validator fail-fast on objects and arrays
const obj = validators.object({ schema: { name: validators.string() }, failFast: true });
const arr = validators.array({ itemValidator: validators.int(), failFast: true });
```

The flag propagates automatically into nested child validators: a `failFast` `Schema` or `ObjectValidator`/`ArrayValidator` stops validating fields/items after the first failure and sets the flag on each child validator before running it, so a nested field also yields at most one error. `failFast` is exposed as a public `failFast` property on `Validator` (default `false`).

### Custom Validators

`.custom()` is immutable: it returns a cloned validator with the custom rule appended and leaves the original unchanged.

```typescript
const isEven = (value: number) => {
  return value % 2 !== 0 ? 'Value must be even' : null;
};

const base = validators.int();
const validator = base.custom(isEven);

console.log(base.validate(3, 'number').valid);    // true - base untouched
console.log(validator.validate(4, 'number').valid); // true
console.log(validator.validate(3, 'number').valid); // false
```

## Security

### Redacting secrets with `sensitive`

Mark a validator `sensitive: true` for fields that may contain secrets (passwords, tokens, keys). The offending value is then redacted: every message that would interpolate it renders the literal `***` instead, and `error.value` is `undefined`, so errors can be logged safely.

```typescript
const schema = new Schema({
  password: validators.string({ minLength: 8, sensitive: true }),
  api_token: validators.string({ choices: ['tok_a', 'tok_b'], sensitive: true }),
});

const result = schema.validate({ password: 'short', api_token: 'hunter2' });
// password:  'password: String length must be at least 8, got 5'  (length messages never embed the input)
// api_token: "api_token: Value must be one of tok_a, tok_b, got '***'"
```

`sensitive` is per-validator and is **not** inherited by nested validators (unlike `failFast`): an inner `email` inside a `sensitive` object validator still reports the raw value unless it is itself marked sensitive.

### Pattern input length cap

Before running any `pattern` check — including the fixed email, URL, UUID, date, datetime, and numeric patterns — inputs longer than `MAX_PATTERN_INPUT_LENGTH` (10,000 UTF-16 code units, i.e. `string.length`) skip the regex entirely and report the normal pattern error. JavaScript's `RegExp` is a backtracking engine, so this bound protects against ReDoS-style worst-case match costs on long inputs.

```typescript
import { MAX_PATTERN_INPUT_LENGTH } from 'universal-validator';

validators.email().validate('a'.repeat(MAX_PATTERN_INPUT_LENGTH + 1), 'f');
// => 'Invalid email address: ...' error, produced without running the regex
```

### Trusted schemas only

`pattern` is executed with the platform `RegExp` engine. A hostile pattern can still cause exponential backtracking even on short inputs — the length cap bounds input size, not pattern complexity. **Schemas and patterns must come from trusted sources.**

### Email/URL are sanity checks only

The email pattern is a permissive sanity check: it does not enforce RFC 5321 length limits and cannot prove deliverability. The URL pattern performs no IDN or port validation and accepts values such as `http://x.`.

## Real-World Examples

### User Registration

```typescript
const registrationSchema = new Schema({
  username: validators.string({ minLength: 3, maxLength: 20 }),
  email: validators.email(),
  password: validators.string({ minLength: 8 }),
  age: validators.int({ minValue: 13, required: false }),
  terms_accepted: validators.boolean(),
});
```

### API Request Validation

```typescript
const apiRequestSchema = new Schema({
  method: validators.string({ choices: ['GET', 'POST', 'PUT', 'DELETE'] }),
  url: validators.url(),
  headers: validators.object({ required: false }),
  body: validators.object({ required: false }),
  timeout: validators.number({ minValue: 0, required: false }),
});
```

### Configuration Validation

```typescript
const configSchema = new Schema({
  database: validators.object({
    schema: {
      host: validators.string(),
      port: validators.int({ minValue: 1, maxValue: 65535 }),
      username: validators.string(),
      password: validators.string(),
      ssl: validators.boolean(),
    },
  }),
  cache: validators.object({
    schema: {
      enabled: validators.boolean(),
      ttl: validators.int({ minValue: 0 }),
      max_size: validators.int({ minValue: 1 }),
    },
  }),
  features: validators.array({ itemValidator: validators.string() }),
});
```

## Error Handling

```typescript
import { Schema, validators, ValidationErrors } from 'universal-validator';

const schema = new Schema({
  email: validators.email(),
});

// Option 1: Check result
const result = schema.validate({ email: 'invalid' });
if (!result.valid) {
  for (const error of result.errors) {
    console.log(`${error.field}: ${error.message}`);
  }
}

// Option 2: Throw exception
try {
  schema.validateOrThrow({ email: 'invalid' });
} catch (error) {
  if (error instanceof ValidationErrors) {
    console.log(`Validation failed:\n${error.message}`);
    console.log(error.errors);
  }
}
```

`validateOrThrow` throws a `ValidationErrors` aggregate. Its `message` joins `field: message` lines with newlines, and `errors` exposes the full readonly `ValidationError` array.

## Error Path Format

Validation errors include a machine-readable `field` path:

- Top-level type errors on a `Schema` input use the field name `root`.
- Object fields are joined with `.`: `user.address.city`.
- Array items use bracket indices: `tags[0]`, `users[1].email`.

## API Reference

### Base

- `Validator<T>` - abstract generic base class
- `BaseValidatorOptions` - `{ required?, nullable?, failFast?, sensitive? }`
- `MAX_PATTERN_INPUT_LENGTH` - max string length (UTF-16 code units) a pattern check will run against (10,000)
- `ValidationResult` - `{ valid: boolean; errors: readonly ValidationError[] }`
- `ValidationError` - `{ field, message, value?, code }`
- `ValidationErrors` - `{ errors: readonly ValidationError[]; message }`
- `Codes` - object of canonical error-code strings

### Validators

- `validators.string(options?)` - `StringValidator`
- `validators.number(options?)` - `NumberValidator`
- `validators.int(options?)` - `NumberValidator` with `integer: true`
- `validators.boolean(options?)` - `BooleanValidator`
- `validators.email(options?)` - `EmailValidator`
- `validators.url(options?)` - `UrlValidator`
- `validators.uuid(options?)` - `UuidValidator`
- `validators.date(options?)` - `DateValidator`
- `validators.datetime(options?)` - `DateTimeValidator`
- `validators.numeric(options?)` - `NumericValidator`
- `validators.notEmpty(options?)` - `NotEmptyValidator`
- `validators.enum({ values, ...options? })` - `EnumValidator`
- `validators.oneOf(...validators)` or `validators.oneOf({ validators: [...], ...options? })` - `OneOfValidator`
- `validators.regex(pattern, options?)` or `validators.regex({ pattern, ...options? })` - `StringValidator`
- `validators.array(options?)` - `ArrayValidator<T>`
- `validators.object(options?)` - `ObjectValidator`

### Schema

- `new Schema(schema, { strict?, failFast? })`
- `schema.validate(data)` - validate data and return `ValidationResult`
- `schema.validateOrThrow(data)` - validate data and throw `ValidationErrors` if invalid

### ValidationResult

- `result.valid` - boolean indicating if validation passed
- `result.errors` - readonly array of `ValidationError` objects

### ValidationError

- `error.field` - field name/path that failed validation
- `error.message` - error message (prefixed with `<field>: `)
- `error.value` - value that failed validation
- `error.code` - machine-readable error code (see `Codes`)

### Error Codes (`Codes`)

Every `ValidationError` carries a `code` suitable for programmatic handling:

```typescript
import { Codes } from 'universal-validator';

for (const error of result.errors) {
  if (error.code === Codes.Required) {
    // handle missing/null field
  }
}
```

| `Codes` member | Code string | Meaning |
|---|---|---|
| `Codes.Required` | `required` | Required field absent, or null/undefined when `required && !nullable` |
| `Codes.UnknownField` | `unknown_field` | Strict-mode key not in schema |
| `Codes.Type` | `type` | Wrong runtime type |
| `Codes.MinLength` | `min_length` | String/array too short |
| `Codes.MaxLength` | `max_length` | String/array too long |
| `Codes.MinValue` | `min_value` | Below numeric minimum |
| `Codes.MaxValue` | `max_value` | Above numeric maximum |
| `Codes.Pattern` | `pattern` | Regex pattern mismatch |
| `Codes.Choices` | `choices` | Value not in allowed choices |
| `Codes.Email` | `email` | Email pattern mismatch |
| `Codes.Url` | `url` | URL pattern mismatch |
| `Codes.Custom` | `custom` | Custom validator rejection |
| `Codes.Uuid` | `uuid` | UUID format mismatch |
| `Codes.Date` | `date` | Date format mismatch |
| `Codes.DateTime` | `datetime` | Datetime format mismatch |
| `Codes.Numeric` | `numeric` | String is not numeric |
| `Codes.NotEmpty` | `not_empty` | String/array/object is empty |
| `Codes.Enum` | `enum` | Value not in enum `values` |
| `Codes.OneOf` | `one_of` | No `oneOf` branch matched |

## Presence & Null Semantics

A key that is present with a `null` or `undefined` value counts as present and is validated as a null value. It is not treated as a missing key. Both produce `Field is required` when `required && !nullable`, but through the validate path, not the missing-key path.

## Coercion and Transform

This library intentionally does not coerce or transform input. Trimming, case-folding, parsing strings into numbers or dates, and similar operations are the caller's responsibility before validation.

## Testing

```bash
npm test
npm run test:coverage
npm run test:watch
```

## Building

```bash
npm run build
```

## License

MIT License

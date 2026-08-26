# Universal Data Validator (TypeScript)

A comprehensive data validation library for TypeScript that works across API, database, and form contexts.

## Features

- ✅ **Rich Validator Types**: String, number, int, boolean, email, URL, array, object
- ✅ **Schema-Based Validation**: Define complex data structures
- ✅ **Custom Validators**: Add your own validation logic
- ✅ **Nested Validation**: Validate nested objects and arrays
- ✅ **Clear Error Messages**: Detailed error reporting with field paths
- ✅ **Type Safety**: Full TypeScript support with type inference
- ✅ **Zero Dependencies**: No external dependencies required
- ✅ **Production Ready**: Comprehensive test coverage

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

## Usage Examples

### String Validation

```typescript
import { validators } from 'universal-validator';

// Basic string
const validator = validators.string();

// String with length constraints
const usernameValidator = validators.string({ minLength: 3, maxLength: 20 });

// String with pattern
const phoneValidator = validators.string({ pattern: /^\d{3}-\d{4}$/ });

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
    zip: validators.string({ pattern: /^\d{5}$/ }),
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
  console.log('✅ Data is valid!');
} else {
  console.log('❌ Validation errors:');
  for (const error of result.errors) {
    console.log(`  ${error.field}: ${error.message}`);
  }
}
```

### Custom Validators

```typescript
const isEven = (value: number) => {
  return value % 2 !== 0 ? 'Value must be even' : null;
};

const validator = validators.int().custom(isEven);

const result = validator.validate(4, 'number');
console.log(result.valid); // true

const result2 = validator.validate(3, 'number');
console.log(result2.valid); // false
```

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
import { Schema, validators, ValidationError } from 'universal-validator';

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
  if (error instanceof ValidationError) {
    console.log(`Validation failed: ${error.message}`);
  }
}
```

## API Reference

### Validators

- `validators.string(options?)` - String validator
- `validators.number(options?)` - Number validator
- `validators.int(options?)` - Integer validator
- `validators.boolean(options?)` - Boolean validator
- `validators.email(options?)` - Email validator
- `validators.url(options?)` - URL validator
- `validators.array(options?)` - Array validator
- `validators.object(options?)` - Object validator

### Schema

- `new Schema(schema)` - Create a schema
- `schema.validate(data)` - Validate data and return ValidationResult
- `schema.validateOrThrow(data)` - Validate data and throw ValidationError if invalid

### ValidationResult

- `result.valid` - Boolean indicating if validation passed
- `result.errors` - Array of ValidationError objects

### ValidationError

- `error.field` - Field name that failed validation
- `error.message` - Error message
- `error.value` - Value that failed validation

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

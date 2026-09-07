/**
 * Tests for Universal Data Validator Library
 */

import { describe, it, expect } from 'vitest';
import { Schema, validators, ValidationErrors, Codes, MAX_PATTERN_INPUT_LENGTH } from './index';

describe('StringValidator', () => {
  it('validates valid string', () => {
    const validator = validators.string();
    const result = validator.validate('hello', 'field');
    expect(result.valid).toBe(true);
    expect(result.errors).toHaveLength(0);
  });

  it('rejects invalid type', () => {
    const validator = validators.string();
    const result = validator.validate(123, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toContain('Expected string');
    expect(result.errors[0].code).toBe(Codes.Type);
  });

  it('validates min length', () => {
    const validator = validators.string({ minLength: 5 });
    
    expect(validator.validate('hello', 'field').valid).toBe(true);
    expect(validator.validate('hi', 'field').valid).toBe(false);
  });

  it('validates max length', () => {
    const validator = validators.string({ maxLength: 5 });
    
    expect(validator.validate('hello', 'field').valid).toBe(true);
    expect(validator.validate('hello world', 'field').valid).toBe(false);
  });

  it('validates pattern', () => {
    const validator = validators.string({ pattern: /^\d{3}-\d{4}$/ });
    
    expect(validator.validate('123-4567', 'field').valid).toBe(true);
    expect(validator.validate('invalid', 'field').valid).toBe(false);
  });

  it('validates choices', () => {
    const validator = validators.string({ choices: ['red', 'green', 'blue'] });
    
    expect(validator.validate('red', 'field').valid).toBe(true);
    expect(validator.validate('yellow', 'field').valid).toBe(false);
  });

  it('handles required field', () => {
    const validator = validators.string({ required: true });
    const result = validator.validate(null, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toContain('required');
    expect(result.errors[0].code).toBe(Codes.Required);
  });

  it('handles optional field', () => {
    const validator = validators.string({ required: false });
    const result = validator.validate(null, 'field');
    expect(result.valid).toBe(true);
  });
});

describe('NumberValidator', () => {
  it('validates valid number', () => {
    const validator = validators.number();
    const result = validator.validate(42, 'field');
    expect(result.valid).toBe(true);
  });

  it('rejects invalid type', () => {
    const validator = validators.number();
    const result = validator.validate('not a number', 'field');
    expect(result.valid).toBe(false);
  });

  it('validates min value', () => {
    const validator = validators.number({ minValue: 0 });
    
    expect(validator.validate(5, 'field').valid).toBe(true);
    expect(validator.validate(-1, 'field').valid).toBe(false);
  });

  it('validates max value', () => {
    const validator = validators.number({ maxValue: 100 });
    
    expect(validator.validate(50, 'field').valid).toBe(true);
    expect(validator.validate(101, 'field').valid).toBe(false);
  });
});

describe('IntValidator', () => {
  it('validates integer', () => {
    const validator = validators.int();
    
    expect(validator.validate(42, 'field').valid).toBe(true);
    expect(validator.validate(3.14, 'field').valid).toBe(false);
  });

  it('validates range', () => {
    const validator = validators.int({ minValue: 0, maxValue: 100 });
    
    expect(validator.validate(50, 'field').valid).toBe(true);
    expect(validator.validate(-1, 'field').valid).toBe(false);
    expect(validator.validate(101, 'field').valid).toBe(false);
  });
});

describe('BooleanValidator', () => {
  it('validates boolean', () => {
    const validator = validators.boolean();
    
    expect(validator.validate(true, 'field').valid).toBe(true);
    expect(validator.validate(false, 'field').valid).toBe(true);
    expect(validator.validate('not a bool', 'field').valid).toBe(false);
  });
});

describe('EmailValidator', () => {
  it('validates valid emails', () => {
    const validator = validators.email();
    
    const validEmails = [
      'user@example.com',
      'test.user@example.com',
      'user+tag@example.co.uk',
      'user123@test-domain.com',
    ];

    for (const email of validEmails) {
      const result = validator.validate(email, 'field');
      expect(result.valid).toBe(true);
    }
  });

  it('rejects invalid emails', () => {
    const validator = validators.email();
    
    const invalidEmails = [
      'not-an-email',
      '@example.com',
      'user@',
      'user @example.com',
    ];

    for (const email of invalidEmails) {
      const result = validator.validate(email, 'field');
      expect(result.valid).toBe(false);
    }
  });
});

describe('UrlValidator', () => {
  it('validates valid URLs', () => {
    const validator = validators.url();
    
    const validUrls = [
      'http://example.com',
      'https://example.com',
      'https://example.com/path',
      'https://example.com/path?query=value',
    ];

    for (const url of validUrls) {
      const result = validator.validate(url, 'field');
      expect(result.valid).toBe(true);
    }
  });

  it('rejects invalid URLs', () => {
    const validator = validators.url();
    
    const invalidUrls = [
      'not-a-url',
      'example.com',
      'http://',
    ];

    for (const url of invalidUrls) {
      const result = validator.validate(url, 'field');
      expect(result.valid).toBe(false);
    }
  });
});

describe('ArrayValidator', () => {
  it('validates array', () => {
    const validator = validators.array();
    const result = validator.validate(['a', 'b', 'c'], 'field');
    expect(result.valid).toBe(true);
  });

  it('rejects invalid type', () => {
    const validator = validators.array();
    const result = validator.validate('not an array', 'field');
    expect(result.valid).toBe(false);
  });

  it('validates min length', () => {
    const validator = validators.array({ minLength: 2 });
    
    expect(validator.validate(['a', 'b'], 'field').valid).toBe(true);
    expect(validator.validate(['a'], 'field').valid).toBe(false);
  });

  it('validates max length', () => {
    const validator = validators.array({ maxLength: 3 });
    
    expect(validator.validate(['a', 'b'], 'field').valid).toBe(true);
    expect(validator.validate(['a', 'b', 'c', 'd'], 'field').valid).toBe(false);
  });

  it('validates items', () => {
    const validator = validators.array({ itemValidator: validators.int() });
    
    expect(validator.validate([1, 2, 3], 'field').valid).toBe(true);
    expect(validator.validate([1, 'two', 3], 'field').valid).toBe(false);
  });
});

describe('ObjectValidator', () => {
  it('validates object', () => {
    const validator = validators.object();
    const result = validator.validate({ key: 'value' }, 'field');
    expect(result.valid).toBe(true);
  });

  it('rejects invalid type', () => {
    const validator = validators.object();
    const result = validator.validate('not an object', 'field');
    expect(result.valid).toBe(false);
  });

  it('validates schema', () => {
    const validator = validators.object({
      schema: {
        name: validators.string(),
        age: validators.int({ minValue: 0 }),
      },
    });

    expect(validator.validate({ name: 'John', age: 30 }, 'field').valid).toBe(true);
    expect(validator.validate({ name: 'John', age: -1 }, 'field').valid).toBe(false);
  });

  it('checks required fields', () => {
    const validator = validators.object({
      schema: {
        name: validators.string({ required: true }),
      },
    });

    const result = validator.validate({}, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toContain('required');
  });
});

describe('Schema', () => {
  it('validates simple schema', () => {
    const schema = new Schema({
      username: validators.string({ minLength: 3 }),
      email: validators.email(),
      age: validators.int({ minValue: 0, maxValue: 120 }),
    });

    const data = {
      username: 'john_doe',
      email: 'john@example.com',
      age: 30,
    };

    const result = schema.validate(data);
    expect(result.valid).toBe(true);
  });

  it('collects multiple errors', () => {
    const schema = new Schema({
      username: validators.string({ minLength: 3 }),
      email: validators.email(),
      age: validators.int({ minValue: 0 }),
    });

    const data = {
      username: 'ab',
      email: 'invalid-email',
      age: -1,
    };

    const result = schema.validate(data);
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThanOrEqual(3);
  });

  it('validates nested schema', () => {
    const schema = new Schema({
      user: validators.object({
        schema: {
          name: validators.string(),
          email: validators.email(),
        },
      }),
      settings: validators.object({
        schema: {
          theme: validators.string({ choices: ['light', 'dark'] }),
          notifications: validators.boolean(),
        },
      }),
    });

    const data = {
      user: {
        name: 'John',
        email: 'john@example.com',
      },
      settings: {
        theme: 'dark',
        notifications: true,
      },
    };

    const result = schema.validate(data);
    expect(result.valid).toBe(true);
  });

  it('validates array of objects', () => {
    const schema = new Schema({
      users: validators.array({
        itemValidator: validators.object({
          schema: {
            name: validators.string(),
            age: validators.int({ minValue: 0 }),
          },
        }),
      }),
    });

    const data = {
      users: [
        { name: 'John', age: 30 },
        { name: 'Jane', age: 25 },
      ],
    };

    const result = schema.validate(data);
    expect(result.valid).toBe(true);
  });

  it('throws on validateOrThrow', () => {
    const schema = new Schema({
      email: validators.email(),
    });

    expect(() => {
      schema.validateOrThrow({ email: 'invalid' });
    }).toThrow(ValidationErrors);
  });
});

describe('Custom Validators', () => {
  it('applies custom validator', () => {
    const isEven = (value: number) => {
      return value % 2 !== 0 ? 'Value must be even' : null;
    };

    const validator = validators.int().custom(isEven);

    expect(validator.validate(4, 'field').valid).toBe(true);
    expect(validator.validate(3, 'field').valid).toBe(false);
  });

  it('applies multiple custom validators', () => {
    const isPositive = (value: number) => (value <= 0 ? 'Must be positive' : null);
    const isEven = (value: number) => (value % 2 !== 0 ? 'Must be even' : null);

    const validator = validators.int().custom(isPositive).custom(isEven);

    expect(validator.validate(4, 'field').valid).toBe(true);
    expect(validator.validate(-2, 'field').valid).toBe(false);
  });
});

describe('Real-world Scenarios', () => {
  it('validates user registration', () => {
    const schema = new Schema({
      username: validators.string({ minLength: 3, maxLength: 20 }),
      email: validators.email(),
      password: validators.string({ minLength: 8 }),
      age: validators.int({ minValue: 13, required: false }),
      terms_accepted: validators.boolean(),
    });

    const validData = {
      username: 'john_doe',
      email: 'john@example.com',
      password: 'secure_password_123',
      age: 25,
      terms_accepted: true,
    };

    expect(schema.validate(validData).valid).toBe(true);

    const invalidData = {
      username: 'ab',
      email: 'invalid',
      password: 'short',
      terms_accepted: false,
    };

    const result = schema.validate(invalidData);
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThanOrEqual(3);
  });

  it('validates API request', () => {
    const schema = new Schema({
      method: validators.string({ choices: ['GET', 'POST', 'PUT', 'DELETE'] }),
      url: validators.url(),
      headers: validators.object({ required: false }),
      body: validators.object({ required: false }),
      timeout: validators.number({ minValue: 0, required: false }),
    });

    const data = {
      method: 'POST',
      url: 'https://api.example.com/users',
      headers: { 'Content-Type': 'application/json' },
      body: { name: 'John' },
      timeout: 30.0,
    };

    expect(schema.validate(data).valid).toBe(true);
  });
});

describe('Strict Mode', () => {
  it('reports unknown top-level keys when strict', () => {
    const schema = new Schema(
      {
        name: validators.string(),
      },
      { strict: true }
    );

    const result = schema.validate({ name: 'John', extra: 'nope' });
    expect(result.valid).toBe(false);
    const unknown = result.errors.find(e => e.field === 'extra');
    expect(unknown).toBeDefined();
    expect(unknown!.message).toContain('Unknown field');
    expect(unknown!.code).toBe(Codes.UnknownField);
  });

  it('reports unknown nested keys with path field.key', () => {
    const schema = new Schema(
      {
        user: validators.object({
          schema: { name: validators.string() },
          strict: true,
        }),
      },
      { strict: true }
    );

    const result = schema.validate({
      user: { name: 'John', extra: 1 },
      top_extra: 2,
    });

    expect(result.valid).toBe(false);
    const fields = result.errors.map(e => e.field);
    expect(fields).toContain('user.extra');
    expect(fields).toContain('top_extra');
  });

  it('ignores unknown keys by default', () => {
    const schema = new Schema({
      name: validators.string(),
    });

    expect(schema.validate({ name: 'John', extra: 'ok' }).valid).toBe(true);
  });

  it('nested object ignores unknown keys by default', () => {
    const validator = validators.object({
      schema: { name: validators.string() },
    });

    expect(validator.validate({ name: 'John', extra: 1 }, 'field').valid).toBe(true);
  });
});

describe('Email/URL error pass-through', () => {
  it('preserves non-pattern errors (minLength) instead of only "Invalid email"', () => {
    const validator = validators.email({ minLength: 5 });
    const result = validator.validate('a@b', 'email');

    expect(result.valid).toBe(false);
    const messages = result.errors.map(e => e.message);
    expect(messages.some(m => m.includes('at least 5'))).toBe(true);
  });

  it('replaces only the pattern error with friendly message', () => {
    const validator = validators.email({ minLength: 5 });
    const result = validator.validate('not-an-email', 'email');

    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].message).toContain('Invalid email address');
    expect(result.errors[0].code).toBe(Codes.Email);
  });

  it('composes email with maxLength', () => {
    const validator = validators.email({ maxLength: 20 });

    expect(validator.validate('user@example.com', 'email').valid).toBe(true);

    const longEmail = 'averylongusername@averylongdomainname.com';
    const result = validator.validate(longEmail, 'email');
    expect(result.valid).toBe(false);
    expect(result.errors.some(e => e.message.includes('at most 20'))).toBe(true);
    expect(result.errors.every(e => !e.message.includes('Invalid email'))).toBe(true);
  });

  it('composes email with choices', () => {
    const validator = validators.email({ choices: ['a@b.com'] });

    expect(validator.validate('a@b.com', 'email').valid).toBe(true);
    const result = validator.validate('x@y.com', 'email');
    expect(result.valid).toBe(false);
    expect(result.errors.some(e => e.message.includes('one of'))).toBe(true);
  });

  it('url validator preserves non-pattern errors', () => {
    const validator = validators.url({ minLength: 20 });
    const result = validator.validate('http://a.b', 'url');

    expect(result.valid).toBe(false);
    const messages = result.errors.map(e => e.message);
    expect(messages.some(m => m.includes('at least 20'))).toBe(true);
    expect(messages.every(m => !m.includes('Invalid URL'))).toBe(true);
  });

  it('rewrites url pattern error with url code', () => {
    const validator = validators.url();
    const result = validator.validate('not-a-url', 'url');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Url);
  });
});

describe('Fail-fast mode', () => {
  it('Schema returns exactly one error when multiple fields fail', () => {
    const schema = new Schema(
      {
        username: validators.string({ minLength: 3 }),
        email: validators.email(),
        age: validators.int({ minValue: 0 }),
      },
      { failFast: true }
    );

    const result = schema.validate({ username: 'ab', email: 'bad', age: -1 });
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
  });

  it('default (failFast off) collects all errors', () => {
    const schema = new Schema({
      username: validators.string({ minLength: 3 }),
      email: validators.email(),
      age: validators.int({ minValue: 0 }),
    });

    const result = schema.validate({ username: 'ab', email: 'bad', age: -1 });
    expect(result.valid).toBe(false);
    expect(result.errors.length).toBeGreaterThanOrEqual(3);
  });

  it('ObjectValidator stops at the first field error', () => {
    const validator = validators.object({
      failFast: true,
      schema: {
        a: validators.string({ minLength: 3 }),
        b: validators.int({ minValue: 0 }),
      },
    });

    const result = validator.validate({ a: 'x', b: -1 }, 'obj');
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
  });

  it('propagates fail-fast into nested object validators via Schema', () => {
    const schema = new Schema(
      {
        user: validators.object({
          schema: {
            name: validators.string({ minLength: 3 }),
            email: validators.email(),
          },
        }),
      },
      { failFast: true }
    );

    const result = schema.validate({ user: { name: 'x', email: 'bad' } });
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
  });

  it('ArrayValidator stops at the first failing item', () => {
    const validator = validators.array({
      itemValidator: validators.int({ minValue: 0 }),
      failFast: true,
    });

    const result = validator.validate([-1, -2, -3], 'nums');
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].field).toBe('nums[0]');
  });

  it('propagates fail-fast into array items via Schema', () => {
    const schema = new Schema(
      {
        nums: validators.array({ itemValidator: validators.int({ minValue: 0 }) }),
      },
      { failFast: true }
    );

    const result = schema.validate({ nums: [-1, -2] });
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
  });

  it('missing required fields also stop after the first', () => {
    const schema = new Schema(
      {
        a: validators.string(),
        b: validators.string(),
      },
      { failFast: true }
    );

    const result = schema.validate({});
    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].code).toBe(Codes.Required);
  });
});

describe('Choices message', () => {
  it('preserves declaration order in the choices error message', () => {
    const validator = validators.string({ choices: ['zeta', 'alpha', 'mid'] });
    const result = validator.validate('nope', 'field');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Choices);
    expect(result.errors[0].message).toContain(
      'Value must be one of zeta, alpha, mid'
    );
  });
});

describe('Presence & null semantics', () => {
  it('a key present with undefined counts as present (required error via validate path)', () => {
    const schema = new Schema({ name: validators.string() });
    const result = schema.validate({ name: undefined });

    expect(result.valid).toBe(false);
    expect(result.errors).toHaveLength(1);
    expect(result.errors[0].field).toBe('name');
    expect(result.errors[0].code).toBe(Codes.Required);
  });

  it('ObjectValidator treats present-undefined as present', () => {
    const validator = validators.object({
      schema: { name: validators.string() },
    });
    const result = validator.validate({ name: undefined }, 'field');

    expect(result.valid).toBe(false);
    expect(result.errors[0].field).toBe('field.name');
    expect(result.errors[0].code).toBe(Codes.Required);
  });
});

describe('Custom validator immutability', () => {
  it('does not mutate the original validator', () => {
    const base = validators.int();
    const customized = base.custom((value) =>
      value % 2 !== 0 ? 'Value must be even' : null
    );

    expect(base.validate(3, 'field').valid).toBe(true);
    expect(customized.validate(3, 'field').valid).toBe(false);
    expect(customized.validate(4, 'field').valid).toBe(true);
  });
});

describe('UUID validator', () => {
  it('accepts a valid UUID', () => {
    const validator = validators.uuid();
    expect(validator.validate('550e8400-e29b-41d4-a716-446655440000', 'field').valid).toBe(true);
  });

  it('rejects an invalid UUID', () => {
    const validator = validators.uuid();
    const result = validator.validate('not-a-uuid', 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Uuid);
    expect(result.errors[0].message).toContain('Invalid UUID');
  });

  it('rejects non-string values', () => {
    const validator = validators.uuid();
    const result = validator.validate(42, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Type);
  });
});

describe('Date validator', () => {
  it('accepts a valid date', () => {
    const validator = validators.date();
    expect(validator.validate('2026-09-07', 'field').valid).toBe(true);
  });

  it('rejects an invalid date', () => {
    const validator = validators.date();
    const result = validator.validate('2026-13-07', 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Date);
    expect(result.errors[0].message).toContain('Invalid date');
  });
});

describe('Datetime validator', () => {
  it('accepts a valid datetime', () => {
    const validator = validators.datetime();
    expect(validator.validate('2026-09-07T10:30:00Z', 'field').valid).toBe(true);
    expect(validator.validate('2026-09-07T10:30:00.5+05:30', 'field').valid).toBe(true);
  });

  it('rejects an invalid datetime', () => {
    const validator = validators.datetime();
    const result = validator.validate('2026-09-07 10:30', 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.DateTime);
    expect(result.errors[0].message).toContain('Invalid datetime');
  });
});

describe('Numeric validator', () => {
  it('accepts numeric strings', () => {
    const validator = validators.numeric();
    expect(validator.validate('-12.5', 'field').valid).toBe(true);
    expect(validator.validate('1e3', 'field').valid).toBe(true);
  });

  it('rejects non-numeric strings', () => {
    const validator = validators.numeric();
    const result = validator.validate('12a', 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Numeric);
    expect(result.errors[0].message).toContain('Expected numeric string');
  });

  it('rejects non-string values', () => {
    const validator = validators.numeric();
    const result = validator.validate(12, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Type);
  });
});

describe('Not-empty validator', () => {
  it('accepts non-empty string, array, and object', () => {
    const validator = validators.notEmpty();
    expect(validator.validate('x', 'field').valid).toBe(true);
    expect(validator.validate([1], 'field').valid).toBe(true);
    expect(validator.validate({ k: 1 }, 'field').valid).toBe(true);
  });

  it('rejects empty string, array, and object', () => {
    const validator = validators.notEmpty();
    expect(validator.validate('', 'field').valid).toBe(false);
    expect(validator.validate([], 'field').valid).toBe(false);
    expect(validator.validate({}, 'field').valid).toBe(false);
  });

  it('rejects non-container values', () => {
    const validator = validators.notEmpty();
    const result = validator.validate(5, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Type);
    expect(result.errors[0].message).toContain('Expected string, array, or object');
  });
});

describe('Enum validator', () => {
  it('accepts values from the allowed set', () => {
    const validator = validators.enum({ values: ['a', 1, true] });
    expect(validator.validate('a', 'field').valid).toBe(true);
    expect(validator.validate(1, 'field').valid).toBe(true);
    expect(validator.validate(true, 'field').valid).toBe(true);
  });

  it('rejects values not in the set', () => {
    const validator = validators.enum({ values: ['a', 1, true] });
    const result = validator.validate('b', 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Enum);
    expect(result.errors[0].message).toContain('Value must be one of "a", 1, true');
  });
});

describe('One-of validator', () => {
  it('validates when any branch matches', () => {
    const validator = validators.oneOf(
      validators.string({ minLength: 3 }),
      validators.int({ minValue: 10 })
    );
    expect(validator.validate('abc', 'field').valid).toBe(true);
    expect(validator.validate(15, 'field').valid).toBe(true);
  });

  it('fails when no branch matches', () => {
    const validator = validators.oneOf(
      validators.string({ minLength: 3 }),
      validators.int({ minValue: 10 })
    );
    const result = validator.validate(5, 'field');
    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.OneOf);
    expect(result.errors[0].message).toContain('Value does not match any allowed schema');
  });
});

describe('Regex factory', () => {
  it('creates a full-match pattern validator from a RegExp', () => {
    const validator = validators.regex(/^[0-9]{3}-[0-9]{4}$/);
    expect(validator.validate('123-4567', 'field').valid).toBe(true);
    expect(validator.validate('123-45678', 'field').valid).toBe(false);
  });

  it('creates a full-match pattern validator from a string', () => {
    const validator = validators.regex('^[0-9]{3}-[0-9]{4}$');
    expect(validator.validate('123-4567', 'field').valid).toBe(true);
    expect(validator.validate('invalid', 'field').valid).toBe(false);
  });
});

describe('Sensitive option', () => {
  it('redacts the value in email error message and error.value', () => {
    const validator = validators.email({ sensitive: true });
    const result = validator.validate('p@ssw0rd', 'secret');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Email);
    expect(result.errors[0].message).toBe('secret: Invalid email address: ***');
    expect(result.errors[0].value).toBeUndefined();
  });

  it('redacts the value in choices message (quotes kept)', () => {
    const validator = validators.string({ choices: ['a', 'b'], sensitive: true });
    const result = validator.validate('hunter2', 'token');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Choices);
    expect(result.errors[0].message).toBe("token: Value must be one of a, b, got '***'");
    expect(result.errors[0].value).toBeUndefined();
  });

  it('leaves messages that do not embed the input unchanged', () => {
    const validator = validators.string({ minLength: 5, sensitive: true });
    const result = validator.validate('abc', 'secret');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.MinLength);
    expect(result.errors[0].message).toBe('secret: String length must be at least 5, got 3');
  });

  it('redacts enum message bare (no quotes)', () => {
    const validator = validators.enum({ values: ['a', 1], sensitive: true });
    const result = validator.validate('b', 'field');

    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toBe('field: Value must be one of "a", 1, got ***');
    expect(result.errors[0].value).toBeUndefined();
  });

  it('is not inherited by nested validators', () => {
    const validator = validators.object({
      sensitive: true,
      schema: { inner: validators.email() },
    });
    const result = validator.validate({ inner: 'bad' }, 'obj');

    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toContain('bad');
    expect(result.errors[0].value).toBe('bad');
  });

  it('survives .custom() cloning', () => {
    const base = validators.email({ sensitive: true });
    const clone = base.custom(() => null);
    const result = clone.validate('secret-value', 'field');

    expect(result.valid).toBe(false);
    expect(result.errors[0].message).toContain('***');
    expect(result.errors[0].value).toBeUndefined();
  });
});

describe('Pattern input length cap', () => {
  it('skips the regex and reports a pattern error for oversized input', () => {
    const validator = validators.string({ pattern: /^a+$/ });
    const big = 'a'.repeat(MAX_PATTERN_INPUT_LENGTH + 1);
    const result = validator.validate(big, 'field');

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Pattern);
  });

  it('applies the cap to the fixed email pattern', () => {
    const validator = validators.email();
    const big = 'a'.repeat(MAX_PATTERN_INPUT_LENGTH + 1);
    const start = Date.now();
    const result = validator.validate(big, 'field');
    const elapsed = Date.now() - start;

    expect(result.valid).toBe(false);
    expect(result.errors[0].code).toBe(Codes.Email);
    expect(elapsed).toBeLessThan(1000);
  });

  it('still matches inputs at the cap boundary', () => {
    const validator = validators.string({ pattern: /^a+$/ });
    const exact = 'a'.repeat(MAX_PATTERN_INPUT_LENGTH);
    expect(validator.validate(exact, 'field').valid).toBe(true);
  });
});

/**
 * Tests for Universal Data Validator Library
 */

import { describe, it, expect } from 'vitest';
import { Schema, validators, ValidationError } from './index';

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
    }).toThrow(ValidationError);
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

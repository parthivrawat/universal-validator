/**
 * Universal Data Validator Library
 * 
 * A comprehensive data validation library that works across API, database, and form contexts.
 * 
 * @author Parthiv Rawat
 * @license MIT
 */

/**
 * Validation error class
 */
export class ValidationError extends Error {
  constructor(
    public readonly field: string,
    message: string,
    public readonly value?: any
  ) {
    super(`${field}: ${message}`);
    this.name = 'ValidationError';
  }
}

/**
 * Validation result
 */
export class ValidationResult {
  public valid: boolean = true;
  public errors: ValidationError[] = [];

  addError(error: ValidationError): void {
    this.valid = false;
    this.errors.push(error);
  }

  toString(): string {
    if (this.valid) {
      return 'ValidationResult(valid=true)';
    }
    return `ValidationResult(valid=false, errors=${this.errors.length})`;
  }
}

/**
 * Base validator class
 */
export abstract class Validator<T = any> {
  protected customValidators: Array<(value: any) => string | null> = [];

  constructor(
    public readonly required: boolean = true,
    public readonly nullable: boolean = false
  ) {}

  /**
   * Validate a value
   */
  validate(value: any, field: string = 'field'): ValidationResult {
    const result = new ValidationResult();

    if (value === null || value === undefined) {
      if (this.required && !this.nullable) {
        result.addError(new ValidationError(field, 'Field is required'));
      }
      return result;
    }

    const typeError = this.validateType(value, field);
    if (typeError) {
      result.addError(typeError);
      return result;
    }

    const constraintErrors = this.validateConstraints(value, field);
    for (const error of constraintErrors) {
      result.addError(error);
    }

    for (const customValidator of this.customValidators) {
      const errorMsg = customValidator(value);
      if (errorMsg) {
        result.addError(new ValidationError(field, errorMsg, value));
      }
    }

    return result;
  }

  /**
   * Validate the type of the value
   */
  protected abstract validateType(value: any, field: string): ValidationError | null;

  /**
   * Validate constraints on the value
   */
  protected abstract validateConstraints(value: any, field: string): ValidationError[];

  /**
   * Add a custom validator function
   */
  custom(validator: (value: any) => string | null): this {
    this.customValidators.push(validator);
    return this;
  }
}

/**
 * String validator
 */
export class StringValidator extends Validator<string> {
  constructor(
    private readonly minLength?: number,
    private readonly maxLength?: number,
    private readonly pattern?: RegExp,
    private readonly choices?: string[],
    required: boolean = true,
    nullable: boolean = false
  ) {
    super(required, nullable);
  }

  protected validateType(value: any, field: string): ValidationError | null {
    if (typeof value !== 'string') {
      return new ValidationError(field, `Expected string, got ${typeof value}`, value);
    }
    return null;
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.minLength !== undefined && value.length < this.minLength) {
      errors.push(
        new ValidationError(
          field,
          `String length must be at least ${this.minLength}, got ${value.length}`,
          value
        )
      );
    }

    if (this.maxLength !== undefined && value.length > this.maxLength) {
      errors.push(
        new ValidationError(
          field,
          `String length must be at most ${this.maxLength}, got ${value.length}`,
          value
        )
      );
    }

    if (this.pattern && !this.pattern.test(value)) {
      errors.push(
        new ValidationError(field, `String does not match pattern ${this.pattern}`, value)
      );
    }

    if (this.choices && !this.choices.includes(value)) {
      errors.push(
        new ValidationError(
          field,
          `Value must be one of ${this.choices.join(', ')}, got '${value}'`,
          value
        )
      );
    }

    return errors;
  }
}

/**
 * Number validator (for integers and floats)
 */
export class NumberValidator extends Validator<number> {
  constructor(
    private readonly minValue?: number,
    private readonly maxValue?: number,
    private readonly integer: boolean = false,
    required: boolean = true,
    nullable: boolean = false
  ) {
    super(required, nullable);
  }

  protected validateType(value: any, field: string): ValidationError | null {
    if (typeof value !== 'number' || isNaN(value)) {
      return new ValidationError(field, `Expected number, got ${typeof value}`, value);
    }

    if (this.integer && !Number.isInteger(value)) {
      return new ValidationError(field, `Expected integer, got ${value}`, value);
    }

    return null;
  }

  protected validateConstraints(value: number, field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.minValue !== undefined && value < this.minValue) {
      errors.push(
        new ValidationError(
          field,
          `Value must be at least ${this.minValue}, got ${value}`,
          value
        )
      );
    }

    if (this.maxValue !== undefined && value > this.maxValue) {
      errors.push(
        new ValidationError(
          field,
          `Value must be at most ${this.maxValue}, got ${value}`,
          value
        )
      );
    }

    return errors;
  }
}

/**
 * Boolean validator
 */
export class BooleanValidator extends Validator<boolean> {
  protected validateType(value: any, field: string): ValidationError | null {
    if (typeof value !== 'boolean') {
      return new ValidationError(field, `Expected boolean, got ${typeof value}`, value);
    }
    return null;
  }

  protected validateConstraints(value: boolean, field: string): ValidationError[] {
    return [];
  }
}

/**
 * Email validator
 */
export class EmailValidator extends StringValidator {
  constructor(required: boolean = true, nullable: boolean = false) {
    const emailPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    super(undefined, undefined, emailPattern, undefined, required, nullable);
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors = super.validateConstraints(value, field);
    
    if (errors.length > 0 && this.pattern) {
      return [new ValidationError(field, `Invalid email address: ${value}`, value)];
    }
    
    return errors;
  }
}

/**
 * URL validator
 */
export class UrlValidator extends StringValidator {
  constructor(required: boolean = true, nullable: boolean = false) {
    const urlPattern = /^https?:\/\/[^\s/$.?#].[^\s]*$/;
    super(undefined, undefined, urlPattern, undefined, required, nullable);
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors = super.validateConstraints(value, field);
    
    if (errors.length > 0 && this.pattern) {
      return [new ValidationError(field, `Invalid URL: ${value}`, value)];
    }
    
    return errors;
  }
}

/**
 * Array validator
 */
export class ArrayValidator extends Validator<any[]> {
  constructor(
    private readonly itemValidator?: Validator,
    private readonly minLength?: number,
    private readonly maxLength?: number,
    required: boolean = true,
    nullable: boolean = false
  ) {
    super(required, nullable);
  }

  protected validateType(value: any, field: string): ValidationError | null {
    if (!Array.isArray(value)) {
      return new ValidationError(field, `Expected array, got ${typeof value}`, value);
    }
    return null;
  }

  protected validateConstraints(value: any[], field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.minLength !== undefined && value.length < this.minLength) {
      errors.push(
        new ValidationError(
          field,
          `Array length must be at least ${this.minLength}, got ${value.length}`,
          value
        )
      );
    }

    if (this.maxLength !== undefined && value.length > this.maxLength) {
      errors.push(
        new ValidationError(
          field,
          `Array length must be at most ${this.maxLength}, got ${value.length}`,
          value
        )
      );
    }

    if (this.itemValidator) {
      value.forEach((item, index) => {
        const itemResult = this.itemValidator!.validate(item, `${field}[${index}]`);
        errors.push(...itemResult.errors);
      });
    }

    return errors;
  }
}

/**
 * Object validator
 */
export class ObjectValidator extends Validator<Record<string, any>> {
  constructor(
    private readonly schema?: Record<string, Validator>,
    required: boolean = true,
    nullable: boolean = false
  ) {
    super(required, nullable);
  }

  protected validateType(value: any, field: string): ValidationError | null {
    if (typeof value !== 'object' || value === null || Array.isArray(value)) {
      return new ValidationError(field, `Expected object, got ${typeof value}`, value);
    }
    return null;
  }

  protected validateConstraints(value: Record<string, any>, field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.schema) {
      for (const [key, validator] of Object.entries(this.schema)) {
        if (key in value) {
          const result = validator.validate(value[key], `${field}.${key}`);
          errors.push(...result.errors);
        } else if (validator.required) {
          errors.push(new ValidationError(`${field}.${key}`, 'Field is required'));
        }
      }
    }

    return errors;
  }
}

/**
 * Schema for validating complex data structures
 */
export class Schema {
  constructor(private readonly schema: Record<string, Validator>) {}

  /**
   * Validate data against the schema
   */
  validate(data: Record<string, any>): ValidationResult {
    const result = new ValidationResult();

    if (typeof data !== 'object' || data === null || Array.isArray(data)) {
      result.addError(new ValidationError('root', `Expected object, got ${typeof data}`));
      return result;
    }

    for (const [field, validator] of Object.entries(this.schema)) {
      if (field in data) {
        const fieldResult = validator.validate(data[field], field);
        for (const error of fieldResult.errors) {
          result.addError(error);
        }
      } else if (validator.required) {
        result.addError(new ValidationError(field, 'Field is required'));
      }
    }

    return result;
  }

  /**
   * Validate data and throw exception if invalid
   */
  validateOrThrow(data: Record<string, any>): void {
    const result = this.validate(data);
    if (!result.valid) {
      const errorMessages = result.errors.map(e => `${e.field}: ${e.message}`).join('\n');
      throw new ValidationError('validation', errorMessages);
    }
  }
}

/**
 * Validator factory functions
 */
export const validators = {
  string(options?: {
    minLength?: number;
    maxLength?: number;
    pattern?: RegExp;
    choices?: string[];
    required?: boolean;
    nullable?: boolean;
  }): StringValidator {
    return new StringValidator(
      options?.minLength,
      options?.maxLength,
      options?.pattern,
      options?.choices,
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  number(options?: {
    minValue?: number;
    maxValue?: number;
    required?: boolean;
    nullable?: boolean;
  }): NumberValidator {
    return new NumberValidator(
      options?.minValue,
      options?.maxValue,
      false,
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  int(options?: {
    minValue?: number;
    maxValue?: number;
    required?: boolean;
    nullable?: boolean;
  }): NumberValidator {
    return new NumberValidator(
      options?.minValue,
      options?.maxValue,
      true,
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  boolean(options?: {
    required?: boolean;
    nullable?: boolean;
  }): BooleanValidator {
    return new BooleanValidator(
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  email(options?: {
    required?: boolean;
    nullable?: boolean;
  }): EmailValidator {
    return new EmailValidator(
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  url(options?: {
    required?: boolean;
    nullable?: boolean;
  }): UrlValidator {
    return new UrlValidator(
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  array(options?: {
    itemValidator?: Validator;
    minLength?: number;
    maxLength?: number;
    required?: boolean;
    nullable?: boolean;
  }): ArrayValidator {
    return new ArrayValidator(
      options?.itemValidator,
      options?.minLength,
      options?.maxLength,
      options?.required ?? true,
      options?.nullable ?? false
    );
  },

  object(options?: {
    schema?: Record<string, Validator>;
    required?: boolean;
    nullable?: boolean;
  }): ObjectValidator {
    return new ObjectValidator(
      options?.schema,
      options?.required ?? true,
      options?.nullable ?? false
    );
  },
};

export default validators;

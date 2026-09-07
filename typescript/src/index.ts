/**
 * Universal Data Validator Library
 * 
 * A comprehensive data validation library that works across API, database, and form contexts.
 * 
 * @author Parthiv Rawat
 * @license MIT
 */

/**
 * Map a runtime value to its canonical JSON-kind name.
 */
function kindName(value: unknown): string {
  if (value === null) {
    return 'null';
  }
  if (typeof value === 'boolean') {
    return 'boolean';
  }
  if (typeof value === 'number') {
    return Number.isInteger(value) ? 'integer' : 'number';
  }
  if (typeof value === 'string') {
    return 'string';
  }
  if (Array.isArray(value)) {
    return 'array';
  }
  if (typeof value === 'object') {
    return 'object';
  }
  return 'unknown';
}

/**
 * Canonical machine-readable error codes (see ../testdata/SPEC.md).
 */
export const Codes = {
  Required: 'required',
  UnknownField: 'unknown_field',
  Type: 'type',
  MinLength: 'min_length',
  MaxLength: 'max_length',
  MinValue: 'min_value',
  MaxValue: 'max_value',
  Pattern: 'pattern',
  Choices: 'choices',
  Email: 'email',
  Url: 'url',
  Custom: 'custom',
  Uuid: 'uuid',
  Date: 'date',
  DateTime: 'datetime',
  Numeric: 'numeric',
  NotEmpty: 'not_empty',
  Enum: 'enum',
  OneOf: 'one_of',
} as const;

/**
 * Union of all canonical error codes.
 */
export type ErrorCode = typeof Codes[keyof typeof Codes];

/**
 * Single validation error.
 *
 * Kept as one class keyed by code rather than a full discriminated union to
 * keep the public surface small while still providing type-safe codes.
 */
export class ValidationError extends Error {
  public readonly name = 'ValidationError';
  constructor(
    public readonly field: string,
    message: string,
    public readonly value?: unknown,
    public readonly code: ErrorCode = Codes.Custom
  ) {
    super(`${field}: ${message}`);
  }
}

/**
 * Aggregate error thrown by validateOrThrow.
 */
export class ValidationErrors extends Error {
  public readonly name = 'ValidationErrors';
  constructor(public readonly errors: readonly ValidationError[]) {
    super(errors.map((error) => error.message).join('\n'));
  }
}

/**
 * Validation result.
 */
export class ValidationResult {
  private _valid = true;
  private _errors: ValidationError[] = [];

  public get valid(): boolean {
    return this._valid;
  }

  public get errors(): readonly ValidationError[] {
    return this._errors;
  }

  addError(error: ValidationError): void {
    this._valid = false;
    this._errors.push(error);
  }

  toString(): string {
    if (this._valid) {
      return 'ValidationResult(valid=true)';
    }
    return `ValidationResult(valid=false, errors=${this._errors.length})`;
  }
}

/**
 * Maximum input length (in UTF-16 code units, i.e. `string.length`) that a
 * pattern check will run against. Inputs longer than this skip the regex and
 * report the normal pattern error. JavaScript's `RegExp` is a backtracking
 * engine, so this bounds worst-case match cost (ReDoS protection).
 */
export const MAX_PATTERN_INPUT_LENGTH = 10_000;

/**
 * Shared base options for all validators.
 */
export interface BaseValidatorOptions {
  required?: boolean;
  nullable?: boolean;
  failFast?: boolean;
  /**
   * When true, the offending value is redacted: messages that would embed it
   * render `***` instead, and `ValidationError.value` is `undefined`.
   * Per-validator — NOT inherited by nested validators.
   */
  sensitive?: boolean;
}

/**
 * Base validator class.
 */
export abstract class Validator<T> {
  public readonly required: boolean;
  public readonly nullable: boolean;
  public failFast: boolean;
  protected readonly sensitive: boolean;
  protected customValidators: Array<(value: unknown) => string | null> = [];

  constructor(options: BaseValidatorOptions = {}) {
    this.required = options.required ?? true;
    this.nullable = options.nullable ?? false;
    this.failFast = options.failFast ?? false;
    this.sensitive = options.sensitive ?? false;
  }

  /**
   * Value to interpolate into an error message: `***` when sensitive.
   */
  protected displayValue(value: unknown): unknown {
    return this.sensitive ? '***' : value;
  }

  /**
   * Value to store on a `ValidationError`: `undefined` when sensitive.
   */
  protected errorValue(value: unknown): unknown {
    return this.sensitive ? undefined : value;
  }

  /**
   * Validate a value.
   */
  validate(value: unknown, field: string = 'field'): ValidationResult {
    const result = new ValidationResult();

    if (value === null || value === undefined) {
      if (this.required && !this.nullable) {
        result.addError(
          new ValidationError(field, 'Field is required', this.errorValue(value), Codes.Required)
        );
      }
      return result;
    }

    const typeError = this.validateType(value, field);
    if (typeError) {
      result.addError(typeError);
      return result;
    }

    const typedValue = value as T;
    const constraintErrors = this.validateConstraints(typedValue, field);
    for (const error of constraintErrors) {
      result.addError(error);
      if (this.failFast) {
        return result;
      }
    }

    for (const customValidator of this.customValidators) {
      const errorMsg = customValidator(typedValue);
      if (errorMsg) {
        result.addError(new ValidationError(field, errorMsg, this.errorValue(typedValue), Codes.Custom));
        if (this.failFast) {
          return result;
        }
      }
    }

    return result;
  }

  /**
   * Validate the type of the value.
   */
  protected abstract validateType(value: unknown, field: string): ValidationError | null;

  /**
   * Validate constraints on the value.
   */
  protected abstract validateConstraints(value: T, field: string): ValidationError[];

  /**
   * Add a custom validator function. Returns a cloned validator; the receiver is
   * not mutated.
   */
  custom(validator: (value: T) => string | null): this {
    const clone = Object.create(Object.getPrototypeOf(this)) as this;
    Object.assign(clone, this);
    // Custom validators run after the type check, so the value is known to be T.
    const adapted = (value: unknown) => validator(value as T);
    clone.customValidators = [...this.customValidators, adapted];
    return clone;
  }
}

/**
 * String validator options.
 */
export interface StringValidatorOptions extends BaseValidatorOptions {
  minLength?: number;
  maxLength?: number;
  pattern?: RegExp;
  choices?: string[];
}

/**
 * String validator.
 */
export class StringValidator extends Validator<string> {
  private readonly fullPattern?: RegExp;
  private readonly choicesSet?: Set<string>;

  constructor(protected readonly options: StringValidatorOptions = {}) {
    super(options);
    if (options.pattern) {
      const flags = options.pattern.flags.replace(/[gy]/g, '');
      this.fullPattern = new RegExp(`^(?:${options.pattern.source})$`, flags);
    }
    if (options.choices) {
      this.choicesSet = new Set(options.choices);
    }
  }

  /**
   * Check whether a constraint error is the pattern-mismatch error
   * produced by this validator's regex check.
   */
  protected isPatternError(error: ValidationError): boolean {
    return error.code === Codes.Pattern;
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (typeof value !== 'string') {
      return new ValidationError(
        field,
        `Expected string, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }
    return null;
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const valueLength = [...value].length;

    if (this.options.minLength !== undefined && valueLength < this.options.minLength) {
      errors.push(
        new ValidationError(
          field,
          `String length must be at least ${this.options.minLength}, got ${valueLength}`,
          this.errorValue(value),
          Codes.MinLength
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    if (this.options.maxLength !== undefined && valueLength > this.options.maxLength) {
      errors.push(
        new ValidationError(
          field,
          `String length must be at most ${this.options.maxLength}, got ${valueLength}`,
          this.errorValue(value),
          Codes.MaxLength
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    // Skip the match entirely for oversized inputs: `RegExp` is a
    // backtracking engine, so MAX_PATTERN_INPUT_LENGTH bounds worst-case
    // match cost. `value.length` counts UTF-16 code units.
    const patternSkipped = value.length > MAX_PATTERN_INPUT_LENGTH;
    if (this.fullPattern && (patternSkipped || !this.fullPattern.test(value))) {
      errors.push(
        new ValidationError(
          field,
          `String does not match pattern ${this.options.pattern!.source}`,
          this.errorValue(value),
          Codes.Pattern
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    if (this.choicesSet && !this.choicesSet.has(value)) {
      errors.push(
        new ValidationError(
          field,
          `Value must be one of ${this.options.choices!.join(', ')}, got '${this.displayValue(value)}'`,
          this.errorValue(value),
          Codes.Choices
        )
      );
    }

    return errors;
  }
}

/**
 * Number validator options.
 */
export interface NumberValidatorOptions extends BaseValidatorOptions {
  minValue?: number;
  maxValue?: number;
  integer?: boolean;
}

/**
 * Number validator (for integers and floats).
 */
export class NumberValidator extends Validator<number> {
  constructor(private readonly options: NumberValidatorOptions = {}) {
    super(options);
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (typeof value !== 'number' || Number.isNaN(value)) {
      const expected = this.options.integer ? 'integer' : 'number';
      return new ValidationError(
        field,
        `Expected ${expected}, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }

    if (this.options.integer && !Number.isInteger(value)) {
      return new ValidationError(
        field,
        `Expected integer, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }

    return null;
  }

  protected validateConstraints(value: number, field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.options.minValue !== undefined && value < this.options.minValue) {
      errors.push(
        new ValidationError(
          field,
          `Value must be at least ${this.options.minValue}, got ${this.displayValue(value)}`,
          this.errorValue(value),
          Codes.MinValue
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    if (this.options.maxValue !== undefined && value > this.options.maxValue) {
      errors.push(
        new ValidationError(
          field,
          `Value must be at most ${this.options.maxValue}, got ${this.displayValue(value)}`,
          this.errorValue(value),
          Codes.MaxValue
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    return errors;
  }
}

/**
 * Boolean validator.
 */
export class BooleanValidator extends Validator<boolean> {
  constructor(options: BaseValidatorOptions = {}) {
    super(options);
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (typeof value !== 'boolean') {
      return new ValidationError(
        field,
        `Expected boolean, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }
    return null;
  }

  protected validateConstraints(_value: boolean, _field: string): ValidationError[] {
    return [];
  }
}

/**
 * Email validator options.
 */
export interface EmailValidatorOptions extends BaseValidatorOptions {
  minLength?: number;
  maxLength?: number;
  choices?: string[];
}

/**
 * Email validator.
 */
export class EmailValidator extends StringValidator {
  private static readonly emailPattern = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+[.][a-zA-Z]{2,}$/;

  constructor(options: EmailValidatorOptions = {}) {
    super({ ...options, pattern: EmailValidator.emailPattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors = super.validateConstraints(value, field);

    return errors.map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Invalid email address: ${this.displayValue(value)}`, this.errorValue(value), Codes.Email)
        : error
    );
  }
}

/**
 * URL validator options.
 */
export interface UrlValidatorOptions extends BaseValidatorOptions {
  minLength?: number;
  maxLength?: number;
  choices?: string[];
}

/**
 * URL validator.
 */
export class UrlValidator extends StringValidator {
  private static readonly urlPattern = /^https?:[/][/][^ /$.?#].[^ ]*$/;

  constructor(options: UrlValidatorOptions = {}) {
    super({ ...options, pattern: UrlValidator.urlPattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    const errors = super.validateConstraints(value, field);

    return errors.map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Invalid URL: ${this.displayValue(value)}`, this.errorValue(value), Codes.Url)
        : error
    );
  }
}

/**
 * Array validator options.
 */
export interface ArrayValidatorOptions<T = unknown> extends BaseValidatorOptions {
  itemValidator?: Validator<T>;
  minLength?: number;
  maxLength?: number;
}

/**
 * Array validator.
 */
export class ArrayValidator<T = unknown> extends Validator<T[]> {
  constructor(private readonly options: ArrayValidatorOptions<T> = {}) {
    super(options);
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (!Array.isArray(value)) {
      return new ValidationError(
        field,
        `Expected array, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }
    return null;
  }

  protected validateConstraints(value: T[], field: string): ValidationError[] {
    const errors: ValidationError[] = [];

    if (this.options.minLength !== undefined && value.length < this.options.minLength) {
      errors.push(
        new ValidationError(
          field,
          `Array length must be at least ${this.options.minLength}, got ${value.length}`,
          this.errorValue(value),
          Codes.MinLength
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    if (this.options.maxLength !== undefined && value.length > this.options.maxLength) {
      errors.push(
        new ValidationError(
          field,
          `Array length must be at most ${this.options.maxLength}, got ${value.length}`,
          this.errorValue(value),
          Codes.MaxLength
        )
      );
      if (this.failFast) {
        return errors;
      }
    }

    if (this.options.itemValidator) {
      const itemValidator = this.options.itemValidator;
      itemValidator.failFast = this.failFast;
      for (let index = 0; index < value.length; index++) {
        const itemResult = itemValidator.validate(value[index], `${field}[${index}]`);
        for (const error of itemResult.errors) {
          errors.push(error);
        }
        if (this.failFast && errors.length > 0) {
          return errors;
        }
      }
    }

    return errors;
  }
}

/**
 * Object validator options.
 */
export interface ObjectValidatorOptions extends BaseValidatorOptions {
  schema?: Record<string, Validator<unknown>>;
  strict?: boolean;
}

/**
 * Object validator.
 */
export class ObjectValidator extends Validator<Record<string, unknown>> {
  constructor(private readonly options: ObjectValidatorOptions = {}) {
    super(options);
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (typeof value !== 'object' || value === null || Array.isArray(value)) {
      return new ValidationError(
        field,
        `Expected object, got ${kindName(value)}`,
        this.errorValue(value),
        Codes.Type
      );
    }
    return null;
  }

  protected validateConstraints(value: Record<string, unknown>, field: string): ValidationError[] {
    const errors: ValidationError[] = [];
    const schema = this.options.schema;

    if (schema) {
      for (const [key, validator] of Object.entries(schema)) {
        validator.failFast = this.failFast;
        if (key in value) {
          const result = validator.validate(value[key], `${field}.${key}`);
          for (const error of result.errors) {
            errors.push(error);
          }
        } else if (validator.required) {
          errors.push(
            new ValidationError(`${field}.${key}`, 'Field is required', undefined, Codes.Required)
          );
        }
        if (this.failFast && errors.length > 0) {
          return errors;
        }
      }

      if (this.options.strict) {
        for (const key of Object.keys(value)) {
          if (!(key in schema)) {
            errors.push(
              new ValidationError(`${field}.${key}`, 'Unknown field', this.errorValue(value[key]), Codes.UnknownField)
            );
            if (this.failFast) {
              return errors;
            }
          }
        }
      }
    }

    return errors;
  }
}

/**
 * UUID validator.
 */
export class UuidValidator extends StringValidator {
  private static readonly uuidPattern = /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

  constructor(options: BaseValidatorOptions = {}) {
    super({ ...options, pattern: UuidValidator.uuidPattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    return super.validateConstraints(value, field).map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Invalid UUID: ${this.displayValue(value)}`, this.errorValue(value), Codes.Uuid)
        : error
    );
  }
}

/**
 * Date validator.
 */
export class DateValidator extends StringValidator {
  private static readonly datePattern = /^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])$/;

  constructor(options: BaseValidatorOptions = {}) {
    super({ ...options, pattern: DateValidator.datePattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    return super.validateConstraints(value, field).map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Invalid date: ${this.displayValue(value)}`, this.errorValue(value), Codes.Date)
        : error
    );
  }
}

/**
 * Datetime validator.
 */
export class DateTimeValidator extends StringValidator {
  private static readonly datetimePattern = /^[0-9]{4}-(0[1-9]|1[0-2])-(0[1-9]|[12][0-9]|3[01])T([01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]([.][0-9]+)?(Z|[+-]([01][0-9]|2[0-3]):[0-5][0-9])$/;

  constructor(options: BaseValidatorOptions = {}) {
    super({ ...options, pattern: DateTimeValidator.datetimePattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    return super.validateConstraints(value, field).map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Invalid datetime: ${this.displayValue(value)}`, this.errorValue(value), Codes.DateTime)
        : error
    );
  }
}

/**
 * Numeric validator.
 */
export class NumericValidator extends StringValidator {
  private static readonly numericPattern = /^-?(0|[1-9][0-9]*)([.][0-9]+)?([eE][+-]?[0-9]+)?$/;

  constructor(options: BaseValidatorOptions = {}) {
    super({ ...options, pattern: NumericValidator.numericPattern });
  }

  protected validateConstraints(value: string, field: string): ValidationError[] {
    return super.validateConstraints(value, field).map((error) =>
      this.isPatternError(error)
        ? new ValidationError(field, `Expected numeric string, got ${this.displayValue(value)}`, this.errorValue(value), Codes.Numeric)
        : error
    );
  }
}

/**
 * Not-empty validator.
 */
export class NotEmptyValidator extends Validator<unknown> {
  constructor(options: BaseValidatorOptions = {}) {
    super(options);
  }

  protected validateType(value: unknown, field: string): ValidationError | null {
    if (
      typeof value === 'string' ||
      Array.isArray(value) ||
      (typeof value === 'object' && value !== null)
    ) {
      return null;
    }
    return new ValidationError(
      field,
      `Expected string, array, or object, got ${kindName(value)}`,
      this.errorValue(value),
      Codes.Type
    );
  }

  protected validateConstraints(value: unknown, field: string): ValidationError[] {
    if (typeof value === 'string' && value.length === 0) {
      return [new ValidationError(field, 'Value must not be empty', this.errorValue(value), Codes.NotEmpty)];
    }
    if (Array.isArray(value) && value.length === 0) {
      return [new ValidationError(field, 'Value must not be empty', this.errorValue(value), Codes.NotEmpty)];
    }
    if (
      typeof value === 'object' &&
      value !== null &&
      !Array.isArray(value) &&
      Object.keys(value as Record<string, unknown>).length === 0
    ) {
      return [new ValidationError(field, 'Value must not be empty', this.errorValue(value), Codes.NotEmpty)];
    }
    return [];
  }
}

/**
 * Deep equality for JSON-shaped values.
 */
function deepEqual(a: unknown, b: unknown): boolean {
  if (Object.is(a, b)) {
    return true;
  }
  if (a === null || b === null) {
    return false;
  }
  if (typeof a !== typeof b) {
    return false;
  }
  if (typeof a !== 'object') {
    return false;
  }
  if (Array.isArray(a) !== Array.isArray(b)) {
    return false;
  }
  if (Array.isArray(a) && Array.isArray(b)) {
    if (a.length !== b.length) {
      return false;
    }
    for (let i = 0; i < a.length; i++) {
      if (!deepEqual(a[i], b[i])) {
        return false;
      }
    }
    return true;
  }
  const aObj = a as Record<string, unknown>;
  const bObj = b as Record<string, unknown>;
  const aKeys = Object.keys(aObj);
  const bKeys = Object.keys(bObj);
  if (aKeys.length !== bKeys.length) {
    return false;
  }
  for (const key of aKeys) {
    if (!Object.prototype.hasOwnProperty.call(bObj, key)) {
      return false;
    }
    if (!deepEqual(aObj[key], bObj[key])) {
      return false;
    }
  }
  return true;
}

/**
 * Enum validator options.
 */
export interface EnumValidatorOptions<T = unknown> extends BaseValidatorOptions {
  values?: T[];
}

/**
 * Enum validator.
 */
export class EnumValidator<T = unknown> extends Validator<T> {
  private readonly values: T[];

  constructor(options: EnumValidatorOptions<T> = {}) {
    super(options);
    this.values = options.values ?? [];
  }

  protected validateType(_value: unknown, _field: string): ValidationError | null {
    return null;
  }

  protected validateConstraints(value: T, field: string): ValidationError[] {
    if (this.values.some((v) => deepEqual(v, value))) {
      return [];
    }

    const valuesString = this.values
      .map((v) => JSON.stringify(v) ?? String(v))
      .join(', ');
    // Bare `***` in the JSON.stringify slot when sensitive.
    const valueString = this.sensitive
      ? '***'
      : JSON.stringify(value) ?? String(value);

    return [
      new ValidationError(
        field,
        `Value must be one of ${valuesString}, got ${valueString}`,
        this.errorValue(value),
        Codes.Enum
      ),
    ];
  }
}

/**
 * One-of validator options.
 */
export interface OneOfValidatorOptions<T = unknown> extends BaseValidatorOptions {
  validators: Validator<T>[];
}

/**
 * One-of validator.
 */
export class OneOfValidator<T = unknown> extends Validator<T> {
  private readonly branches: Validator<T>[];

  constructor(options: OneOfValidatorOptions<T> = { validators: [] }) {
    super(options);
    this.branches = options.validators ?? [];
  }

  protected validateType(_value: unknown, _field: string): ValidationError | null {
    return null;
  }

  protected validateConstraints(value: T, field: string): ValidationError[] {
    for (const branch of this.branches) {
      branch.failFast = this.failFast;
      const branchResult = branch.validate(value, field);
      if (branchResult.errors.length === 0) {
        return [];
      }
      if (this.failFast) {
        break;
      }
    }
    return [
      new ValidationError(
        field,
        'Value does not match any allowed schema',
        this.errorValue(value),
        Codes.OneOf
      ),
    ];
  }
}

/**
 * Schema for validating complex data structures.
 */
export class Schema {
  private readonly schema: Record<string, Validator<unknown>>;
  private readonly strict: boolean;
  private readonly failFast: boolean;

  constructor(
    schema: Record<string, Validator<unknown>>,
    options?: { strict?: boolean; failFast?: boolean }
  ) {
    this.schema = schema;
    this.strict = options?.strict ?? false;
    this.failFast = options?.failFast ?? false;
  }

  /**
   * Validate data against the schema.
   */
  validate(data: unknown): ValidationResult {
    const result = new ValidationResult();

    if (typeof data !== 'object' || data === null || Array.isArray(data)) {
      result.addError(
        new ValidationError('root', `Expected object, got ${kindName(data)}`, data, Codes.Type)
      );
      return result;
    }

    const record = data as Record<string, unknown>;

    for (const [field, validator] of Object.entries(this.schema)) {
      validator.failFast = this.failFast;
      if (field in record) {
        const fieldResult = validator.validate(record[field], field);
        for (const error of fieldResult.errors) {
          result.addError(error);
        }
      } else if (validator.required) {
        result.addError(
          new ValidationError(field, 'Field is required', undefined, Codes.Required)
        );
      }
      if (this.failFast && !result.valid) {
        return result;
      }
    }

    if (this.strict) {
      for (const key of Object.keys(record)) {
        if (!(key in this.schema)) {
          result.addError(
            new ValidationError(key, 'Unknown field', record[key], Codes.UnknownField)
          );
          if (this.failFast) {
            return result;
          }
        }
      }
    }

    return result;
  }

  /**
   * Validate data and throw an aggregate exception if invalid.
   */
  validateOrThrow(data: unknown): void {
    const result = this.validate(data);
    if (!result.valid) {
      throw new ValidationErrors(result.errors);
    }
  }
}

/**
 * Regex validator options.
 */
export interface RegexValidatorOptions extends Omit<StringValidatorOptions, 'pattern'> {
  pattern: RegExp | string;
}

/**
 * Validator factory functions.
 */
function string(options?: StringValidatorOptions): StringValidator {
  return new StringValidator(options);
}

function number(options?: NumberValidatorOptions): NumberValidator {
  return new NumberValidator(options);
}

function int(options?: NumberValidatorOptions): NumberValidator {
  return new NumberValidator({ ...options, integer: true });
}

function boolean(options?: BaseValidatorOptions): BooleanValidator {
  return new BooleanValidator(options);
}

function email(options?: EmailValidatorOptions): EmailValidator {
  return new EmailValidator(options);
}

function url(options?: UrlValidatorOptions): UrlValidator {
  return new UrlValidator(options);
}

function array<T = unknown>(options?: ArrayValidatorOptions<T>): ArrayValidator<T> {
  return new ArrayValidator(options);
}

function object(options?: ObjectValidatorOptions): ObjectValidator {
  return new ObjectValidator(options);
}

function uuid(options?: BaseValidatorOptions): UuidValidator {
  return new UuidValidator(options);
}

function date(options?: BaseValidatorOptions): DateValidator {
  return new DateValidator(options);
}

function datetime(options?: BaseValidatorOptions): DateTimeValidator {
  return new DateTimeValidator(options);
}

function numeric(options?: BaseValidatorOptions): NumericValidator {
  return new NumericValidator(options);
}

function notEmpty(options?: BaseValidatorOptions): NotEmptyValidator {
  return new NotEmptyValidator(options);
}

function enumValidator<T = unknown>(options?: EnumValidatorOptions<T>): EnumValidator<T> {
  return new EnumValidator(options);
}

function oneOf<T = unknown>(...branches: Validator<T>[]): OneOfValidator<T>;
function oneOf<T = unknown>(options: OneOfValidatorOptions<T>): OneOfValidator<T>;
function oneOf<T = unknown>(
  first?: Validator<T> | OneOfValidatorOptions<T>,
  ...rest: Validator<T>[]
): OneOfValidator<T> {
  if (first === undefined) {
    return new OneOfValidator<T>({ validators: [] });
  }
  if (first instanceof Validator) {
    return new OneOfValidator<T>({ validators: [first, ...rest] });
  }
  return new OneOfValidator<T>(first);
}

function regex(pattern: RegExp | string, options?: StringValidatorOptions): StringValidator;
function regex(options: RegexValidatorOptions): StringValidator;
function regex(
  patternOrOptions: RegExp | string | RegexValidatorOptions,
  options?: StringValidatorOptions
): StringValidator {
  if (patternOrOptions instanceof RegExp) {
    return new StringValidator({ ...options, pattern: patternOrOptions });
  }
  if (typeof patternOrOptions === 'string') {
    return new StringValidator({ ...options, pattern: new RegExp(patternOrOptions) });
  }
  const opts = patternOrOptions as RegexValidatorOptions;
  const pattern = typeof opts.pattern === 'string' ? new RegExp(opts.pattern) : opts.pattern;
  return new StringValidator({ ...opts, pattern });
}

export const validators = {
  string,
  number,
  int,
  boolean,
  email,
  url,
  array,
  object,
  uuid,
  date,
  datetime,
  numeric,
  notEmpty,
  enum: enumValidator,
  oneOf,
  regex,
};

export default validators;

/**
 * Conformance tests driven by the shared canonical vectors in
 * ../testdata/vectors/cases.json (see ../testdata/SPEC.md).
 */

import * as fs from 'fs';
import * as path from 'path';
import { fileURLToPath } from 'url';
import { describe, it, expect } from 'vitest';
import { Schema, Validator, ValidationError, validators } from './index';

const here = path.dirname(fileURLToPath(import.meta.url));
const casesPath = path.resolve(here, '../../testdata/vectors/cases.json');

interface ExpectedError {
  field: string;
  code: string;
  message: string;
}

interface ConformanceCase {
  name: string;
  schema: Record<string, any>;
  input: any;
  expected: ExpectedError[];
}

const cases: ConformanceCase[] = JSON.parse(fs.readFileSync(casesPath, 'utf-8'));

/**
 * Build a Validator from a spec DSL object (see SPEC.md).
 */
function buildValidator(spec: any): Validator<unknown> {
  const common = {
    required: spec.required ?? true,
    nullable: spec.nullable ?? false,
    sensitive: spec.sensitive ?? false,
  };

  switch (spec.type) {
    case 'string':
      return validators.string({
        ...common,
        minLength: spec.minLength,
        maxLength: spec.maxLength,
        pattern: spec.pattern !== undefined ? new RegExp(spec.pattern) : undefined,
        choices: spec.choices,
      });
    case 'integer':
      return validators.int({
        ...common,
        minValue: spec.minValue,
        maxValue: spec.maxValue,
      });
    case 'number':
      return validators.number({
        ...common,
        minValue: spec.minValue,
        maxValue: spec.maxValue,
      });
    case 'boolean':
      return validators.boolean(common);
    case 'email':
      return validators.email({
        ...common,
        minLength: spec.minLength,
        maxLength: spec.maxLength,
        choices: spec.choices,
      });
    case 'url':
      return validators.url({
        ...common,
        minLength: spec.minLength,
        maxLength: spec.maxLength,
        choices: spec.choices,
      });
    case 'array':
      return validators.array({
        ...common,
        itemValidator: spec.items ? buildValidator(spec.items) : undefined,
        minLength: spec.minLength,
        maxLength: spec.maxLength,
      });
    case 'object': {
      const schema: Record<string, Validator<unknown>> = {};
      for (const [key, fieldSpec] of Object.entries(spec.fields ?? {})) {
        schema[key] = buildValidator(fieldSpec);
      }
      return validators.object({
        ...common,
        schema,
        strict: spec.strict ?? false,
      });
    }
    case 'uuid':
      return validators.uuid(common);
    case 'date':
      return validators.date(common);
    case 'datetime':
      return validators.datetime(common);
    case 'numeric':
      return validators.numeric(common);
    case 'not_empty':
      return validators.notEmpty(common);
    case 'enum':
      return validators.enum({ ...common, values: spec.values });
    case 'one_of':
      return validators.oneOf({
        ...common,
        validators: (spec.of ?? []).map(buildValidator),
      });
    default:
      throw new Error(`Unknown validator type: ${spec.type}`);
  }
}

/**
 * Build a Schema from a schema spec: { "fields": {...}, "strict": false }.
 */
function buildSchema(spec: any): Schema {
  const fields: Record<string, Validator<unknown>> = {};
  for (const [key, fieldSpec] of Object.entries(spec.fields ?? {})) {
    fields[key] = buildValidator(fieldSpec);
  }
  return new Schema(fields, { strict: spec.strict ?? false });
}

/**
 * ValidationError.message is stored as "<field>: <message>"; recover the
 * raw message for comparison with the canonical vectors.
 */
function rawMessage(error: ValidationError): string {
  const prefix = `${error.field}: `;
  return error.message.startsWith(prefix)
    ? error.message.slice(prefix.length)
    : error.message;
}

function sortTriples(
  triples: Array<[string, string, string]>
): Array<[string, string, string]> {
  return [...triples].sort(
    (a, b) =>
      a[0].localeCompare(b[0]) ||
      a[1].localeCompare(b[1]) ||
      a[2].localeCompare(b[2])
  );
}

describe('Conformance', () => {
  for (const c of cases) {
    it(c.name, () => {
      const schema = buildSchema(c.schema);
      const result = schema.validate(c.input);

      const actual: Array<[string, string, string]> = result.errors.map(e => [
        e.field,
        e.code,
        rawMessage(e),
      ]);
      const expected: Array<[string, string, string]> = c.expected.map(e => [
        e.field,
        e.code,
        e.message,
      ]);

      expect(sortTriples(actual)).toEqual(sortTriples(expected));
    });
  }
});

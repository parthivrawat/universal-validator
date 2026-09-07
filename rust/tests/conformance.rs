//! Conformance test against the shared test corpus in `../testdata/vectors/cases.json`.
//!
//! Implements the `buildValidator`/`buildSchema` test helpers described in
//! `../testdata/SPEC.md` and asserts that `(field, code, message)` triples match exactly.

use serde_json::Value;
use universal_validator::{validators, Schema, Validator};

fn build_validator(spec: &Value) -> Box<dyn Validator> {
    let ty = spec
        .get("type")
        .and_then(Value::as_str)
        .expect("validator spec requires a `type`");
    let optional = spec.get("required").and_then(Value::as_bool) == Some(false);
    let nullable = spec.get("nullable").and_then(Value::as_bool) == Some(true);
    let sensitive = spec.get("sensitive").and_then(Value::as_bool) == Some(true);

    macro_rules! finish {
        ($v:expr) => {{
            let v = $v;
            let v = if optional { v.optional() } else { v };
            let v = if nullable { v.nullable() } else { v };
            let v = if sensitive { v.sensitive(true) } else { v };
            Box::new(v) as Box<dyn Validator>
        }};
    }

    match ty {
        "string" => {
            let mut v = validators::string();
            if let Some(n) = spec.get("minLength").and_then(Value::as_u64) {
                v = v.min(n as usize);
            }
            if let Some(n) = spec.get("maxLength").and_then(Value::as_u64) {
                v = v.max(n as usize);
            }
            if let Some(p) = spec.get("pattern").and_then(Value::as_str) {
                v = v.try_pattern(p).expect("invalid pattern in test spec");
            }
            if let Some(c) = spec.get("choices").and_then(Value::as_array) {
                v = v.choices(
                    c.iter()
                        .map(|x| x.as_str().expect("choice must be a string").to_string())
                        .collect(),
                );
            }
            finish!(v)
        }
        "integer" => {
            let mut v = validators::int();
            if let Some(n) = spec.get("minValue").and_then(Value::as_i64) {
                v = v.min(n);
            }
            if let Some(n) = spec.get("maxValue").and_then(Value::as_i64) {
                v = v.max(n);
            }
            finish!(v)
        }
        "number" => {
            let mut v = validators::float();
            if let Some(n) = spec.get("minValue").and_then(Value::as_f64) {
                v = v.min(n);
            }
            if let Some(n) = spec.get("maxValue").and_then(Value::as_f64) {
                v = v.max(n);
            }
            finish!(v)
        }
        "boolean" => finish!(validators::bool()),
        "email" => finish!(validators::email()),
        "url" => finish!(validators::url()),
        "array" => {
            let mut v = validators::vec();
            if let Some(n) = spec.get("minLength").and_then(Value::as_u64) {
                v = v.min(n as usize);
            }
            if let Some(n) = spec.get("maxLength").and_then(Value::as_u64) {
                v = v.max(n as usize);
            }
            if let Some(items) = spec.get("items") {
                v = v.item(build_validator(items));
            }
            finish!(v)
        }
        "object" => {
            let mut v = validators::map();
            if let Some(fields) = spec.get("fields").and_then(Value::as_object) {
                for (name, field_spec) in fields {
                    v = v.field(name.clone(), build_validator(field_spec));
                }
            }
            if spec.get("strict").and_then(Value::as_bool) == Some(true) {
                v = v.strict(true);
            }
            finish!(v)
        }
        "uuid" => finish!(validators::uuid()),
        "date" => finish!(validators::date()),
        "datetime" => finish!(validators::datetime()),
        "numeric" => finish!(validators::numeric()),
        "not_empty" => finish!(validators::not_empty()),
        "enum" => {
            let values = spec
                .get("values")
                .and_then(Value::as_array)
                .expect("enum spec requires `values`")
                .clone();
            finish!(validators::enum_values(values))
        }
        "one_of" => {
            let branches = spec
                .get("of")
                .and_then(Value::as_array)
                .expect("one_of spec requires `of`")
                .iter()
                .map(build_validator)
                .collect();
            finish!(validators::one_of(branches))
        }
        other => panic!("unknown validator type: {other}"),
    }
}

fn build_schema(spec: &Value) -> Schema {
    let mut schema = Schema::new();
    if let Some(fields) = spec.get("fields").and_then(Value::as_object) {
        for (name, field_spec) in fields {
            schema = schema.field(name.clone(), build_validator(field_spec));
        }
    }
    if spec.get("strict").and_then(Value::as_bool) == Some(true) {
        schema = schema.strict(true);
    }
    schema
}

#[test]
fn conformance_cases() {
    let path = concat!(
        env!("CARGO_MANIFEST_DIR"),
        "/../testdata/vectors/cases.json"
    );
    let raw = std::fs::read_to_string(path).expect("failed to read test vectors");
    let cases: Vec<Value> = serde_json::from_str(&raw).expect("failed to parse test vectors");
    assert!(!cases.is_empty(), "no conformance cases loaded");

    for case in &cases {
        let name = case["name"].as_str().unwrap_or("<unnamed>");
        let schema = build_schema(&case["schema"]);
        let result = schema.validate(&case["input"]);

        let mut actual: Vec<(String, String, String)> = result
            .errors()
            .iter()
            .map(|e| (e.field.clone(), e.code.clone(), e.message.clone()))
            .collect();
        actual.sort();

        let mut expected: Vec<(String, String, String)> = case["expected"]
            .as_array()
            .expect("case `expected` must be an array")
            .iter()
            .map(|e| {
                (
                    e["field"].as_str().unwrap().to_string(),
                    e["code"].as_str().unwrap().to_string(),
                    e["message"].as_str().unwrap().to_string(),
                )
            })
            .collect();
        expected.sort();

        assert_eq!(actual, expected, "conformance case `{name}` failed");
    }
}

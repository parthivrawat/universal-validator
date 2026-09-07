package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"testing"
)

type testCase struct {
	Name     string                 `json:"name"`
	Schema   map[string]interface{} `json:"schema"`
	Input    map[string]interface{} `json:"input"`
	Expected []testError            `json:"expected"`
}

type testError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorPair struct {
	Field   string
	Code    string
	Message string
}

func getInt(spec map[string]interface{}, key string) (*int, bool) {
	v, ok := spec[key]
	if !ok {
		return nil, false
	}
	switch n := v.(type) {
	case float64:
		i := int(n)
		return &i, true
	case int:
		i := n
		return &i, true
	}
	return nil, false
}

func getFloat(spec map[string]interface{}, key string) (*float64, bool) {
	v, ok := spec[key]
	if !ok {
		return nil, false
	}
	switch n := v.(type) {
	case float64:
		return &n, true
	case int:
		f := float64(n)
		return &f, true
	}
	return nil, false
}

func getStringSlice(spec map[string]interface{}, key string) ([]string, bool) {
	v, ok := spec[key].([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]string, len(v))
	for i, c := range v {
		s, ok := c.(string)
		if !ok {
			return nil, false
		}
		out[i] = s
	}
	return out, true
}

func getInterfaceSlice(spec map[string]interface{}, key string) ([]interface{}, bool) {
	v, ok := spec[key].([]interface{})
	if !ok {
		return nil, false
	}
	out := make([]interface{}, len(v))
	copy(out, v)
	return out, true
}

func buildValidator(spec map[string]interface{}) Validator {
	typ, _ := spec["type"].(string)

	required := true
	if r, ok := spec["required"].(bool); ok {
		required = r
	}
	optional := !required

	nullable := false
	if n, ok := spec["nullable"].(bool); ok {
		nullable = n
	}

	baseOpts := []Option{}
	if optional {
		baseOpts = append(baseOpts, Optional())
	}
	if nullable {
		baseOpts = append(baseOpts, Nullable())
	}
	if s, ok := spec["sensitive"].(bool); ok && s {
		baseOpts = append(baseOpts, Sensitive())
	}

	switch typ {
	case "string":
		opts := append([]Option(nil), baseOpts...)
		if v, ok := getInt(spec, "minLength"); ok {
			opts = append(opts, MinLength(*v))
		}
		if v, ok := getInt(spec, "maxLength"); ok {
			opts = append(opts, MaxLength(*v))
		}
		if p, ok := spec["pattern"].(string); ok {
			opts = append(opts, Pattern(regexp.MustCompile(p)))
		}
		if c, ok := getStringSlice(spec, "choices"); ok {
			opts = append(opts, Choices(c...))
		}
		return String(opts...)
	case "integer":
		opts := append([]Option(nil), baseOpts...)
		if v, ok := getInt(spec, "minValue"); ok {
			opts = append(opts, MinValue(*v))
		}
		if v, ok := getInt(spec, "maxValue"); ok {
			opts = append(opts, MaxValue(*v))
		}
		return Int(opts...)
	case "number":
		opts := append([]Option(nil), baseOpts...)
		if v, ok := getFloat(spec, "minValue"); ok {
			opts = append(opts, MinFloat(*v))
		}
		if v, ok := getFloat(spec, "maxValue"); ok {
			opts = append(opts, MaxFloat(*v))
		}
		return Float(opts...)
	case "boolean":
		return Bool(baseOpts...)
	case "email":
		return Email(baseOpts...)
	case "url":
		return URL(baseOpts...)
	case "uuid":
		return UUID(baseOpts...)
	case "date":
		return Date(baseOpts...)
	case "datetime":
		return DateTime(baseOpts...)
	case "numeric":
		return Numeric(baseOpts...)
	case "not_empty":
		return NotEmpty(baseOpts...)
	case "enum":
		values, ok := getInterfaceSlice(spec, "values")
		if !ok {
			panic(fmt.Sprintf("enum missing values: %v", spec))
		}
		return Enum(values, baseOpts...)
	case "one_of":
		raw, ok := spec["of"].([]interface{})
		if !ok {
			panic(fmt.Sprintf("one_of missing of: %v", spec))
		}
		branches := make([]Validator, len(raw))
		for i, r := range raw {
			branches[i] = buildValidator(r.(map[string]interface{}))
		}
		return OneOf(branches, baseOpts...)
	case "array":
		opts := append([]Option(nil), baseOpts...)
		if v, ok := getInt(spec, "minLength"); ok {
			opts = append(opts, MinLength(*v))
		}
		if v, ok := getInt(spec, "maxLength"); ok {
			opts = append(opts, MaxLength(*v))
		}
		if it, ok := spec["items"].(map[string]interface{}); ok {
			opts = append(opts, Items(buildValidator(it)))
		}
		return Slice(opts...)
	case "object":
		opts := append([]Option(nil), baseOpts...)
		if fields, ok := spec["fields"].(map[string]interface{}); ok {
			schema := make(map[string]Validator)
			for k, v := range fields {
				schema[k] = buildValidator(v.(map[string]interface{}))
			}
			opts = append(opts, Fields(schema))
		}
		if s, ok := spec["strict"].(bool); ok && s {
			opts = append(opts, Strict())
		}
		return Map(opts...)
	default:
		panic(fmt.Sprintf("unknown validator type: %s", typ))
	}
}

// buildSchema translates the top-level conformance spec DSL into a Schema.
func buildSchema(spec map[string]interface{}) *Schema {
	fields := make(map[string]Validator)
	if f, ok := spec["fields"].(map[string]interface{}); ok {
		for k, v := range f {
			fields[k] = buildValidator(v.(map[string]interface{}))
		}
	}

	opts := []Option{}
	if s, ok := spec["strict"].(bool); ok && s {
		opts = append(opts, Strict())
	}
	return NewSchema(fields, opts...)
}

func TestConformance(t *testing.T) {
	f, err := os.Open("../testdata/vectors/cases.json")
	if err != nil {
		t.Fatalf("open cases.json: %v", err)
	}
	defer f.Close()

	var cases []testCase
	if err := json.NewDecoder(f).Decode(&cases); err != nil {
		t.Fatalf("decode cases.json: %v", err)
	}

	ctx := context.Background()
	for _, tc := range cases {
		schema := buildSchema(tc.Schema)
		result := schema.Validate(ctx, tc.Input)

		actual := make([]errorPair, 0, len(result.Errors))
		for _, e := range result.Errors {
			actual = append(actual, errorPair{Field: e.Field, Code: e.Code, Message: e.Message})
		}
		sort.Slice(actual, func(i, j int) bool {
			if actual[i].Field != actual[j].Field {
				return actual[i].Field < actual[j].Field
			}
			if actual[i].Code != actual[j].Code {
				return actual[i].Code < actual[j].Code
			}
			return actual[i].Message < actual[j].Message
		})

		expected := make([]errorPair, 0, len(tc.Expected))
		for _, e := range tc.Expected {
			expected = append(expected, errorPair{Field: e.Field, Code: e.Code, Message: e.Message})
		}
		sort.Slice(expected, func(i, j int) bool {
			if expected[i].Field != expected[j].Field {
				return expected[i].Field < expected[j].Field
			}
			if expected[i].Code != expected[j].Code {
				return expected[i].Code < expected[j].Code
			}
			return expected[i].Message < expected[j].Message
		})

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("case %q: actual %v, expected %v", tc.Name, actual, expected)
		}
	}
}

package rez

import (
	"reflect"
	"testing"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
	"github.com/stretchr/testify/assert"
)

type Embedded struct {
	X int
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		schema   *api.Schema
		value    any
		failures []Validation
		options  ValidationOptions
	}{{
		name: "minimum on",
		schema: &api.Schema{
			Minimum: ptrTo(0),
		},
		value: 0,
	}, {
		name: "minimum fail",
		schema: &api.Schema{
			Minimum: ptrTo(0),
		},
		value: -1,
		failures: []Validation{{
			Rule:    ValidationRuleMinimum,
			Path:    []string{},
			Message: "-1 is below the minimum of 0",
		}},
	}, {
		name: "minimum over",
		schema: &api.Schema{
			Minimum: ptrTo(0),
		},
		value: 1,
	}, {
		name: "minimum exclusive",
		schema: &api.Schema{
			Minimum:          ptrTo(0),
			ExclusiveMinimum: true,
		},
		value: 0.001,
	}, {
		name: "minimum exclusive fail",
		schema: &api.Schema{
			Minimum:          ptrTo(0),
			ExclusiveMinimum: true,
		},
		value: 0,
		failures: []Validation{{
			Rule:    ValidationRuleMinimum,
			Path:    []string{},
			Message: "0 is below the minimum of 0",
		}},
	}, {
		name: "maximum on",
		schema: &api.Schema{
			Maximum: ptrTo(0),
		},
		value: 0,
	}, {
		name: "maximum fail",
		schema: &api.Schema{
			Maximum: ptrTo(0),
		},
		value: 1,
		failures: []Validation{{
			Rule:    ValidationRuleMaximum,
			Path:    []string{},
			Message: "1 exceeds the maximum of 0",
		}},
	}, {
		name: "maximum under",
		schema: &api.Schema{
			Maximum: ptrTo(0),
		},
		value: -1,
	}, {
		name: "maximum exclusive",
		schema: &api.Schema{
			Maximum:          ptrTo(0),
			ExclusiveMaximum: true,
		},
		value: -0.001,
	}, {
		name: "maximum exclusive fail",
		schema: &api.Schema{
			Maximum:          ptrTo(0),
			ExclusiveMaximum: true,
		},
		value: 0,
		failures: []Validation{{
			Rule:    ValidationRuleMaximum,
			Path:    []string{},
			Message: "0 exceeds the maximum of 0",
		}},
	}, {
		name: "multipleof zero",
		schema: &api.Schema{
			MultipleOf: 2,
		},
		value: 0,
	}, {
		name: "multipleof zero",
		schema: &api.Schema{
			MultipleOf: 2,
		},
		value: 2,
	}, {
		name: "multipleof fail",
		schema: &api.Schema{
			MultipleOf: 2,
		},
		value: 1,
		failures: []Validation{{
			Rule:    ValidationRuleMultipleOf,
			Path:    []string{},
			Message: "1 is not a multiple of 2",
		}},
	}, {
		name: "minlength on",
		schema: &api.Schema{
			MinLength: ptrTo(2),
		},
		value: "ab",
	}, {
		name: "minlength over",
		schema: &api.Schema{
			MinLength: ptrTo(2),
		},
		value: "abc",
	}, {
		name: "minlength under",
		schema: &api.Schema{
			MinLength: ptrTo(2),
		},
		value: "a",
		failures: []Validation{{
			Rule:    ValidationRuleMinLength,
			Path:    []string{},
			Message: "1 does not meet the minimum length of 2",
		}},
	}, {
		name: "maxlength on",
		schema: &api.Schema{
			MaxLength: 2,
		},
		value: "ab",
	}, {
		name: "maxlength under",
		schema: &api.Schema{
			MaxLength: 2,
		},
		value: "a",
	}, {
		name: "maxlength over",
		schema: &api.Schema{
			MaxLength: 2,
		},
		value: "abc",
		failures: []Validation{{
			Rule:    ValidationRuleMaxLength,
			Path:    []string{},
			Message: "3 exceeds the maximum length of 2",
		}},
	}, {
		name: "minitems on",
		schema: &api.Schema{
			MinItems: ptrTo(2),
		},
		value: []string{"a", "b"},
	}, {
		name: "minitems over",
		schema: &api.Schema{
			MinItems: ptrTo(2),
		},
		value: []string{"a", "b", "c"},
	}, {
		name: "minitems under",
		schema: &api.Schema{
			MinItems: ptrTo(2),
		},
		value: []string{"a"},
		failures: []Validation{{
			Rule:    ValidationRuleMinItems,
			Path:    []string{},
			Message: "1 does not meet the minimum items of 2",
		}},
	}, {
		name: "maxitems on",
		schema: &api.Schema{
			MaxItems: 2,
		},
		value: []string{"a", "b"},
	}, {
		name: "maxitems under",
		schema: &api.Schema{
			MaxItems: 2,
		},
		value: []string{"a"},
	}, {
		name: "maxitems over",
		schema: &api.Schema{
			MaxItems: 2,
		},
		value: []string{"a", "b", "c"},
		failures: []Validation{{
			Rule:    ValidationRuleMaxItems,
			Path:    []string{},
			Message: "3 exceeds the maximum items of 2",
		}},
	}, {
		name: "items valid",
		schema: &api.Schema{
			Items: &api.Schema{
				MultipleOf: 2,
			},
		},
		value: []int{0, 2, 4},
	}, {
		name: "items fail",
		schema: &api.Schema{
			Items: &api.Schema{
				MultipleOf: 2,
			},
		},
		value: []int{0, 2, 3},
		failures: []Validation{{
			Rule:    ValidationRuleMultipleOf,
			Path:    []string{"2"},
			Message: "3 is not a multiple of 2",
		}},
	}, {
		name: "uniqueitems",
		schema: &api.Schema{
			UniqueItems: true,
		},
		value: []int{0, 2, 3},
	}, {
		name: "uniqueitems fail",
		schema: &api.Schema{
			UniqueItems: true,
		},
		value: []int{0, 2, 0},
		failures: []Validation{{
			Rule:    ValidationRuleUniqueItems,
			Path:    []string{},
			Message: "0 is not a unique item",
		}},
	}, {
		name: "minproperties on",
		schema: &api.Schema{
			MinProperties: ptrTo(2),
		},
		value: map[string]int{"a": 0, "b": 1},
	}, {
		name: "minproperties over",
		schema: &api.Schema{
			MinProperties: ptrTo(2),
		},
		value: map[string]int{"a": 0, "b": 1, "c": 2},
	}, {
		name: "minproperties under",
		schema: &api.Schema{
			MinProperties: ptrTo(2),
		},
		value: map[string]int{"a": 0},
		failures: []Validation{{
			Rule:    ValidationRuleMinProperties,
			Path:    []string{},
			Message: "1 does not meet the minimum properties of 2",
		}},
	}, {
		name: "maxproperties on",
		schema: &api.Schema{
			MaxProperties: 2,
		},
		value: map[string]int{"a": 0, "b": 1},
	}, {
		name: "maxproperties under",
		schema: &api.Schema{
			MaxProperties: 2,
		},
		value: map[string]int{"a": 0},
	}, {
		name: "maxproperties over",
		schema: &api.Schema{
			MaxProperties: 2,
		},
		value: map[string]int{"a": 0, "b": 1, "c": 2},
		failures: []Validation{{
			Rule:    ValidationRuleMaxProperties,
			Path:    []string{},
			Message: "3 exceeds the maximum properties of 2",
		}},
	}, {
		name: "additionalproperties",
		schema: &api.Schema{
			AdditionalProperties: api.SchemaForSchema(&api.Schema{
				MultipleOf: 2,
			}),
		},
		value: map[string]int{"a": 0, "b": 2},
	}, {
		name: "additionalproperties fail",
		schema: &api.Schema{
			AdditionalProperties: api.SchemaForSchema(&api.Schema{
				MultipleOf: 2,
			}),
		},
		value: map[string]int{"a": 0, "b": 1},
		failures: []Validation{{
			Rule:    ValidationRuleMultipleOf,
			Path:    []string{"b"},
			Message: "1 is not a multiple of 2",
		}},
	}, {
		name: "pattern",
		schema: &api.Schema{
			Pattern: `^[a-z]+\d$`,
		},
		value: "abc1",
	}, {
		name: "pattern fail",
		schema: &api.Schema{
			Pattern: `^[a-z]+\d$`,
		},
		value: "abc",
		failures: []Validation{{
			Rule:    ValidationRulePattern,
			Path:    []string{},
			Message: `abc does not match the pattern ^[a-z]+\d$`,
		}},
	}, {
		name: "format",
		schema: &api.Schema{
			Format: "email",
		},
		value: "abc",
	}, {
		name: "format",
		schema: &api.Schema{
			Format: "email",
		},
		value: "abc@gmail.com",
	}, {
		name: "format",
		schema: &api.Schema{
			Format: "email",
		},
		value:   "abc@gmail.com",
		options: ValidationOptions{EnforceFormat: true},
	}, {
		name: "format",
		schema: &api.Schema{
			Format: "email",
		},
		value:   "abc",
		options: ValidationOptions{EnforceFormat: true},
		failures: []Validation{{
			Rule:    ValidationRuleFormat,
			Path:    []string{},
			Message: `abc does not match the format email`,
		}},
	}, {
		name: "enum string",
		schema: &api.Schema{
			Enum: []any{"a", "b"},
		},
		value: "a",
	}, {
		name: "enum string fail",
		schema: &api.Schema{
			Enum: []any{"a", "b"},
		},
		value: "c",
		failures: []Validation{{
			Rule:    ValidationRuleEnum,
			Path:    []string{},
			Message: `c does not match one of the enum values [a b]`,
		}},
	}, {
		name: "enum int",
		schema: &api.Schema{
			Enum: []any{1, 2},
		},
		value: 1,
	}, {
		name: "enum int fail",
		schema: &api.Schema{
			Enum: []any{1, 2},
		},
		value: 3,
		failures: []Validation{{
			Rule:    ValidationRuleEnum,
			Path:    []string{},
			Message: `3 does not match one of the enum values [1 2]`,
		}},
	}, {
		name: "oneof",
		schema: &api.Schema{
			OneOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 3,
	}, {
		name: "oneof #2",
		schema: &api.Schema{
			OneOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 9,
	}, {
		name: "oneof fail both",
		schema: &api.Schema{
			OneOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 6,
		failures: []Validation{{
			Rule:    ValidationRuleOneOf,
			Path:    []string{},
			Message: `6 does not match one of the possible schemas`,
		}},
	}, {
		name: "oneof fail neither",
		schema: &api.Schema{
			OneOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 5,
		failures: []Validation{{
			Rule:    ValidationRuleOneOf,
			Path:    []string{},
			Message: `5 does not match one of the possible schemas`,
		}},
	}, {
		name: "allof",
		schema: &api.Schema{
			AllOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 6,
	}, {
		name: "allof one",
		schema: &api.Schema{
			AllOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 2,
		failures: []Validation{{
			Rule:    ValidationRuleAllOf,
			Path:    []string{},
			Message: `2 does not match all of the possible schemas`,
		}},
	}, {
		name: "allof zero",
		schema: &api.Schema{
			AllOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 1,
		failures: []Validation{{
			Rule:    ValidationRuleAllOf,
			Path:    []string{},
			Message: `1 does not match all of the possible schemas`,
		}},
	}, {
		name: "anyof one",
		schema: &api.Schema{
			AnyOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 2,
	}, {
		name: "anyof other",
		schema: &api.Schema{
			AnyOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 3,
	}, {
		name: "anyof both",
		schema: &api.Schema{
			AnyOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 6,
	}, {
		name: "anyof neither",
		schema: &api.Schema{
			AnyOf: []api.Schema{{
				MultipleOf: 2,
			}, {
				MultipleOf: 3,
			}},
		},
		value: 5,
		failures: []Validation{{
			Rule:    ValidationRuleAnyOf,
			Path:    []string{},
			Message: `5 does not match any of the possible schemas`,
		}},
	}, {
		name: "not",
		schema: &api.Schema{
			Not: &api.Schema{
				MultipleOf: 2,
			},
		},
		value: 5,
	}, {
		name: "not",
		schema: &api.Schema{
			Not: &api.Schema{
				MultipleOf: 2,
			},
		},
		value: 2,
		failures: []Validation{{
			Rule:    ValidationRuleNot,
			Path:    []string{},
			Message: `2 matches the not schema`,
		}},
	}, {
		name: "skip",
		schema: &api.Schema{
			MultipleOf: 2,
		},
		value:   1,
		options: ValidationOptions{Skip: true},
	}, {
		name: "nullable",
		schema: &api.Schema{
			Nullable: true,
		},
		value: nil,
	}, {
		name: "type null",
		schema: &api.Schema{
			Type: api.DataTypeNull,
		},
		value: nil,
	}, {
		name: "skip deprecated",
		schema: &api.Schema{
			Deprecated: true,
			MultipleOf: 2,
		},
		value:   1,
		options: ValidationOptions{SkipDeprecated: true},
	}, {
		name: "fail deprecated",
		schema: &api.Schema{
			Deprecated: true,
		},
		value:   0,
		options: ValidationOptions{FailDeprecated: true},
	}, {
		name: "fail deprecated",
		schema: &api.Schema{
			Deprecated: true,
		},
		value:   1,
		options: ValidationOptions{FailDeprecated: true},
		failures: []Validation{{
			Rule: ValidationRuleDeprecated,
			Path: []string{},
		}},
	}, {
		name:   "struct empty",
		schema: &api.Schema{},
		value:  struct{}{},
	}, {
		name: "struct pass",
		schema: &api.Schema{
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ X int }{X: 4},
	}, {
		name: "struct fail",
		schema: &api.Schema{
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ X int }{X: 3},
		failures: []Validation{{
			Rule:    ValidationRuleMultipleOf,
			Path:    []string{"X"},
			Message: "3 is not a multiple of 2",
		}},
	}, {
		name: "struct required given",
		schema: &api.Schema{
			Required: []string{"X"},
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ X *int }{X: ptrTo(2)},
	}, {
		name: "struct not required nil",
		schema: &api.Schema{
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ X *int }{X: nil},
	}, {
		name: "struct required fail",
		schema: &api.Schema{
			Required: []string{"X"},
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ X *int }{X: nil},
		failures: []Validation{{
			Rule:    ValidationRuleRequired,
			Path:    []string{},
			Message: "X is a required field",
		}},
	}, {
		name: "struct skip without schema",
		schema: &api.Schema{
			Properties: map[string]api.Schema{},
		},
		value: struct{ X int }{X: 0},
	}, {
		name: "struct anonymous",
		schema: &api.Schema{
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ Embedded }{Embedded{X: 4}},
	}, {
		name: "struct anonymous fail",
		schema: &api.Schema{
			Properties: map[string]api.Schema{
				"X": {MultipleOf: 2},
			},
		},
		value: struct{ Embedded }{Embedded{X: 3}},
		failures: []Validation{{
			Rule:    ValidationRuleMultipleOf,
			Path:    []string{"Embedded", "X"},
			Message: "3 is not a multiple of 2",
		}},
	}}

	for _, test := range tests {
		vp := testValidationProvider{
			options: map[reflect.Type]ValidationOptions{
				reflect.TypeOf(test.value): test.options,
			},
		}
		s := deps.New()
		v := NewValidator(vp, s)

		Validate(test.schema, test.value, v)

		if v.IsValid() && len(test.failures) > 0 {
			t.Errorf("%s passed but expected %d failures", test.name, len(test.failures))
		} else if !v.IsValid() && len(test.failures) == 0 {
			t.Errorf("%s failed but expected to pass", test.name)
		} else if len(test.failures) > 0 {
			assert.Equal(t, test.failures, *v.Validations, test.name)
		}
	}
}

type testValidationProvider struct {
	options map[reflect.Type]ValidationOptions
}

func (vp testValidationProvider) ValidationOptions(typ reflect.Type) ValidationOptions {
	return vp.options[typ]
}

func ptrTo[V any](value V) *V {
	return &value
}

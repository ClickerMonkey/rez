package rez

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
)

// Options for validating a particular type.
type ValidationOptions struct {
	// If all validation should be skipped.
	Skip bool
	// If format is set, if the string representation of the
	// value should be validated against the specified format (if its supported).
	EnforceFormat bool
	// If specifying a deprecated value should result in a validation error.
	FailDeprecated bool
}

// A validation failure
type Validation struct {
	// The path to the offending value.
	Path []string `json:"path,omitempty"`
	// The name of the schema, if any.
	Schema *string `json:"schema,omitempty"`
	// The validation rule that caused the failure.
	Rule ValidationRule `json:"rule,omitempty"`
	// A message with more details.
	Message string `json:"message,omitempty"`
}

// A validator for a specific element being validated.
// Validators all share Validations and Providers. Adding a validation
// to one adds it to the others.
type Validator struct {
	Path        []string           `json:"-"`
	Validations *[]Validation      `json:"validations"`
	Provider    ValidationProvider `json:"-"`
	Scope       *deps.Scope        `json:"-"`
}

var _ error = Validator{}
var _ HasStatus = Validator{}

// Validator implements error
func (v Validator) Error() string {
	n := len(*v.Validations)
	if n == 0 {
		return "No validation errors"
	} else if n == 1 {
		return "1 validation error"
	} else {
		return fmt.Sprintf("%d validation errors", n)
	}
}

// 400 is used, this also signals to return the error as JSON.
func (v Validator) HTTPStatus() int {
	return http.StatusBadRequest
}
func (v Validator) HTTPStatuses() []int {
	return []int{http.StatusBadRequest}
}

// Returns a child validator with the added path node. Validations are shared.
func (v Validator) Next(path string) *Validator {
	return &Validator{
		Path:        append(v.Path[:], path),
		Validations: v.Validations,
		Provider:    v.Provider,
		Scope:       v.Scope,
	}
}

// Adds a validation error to the validator. If no path is specified on the
// validation then the path of the validator is applied.
func (v *Validator) Add(msg Validation) {
	if msg.Path == nil {
		msg.Path = v.Path
	}
	*v.Validations = append(*v.Validations, msg)
}

// Creates a validator with the same path but does not share validations.
func (v Validator) Detach() Validator {
	validations := make([]Validation, 0)
	return Validator{
		Path:        v.Path[:],
		Validations: &validations,
		Provider:    v.Provider,
		Scope:       v.Scope,
	}
}

// Attaches a validtor to this validator by adding its validations to this one.
func (v *Validator) Attach(detached Validator) {
	*v.Validations = append(*v.Validations, *detached.Validations...)
}

// If a type implements this interface then it handles all validation
// for the type.
type CanValidateFull interface {
	FullValidate(v *Validator)
}

// If a type implements this interface then it handles validation
// after the normal validation process has occurred.
type CanValidatePost interface {
	PostValidate(v *Validator)
}

// An interface which helps the validation process.
type ValidationProvider interface {
	ValidationOptions(reflect.Type) ValidationOptions
}

// A rule that was broken that caused the validation failure.
type ValidationRule string

const (
	ValidationRuleType          ValidationRule = "type"
	ValidationRuleMultipleOf    ValidationRule = "multipleOf"
	ValidationRuleMaximum       ValidationRule = "maximum"
	ValidationRuleMinimum       ValidationRule = "minimum"
	ValidationRuleMaxLength     ValidationRule = "maxLength"
	ValidationRuleMinLength     ValidationRule = "minLength"
	ValidationRulePattern       ValidationRule = "pattern"
	ValidationRuleFormat        ValidationRule = "format"
	ValidationRuleMaxItems      ValidationRule = "maxItems"
	ValidationRuleMinItems      ValidationRule = "minItems"
	ValidationRuleUniqueItems   ValidationRule = "uniqueItems"
	ValidationRuleMaxProperties ValidationRule = "maxProperties"
	ValidationRuleMinProperties ValidationRule = "minProperties"
	ValidationRuleRequired      ValidationRule = "required"
	ValidationRuleDeprecated    ValidationRule = "deprecated"
	ValidationRuleEnum          ValidationRule = "enum"
	ValidationRuleNullable      ValidationRule = "nullable"
	ValidationRuleOneOf         ValidationRule = "oneOf"
	ValidationRuleAllOf         ValidationRule = "allOf"
	ValidationRuleAnyOf         ValidationRule = "anyOf"
	ValidationRuleNot           ValidationRule = "not"
	ValidationRuleCustom        ValidationRule = "custom"
)

// Creates a new validator for the given provider and scope.
func NewValidator(provider ValidationProvider, scope *deps.Scope) *Validator {
	validations := make([]Validation, 0)
	return &Validator{
		Path:        make([]string, 0),
		Validations: &validations,
		Provider:    provider,
		Scope:       scope,
	}
}

// Validates a value against a schema.
func Validate(givenSchema *api.Schema, rawValue any, v *Validator) {
	val := concrete(rawValue)
	typ := concreteType(rawValue)

	if full, ok := rawValue.(CanValidateFull); ok {
		full.FullValidate(v)
		return
	}

	schema := givenSchema.ResolveReference()
	if schema == nil {
		return
	}

	options := v.Provider.ValidationOptions(typ)
	if options.Skip {
		return
	}

	if isNull(val) {
		if !schema.Nullable && schema.Type != api.DataTypeNull {
			v.Add(Validation{
				Rule:   ValidationRuleNullable,
				Schema: schema.GetName(),
			})
		}
		return
	}

	if schema.Deprecated && !isZero(val) {
		v.Add(Validation{
			Rule:   ValidationRuleDeprecated,
			Schema: schema.GetName(),
		})
	}

	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		validateNumber(schema, float64(val.Int()), rawValue, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		validateNumber(schema, float64(val.Uint()), rawValue, v)
	case reflect.Float32, reflect.Float64:
		validateNumber(schema, val.Float(), rawValue, v)
	case reflect.String:
		len := val.Len()

		if schema.MinLength != nil && len < *schema.MinLength {
			v.Add(Validation{
				Rule:    ValidationRuleMinLength,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the minimum length of %d", len, *schema.MinLength),
			})
		}
		if schema.MaxLength != 0 && len > schema.MaxLength {
			v.Add(Validation{
				Rule:    ValidationRuleMaxLength,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the maximum items of %d", len, schema.MaxLength),
			})
		}

	case reflect.Array, reflect.Slice:
		len := val.Len()

		if schema.MinItems != nil && len < *schema.MinItems {
			v.Add(Validation{
				Rule:    ValidationRuleMinItems,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the minimum items of %d", len, *schema.MinItems),
			})
		}
		if schema.MaxItems != 0 && len > schema.MaxItems {
			v.Add(Validation{
				Rule:    ValidationRuleMaxItems,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the maximum items of %d", len, schema.MaxItems),
			})
		}
		if schema.Items != nil {
			for i := 0; i < len; i++ {
				item := val.Index(i).Interface()
				itemValidator := v.Next(strconv.Itoa(i))
				Validate(schema.Items, item, itemValidator)
			}
		}
		if schema.UniqueItems {
			found := make(map[string]struct{})
			for i := 0; i < len; i++ {
				item := val.Index(i).Interface()
				asString := toString(item)
				if _, exists := found[asString]; exists {
					v.Add(Validation{
						Rule:    ValidationRuleUniqueItems,
						Schema:  schema.GetName(),
						Message: fmt.Sprintf("%v is not a unique item", item),
					})
					break
				} else {
					found[asString] = struct{}{}
				}
			}
		}
	case reflect.Map:
		len := val.Len()

		if schema.MinProperties != nil && len < *schema.MinProperties {
			v.Add(Validation{
				Rule:    ValidationRuleMinProperties,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the minimum properties of %d", len, *schema.MinProperties),
			})
		}
		if schema.MaxProperties != 0 && len > schema.MaxProperties {
			v.Add(Validation{
				Rule:    ValidationRuleMaxProperties,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%d does not meet the maximum properties of %d", len, schema.MaxProperties),
			})
		}
		if schema.AdditionalProperties != nil && schema.AdditionalProperties.Schema != nil {
			mapValueSchema := schema.AdditionalProperties.Schema
			iter := val.MapRange()
			for iter.Next() {
				mapKey := iter.Key().Interface()
				mapValue := iter.Value().Interface()
				valueValidator := v.Next(toString(mapKey))
				Validate(mapValueSchema, mapValue, valueValidator)
			}
		}

	case reflect.Struct:
		if schema.Properties == nil {
			break
		}

		for i := 0; i < typ.NumField(); i++ {
			fieldType := typ.Field(i)
			field := val.Field(i)

			propertyName, _, skip := api.GetJSONOptions(fieldType)
			if skip {
				continue
			}

			required := false
			if len(schema.Required) > 0 {
				for _, p := range schema.Required {
					if strings.EqualFold(propertyName, p) {
						required = true
					}
				}
			}

			if isNull(field) {
				if required {
					v.Add(Validation{
						Rule:    ValidationRuleRequired,
						Schema:  schema.GetName(),
						Message: fmt.Sprintf("%s is a required field", propertyName),
					})
				}
				continue
			}

			propertyValidator := v.Next(propertyName)

			if fieldType.Anonymous {
				Validate(schema, field.Interface(), propertyValidator)
			} else {
				if propertySchema, ok := schema.Properties[propertyName]; ok {
					Validate(&propertySchema, field.Interface(), propertyValidator)
				}
			}
		}
	}

	if schema.Pattern != "" {
		r := schema.GetPattern()
		if r != nil {
			asString := toString(rawValue)
			if !r.MatchString(asString) {
				v.Add(Validation{
					Rule:    ValidationRulePattern,
					Schema:  schema.GetName(),
					Message: fmt.Sprintf("%s does not match the pattern %s", asString, schema.Pattern),
				})
			}
		}
	}
	if schema.Format != "" && options.EnforceFormat {
		if r, exists := api.FormatRegex[schema.Format]; exists {
			asString := toString(rawValue)
			if !r.MatchString(asString) {
				v.Add(Validation{
					Rule:    ValidationRuleFormat,
					Schema:  schema.GetName(),
					Message: fmt.Sprintf("%s does not match the format %s", asString, schema.Format),
				})
			}
		}
	}
	if len(schema.Enum) > 0 {
		invalid := true
		for _, enumValue := range schema.Enum {
			if isTextuallyEqual(enumValue, rawValue) {
				invalid = false
				break
			}
		}
		if invalid {
			v.Add(Validation{
				Rule:    ValidationRuleEnum,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%v does not match one of the enum values %+v", rawValue, schema.Enum),
			})
		}
	}
	if len(schema.OneOf) > 0 {
		matches := 0
		for _, oneOf := range schema.OneOf {
			detached := v.Detach()
			Validate(&oneOf, rawValue, &detached)
			if len(*detached.Validations) == 0 {
				matches++
				if matches > 1 {
					break
				}
			}
		}
		if matches != 0 {
			v.Add(Validation{
				Rule:    ValidationRuleOneOf,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%v does not match one of the possible schemas", rawValue),
			})
		}
	}
	if len(schema.AllOf) > 0 {
		for _, allOf := range schema.AllOf {
			detached := v.Detach()
			Validate(&allOf, rawValue, &detached)
			if len(*detached.Validations) != 0 {
				v.Add(Validation{
					Rule:    ValidationRuleAllOf,
					Schema:  schema.GetName(),
					Message: fmt.Sprintf("%v does not match all of the possible schemas", rawValue),
				})
				break
			}
		}
	}
	if len(schema.AnyOf) > 0 {
		valid := false
		for _, anyOf := range schema.AnyOf {
			detached := v.Detach()
			Validate(&anyOf, rawValue, &detached)
			if len(*detached.Validations) == 0 {
				valid = true
				break
			}
		}
		if !valid {
			v.Add(Validation{
				Rule:    ValidationRuleAnyOf,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%v does not match any of the possible schemas", rawValue),
			})
		}
	}
	if schema.Not != nil {
		detached := v.Detach()
		Validate(schema.Not, rawValue, &detached)
		if len(*detached.Validations) == 0 {
			v.Add(Validation{
				Rule:    ValidationRuleNot,
				Schema:  schema.GetName(),
				Message: fmt.Sprintf("%v matches the not schema", rawValue),
			})
		}
	}

	if post, ok := rawValue.(CanValidatePost); ok {
		post.PostValidate(v)
	}
}

func validateNumber(s *api.Schema, value float64, rawValue any, v *Validator) {
	if s.Maximum != nil {
		invalid := (s.ExclusiveMaximum && value >= float64(*s.Maximum)) || (!s.ExclusiveMaximum && value > float64(*s.Maximum))
		if invalid {
			v.Add(Validation{
				Rule:    ValidationRuleMaximum,
				Schema:  s.GetName(),
				Message: fmt.Sprintf("%v exceeds the maximum of %d", rawValue, *s.Maximum),
			})
		}
	}
	if s.Minimum != nil {
		invalid := (s.ExclusiveMinimum && value <= float64(*s.Minimum)) || (!s.ExclusiveMinimum && value < float64(*s.Minimum))
		if invalid {
			v.Add(Validation{
				Rule:    ValidationRuleMinimum,
				Schema:  s.GetName(),
				Message: fmt.Sprintf("%v is below the minimum of %d", rawValue, *s.Minimum),
			})
		}
	}
	if s.MultipleOf != 0 {
		invalid := int(value)%s.MultipleOf != 0
		if invalid {
			v.Add(Validation{
				Rule:    ValidationRuleMultipleOf,
				Schema:  s.GetName(),
				Message: fmt.Sprintf("%v is not a multiple of %d", rawValue, s.MultipleOf),
			})
		}
	}
}

func concrete(val any) reflect.Value {
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv
}

func concreteType(val any) reflect.Type {
	typ := reflect.TypeOf(val)
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return typ
}

func isNull(val reflect.Value) bool {
	if !val.IsValid() {
		return true
	}
	switch val.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return val.IsNil()
	}
	return false
}

func isZero(val reflect.Value) bool {
	zero := reflect.New(val.Type()).Elem().Interface()
	return isTextuallyEqual(val.Interface(), zero)
}

func isTextuallyEqual(x any, y any) bool {
	return toString(x) == toString(y)
}

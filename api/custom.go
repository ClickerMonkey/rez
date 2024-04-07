package api

import (
	"reflect"
	"regexp"
	"strings"
)

// A type which has a specific name. Any type that implements this
// will be promoted to a named schema.
type HasName interface {
	APIName() string
}

// A type which has another type it actually marshals to and from.
// You can override the schema type by specifying a full or base schema
// but if the type is fundamentally different then the wrong validation and wrong
// examples can be generated.
type HasSchemaType interface {
	APISchemaType() any
}

// A type which has a specific description.
type HasDescription interface {
	APIDescription() string
}

// A type which can optionally provide schema defaults which can
// be built on with the default schema building behavior.
type HasBaseSchema interface {
	APIBaseSchema() *Schema
}

// A type which has its schema fully defined on it. This schema
// will be promoted to a named schema with exactly what is returned.
type HasFullSchema interface {
	APIFullSchema() *Schema
}

// A type which only accepts specific values. This values will be used
// in documentation and validation.
type HasEnum interface {
	APIEnum() []any
}

// A type which has examples defined on it that will be used in the request, response,
// parameter, or header. If there are no examples for the requested contentType, return nil.
type HasExamples interface {
	APIExamples(contentType ContentType) Examples
}

// A type which has an example defined on it that will be used in the type's schema.
type HasExample interface {
	APIExample() *any
}

// A function type can have an operation fully defined on it. The normal
// operation building logic will be skipped and this definition will be used.
// For operation augmentation see HasAPIOperationUpdate.
type HasOperation interface {
	APIOperation() Operation
}

// A function type can alter the detected operation after its been built based
// on the functions argument types and return types.
type HasOperationUpdate interface {
	APIOperationUpdate(op *Operation)
}

// Gets reflect.Type of the given type or value.
// This accounts for if the type/value implements the HasSchemaType.
func GetSchemaType(valueOrType any) reflect.Type {
	customType := getValueOrType(valueOrType, func(value HasSchemaType) any {
		return value.APISchemaType()
	})
	if customType != nil {
		valueOrType = customType
	}
	return GetType(valueOrType)
}

// Gets the Name of the given type. If it implements HasName then that name is returned.
func GetName(valueOrType any) string {
	name := getValueOrType(valueOrType, func(value HasName) string {
		return value.APIName()
	})
	if name == "" {
		name = GetType(valueOrType).Name()
	}
	return fixName(name)
}

var hasTypeNameType = reflect.TypeOf((*HasName)(nil)).Elem()

// Returns whether the given type implements HasName.
func IsNamedType(valueOrType any) bool {
	return GetType(valueOrType).Implements(hasTypeNameType)
}

// Gets the name of a given type, including the package path.
func GetNameQualified(valueOrType any) string {
	name := GetName(valueOrType)
	pkg := GetType(valueOrType).PkgPath()
	return fixName(pkg) + name
}

var nameSplitter = regexp.MustCompile(`[^-\w]+`)

func fixName(name string) string {
	parts := nameSplitter.Split(name, -1)
	fixedParts := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		fixedParts += strings.ToUpper(part[0:1]) + part[1:]
	}
	return fixedParts
}

// Gets the defined schema for the given type and if its the full schema, or
// returns nil if the type does not implement HasFullSchema or HasBaseSchema.
func GetSchema(valueOrType any) (*Schema, bool) {
	full := getValueOrType(valueOrType, func(value HasFullSchema) *Schema {
		return value.APIFullSchema()
	})

	if full != nil {
		return full, true
	}

	base := getValueOrType(valueOrType, func(value HasBaseSchema) *Schema {
		return value.APIBaseSchema()
	})

	if base != nil {
		return base, false
	}

	return nil, false
}

// Gets the description defined on the given type, or "" if the type does not
// implement HasDescription.
func GetDescription(valueOrType any) string {
	return getValueOrType(valueOrType, func(value HasDescription) string {
		return value.APIDescription()
	})
}

// Gets the enum values defined on the given type, or nil if the type does not
// implement HasEnum.
func GetEnums(valueOrType any) []any {
	return getValueOrType(valueOrType, func(value HasEnum) []any {
		return value.APIEnum()
	})
}

// Gets the examples defined on the given type with the given contentType or returns
// nil if the type does not implement HasExamples.
func GetExamples(valueOrType any, contentType ContentType) Examples {
	return getValueOrType(valueOrType, func(value HasExamples) Examples {
		return value.APIExamples(contentType)
	})
}

// Gets the JSON example defined on the given value or type or returns nil if
// the type does not implement HasExample.
func GetExample(valueOrType any) *any {
	return getValueOrType(valueOrType, func(value HasExample) *any {
		return value.APIExample()
	})
}

// Gets the operation on the function instance or type or returns nil if the
// type does not implement HasAPIOperation.
func GetOperation(fnOrType any) *Operation {
	return getValueOrType(fnOrType, func(value HasOperation) *Operation {
		op := value.APIOperation()
		return &op
	})
}

// Gets the operation on the function instance or type or returns nil if the
// type does not implement HasAPIOperation.
func GetOperationUpdate(fnOrType any, op *Operation) bool {
	return getValueOrType(fnOrType, func(value HasOperationUpdate) bool {
		value.APIOperationUpdate(op)
		return true
	})
}

// A value or type can be given, and if they implement V then it passes it to
// get and returns the result of that function. Otherwise the zero value of R is returned.
func getValueOrType[V any, R any](valueOrType any, get func(value V) R) R {
	if value, ok := valueOrType.(V); ok {
		return get(value)
	} else if typ, ok := valueOrType.(reflect.Type); ok {
		val := reflect.New(typ).Elem().Interface()
		if value, ok := val.(V); ok {
			return get(value)
		}
	}
	var invalid R
	return invalid
}

// Given a value or type, return the type.
func GetType(valueOrType any) reflect.Type {
	if typ, ok := valueOrType.(reflect.Type); ok {
		return typ
	}
	return reflect.TypeOf(valueOrType)
}

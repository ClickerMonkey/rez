package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Sets the full schema on the given builder with the defined generic type.
// A schema can be defined in its entirety and schema building won't try to detect it.
// The full schema should be defined before any building is done.
func SetFullSchema[V any](build *Builder, s *Schema) {
	build.FullSchema[typeOf[V]()] = s
}

// Sets the base schema on the given builder with the defined generic type.
// A schema can be defined with starting values, when it's first built it will
// pull in the base schema. The base schema should be defined before any building is done.
func SetBaseSchema[V any](build *Builder, s *Schema) {
	build.BaseSchema[typeOf[V]()] = s
}

// Builds an OpenAPI document, creating schemas from types, and deals with
// named schema collisions.
type Builder struct {
	// The base document that will be used when Build() is called.
	Document Document
	// If a field is nullable (pointer) should it be marked as optional? (not required)
	NullableIsOptional bool
	// If a field is optional (not required) should we accept null
	OptionalIsNullable bool
	// All collisions that occurred on the last Build().
	Collisions map[reflect.Type]*Schema
	// A schema can be defined with starting values, when it's first built it will
	// pull in the base schema. The base schema should be defined before any building is done.
	BaseSchema map[reflect.Type]*Schema
	// A schema can be defined in its entirety and schema building won't try to detect it.
	// The full schema should be defined before any building is done.
	FullSchema map[reflect.Type]*Schema

	schemas         map[reflect.Type]*Schema
	paths           map[string]*Path
	responses       map[string]*Response
	parameters      map[string]*Parameter
	examples        map[string]*Example
	requestBodies   map[string]*RequestBody
	headers         map[string]*Header
	securitySchemes map[string]*Security
	links           map[string]*Link
}

// Creates a new empty builder.
func NewBuilder() *Builder {
	return &Builder{
		BaseSchema: make(map[reflect.Type]*Schema),
		FullSchema: make(map[reflect.Type]*Schema),
		Collisions: make(map[reflect.Type]*Schema),

		schemas:         make(map[reflect.Type]*Schema),
		paths:           make(map[string]*Path),
		responses:       make(map[string]*Response),
		parameters:      make(map[string]*Parameter),
		examples:        make(map[string]*Example),
		requestBodies:   make(map[string]*RequestBody),
		headers:         make(map[string]*Header),
		securitySchemes: make(map[string]*Security),
		links:           make(map[string]*Link),
	}
}

// Builds the document and returns it as JSON
func (build *Builder) BuildJSON() []byte {
	json, _ := json.Marshal(build.Build())
	return json
}

// Builds the current document.
func (build *Builder) Build() Document {
	doc := build.Document
	if doc.Components == nil {
		doc.Components = &Component{}
	}
	if doc.Paths == nil {
		doc.Paths = make(map[string]Path)
	}

	build.Collisions = make(map[reflect.Type]*Schema)
	if len(build.schemas) > 0 {
		if doc.Components.Schemas == nil {
			doc.Components.Schemas = make(map[string]Schema)
		}
		for typ, schema := range build.schemas {
			build.setDocumentSchema(&doc, schema, typ)
		}
	}

	if len(build.responses) > 0 {
		if doc.Components.Responses == nil {
			doc.Components.Responses = Responses{}
		}
		for name, response := range build.responses {
			doc.Components.Responses[name] = response
		}
	}

	if len(build.parameters) > 0 {
		if doc.Components.Parameters == nil {
			doc.Components.Parameters = make(map[string]Parameter)
		}
		for name, param := range build.parameters {
			doc.Components.Parameters[name] = *param
		}
	}

	if len(build.examples) > 0 {
		if doc.Components.Examples == nil {
			doc.Components.Examples = Examples{}
		}
		for name, ex := range build.examples {
			doc.Components.Examples[name] = *ex
		}
	}

	if len(build.requestBodies) > 0 {
		if doc.Components.RequestBodies == nil {
			doc.Components.RequestBodies = make(map[string]RequestBody)
		}
		for name, body := range build.requestBodies {
			doc.Components.RequestBodies[name] = *body
		}
	}

	if len(build.headers) > 0 {
		if doc.Components.Headers == nil {
			doc.Components.Headers = Headers{}
		}
		for name, header := range build.headers {
			doc.Components.Headers[name] = header
		}
	}

	if len(build.securitySchemes) > 0 {
		if doc.Components.SecuritySchemes == nil {
			doc.Components.SecuritySchemes = make(map[string]Security)
		}
		for name, scheme := range build.securitySchemes {
			doc.Components.SecuritySchemes[name] = *scheme
		}
	}

	if len(build.links) > 0 {
		if doc.Components.Links == nil {
			doc.Components.Links = Links{}
		}
		for name, link := range build.links {
			doc.Components.Links[name] = *link
		}
	}

	for url, path := range build.paths {
		doc.Paths[url] = *path
	}

	return doc
}

// Adds the schema to the given document, and if a name collision occurs it calculates a
// fully qualified name including the package path and uses that. If a collision still exists
// with the full name its recorded in Builder.Collisions and the schema is not added to the document.
func (build *Builder) setDocumentSchema(doc *Document, s *Schema, typ reflect.Type) {
	name := GetName(typ)
	if _, exists := doc.Components.Schemas[name]; exists {
		name = GetNameQualified(typ)
	}

	s.named.Ref = RefTo(s, name).Ref

	if _, exists := doc.Components.Schemas[name]; exists {
		build.Collisions[typ] = s
	} else {
		doc.Components.Schemas[name] = *s
	}
}

// Adds a tag to the builder.
func (build *Builder) AddTag(tag Tag) {
	build.Document.Tags = append(build.Document.Tags, tag)
}

// Adds a path to the builder. If it already exists it may break references that were created
// with Builder.RefPath()
func (build *Builder) AddPath(url string, path *Path) {
	path.named = RefTo(path, url)
	build.paths[url] = path
}

// Gets the path at the given URL.
func (build *Builder) GetPath(url string) *Path {
	return build.paths[url]
}

// Gets a reference to the path at the given URL, or returns nil if none exists.
func (build *Builder) RefPath(url string) *Path {
	path := build.paths[url]
	if path != nil {
		path = &Path{Reference: path.named}
	}
	return path
}

// Adds a named response to the builder.
func (build *Builder) AddResponse(name string, response *Response) {
	response.named = RefTo(response, name)
	build.responses[name] = response
}

// Gets a named response or returns nil if it doesn't exist.
func (build *Builder) GetResponse(name string) *Response {
	return build.responses[name]
}

// Gets a reference to the response with the given name, or returns nil if none exists.
func (build *Builder) RefResponse(name string) *Response {
	response := build.responses[name]
	if response != nil {
		response = &Response{Reference: response.named}
	}
	return response
}

// Adds a named parameter to the builder.
func (build *Builder) AddParameter(name string, param *Parameter) {
	param.named = RefTo(param, name)
	build.parameters[name] = param
}

// Gets a named parameter or returns nil if it doesn't exist.
func (build *Builder) GetParameter(name string) *Parameter {
	return build.parameters[name]
}

// Gets a reference to the parameter with the given name, or returns nil if none exists.
func (build *Builder) RefParameter(name string) *Parameter {
	param := build.parameters[name]
	if param != nil {
		param = &Parameter{Reference: param.named}
	}
	return param
}

// Adds a named example to the builder.
func (build *Builder) AddExample(name string, ex *Example) {
	ex.named = RefTo(ex, name)
	build.examples[name] = ex
}

// Gets a named example or returns nil if it doesn't exist.
func (build *Builder) GetExample(name string) *Example {
	return build.examples[name]
}

// Gets a reference to the example with the given name, or returns nil if none exists.
func (build *Builder) RefExample(name string) *Example {
	ex := build.examples[name]
	if ex != nil {
		ex = &Example{Reference: ex.named}
	}
	return ex
}

// Adds a named request body to the builder.
func (build *Builder) AddRequestBody(name string, body *RequestBody) {
	body.named = RefTo(body, name)
	build.requestBodies[name] = body
}

// Gets a named request body or returns nil if it doesn't exist.
func (build *Builder) GetRequestBody(name string) *RequestBody {
	return build.requestBodies[name]
}

// Gets a reference to the request body with the given name, or returns nil if none exists.
func (build *Builder) RefRequestBody(name string) *RequestBody {
	body := build.requestBodies[name]
	if body != nil {
		body = &RequestBody{Reference: body.named}
	}
	return body
}

// Adds a named header to the builder.
func (build *Builder) AddHeader(name string, header *Header) {
	header.named = RefTo(header, name)
	build.headers[name] = header
}

// Gets a named header or returns nil if it doesn't exist.
func (build *Builder) GetHeader(name string) *Header {
	return build.headers[name]
}

// Gets a reference to the header with the given name, or returns nil if none exists.
func (build *Builder) RefHeader(name string) *Header {
	header := build.headers[name]
	if header != nil {
		header = &Header{Reference: header.named}
	}
	return header
}

// Adds a named security to the builder.
func (build *Builder) AddSecurity(name string, security *Security) {
	security.named = RefTo(security, name)
	build.securitySchemes[name] = security
}

// Gets a named security or returns nil if it doesn't exist.
func (build *Builder) GetSecurity(name string) *Security {
	return build.securitySchemes[name]
}

// Gets a reference to the security with the given name, or returns nil if none exists.
func (build *Builder) RefSecurity(name string) *Security {
	security := build.securitySchemes[name]
	if security != nil {
		security = &Security{Reference: security.named}
	}
	return security
}

// Adds a named link to the builder.
func (build *Builder) AddLink(name string, link *Link) {
	link.named = RefTo(link, name)
	build.links[name] = link
}

// Gets a named link or returns nil if it doesn't exist.
func (build *Builder) GetLink(name string) *Link {
	return build.links[name]
}

// Gets a reference to the link with the given name, or returns nil if none exists.
func (build *Builder) RefLink(name string) *Link {
	link := build.links[name]
	if link != nil {
		link = &Link{Reference: link.named}
	}
	return link
}

// Adds a schema to the builder for the given type. This is equivalent to calling
// builder.GetSchema() without getting the schema.
func (build *Builder) AddSchema(typ reflect.Type) {
	build.GetSchema(typ)
}

// Gets or builds the schema for the given type. If a schema could not be determined,
// nil is returned.
func (build *Builder) GetSchema(typ reflect.Type) *Schema {
	if schema, ok := build.schemas[typ]; ok {
		return schema
	}

	return build.BuildSchema(typ, true)
}

// Saves the schema for the given type. This adds the named field on the schema which signals
// that a schema can only be referenced and not directly modified.
func (build *Builder) setSchema(typ reflect.Type, s *Schema) {
	s.named = &Reference{} // Ref is built later to avoid collisions
	s.typ = typ
	build.schemas[typ] = s
}

// Builds a schema for the given type, and potentially adds it to the builder. This should
// not be called if the type has a named schema defined for it and addToDocument = true.
func (build *Builder) BuildSchema(typ reflect.Type, addToDocument bool) *Schema {
	s := &Schema{}

	// If its a pointer type, return the nullable type.
	if concrete := getConcrete(typ); concrete != typ {
		concreteSchema := build.GetSchema(concrete)
		if concreteSchema == nil {
			return nil
		}

		return build.makeNullable(concrete, concreteSchema)
	}

	// Can/should this schema be promoted to the top?
	isDefined := typ.Kind() == reflect.Struct || IsNamedType(typ)
	continueDefining := true

	// Types can have custom schemas defined by the user
	if fullSchema := build.FullSchema[typ]; fullSchema != nil {
		*s = *fullSchema
		continueDefining = false
	} else if override := build.BaseSchema[typ]; override != nil {
		*s = *override
	} else if custom, full := GetSchema(typ); custom != nil {
		*s = *custom
		isDefined = true
		continueDefining = !full
	}

	// Promote before we build any more schemas (to account circular references)
	if isDefined && addToDocument {
		build.setSchema(typ, s)
	}

	// We have a full schema, move on
	if !continueDefining {
		return s
	}

	// If the user did not supply an enum via APISchema, try APIEnums
	if s.Enum == nil {
		s.Enum = GetEnums(typ)
	}

	// If the user did not supply an example, try APIExample
	if s.Example == nil {
		s.Example = GetExample(typ)
	}

	// If the user did not supply a description, try APIDescription
	if s.Description == "" {
		s.Description = GetDescription(typ)
	}

	// Coalesce ensures we don't override non-zero values returned by APISchema

	switch typ.Kind() {
	// Unsupported types
	case reflect.Func, reflect.Chan:
		return nil
	case reflect.Interface:
		// No type
	case reflect.Array:
		len := typ.Len()
		s.Type = MergeValue(s.Type, DataTypeArray)
		s.MinItems = MergeValue(s.MinItems, &len)
		s.MaxItems = MergeValue(s.MaxItems, len)
		s.Items = MergeValue(s.Items, build.GetSchema(typ.Elem()))
	case reflect.Slice:
		s.Type = MergeValue(s.Type, DataTypeArray)
		s.Items = MergeValue(s.Items, build.GetSchema(typ.Elem()))
	case reflect.Map:
		s.Type = MergeValue(s.Type, DataTypeObject)
		s.AdditionalProperties = MergeValue(s.AdditionalProperties, SchemaForSchema(build.GetSchema(typ.Elem())))
	case reflect.String:
		s.Type = MergeValue(s.Type, DataTypeString)
	case reflect.Bool:
		s.Type = MergeValue(s.Type, DataTypeBoolean)
	case reflect.Complex128, reflect.Complex64, reflect.Float32, reflect.Float64:
		s.Type = MergeValue(s.Type, DataTypeNumber)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		s.Type = MergeValue(s.Type, DataTypeInteger)
	case reflect.Struct:
		s.Type = MergeValue(s.Type, DataTypeObject)
		// If required wasn't explicitly required via APISchema
		if s.Required == nil {
			s.Required = make([]string, 0)
		}
		// If properties was explicitly given via APISchema, we won't build them
		if s.Properties != nil {
			break
		}

		// Determine the properties via reflection
		s.Properties = make(map[string]Schema)
		s.AdditionalProperties = SchemaForBool(false)

		// Recursive function to populate properties
		build.addProperties(s, false, typ)
	}

	return s
}

// Makes the given schema nullable. If the given schema is named, it creates a new
// schema referring to the named one.
func (build *Builder) makeNullable(typ reflect.Type, s *Schema) *Schema {
	if s.named != nil {
		s = &Schema{
			OneOf: []Schema{
				*s.AsReference(),
				{Type: DataTypeNull},
			},
		}
	} else if len(s.AllOf) == 1 {
		s.OneOf = []Schema{
			s.AllOf[0],
			{Type: DataTypeNull},
		}
		s.AllOf = nil
	} else {
		s.Nullable = true
	}
	return s
}

// Returns whether the given schema points to a nullable value.
func (build *Builder) IsNullable(s Schema) bool {
	return s.Nullable || (len(s.OneOf) == 2 && s.OneOf[1].Type == DataTypeNull)
}

// Gets the inner schema of the given schema or returns nil if there is none.
// A schema has an inner schema if its a schema that refers to a named schema but that
// schema is nullable or has extra metadata added to the named schema.
func (build *Builder) GetInnerSchema(s Schema) *Schema {
	if len(s.OneOf) > 0 {
		return &s.OneOf[0]
	}
	if len(s.AllOf) == 1 {
		return &s.AllOf[0]
	}
	return nil
}

// Adds properties to objectSchema that exist on the struct type typ.
func (build *Builder) addProperties(objectSchema *Schema, parentOptional bool, typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// A field is required if its not nullable and there is no omitempty or required is specified in the api tag.

		property, optional, skip := GetJSONOptions(field)
		if skip {
			continue
		}
		optional = optional || parentOptional

		// Embedded struct?
		if field.Anonymous {
			build.addProperties(objectSchema, optional, field.Type)
		} else {
			propertySchema := build.GetSchema(field.Type)

			if propertySchema == nil {
				continue
			}

			nullable := propertySchema.Nullable
			if optional && build.OptionalIsNullable {
				nullable = true
			}
			if nullable && build.NullableIsOptional {
				optional = true
			}

			if nullable && !build.IsNullable(*propertySchema) {
				propertySchema = build.makeNullable(field.Type, propertySchema)
			}

			if api := field.Tag.Get("api"); api != "" {
				// If its a named schema, wrap it
				if propertySchema.named != nil {
					propertySchema = &Schema{
						AllOf: []Schema{*propertySchema.AsReference()},
					}
				}
				ApplyOptions(propertySchema, api)
			}

			propertySchemaFinal := propertySchema.AsReference()

			objectSchema.Properties[property] = *propertySchemaFinal

			if !optional {
				objectSchema.Required = append(objectSchema.Required, property)
			}
		}
	}
}

// Returns the type for the given generic parameter.
func typeOf[V any]() reflect.Type {
	return reflect.TypeOf((*V)(nil)).Elem()
}

// Gets the concrete (non pointer) type that typ potentially points to.
func getConcrete(typ reflect.Type) reflect.Type {
	concrete := typ
	for concrete.Kind() == reflect.Pointer {
		concrete = concrete.Elem()
	}
	return concrete
}

// Gets the json options from the given struct field.
func GetJSONOptions(field reflect.StructField) (property string, optional bool, skip bool) {
	property = field.Name
	optional = false
	skip = false

	options := strings.Split(field.Tag.Get("json"), ",")
	if len(options) > 0 {
		if len(options) > 1 && strings.EqualFold(options[1], "omitempty") {
			optional = true
		}
		if options[0] == "-" {
			skip = true
		} else if options[0] != "" {
			property = options[0]
		}
	}
	if !field.IsExported() {
		skip = true
	}

	return
}

// Escapes a value so it can be used in a reference path.
func EscapePathPart(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

// Splits s by delim, unless delim is preceded by escape.
func splitWithEscape(s string, delim string, escape string) []string {
	s = strings.ReplaceAll(s, escape+delim, "\x00")
	tokens := strings.Split(s, delim)
	for i, token := range tokens {
		tokens[i] = strings.ReplaceAll(token, "\x00", delim)
	}
	return tokens
}

// Applies the key=value & flag options found in tag to the given schema.
// Example format of tag: "title=Hi,required". Full list of examples:
//   - title=Example title which can have\, commas
//   - desc=A description of the schema.
//   - description=A description of the schema.
//   - format=date
//   - pattern=\d{1\,3}.\d{1\,3}.\d{1\,3}.\d{1\,3}
//   - deprecated
//   - required
//   - nullable
//   - null
//   - readonly
//   - writeonly
//   - enum=A|B|C|A\,B\,C
//   - minlength=1
//   - maxlength=1
//   - multipleof=2
//   - maximum=1
//   - max=1
//   - minimum=1
//   - min=1
//   - minitems=1
//   - maxitems=1
//   - exclusivemaximum=true
//   - exclusivemax=1
//   - exclusiveminimum=false
//   - exclusivemin=0
func ApplyOptions(s *Schema, tag string) {
	apiOptions := splitWithEscape(tag, ",", "\\")

	for _, optionRaw := range apiOptions {
		keyValue := strings.SplitN(optionRaw, "=", 2)
		key := strings.TrimSpace(keyValue[0])
		value := key
		if len(keyValue) > 1 {
			value = keyValue[1]
		}

		var parseInt *int
		var parseBool *bool

		defaultInt := 0

		switch strings.ToLower(key) {
		case "title":
			s.Title = value
		case "desc", "description":
			s.Description = value
		case "format":
			s.Format = value
		case "pattern":
			s.Pattern = value
		case "deprecated":
			s.Deprecated = true
		case "required":
			s.Nullable = false
		case "nullable", "null":
			s.Nullable = true
		case "readonly":
			s.ReadOnly = true
		case "writeonly":
			s.WriteOnly = true
		case "enum":
			s.Enum = make([]any, 0)
			enumConstants := splitWithEscape(value, "|", "\\")
			for _, enumConstant := range enumConstants {
				if enumConstant != "" {
					s.Enum = append(s.Enum, enumConstant)
				}
			}
		case "minlength":
			s.MinLength = &defaultInt
			parseInt = s.MinLength
		case "maxlength":
			parseInt = &s.MaxLength
		case "minitems":
			s.MinItems = &defaultInt
			parseInt = s.MinItems
		case "maxitems":
			parseInt = &s.MaxItems
		case "multipleof":
			parseInt = &s.MultipleOf
		case "maximum", "max":
			s.Maximum = &defaultInt
			parseInt = s.Maximum
		case "minimum", "min":
			s.Minimum = &defaultInt
			parseInt = s.Minimum
		case "exclusivemaximum", "exclusivemax":
			parseBool = &s.ExclusiveMaximum
		case "exclusiveminimum", "exclusivemin":
			parseBool = &s.ExclusiveMinimum
		}

		if parseInt != nil {
			parsed, err := strconv.Atoi(value)
			if err != nil {
				panic(fmt.Sprintf("Error parsing %s from tag: %s", key, value))
			}
			*parseInt = parsed
		}
		if parseBool != nil {
			if value == key {
				value = "true"
			}
			parsed, err := strconv.ParseBool(value)
			if err != nil {
				panic(fmt.Sprintf("Error parsing %s from tag: %s", key, value))
			}
			*parseBool = parsed
		}
	}
}

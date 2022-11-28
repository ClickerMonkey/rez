package api

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type HasTypeName interface {
	APITypeName() string
}

type HasBaseSchema interface {
	APIBaseSchema() *Schema
}

type HasFullSchema interface {
	APIFullSchema() *Schema
}

type HasEnum interface {
	APIEnum() []any
}

type Builder struct {
	Document           Document
	NullableIsOptional bool
	OptionalIsNullable bool
	Collisions         map[reflect.Type]*Schema
	BaseSchema         map[reflect.Type]*Schema
	FullSchema         map[reflect.Type]*Schema

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

func (build *Builder) BuildJSON() []byte {
	json, _ := json.Marshal(build.Build())
	return json
}

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

func (build *Builder) setDocumentSchema(doc *Document, s *Schema, typ reflect.Type) {
	name := GetTypeName(typ)
	if _, exists := doc.Components.Schemas[name]; exists {
		name = GetTypeNameQualified(typ)
	}

	s.named.Ref = RefTo(s, name).Ref

	if _, exists := doc.Components.Schemas[name]; exists {
		build.Collisions[typ] = s
	} else {
		doc.Components.Schemas[name] = *s
	}
}

func (build *Builder) AddTag(tag Tag) {
	build.Document.Tags = append(build.Document.Tags, tag)
}

func (build *Builder) AddPath(url string, path *Path) {
	path.named = RefTo(path, url)
	build.paths[url] = path
}

func (build *Builder) GetPath(url string) *Path {
	return build.paths[url]
}

func (build *Builder) RefPath(url string) *Path {
	path := build.paths[url]
	if path != nil {
		path = &Path{Reference: path.named}
	}
	return path
}

func (build *Builder) AddResponse(name string, response *Response) {
	response.named = RefTo(response, name)
	build.responses[name] = response
}

func (build *Builder) GetResponse(name string) *Response {
	return build.responses[name]
}

func (build *Builder) RefResponse(name string) *Response {
	response := build.responses[name]
	if response != nil {
		response = &Response{Reference: response.named}
	}
	return response
}

func (build *Builder) AddParameter(name string, param *Parameter) {
	param.named = RefTo(param, name)
	build.parameters[name] = param
}

func (build *Builder) GetParameter(name string) *Parameter {
	return build.parameters[name]
}

func (build *Builder) RefParameter(name string) *Parameter {
	param := build.parameters[name]
	if param != nil {
		param = &Parameter{Reference: param.named}
	}
	return param
}

func (build *Builder) AddExample(name string, ex *Example) {
	ex.named = RefTo(ex, name)
	build.examples[name] = ex
}

func (build *Builder) GetExample(name string) *Example {
	return build.examples[name]
}

func (build *Builder) RefExample(name string) *Example {
	ex := build.examples[name]
	if ex != nil {
		ex = &Example{Reference: ex.named}
	}
	return ex
}

func (build *Builder) AddRequestBody(name string, body *RequestBody) {
	body.named = RefTo(body, name)
	build.requestBodies[name] = body
}

func (build *Builder) GetRequestBody(name string) *RequestBody {
	return build.requestBodies[name]
}

func (build *Builder) RefRequestBody(name string) *RequestBody {
	body := build.requestBodies[name]
	if body != nil {
		body = &RequestBody{Reference: body.named}
	}
	return body
}

func (build *Builder) AddHeader(name string, header *Header) {
	header.named = RefTo(header, name)
	build.headers[name] = header
}

func (build *Builder) GetHeader(name string) *Header {
	return build.headers[name]
}

func (build *Builder) RefHeader(name string) *Header {
	header := build.headers[name]
	if header != nil {
		header = &Header{Reference: header.named}
	}
	return header
}

func (build *Builder) AddSecurity(name string, security *Security) {
	security.named = RefTo(security, name)
	build.securitySchemes[name] = security
}

func (build *Builder) GetSecurity(name string) *Security {
	return build.securitySchemes[name]
}

func (build *Builder) RefSecurity(name string) *Security {
	security := build.securitySchemes[name]
	if security != nil {
		security = &Security{Reference: security.named}
	}
	return security
}

func (build *Builder) AddLink(name string, link *Link) {
	link.named = RefTo(link, name)
	build.links[name] = link
}

func (build *Builder) GetLink(name string) *Link {
	return build.links[name]
}

func (build *Builder) RefLink(name string) *Link {
	link := build.links[name]
	if link != nil {
		link = &Link{Reference: link.named}
	}
	return link
}

func (build *Builder) AddSchema(typ reflect.Type) {
	build.GetSchema(typ)
}

func (build *Builder) GetSchema(typ reflect.Type) *Schema {
	if schema, ok := build.schemas[typ]; ok {
		return schema
	}

	return build.BuildSchema(typ, true)
}

func (build *Builder) setSchema(typ reflect.Type, s *Schema) {
	s.named = &Reference{} // Ref is built later to avoid collisions
	build.schemas[typ] = s
}

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
	} else if custom, full := GetTypeSchema(typ); custom != nil {
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
		s.Enum = GetTypeEnums(typ)
	}

	// Coalesce ensures we don't override non-zero values returned by APISchema

	switch typ.Kind() {
	// Unsupported types
	case reflect.Func, reflect.Chan:
		return nil
	case reflect.Interface:
		// No type
	case reflect.Array:
		s.Type = coalesce(s.Type, DataTypeArray)
		s.MinItems = coalesce(s.MinItems, typ.Len())
		s.MaxItems = coalesce(s.MaxItems, typ.Len())
		s.Items = coalesce(s.Items, build.GetSchema(typ.Elem()))
	case reflect.Slice:
		s.Type = coalesce(s.Type, DataTypeArray)
		s.Items = coalesce(s.Items, build.GetSchema(typ.Elem()))
	case reflect.Map:
		s.Type = coalesce(s.Type, DataTypeObject)
		s.AdditionalProperties = coalesce(s.AdditionalProperties, SchemaForSchema(build.GetSchema(typ.Elem())))
	case reflect.String:
		s.Type = coalesce(s.Type, DataTypeString)
	case reflect.Bool:
		s.Type = coalesce(s.Type, DataTypeBoolean)
	case reflect.Complex128, reflect.Complex64, reflect.Float32, reflect.Float64:
		s.Type = coalesce(s.Type, DataTypeNumber)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8, reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		s.Type = coalesce(s.Type, DataTypeInteger)
	case reflect.Struct:
		s.Type = coalesce(s.Type, DataTypeObject)
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

func (build *Builder) makeNullable(typ reflect.Type, s *Schema) *Schema {
	if s.named != nil {
		s = &Schema{
			OneOf: []Schema{
				{Reference: s.named},
				{Type: DataTypeNull},
			},
		}
	} else {
		s.Nullable = true
	}
	return s
}

func (build *Builder) isNullable(s Schema) bool {
	return s.Nullable || (len(s.OneOf) == 2 && s.OneOf[1].Nullable)
}

func (build *Builder) getInnerSchema(s Schema) *Schema {
	if len(s.OneOf) > 0 {
		return &s.OneOf[0]
	}
	if len(s.AllOf) == 1 {
		return &s.AllOf[0]
	}
	return nil
}

func (build *Builder) addProperties(objectSchema *Schema, parentOptional bool, typ reflect.Type) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// A field is required if its not nullable and there is no omitempty or required is specified in the api tag.

		property, optional, skip := getJSONOptions(field)
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

			if nullable && !build.isNullable(*propertySchema) {
				propertySchema = build.makeNullable(field.Type, propertySchema)
			}

			if api := field.Tag.Get("api"); api != "" {
				// If its a named schema, wrap it
				if propertySchema.named != nil {
					propertySchema = &Schema{
						AllOf: []Schema{{Reference: propertySchema.named}},
					}
				}
				ApplyOptions(propertySchema, api)
			}

			propertySchemaFinal := propertySchema
			if propertySchemaFinal.named != nil {
				propertySchemaFinal = &Schema{
					Reference: propertySchemaFinal.named,
				}
			}

			objectSchema.Properties[property] = *propertySchemaFinal

			if !optional {
				objectSchema.Required = append(objectSchema.Required, property)
			}
		}
	}
}

func typeOf[V any]() reflect.Type {
	return reflect.TypeOf((*V)(nil)).Elem()
}

func SetFullSchema[V any](build *Builder, s *Schema) {
	build.FullSchema[typeOf[V]()] = s
}

func SetBaseSchema[V any](build *Builder, s *Schema) {
	build.BaseSchema[typeOf[V]()] = s
}

func coalesce[V comparable](a V, b V) V {
	var empty V
	if a == empty {
		return b
	}
	return a
}

func getConcrete(typ reflect.Type) reflect.Type {
	concrete := typ
	for concrete.Kind() == reflect.Pointer {
		concrete = concrete.Elem()
	}
	return concrete
}

func getJSONOptions(field reflect.StructField) (property string, optional bool, skip bool) {
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

func GetTypeName(typ reflect.Type) string {
	val := reflect.New(typ).Elem().Interface()

	if hasTypeName, ok := val.(HasTypeName); ok {
		return hasTypeName.APITypeName()
	}

	return typ.Name()
}

var hasTypeNameType = reflect.TypeOf((*HasTypeName)(nil)).Elem()

func IsNamedType(typ reflect.Type) bool {
	return typ.Implements(hasTypeNameType)
}

var pkgReplace = regexp.MustCompile(`(^|\.|\/)[^\.\/]+`)

func GetTypeNameQualified(typ reflect.Type) string {
	name := GetTypeName(typ)
	pkg := typ.PkgPath()
	if pkg == "" {
		return name
	}

	pkgPrefix := pkgReplace.ReplaceAllStringFunc(pkg, func(s string) string {
		if s[0] == '.' || s[0] == '/' {
			return strings.ToUpper(s[1:2]) + s[2:]
		}
		return strings.ToUpper(s[0:1]) + s[1:]
	})

	return pkgPrefix + name
}

func GetTypeSchema(typ reflect.Type) (*Schema, bool) {
	val := reflect.New(typ).Elem().Interface()

	if hasSchema, ok := val.(HasFullSchema); ok {
		return hasSchema.APIFullSchema(), true
	}
	if hasSchema, ok := val.(HasBaseSchema); ok {
		return hasSchema.APIBaseSchema(), false
	}

	return nil, false
}

func GetTypeEnums(typ reflect.Type) []any {
	val := reflect.New(typ).Elem().Interface()

	if hasEnum, ok := val.(HasEnum); ok {
		return hasEnum.APIEnum()
	}

	return nil
}

func EscapePathPart(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	s = strings.ReplaceAll(s, "/", "~1")
	return s
}

func splitWithEscape(s string, delim string, escape string) []string {
	s = strings.ReplaceAll(s, escape+delim, "\x00")
	tokens := strings.Split(s, delim)
	for i, token := range tokens {
		tokens[i] = strings.ReplaceAll(token, "\x00", delim)
	}
	return tokens
}

func ApplyOptions(s *Schema, tag string) {
	apiOptions := splitWithEscape(tag, ",", "\\")

	for _, optionRaw := range apiOptions {
		keyValue := strings.SplitAfterN(optionRaw, "=", 2)
		key := strings.TrimSpace(keyValue[0])
		value := key
		if len(keyValue) > 1 {
			value = keyValue[1]
		}

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
			s.MinLength, _ = strconv.Atoi(value)
		case "maxlength":
			s.MaxLength, _ = strconv.Atoi(value)
		case "multipleof":
			s.MultipleOf, _ = strconv.Atoi(value)
		case "maximum", "max":
			s.Maximum, _ = strconv.Atoi(value)
		case "minimum", "min":
			s.Minimum, _ = strconv.Atoi(value)
		case "exclusivemaximum", "exclusivemax":
			s.ExclusiveMaximum, _ = strconv.ParseBool(value)
		case "exclusiveminimum", "exclusivemin":
			s.ExclusiveMinimum, _ = strconv.ParseBool(value)
		}
	}
}

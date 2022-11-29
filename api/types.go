package api

import (
	"encoding/json"
	"reflect"
)

// A type which can be merged with another type. Merging is about returning the defined values
// in the given value and the values in the current value. The given value takes priority.
// Slices and maps have different merging strategies based on the implementation and the
// interfaces the value types implement.
type CanMerge[T any] interface {
	// Merging incorporates value into this and returns a new value. This or value are not modified.
	// If value has non-default single values it will overwrite those values in this in the returned value.
	Merge(value T) T
}

// Merges zero or more mergeable values into a single value, where later arguments have higher priority.
func Merge[T CanMerge[T]](base T, next T, additional ...T) T {
	base = base.Merge(next)
	for _, n := range additional {
		base = base.Merge(n)
	}
	return base
}

// A mergeable type that can be uniquely identified
type CanReplace[T any] interface {
	IsUnique(value T) bool
}

// This is the root document object of the OpenAPI document.
type Document struct {
	// REQUIRED. This string MUST be the semantic version number of the OpenAPI Specification version that the OpenAPI document uses. The openapi field SHOULD be used by tooling specifications and clients to interpret the OpenAPI document. This is not related to the API info.version string.
	OpenAPI string `json:"openapi"`
	// REQUIRED. Provides metadata about the API. The metadata MAY be used by tooling as required.
	Info Info `json:"info"`
	// An array of Server Objects, which provide connectivity information to a target server. If the servers property is not provided, or is an empty array, the default value would be a Server Object with a url value of /.
	Servers []Server `json:"servers,omitempty"`
	// REQUIRED. The available paths and operations for the API.
	Paths map[string]Path `json:"paths"`
	// An element to hold various schemas for the specification.
	Components *Component `json:"components,omitempty"`
	// A declaration of which security mechanisms can be used across the API. The list of values includes alternative security requirement objects that can be used. Only one of the security requirement objects need to be satisfied to authorize a request. Individual operations can override this definition. To make security optional, an empty security requirement ({}) can be included in the array.
	Security []Security `json:"security,omitempty"`
	// A list of tags used by the specification with additional metadata. The order of the tags can be used to reflect on their order by the parsing tools. Not all tags that are used by the Operation Object must be declared. The tags that are not declared MAY be organized randomly or based on the tools' logic. Each tag name in the list MUST be unique.
	Tags []Tag `json:"tags,omitempty"`
	// Additional external documentation.
	ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
}

var _ CanMerge[Document] = &Document{}

// Merges documents
func (base Document) Merge(next Document) Document {
	return Document{
		OpenAPI:      MergeValue(base.OpenAPI, next.OpenAPI),
		Info:         *MergeCanMerge(&base.Info, &next.Info),
		Servers:      MergeSliceReplace(base.Servers, next.Servers),
		Paths:        MergeMap(base.Paths, next.Paths),
		Components:   MergeCanMerge(base.Components, next.Components),
		Security:     MergeSliceReplace(base.Security, next.Security),
		Tags:         MergeSliceReplace(base.Tags, next.Tags),
		ExternalDocs: MergeValue(base.ExternalDocs, next.ExternalDocs),
	}
}

// The object provides metadata about the API. The metadata MAY be used by the clients if needed, and MAY be presented in editing or documentation generation tools for convenience.
type Info struct {
	// REQUIRED. The title of the API.
	Title string `json:"title"`
	// A short description of the API. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A URL to the Terms of Service for the API. MUST be in the format of a URL.
	TermsOfService string `json:"termsOfService,omitempty"`
	// The contact information for the exposed API.
	Contact *Contact `json:"contact,omitempty"`
	// The license information for the exposed API.
	License *License `json:"license,omitempty"`
	// REQUIRED. The version of the OpenAPI document (which is distinct from the OpenAPI Specification version or the API implementation version).
	Version string `json:"version"`
}

var _ CanMerge[Info] = &Info{}

// Merges info
func (base Info) Merge(next Info) Info {
	return Info{
		Title:          MergeValue(base.Title, next.Title),
		Description:    MergeValue(base.Description, next.Description),
		TermsOfService: MergeValue(base.TermsOfService, next.TermsOfService),
		Contact:        MergeCanMerge(base.Contact, next.Contact),
		License:        MergeCanMerge(base.License, next.License),
		Version:        MergeValue(base.Version, next.Version),
	}
}

// Contact information for the exposed API.
type Contact struct {
	// The identifying name of the contact person/organization.
	Name string `json:"name,omitempty"`
	// The URL pointing to the contact information. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
	// The email address of the contact person/organization. MUST be in the format of an email address.
	Email string `json:"email,omitempty"`
}

var _ CanMerge[Contact] = &Contact{}

// Merges contact
func (base Contact) Merge(next Contact) Contact {
	return Contact{
		Name:  MergeValue(base.Name, next.Name),
		URL:   MergeValue(base.URL, next.URL),
		Email: MergeValue(base.Email, next.Email),
	}
}

// License information for the exposed API.
type License struct {
	// REQUIRED. The license name used for the API.
	Name string `json:"name"`
	// A URL to the license used for the API. MUST be in the format of a URL.
	URL string `json:"url,omitempty"`
}

var _ CanMerge[License] = &License{}

// Merges license
func (base License) Merge(next License) License {
	return License{
		Name: MergeValue(base.Name, next.Name),
		URL:  MergeValue(base.URL, next.URL),
	}
}

// Helper type to make building server variables cleaner.
type ServerVariables map[string]ServerVariable

// An object representing a Server.
type Server struct {
	// REQUIRED. A URL to the target host. This URL supports Server Variables and MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variable substitutions will be made when a variable is named in {brackets}.
	URL string `json:"url"`
	// An optional string describing the host designated by the URL. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A map between a variable name and its value. The value is used for substitution in the server's URL template.
	Variables ServerVariables `json:"variables,omitempty"`
}

var _ CanReplace[Server] = &Server{}

// Are the servers equal?
func (base Server) IsUnique(next Server) bool {
	return base.URL != next.URL
}

// An object representing a Server Variable for server URL template substitution.
type ServerVariable struct {
	// An enumeration of string values to be used if the substitution options are from a limited set. The array SHOULD NOT be empty.
	Enum []string `json:"enum,omitempty"`
	// REQUIRED. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied. Note this behavior is different than the Schema Object's treatment of default values, because in those cases parameter values are optional. If the enum is defined, the value SHOULD exist in the enum's values.
	Default string `json:"default,omitempty"`
	// An optional description for the server variable. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
}

// Holds a set of reusable objects for different aspects of the OAS. All objects defined within the components object will have no effect on the API unless they are explicitly referenced from properties outside the components object.
type Component struct {
	// An object to hold reusable Schema Objects.
	Schemas map[string]Schema `json:"schemas,omitempty"`
	// An object to hold reusable Response Objects.
	Responses Responses `json:"responses,omitempty"`
	// An object to hold reusable Parameter Objects.
	Parameters map[string]Parameter `json:"parameters,omitempty"`
	// An object to hold reusable Example Objects.
	Examples Examples `json:"examples,omitempty"`
	// An object to hold reusable Request Body Objects.
	RequestBodies map[string]RequestBody `json:"requestBodies,omitempty"`
	// An object to hold reusable Header Objects.
	Headers Headers `json:"headers,omitempty"`
	// An object to hold reusable Security Scheme Objects.
	SecuritySchemes map[string]Security `json:"securitySchemes,omitempty"`
	// An object to hold reusable Link Objects.
	Links Links `json:"links,omitempty"`
	// An object to hold reusable Callback Objects.
	Callbacks Callbacks `json:"callbacks,omitempty"`
}

// Merges components
func (base Component) Merge(next Component) Component {
	return Component{
		Schemas:         MergeMap(base.Schemas, next.Schemas),
		Responses:       MergeMap(base.Responses, next.Responses),
		Parameters:      MergeMap(base.Parameters, next.Parameters),
		Examples:        MergeMap(base.Examples, next.Examples),
		RequestBodies:   MergeMap(base.RequestBodies, next.RequestBodies),
		Headers:         MergeMap(base.Headers, next.Headers),
		SecuritySchemes: MergeMap(base.SecuritySchemes, next.SecuritySchemes),
		Links:           MergeMap(base.Links, next.Links),
		Callbacks:       MergeMap(base.Callbacks, next.Callbacks),
	}
}

// A type which can have a reference, and can apply one.
type HasReference interface {
	GetReference() *Reference
	SetReference(ref string)
	GetReferencePrefix() string
}

// Common content types that can be returned
type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeXML  ContentType = "application/xml"
	ContentTypeText ContentType = "text/plain"
	ContentTypeAny  ContentType = "*/*"
)

// Schema data types
type DataType string

const (
	DataTypeString  DataType = "string"
	DataTypeNumber  DataType = "number"
	DataTypeInteger DataType = "integer"
	DataTypeObject  DataType = "object"
	DataTypeArray   DataType = "array"
	DataTypeBoolean DataType = "boolean"
	DataTypeNull    DataType = "null"
)

// The Schema Object allows the definition of input and output data types. These types can be objects, but also primitives and arrays. This object is an extended subset of the JSON Schema Specification Wright Draft 00.
//
// For more information about the properties, see JSON Schema Core and JSON Schema Validation. Unless stated otherwise, the property definitions follow the JSON Schema.
type Schema struct {
	// Allows for an external definition of this schema.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// The schema type, if there is only one known
	Type DataType `json:"type,omitempty"`
	// The title and description keywords must be strings. A “title” will preferably be short, whereas a “description” will provide a more lengthy explanation about the purpose of the data described by the schema.
	Title string `json:"title,omitempty"`
	// The title and description keywords must be strings. A “title” will preferably be short, whereas a “description” will provide a more lengthy explanation about the purpose of the data described by the schema.
	Description string `json:"description,omitempty"`
	// Numbers can be restricted to a multiple of a given number, using the multipleOf keyword. It may be set to any positive number.
	MultipleOf int `json:"multipleOf,omitempty"`
	// Ranges of numbers are specified using a combination of the minimum and maximum keywords, (or exclusiveMinimum and exclusiveMaximum for expressing exclusive range).
	Maximum int `json:"maximum,omitempty"`
	// Ranges of numbers are specified using a combination of the minimum and maximum keywords, (or exclusiveMinimum and exclusiveMaximum for expressing exclusive range).
	ExclusiveMaximum bool `json:"exclusiveMaximum,omitempty"`
	// Ranges of numbers are specified using a combination of the minimum and maximum keywords, (or exclusiveMinimum and exclusiveMaximum for expressing exclusive range).
	Minimum int `json:"minimum,omitempty"`
	// Ranges of numbers are specified using a combination of the minimum and maximum keywords, (or exclusiveMinimum and exclusiveMaximum for expressing exclusive range).
	ExclusiveMinimum bool `json:"exclusiveMinimum,omitempty"`
	// The length of a string can be constrained using the minLength and maxLength keywords. For both keywords, the value must be a non-negative number.
	MaxLength int `json:"maxLength,omitempty"`
	// The length of a string can be constrained using the minLength and maxLength keywords. For both keywords, the value must be a non-negative number.
	MinLength int `json:"minLength,omitempty"`
	// (This string SHOULD be a valid regular expression, according to the Ecma-262 Edition 5.1 regular expression dialect)
	Pattern string `json:"pattern,omitempty"`
	// The length of the array can be specified using the minItems and maxItems keywords. The value of each keyword must be a non-negative number. These keywords work whether doing list validation or Tuple validation.
	MaxItems int `json:"maxItems,omitempty"`
	// The length of the array can be specified using the minItems and maxItems keywords. The value of each keyword must be a non-negative number. These keywords work whether doing list validation or Tuple validation.
	MinItems int `json:"minItems,omitempty"`
	// A schema can ensure that each of the items in an array is unique. Simply set the uniqueItems keyword to true.
	UniqueItems bool `json:"uniqueItems,omitempty"`
	// The number of properties on an object can be restricted using the minProperties and maxProperties keywords. Each of these must be a non-negative integer.
	MaxProperties int `json:"maxProperties,omitempty"`
	// The number of properties on an object can be restricted using the minProperties and maxProperties keywords. Each of these must be a non-negative integer.
	MinProperties int `json:"minProperties,omitempty"`
	// By default, the properties defined by the properties keyword are not required. However, one can provide a list of required properties using the required keyword.
	Required []string `json:"required,omitempty"`
	// The enum keyword is used to restrict a value to a fixed set of values. It must be an array with at least one element, where each element is unique.
	Enum []any `json:"enum,omitempty"`
	// The format keyword allows for basic semantic identification of certain kinds of string values that are commonly used. For example, because JSON doesn’t have a “DateTime” type, dates need to be encoded as strings. format allows the schema author to indicate that the string value should be interpreted as a date. By default, format is just an annotation and does not effect validation.
	// Examples:
	//   - "date-time": Date and time together, for example, 2018-11-13T20:20:39+00:00.
	//   - "time": Time, for example, 20:20:39+00:00
	//   - "date": Date, for example, 2018-11-13.
	//   - "duration": A duration as defined by the ISO 8601 ABNF for “duration”. For example, P3D expresses a duration of 3 days.
	//   - "email": Internet email address, see RFC 5321, section 4.1.2.
	//   - "idn-email": New in draft 7 The internationalized form of an Internet email address, see RFC 6531.
	//   - "hostname": Internet host name, see RFC 1123, section 2.1.
	//   - "idn-hostname": New in draft 7 An internationalized Internet host name, see RFC5890, section 2.3.2.3.
	//   - "ipv4": IPv4 address, according to dotted-quad ABNF syntax as defined in RFC 2673, section 3.2.
	//   - "ipv6": IPv6 address, as defined in RFC 2373, section 2.2.
	//   - "uuid": New in draft 2019-09 A Universally Unique Identifier as defined by RFC 4122. Example: 3e4666bf-d5e5-4aa7-b8ce-cefe41c7568a
	//   - "uri": A universal resource identifier (URI), according to RFC3986.
	//   - "uri-reference": New in draft 6 A URI Reference (either a URI or a relative-reference), according to RFC3986, section 4.1.
	//   - "iri": New in draft 7 The internationalized equivalent of a “uri”, according to RFC3987.
	//   - "iri-reference": New in draft 7 The internationalized equivalent of a “uri-reference”, according to RFC3987
	//   - "uri-template": New in draft 6 A URI Template (of any level) according to RFC6570. If you don’t already know what a URI Template is, you probably don’t need this value.
	//   - "json-pointer": New in draft 6 A JSON Pointer, according to RFC6901. There is more discussion on the use of JSON Pointer within JSON Schema in Structuring a complex schema. Note that this should be used only when the entire string contains only JSON Pointer content, e.g. /foo/bar. JSON Pointer URI fragments, e.g. #/foo/bar/ should use "uri-reference".
	//   - "relative-json-pointer": New in draft 7 A relative JSON pointer.
	//   - "regex": New in draft 7 A regular expression, which should be valid according to the ECMA 262 dialect.
	Format string `json:"format,omitempty"`
	// The default value for this schema.
	Default *any `json:"default,omitempty"`
	// Must be valid against all of the subschemas
	AllOf []Schema `json:"allOf,omitempty"`
	// Must be valid against exactly one of the subschemas
	OneOf []Schema `json:"oneOf,omitempty"`
	// Must be valid against any of the subschemas
	AnyOf []Schema `json:"anyOf,omitempty"`
	// Must not be valid against the given schema
	Not *Schema `json:"not,omitempty"`
	// List validation is useful for arrays of arbitrary length where each item matches the same schema. For this kind of array, set the items keyword to a single schema that will be used to validate all of the items in the array.
	Items *Schema `json:"items,omitempty"`
	// The properties (key-value pairs) on an object are defined using the properties keyword. The value of properties is an object, where each key is the name of a property and each value is a schema used to validate that property. Any property that doesn’t match any of the property names in the properties keyword is ignored by this keyword.
	Properties map[string]Schema `json:"properties,omitempty"`
	// The additionalProperties keyword is used to control the handling of extra stuff, that is, properties whose names are not listed in the properties keyword.
	AdditionalProperties *BoolSchema `json:"additionalProperties,omitempty"`
	// A true value adds "null" to the allowed type specified by the type keyword, only if type is explicitly defined within the same Schema Object. Other Schema Object constraints retain their defined behavior, and therefore may disallow the use of null as a value. A false value leaves the specified or default type unmodified. The default value is false.
	Nullable bool `json:"nullable,omitempty"`
	// Adds support for polymorphism. The discriminator is an object name that is used to differentiate between other schemas which may satisfy the payload description. See Composition and Inheritance for more details.
	Discriminator *Discriminator `json:"discriminator,omitempty"`
	// Relevant only for Schema "properties" definitions. Declares the property as "read only". This means that it MAY be sent as part of a response but SHOULD NOT be sent as part of the request. If the property is marked as readOnly being true and is in the required list, the required will take effect on the response only. A property MUST NOT be marked as both readOnly and writeOnly being true. Default value is false.
	ReadOnly bool `json:"readOnly,omitempty"`
	// Relevant only for Schema "properties" definitions. Declares the property as "write only". Therefore, it MAY be sent as part of a request but SHOULD NOT be sent as part of the response. If the property is marked as writeOnly being true and is in the required list, the required will take effect on the request only. A property MUST NOT be marked as both readOnly and writeOnly being true. Default value is false.
	WriteOnly bool `json:"writeOnly,omitempty"`
	// This MAY be used only on properties schemas. It has no effect on root schemas. Adds additional metadata to describe the XML representation of this property.
	XML *XML `json:"xml,omitempty"`
	// Additional external documentation for this schema.
	ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
	// A free-form property to include an example of an instance for this schema. To represent examples that cannot be naturally represented in JSON or YAML, a string value can be used to contain the example with escaping where necessary.
	Example *any `json:"example,omitempty"`
	// Specifies that a schema is deprecated and SHOULD be transitioned out of usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
}

var _ HasReference = &Schema{}
var _ CanMerge[Schema] = &Schema{}

func (base Schema) Merge(next Schema) Schema {
	if len(base.OneOf) > 0 {
		copy := base
		copy.OneOf = copy.OneOf[:]
		copy.OneOf = append(copy.OneOf, next)
		return copy
	} else {
		return Schema{
			OneOf: []Schema{
				base,
				next,
			},
		}
	}
}
func (hr Schema) GetReference() *Reference {
	return hr.Reference
}
func (hr *Schema) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Schema) GetReferencePrefix() string {
	return "#/components/schemas/"
}

// When request bodies or response payloads may be one of a number of different schemas, a discriminator object can be used to aid in serialization, deserialization, and validation. The discriminator is a specific object in a schema which is used to inform the consumer of the specification of an alternative schema based on the value associated with it.
//
// When using the discriminator, inline schemas will not be considered.
type Discriminator struct {
	// REQUIRED. The name of the property in the payload that will hold the discriminator value.
	PropertyName string `json:"propertyName"`
	// An object to hold mappings between payload values and schema names or references.
	Mapping map[string]string `json:"mapping"`
}

// A metadata object that allows for more fine-tuned XML model definitions.
//
// When using arrays, XML element names are not inferred (for singular/plural forms) and the name property SHOULD be used to add that information. See examples for expected behavior.
type XML struct {
	// Replaces the name of the element/attribute used for the described schema property. When defined within items, it will affect the name of the individual XML elements within the list. When defined alongside type being array (outside the items), it will affect the wrapping element and only if wrapped is true. If wrapped is false, it will be ignored.
	Name string `json:"name,omitempty"`
	// The URI of the namespace definition. Value MUST be in the form of an absolute URI.
	Namespace string `json:"namespace,omitempty"`
	// The prefix to be used for the name.
	Prefix string `json:"prefix,omitempty"`
	// Declares whether the property definition translates to an attribute instead of an element. Default value is false.
	Attribute bool `json:"attribute,omitempty"`
	// MAY be used only for an array definition. Signifies whether the array is wrapped (for example, <books><book/><book/></books>) or unwrapped (<book/><book/>). Default value is false. The definition takes effect only when defined alongside type being array (outside the items).
	Wrapped bool `json:"wrapped,omitempty"`
}

type Example struct {
	// Allows for an external definition of this example.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// Short description for the example.
	Summary string `json:"summary,omitempty"`
	// Long description for the example. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// Embedded literal example. The value field and externalValue field are mutually exclusive. To represent examples of media types that cannot naturally represented in JSON or YAML, use a string value to contain the example, escaping where necessary.
	Value *any `json:"value,omitempty"`
	// A URL that points to the literal example. This provides the capability to reference examples that cannot easily be included in JSON or YAML documents. The value field and externalValue field are mutually exclusive.
	ExternalValue string `json:"externalValue,omitempty"`
}

var _ HasReference = &Example{}

func (hr Example) GetReference() *Reference {
	return hr.Reference
}
func (hr *Example) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Example) GetReferencePrefix() string {
	return "#/components/examples/"
}

// Describes a single request body.
type RequestBody struct {
	// Allows for an external definition of this request body.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// A brief description of the request body. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// REQUIRED. The content of the request body. The key is a media type or media type range and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. text/plain overrides text/*
	Content Contents `json:"content,omitempty"`
	// Determines if the request body is required in the request. Defaults to false.
	Required bool `json:"required,omitempty"`
}

var _ CanMerge[RequestBody] = &RequestBody{}
var _ HasReference = &RequestBody{}

// Merges request bodies
func (base RequestBody) Merge(next RequestBody) RequestBody {
	return RequestBody{
		Reference:   MergeValue(base.Reference, next.Reference),
		Description: MergeValue(base.Description, next.Description),
		Content:     MergeMap(base.Content, next.Content),
		Required:    MergeValue(base.Required, next.Required),
	}
}
func (hr RequestBody) GetReference() *Reference {
	return hr.Reference
}
func (hr *RequestBody) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr RequestBody) GetReferencePrefix() string {
	return "#/components/requestBodies/"
}

// Similar to a parameter in the header, but reusable.
type Header struct {
	// Allows for an external definition of this header.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// Header shares the same fields
	ParameterBase
}

var _ CanMerge[Header] = &Header{}
var _ HasReference = &Header{}

// Merges parameter
func (base Header) Merge(next Header) Header {
	return Header{
		Reference:     MergeValue(base.Reference, next.Reference),
		ParameterBase: *MergeCanMerge(&base.ParameterBase, &next.ParameterBase),
	}
}
func (hr Header) GetReference() *Reference {
	return hr.Reference
}
func (hr *Header) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}

// Using links, you can describe how various values returned by one operation can be used as input for other operations. This way, links provide a known relationship and traversal mechanism between the operations. The concept of links is somewhat similar to hypermedia, but OpenAPI links do not require the link information present in the actual responses.
type Link struct {
	// Allows for an external definition of this link.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// A relative or absolute URI reference to an OAS operation. This field is mutually exclusive of the operationId field, and MUST point to an Operation Object. Relative operationRef values MAY be used to locate an existing Operation Object in the OpenAPI definition.
	OperationRef string `json:"operationRef,omitempty"`
	// The name of an existing, resolvable OAS operation, as defined with a unique operationId. This field is mutually exclusive of the operationRef field.
	OperationId string `json:"operationId,omitempty"`
	// A map representing parameters to pass to an operation as specified with operationId or identified via operationRef. The key is the parameter name to be used, whereas the value can be a constant or an expression to be evaluated and passed to the linked operation. The parameter name can be qualified using the parameter location [{in}.]{name} for operations that use the same parameter name in different locations (e.g. path.id).
	Parameters map[string]any `json:"parameters,omitempty"`
	// A literal value or {expression} to use as a request body when calling the target operation.
	RequestBody *any `json:"requestBody,omitempty"`
	// A description of the link. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A server object to be used by the target operation.
	Server *Server `json:"server,omitempty"`
}

var _ HasReference = &Link{}

func (hr Link) GetReference() *Reference {
	return hr.Reference
}
func (hr *Link) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Link) GetReferencePrefix() string {
	return "#/components/links/"
}

// A map of possible out-of band callbacks related to the parent operation. Each value in the map is a Path Item Object that describes a set of requests that may be initiated by the API provider and the expected responses. The key value used to identify the path item object is an expression, evaluated at runtime, that identifies a URL to use for the callback operation.
type Callbacks map[string] /*event*/ map[string] /*url*/ Path

// Describes the operations available on a single path. A Path Item MAY be empty, due to ACL constraints. The path itself is still exposed to the documentation viewer but they will not know which operations and parameters are available.
type Path struct {
	// Allows for an external definition of this path item. The referenced structure MUST be in the format of a Path Item Object. In case a Path Item Object field appears both in the defined object and the referenced object, the behavior is undefined.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// An optional, string summary, intended to apply to all operations in this path.
	Summary string `json:"summary,omitempty"`
	// An optional, string description, intended to apply to all operations in this path. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// A definition of a GET operation on this path.
	Get *Operation `json:"get,omitempty"`
	// A definition of a PUT operation on this path.
	Put *Operation `json:"put,omitempty"`
	// A definition of a POST operation on this path.
	Post *Operation `json:"post,omitempty"`
	// A definition of a DELETE operation on this path.
	Delete *Operation `json:"delete,omitempty"`
	// A definition of an OPTIONS operation on this path.
	Options *Operation `json:"options,omitempty"`
	// A definition of a HEAD operation on this path.
	Head *Operation `json:"head,omitempty"`
	// A definition of a PATCH operation on this path.
	Patch *Operation `json:"patch,omitempty"`
	// A definition of a TRACE operation on this path.
	Trace *Operation `json:"trace,omitempty"`
	// An alternative server array to service all operations in this path.
	Servers []Server `json:"servers,omitempty"`
	// A list of parameters that are applicable for all the operations described under this path. These parameters can be overridden at the operation level, but cannot be removed there. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the Reference Object to link to parameters that are defined at the OpenAPI Object's components/parameters.
	Parameters []Parameter `json:"parameters,omitempty"`
}

// Merges paths
func (base Path) Merge(next Path) Path {
	return Path{
		Reference:   MergeValue(base.Reference, next.Reference),
		Summary:     MergeValue(base.Summary, next.Summary),
		Description: MergeValue(base.Description, next.Description),
		Get:         MergeCanMerge(base.Get, next.Get),
		Put:         MergeCanMerge(base.Put, next.Put),
		Post:        MergeCanMerge(base.Post, next.Post),
		Delete:      MergeCanMerge(base.Delete, next.Delete),
		Options:     MergeCanMerge(base.Options, next.Options),
		Head:        MergeCanMerge(base.Head, next.Head),
		Patch:       MergeCanMerge(base.Patch, next.Patch),
		Trace:       MergeCanMerge(base.Trace, next.Trace),
		Servers:     MergeSliceReplace(base.Servers, next.Servers),
		Parameters:  MergeSliceReplace(base.Parameters, next.Parameters),
	}
}

var _ HasReference = &Path{}

func (hr Path) GetReference() *Reference {
	return hr.Reference
}
func (hr *Path) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Path) GetReferencePrefix() string {
	return "#/paths/"
}

// Describes a single API operation on a path.
type Operation struct {
	// A list of tags for API documentation control. Tags can be used for logical grouping of operations by resources or any other qualifier.
	Tags []string `json:"tags,omitempty"`
	// A short summary of what the operation does.
	Summary string `json:"summary,omitempty"`
	// A verbose explanation of the operation behavior. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// Additional external documentation for this operation.
	ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
	// Unique string used to identify the operation. The id MUST be unique among all operations described in the API. The operationId value is case-sensitive. Tools and libraries MAY use the operationId to uniquely identify an operation, therefore, it is RECOMMENDED to follow common programming naming conventions.
	OperationID string `json:"operationId,omitempty"`
	// A list of parameters that are applicable for this operation. If a parameter is already defined at the Path Item, the new definition will override it but can never remove it. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the Reference Object to link to parameters that are defined at the OpenAPI Object's components/parameters.
	Parameters []Parameter `json:"parameters,omitempty"`
	// The request body applicable for this operation. The requestBody is only supported in HTTP methods where the HTTP 1.1 specification RFC7231 has explicitly defined semantics for request bodies. In other cases where the HTTP spec is vague, requestBody SHALL be ignored by consumers.
	RequestBody *RequestBody `json:"requestBody,omitempty"`
	// REQUIRED. The list of possible responses as they are returned from executing this operation.
	Responses Responses `json:"responses,omitempty"`
	// A map of possible out-of band callbacks related to the parent operation. The key is a unique identifier for the Callback Object. Each value in the map is a Callback Object that describes a request that may be initiated by the API provider and the expected responses.
	Callbacks Callbacks `json:"callbacks,omitempty"`
	// Declares this operation to be deprecated. Consumers SHOULD refrain from usage of the declared operation. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
	// A declaration of which security mechanisms can be used for this operation. The list of values includes alternative security requirement objects that can be used. Only one of the security requirement objects need to be satisfied to authorize a request. To make security optional, an empty security requirement ({}) can be included in the array. This definition overrides any declared top-level security. To remove a top-level security declaration, an empty array can be used.
	Security []SecurityRequirement `json:"security,omitempty"`
	// An alternative server array to service this operation. If an alternative server object is specified at the Path Item Object or Root level, it will be overridden by this value.
	Servers []Server `json:"servers,omitempty"`
}

// Merges operations
func (base Operation) Merge(next Operation) Operation {
	return Operation{
		Tags:         MergeSliceUnique(base.Tags, next.Tags),
		Summary:      MergeValue(base.Summary, next.Summary),
		Description:  MergeValue(base.Description, next.Description),
		ExternalDocs: MergeValue(base.ExternalDocs, next.ExternalDocs),
		OperationID:  MergeValue(base.OperationID, next.OperationID),
		Parameters:   MergeSliceReplace(base.Parameters, next.Parameters),
		RequestBody:  MergeCanMerge(base.RequestBody, next.RequestBody),
		Responses:    MergeMap(base.Responses, next.Responses),
		Callbacks:    MergeMap(base.Callbacks, next.Callbacks),
		Deprecated:   MergeValue(base.Deprecated, next.Deprecated),
		Security:     MergeSlice(base.Security, next.Security),
		Servers:      MergeSliceReplace(base.Servers, next.Servers),
	}
}

// Given a struct which has fields representing parameters it will add each
// property in the generated schema as a parameter to this operation.
func (op *Operation) AddParameters(build *Builder, in ParameterIn, typ reflect.Type) {
	schema := build.BuildSchema(typ, false)
	if len(schema.Properties) == 0 {
		return
	}
	for paramName, prop := range schema.Properties {
		param := Parameter{}
		param.Name = paramName
		param.In = in
		param.Deprecated = prop.Deprecated
		param.Example = GetExample(prop)

		examples := GetExamples(prop, ContentTypeJSON)
		if len(examples) > 0 {
			param.Examples = make(map[string]any)
			for name, ex := range examples {
				param.Examples[name] = ex.Value
			}
		}

		if in == ParameterInPath {
			param.Required = true
		}
		if build.IsNullable(prop) {
			param.AllowEmptyValue = true
		} else {
			param.Required = true
		}

		inner := build.GetInnerSchema(prop)
		if inner != nil {
			param.Schema = inner
			param.Description = prop.Description
			param.Example = prop.Example
		} else {
			param.Schema = &prop
		}

		op.Parameters = append(op.Parameters, param)
	}
}

// A simple object to allow referencing other components in the specification, internally and externally.
//
// The Reference Object is defined by JSON Reference and follows the same structure, behavior and rules.
//
// For this specification, reference resolution is accomplished as defined by the JSON Reference specification and not by the JSON Schema specification.
type Reference struct {
	// REQUIRED. The reference string.
	Ref string `json:"$ref,omitempty"`
}

var _ HasReference = &Reference{}

func (hr *Reference) GetReference() *Reference {
	return hr
}
func (hr *Reference) SetReference(ref string) {
	if ref == "" {
		hr = nil
	} else {
		hr = &Reference{Ref: ref}
	}
}
func (hr Reference) GetReferencePrefix() string {
	return ""
}

// Parameter Locations
// There are four possible parameter locations specified by the in field:
//   - path: Used together with Path Templating, where the parameter value is actually part of the operation's URL. This does not include the host or base path of the API. For example, in /items/{itemId}, the path parameter is itemId.
//   - query: Parameters that are appended to the URL. For example, in /items?id=###, the query parameter is id.
//   - header: Custom headers that are expected as part of the request. Note that RFC7230 states header names are case insensitive.
//   - cookie: Used to pass a specific cookie value to the API.
type ParameterIn string

const (
	// Parameters that are appended to the URL. For example, in /items?id=###, the query parameter is id.
	ParameterInQuery ParameterIn = "query"
	// Custom headers that are expected as part of the request. Note that RFC7230 states header names are case insensitive.
	ParameterInHeader ParameterIn = "header"
	// Used together with Path Templating, where the parameter value is actually part of the operation's URL. This does not include the host or base path of the API. For example, in /items/{itemId}, the path parameter is itemId.
	ParameterInPath ParameterIn = "path"
	// Used to pass a specific cookie value to the API.
	ParameterInCookie ParameterIn = "cookie"
)

// In order to support common ways of serializing simple parameters, a set of style values are defined.
// https://swagger.io/docs/specification/serialization/
type Style string

const (
	// type(primitive, array, object), in(path), Path-style parameters defined by RFC6570
	ParameterStyleMatrix Style = "matrix"
	// type(primitive, array, object), in(path), Label style parameters defined by RFC6570
	ParameterStyleLabel Style = "label"
	// type(primitive, array, object), in(query, cookie), Form style parameters defined by RFC6570. This option replaces collectionFormat with a csv (when explode is false) or multi (when explode is true) value from OpenAPI 2.0.
	ParameterStyleForm Style = "form"
	// type(array), in(path, header), Simple style parameters defined by RFC6570. This option replaces collectionFormat with a csv value from OpenAPI 2.0.
	ParameterStyleSimple Style = "simple"
	// type(array), in(query), Space separated array values. This option replaces collectionFormat equal to ssv from OpenAPI 2.0.
	ParameterStyleSpaceDelimited Style = "spaceDelimited"
	// type(array), in(query), Pipe separated array values. This option replaces collectionFormat equal to pipes from OpenAPI 2.0.
	ParameterStylePipeDelimited Style = "pipeDelimited"
	// type(object), in(query), Provides a simple way of rendering nested objects using form parameters.
	ParameterStyleDeepObject Style = "deepObject"
)

// A base type shared between Parameter and Header
type ParameterBase struct {
	// A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// Determines whether this parameter is mandatory. If the parameter location is "path", this property is REQUIRED and its value MUST be true. Otherwise, the property MAY be included and its default value is false.
	Required bool `json:"required,omitempty"`
	// Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is false.
	Deprecated bool `json:"deprecated,omitempty"`
	// Sets the ability to pass empty-valued parameters. This is valid only for query parameters and allows sending a parameter with an empty value. Default value is false. If style is used, and if behavior is n/a (cannot be serialized), the value of allowEmptyValue SHALL be ignored. Use of this property is NOT RECOMMENDED, as it is likely to be removed in a later revision.
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	// Describes how the parameter value will be serialized depending on the type of the parameter value. Default values (based on value of in): for query - form; for path - simple; for header - simple; for cookie - form.
	Style Style `json:"style,omitempty"`
	// When this is true, parameter values of type array or object generate separate parameters for each value of the array or key-value pair of the map. For other types of parameters this property has no effect. When style is form, the default value is true. For all other styles, the default value is false.
	Explode bool `json:"explode,omitempty"`
	// Determines whether the parameter value SHOULD allow reserved characters, as defined by RFC3986 :/?#[]@!$&'()*+,;= to be included without percent-encoding. This property only applies to parameters with an in value of query. The default value is false.
	AllowReserved bool `json:"allowReserved,omitempty"`
	// The schema defining the type used for the parameter.
	Schema *Schema `json:"schema,omitempty"`
	// Example of the parameter's potential value. The example SHOULD match the specified schema and encoding properties if present. The example field is mutually exclusive of the examples field. Furthermore, if referencing a schema that contains an example, the example value SHALL override the example provided by the schema. To represent examples of media types that cannot naturally be represented in JSON or YAML, a string value can contain the example with escaping where necessary.
	Example *any `json:"example,omitempty"`
	// Examples of the parameter's potential value. Each example SHOULD contain a value in the correct format as specified in the parameter encoding. The examples field is mutually exclusive of the example field. Furthermore, if referencing a schema that contains an example, the examples value SHALL override the example provided by the schema.
	Examples map[string]any `json:"examples,omitempty"`
	// A map containing the representations for the parameter. The key is the media type and the value describes it. The map MUST only contain one entry.
	Content Contents `json:"content,omitempty"`
}

var _ CanMerge[ParameterBase] = &ParameterBase{}

// Merges parameter base type
func (base ParameterBase) Merge(next ParameterBase) ParameterBase {
	return ParameterBase{
		Description:     MergeValue(base.Description, next.Description),
		Required:        MergeValue(base.Required, next.Required),
		Deprecated:      MergeValue(base.Deprecated, next.Deprecated),
		AllowEmptyValue: MergeValue(base.AllowEmptyValue, next.AllowEmptyValue),
		Style:           MergeValue(base.Style, next.Style),
		Explode:         MergeValue(base.Explode, next.Explode),
		AllowReserved:   MergeValue(base.AllowReserved, next.AllowReserved),
		Schema:          MergeValue(base.Schema, next.Schema),
		Example:         MergeValue(base.Example, next.Example),
		Examples:        MergeMap(base.Examples, next.Examples),
		Content:         MergeMap(base.Content, next.Content),
	}
}

// Describes a single operation parameter.
//
// A unique parameter is defined by a combination of a name and location.
type Parameter struct {
	// // Allows for an external definition of this parameter.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// REQUIRED. The name of the parameter. Parameter names are case sensitive.
	//   - If in is "path", the name field MUST correspond to a template expression occurring within the path field in the Paths Object. See Path Templating for further information.
	//   - If in is "header" and the name field is "Accept", "Content-Type" or "Authorization", the parameter definition SHALL be ignored.
	//   - For all other cases, the name corresponds to the parameter name used by the in property.
	Name string `json:"name"`
	// REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	In ParameterIn `json:"in"`

	ParameterBase
}

var _ CanMerge[Parameter] = &Parameter{}
var _ CanReplace[Parameter] = &Parameter{}
var _ HasReference = &Parameter{}

func (base Parameter) IsUnique(next Parameter) bool {
	return base.Name != next.Name || base.In != next.In
}

// Merges parameter
func (base Parameter) Merge(next Parameter) Parameter {
	return Parameter{
		Reference:     MergeValue(base.Reference, next.Reference),
		Name:          MergeValue(base.Name, next.Name),
		In:            MergeValue(base.In, next.In),
		ParameterBase: *MergeCanMerge(&base.ParameterBase, &next.ParameterBase),
	}
}
func (hr Parameter) GetReference() *Reference {
	return hr.Reference
}
func (hr *Parameter) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Parameter) GetReferencePrefix() string {
	return "#/components/parameters/"
}

// Each Media Type Object provides schema and examples for the media type identified by its key.
type MediaType struct {
	// The schema defining the content of the request, response, or parameter.
	Schema *Schema `json:"schema,omitempty"`
	// Example of the media type. The example object SHOULD be in the correct format as specified by the media type. The example field is mutually exclusive of the examples field. Furthermore, if referencing a schema which contains an example, the example value SHALL override the example provided by the schema.
	Example *any `json:"example,omitempty"`
	// Examples of the media type. Each example object SHOULD match the media type and specified schema if present. The examples field is mutually exclusive of the example field. Furthermore, if referencing a schema which contains an example, the examples value SHALL override the example provided by the schema.
	Examples Examples `json:"examples,omitempty"`
	// A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property. The encoding object SHALL only apply to requestBody objects when the media type is multipart or application/x-www-form-urlencoded.
	Encoding Encodings `json:"encoding,omitempty"`
}

// A single encoding definition applied to a single schema property.
type Encoding struct {
	// The Content-Type for encoding a specific property. Default value depends on the property type: for string with format being binary – application/octet-stream; for other primitive types – text/plain; for object - application/json; for array – the default is defined based on the inner type. The value can be a specific media type (e.g. application/json), a wildcard media type (e.g. image/*), or a comma-separated list of the two types.
	ContentType string `json:"contentType,omitempty"`
	// A map allowing additional information to be provided as headers, for example Content-Disposition. Content-Type is described separately and SHALL be ignored in this section. This property SHALL be ignored if the request body media type is not a multipart.
	Headers Headers `json:"headers,omitempty"`
	// Describes how a specific property value will be serialized depending on its type. See Parameter Object for details on the style property. The behavior follows the same values as query parameters, including default values. This property SHALL be ignored if the request body media type is not application/x-www-form-urlencoded.
	Style Style `json:"style,omitempty"`
	// When this is true, property values of type array or object generate separate parameters for each value of the array, or key-value-pair of the map. For other types of properties this property has no effect. When style is form, the default value is true. For all other styles, the default value is false. This property SHALL be ignored if the request body media type is not application/x-www-form-urlencoded.
	Explode bool `json:"explode,omitempty"`
	// Determines whether the parameter value SHOULD allow reserved characters, as defined by RFC3986 :/?#[]@!$&'()*+,;= to be included without percent-encoding. The default value is false. This property SHALL be ignored if the request body media type is not application/x-www-form-urlencoded.
	AllowReserved bool `json:"allowReserved,omitempty"`
}

// A container for the expected responses of an operation. The container maps a HTTP response code to the expected response.
//
// The documentation is not necessarily expected to cover all possible HTTP response codes because they may not be known in advance. However, documentation is expected to cover a successful operation response and any known errors.
//
// The default MAY be used as a default response object for all HTTP codes that are not covered individually by the specification.
//
// The Responses Object MUST contain at least one response code, and it SHOULD be the response for a successful operation call.
type Responses map[string]*Response

// Helper type to make building content cleaner.
type Contents map[ContentType]*MediaType

// Helper type to make building headers cleaner.
type Headers map[string]*Header

// Helper type to make building links cleaner.
type Links map[string]Link

// Helper type to make building examples cleaner.
type Examples map[string]Example

// Helper type to make building encodings cleaner.
type Encodings map[string]Encoding

// Describes a single response from an API Operation, including design-time, static links to operations based on the response.
type Response struct {
	// Allows for an external definition of this response.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// REQUIRED. A short description of the response. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description"`
	// Maps a header name to its definition. RFC7230 states header names are case insensitive. If a response header is defined with the name "Content-Type", it SHALL be ignored.
	Headers Headers `json:"headers,omitempty"`
	// A map containing descriptions of potential response payloads. The key is a media type or media type range and the value describes it. For responses that match multiple keys, only the most specific key is applicable. e.g. text/plain overrides text/*
	Content Contents `json:"content,omitempty"`
	// A map of operations links that can be followed from the response. The key of the map is a short name for the link, following the naming constraints of the names for Component Objects.
	Links Links `json:"links,omitempty"`
}

var _ HasReference = &Response{}

// Merges a response
func (base Response) Merge(next Response) Response {
	return Response{
		Reference:   MergeValue(base.Reference, next.Reference),
		Description: MergeValue(base.Description, next.Description),
		Headers:     MergeMap(base.Headers, next.Headers),
		Content:     MergeMap(base.Content, next.Content),
		Links:       MergeMap(base.Links, next.Links),
	}
}
func (hr Response) GetReference() *Reference {
	return hr.Reference
}
func (hr *Response) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Response) GetReferencePrefix() string {
	return "#/components/responses/"
}

// A security type
type SecurityType string

const (
	SecurityTypeHTTP          SecurityType = "http"
	SecurityTypeApiKey        SecurityType = "apiKey"
	SecurityTypeOauth2        SecurityType = "oauth2"
	SecurityTypeOpenIDConnect SecurityType = "openIdConnect"
)

// Defines a security scheme that can be used by the operations. Supported schemes are HTTP authentication, an API key (either as a header, a cookie parameter or as a query parameter), OAuth2's common flows (implicit, password, client credentials and authorization code) as defined in RFC6749, and OpenID Connect Discovery.
type Security struct {
	// Allows for an external definition of this security.
	*Reference

	// Internal flag to note that a schema must be referenced if additional properties want to be applied
	named *Reference

	// REQUIRED. The type of the security scheme. Valid values are "apiKey", "http", "oauth2", "openIdConnect".
	Type SecurityType `json:"type"`
	// A short description for security scheme. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// REQUIRED (with apiKey). The name of the header, query or cookie parameter to be used.
	Name string `json:"name,omitempty"`
	// REQUIRED (with apiKey). The location of the API key. Valid values are "query", "header" or "cookie".
	In ParameterIn `json:"in,omitempty"`
	// REQUIRED (with http). The name of the HTTP Authorization scheme to be used in the Authorization header as defined in RFC7235. The values used SHOULD be registered in the IANA Authentication Scheme registry.
	Scheme string `json:"scheme,omitempty"`
	// (with http) A hint to the client to identify how the bearer token is formatted. Bearer tokens are usually generated by an authorization server, so this information is primarily for documentation purposes.
	BearerFormat string `json:"bearerFormat,omitempty"`
	// REQUIRED (with oauth2). An object containing configuration information for the flow types supported.
	Flows any `json:"flows,omitempty"`
	// REQUIRED (with openIdConnect). OpenId Connect URL to discover OAuth2 configuration values. This MUST be in the form of a URL.
	OpenIdConnectUrl string `json:"openIdConnectUrl,omitempty"`
}

var _ CanReplace[Security] = &Security{}
var _ HasReference = &Security{}

// If the two are equivalent
func (base Security) IsUnique(next Security) bool {
	return base.Type != next.Type || base.Name != next.Name || base.In != next.In
}
func (hr Security) GetReference() *Reference {
	return hr.Reference
}
func (hr *Security) SetReference(ref string) {
	if ref == "" {
		hr.Reference = nil
	} else {
		hr.Reference = &Reference{Ref: ref}
	}
}
func (hr Security) GetReferencePrefix() string {
	return "#/components/securitySchemes/"
}

// Lists the required security schemes to execute this operation. The name used for each property MUST correspond to a security scheme declared in the Security Schemes under the Components Object.
//
// Security Requirement Objects that contain multiple schemes require that all schemes MUST be satisfied for a request to be authorized. This enables support for scenarios where multiple query parameters or HTTP headers are required to convey security information.
//
// When a list of Security Requirement Objects is defined on the OpenAPI Object or Operation Object, only one of the Security Requirement Objects in the list needs to be satisfied to authorize the request.
type SecurityRequirement map[string][]string

// Adds metadata to a single tag that is used by the Operation Object. It is not mandatory to have a Tag Object per tag defined in the Operation Object instances.
type Tag struct {
	// REQUIRED. The name of the tag.
	Name string `json:"name"`
	// A short description for the tag. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
	// Additional external documentation for this tag.
	ExternalDocs *ExternalDoc `json:"externalDocs,omitempty"`
}

var _ CanReplace[Tag] = &Tag{}

// If the two tags are pointing to the same one
func (base Tag) IsUnique(next Tag) bool {
	return base.Name == next.Name
}

// Allows referencing an external resource for extended documentation.
type ExternalDoc struct {
	// REQUIRED. The URL for the target documentation. Value MUST be in the format of a URL.
	URL string `json:"url"`
	// A short description of the target documentation. CommonMark syntax MAY be used for rich text representation.
	Description string `json:"description,omitempty"`
}

// A helper type when a schema or a boolean can be given.
type BoolSchema struct {
	// Non-nil if bool should be used
	Bool *bool
	// Non-nil of schema should be used
	Schema *Schema
}

// Returns a BoolSchema for bool
func SchemaForBool(b bool) *BoolSchema {
	return &BoolSchema{Bool: &b}
}

// Returns a BoolSchema for Schema
func SchemaForSchema(s *Schema) *BoolSchema {
	return &BoolSchema{Schema: s}
}

var _ json.Marshaler = &BoolSchema{}
var _ json.Unmarshaler = &BoolSchema{}

func (bs BoolSchema) MarshalJSON() ([]byte, error) {
	if bs.Bool != nil {
		return json.Marshal(*bs.Bool)
	}
	return json.Marshal(bs.Schema)
}

func (bs *BoolSchema) UnmarshalJSON(data []byte) error {
	val := any(0)
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	if b, ok := val.(bool); ok {
		*bs = BoolSchema{Bool: &b}
	} else {
		schema := val.(Schema)
		*bs = BoolSchema{Schema: &schema}
	}
	return nil
}

// A utility function for all the *any fields in the api types.
func Any(val any) *any {
	return &val
}

// Returns a reference to the given type with the given name.
// The type needs to implement HasReference.
func Ref[V any](name string) V {
	var empty V
	asAny := any(&empty)
	if hasRef, ok := asAny.(HasReference); ok {
		hasRef.SetReference(RefTo(hasRef, name).Ref)
	}
	return empty
}

// Returns a reference to the given resource with the given name.
func RefTo(canRef HasReference, name string) *Reference {
	return &Reference{
		Ref: canRef.GetReferencePrefix() + EscapePathPart(name),
	}
}

// Merges base & next. If next is empty, base is returned. Otherwise next is returned.
func MergeValue[V comparable](base V, next V) V {
	var empty V
	if next == empty {
		return base
	}
	return next
}

// Merges base and next which implement CanMerge. This handles potential nils.
func MergeCanMerge[V CanMerge[V]](base *V, next *V) *V {
	if base == nil && next == nil {
		return nil
	}
	var merged V

	if base == nil {
		merged = (*next).Merge(merged)
	} else if next == nil {
		merged = (*base).Merge(merged)
	} else {
		merged = (*base).Merge(*next)
	}

	return &merged
}

// Merges the two slices together, base first foolowed by next.
func MergeSlice[V any](base []V, next []V) []V {
	merged := make([]V, 0, len(base)+len(next))
	if len(base) > 0 {
		merged = append(merged, base...)
	}
	if len(next) > 0 {
		merged = append(merged, next...)
	}
	return merged
}

// Merges the two slices together, avoiding duplicate values.
func MergeSliceUnique[V comparable](base []V, next []V) []V {
	merged := make([]V, 0, len(base)+len(next))
	if len(base) > 0 {
		merged = append(merged, base...)
	}
	if len(next) > 0 {
		for _, v := range next {
			unique := true
			for _, existing := range merged {
				if v == existing {
					unique = false
					break
				}
			}
			if unique {
				merged = append(merged, v)
			}
		}
	}
	return merged
}

// Merges the two slices together, avoiding duplicates based on the
// implementation of CanReplace.
func MergeSliceReplace[V CanReplace[V]](base []V, next []V) []V {
	merged := make([]V, 0, len(base)+len(next))
	if len(base) > 0 {
		merged = append(merged, base...)
	}
	if len(next) > 0 {
		for _, v := range next {
			duplicate := false
			for _, existing := range merged {
				if !existing.IsUnique(v) {
					duplicate = true
					break
				}
			}
			if !duplicate {
				merged = append(merged, v)
			}
		}
	}
	return merged
}

// Merges two maps. The value types can implement CanReplace and CanMerge
// and the merging process will respect those implementations when dealing
// with shared keys.
func MergeMap[K comparable, V any](base map[K]V, next map[K]V) map[K]V {
	merged := make(map[K]V, len(base)+len(next))
	if len(base) > 0 {
		for k, v := range base {
			merged[k] = v
		}
	}
	if len(next) > 0 {
		for k, v := range next {
			overwrite := true

			if existingValue, exists := merged[k]; exists {
				baseAny := any(existingValue)
				baseReplace, isBaseReplace := baseAny.(CanReplace[V])
				baseMerge, isBaseMerge := baseAny.(CanMerge[V])
				merge := isBaseMerge

				if isBaseReplace {
					merge = !baseReplace.IsUnique(v)
				}

				if merge {
					merged[k] = baseMerge.Merge(v)
					overwrite = false
				}
			}

			if overwrite {
				merged[k] = v
			}
		}
	}
	return merged
}

package rez

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

// An empty type to signal that a request has no parameters or body.
type None struct{}

// A value that holds types, so the dependency injection system can pick up on
// the types that represent the parts of a request.
type RequestTypes struct {
	Body   reflect.Type
	Path   reflect.Type
	Query  reflect.Type
	Header reflect.Type
}

// A type which has one or more injectable request types. This is how the dependency injected functions
// are inspected for types which are added to the Open API document.
type Injectable interface {
	deps.Dynamic

	APIRequestTypes() RequestTypes
	APIValidate(op *api.Operation, v *Validator)
}

// A function parameter that is injected with path parameters.
type Path[P any] struct {
	Value P
}

var _ Injectable = &Path[int]{}

func (p Path[P]) APIRequestTypes() RequestTypes {
	return RequestTypes{Path: deps.TypeOf[P]()}
}
func (p Path[P]) APIValidate(op *api.Operation, v *Validator) {
	schema := op.GetParametersSchema(api.ParameterInPath)
	Validate(&schema, p.Value, v.Next("path"))
}
func (r *Path[P]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getPath(&r.Value, request)
	if err != nil {
		return err
	}
	return ValidateInjectable(r, scope)
}

// A function parameter that is injected with query parameters.
type Query[Q any] struct {
	Value Q
}

var _ Injectable = &Query[int]{}

func (q Query[Q]) APIRequestTypes() RequestTypes {
	return RequestTypes{Query: deps.TypeOf[Q]()}
}
func (p Query[Q]) APIValidate(op *api.Operation, v *Validator) {
	schema := op.GetParametersSchema(api.ParameterInQuery)
	Validate(&schema, p.Value, v.Next("query"))
}
func (r *Query[Q]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getQuery(&r.Value, request)
	if err != nil {
		return err
	}
	return ValidateInjectable(r, scope)
}

// A function parameter that is injected with the request body.
type Body[B any] struct {
	Value B
}

var _ Injectable = &Body[int]{}

func (b Body[B]) APIRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[B]()}
}
func (p Body[B]) APIValidate(op *api.Operation, v *Validator) {
	if op.RequestBody != nil && op.RequestBody.Content != nil && op.RequestBody.Content[api.ContentTypeJSON] != nil {
		media := op.RequestBody.Content[api.ContentTypeJSON]
		if media.Schema != nil {
			Validate(media.Schema, p.Value, v.Next("body"))
		}
	}
}
func (b *Body[B]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getBody(&b.Value, request)
	if err != nil {
		return err
	}
	return ValidateInjectable(b, scope)
}

// A function parameter that is injected with the body, path, and query parameters.
type Request[B any, P any, Q any] struct {
	Body  B
	Path  P
	Query Q
}

var _ Injectable = &Request[int, int, int]{}

func (r Request[B, P, Q]) APIRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[B](), Path: deps.TypeOf[P](), Query: deps.TypeOf[Q]()}
}
func (p Request[B, P, Q]) APIValidate(op *api.Operation, v *Validator) {
	if op.RequestBody != nil && op.RequestBody.Content != nil && op.RequestBody.Content[api.ContentTypeJSON] != nil {
		media := op.RequestBody.Content[api.ContentTypeJSON]
		if media.Schema != nil {
			Validate(media.Schema, p.Body, v.Next("body"))
		}
	}
	pathSchema := op.GetParametersSchema(api.ParameterInPath)
	Validate(&pathSchema, p.Path, v.Next("path"))
	querySchema := op.GetParametersSchema(api.ParameterInQuery)
	Validate(&querySchema, p.Query, v.Next("query"))
}
func (r *Request[B, P, Q]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getBody(&r.Body, request)
	if err != nil {
		return err
	}
	err = getPath(&r.Path, request)
	if err != nil {
		return err
	}
	err = getQuery(&r.Query, request)
	if err != nil {
		return err
	}
	return ValidateInjectable(r, scope)
}

// A function parameter that is injected with the request headers.
type Header[H any] struct {
	Value H
}

var _ Injectable = &Header[int]{}

func (h Header[H]) APIRequestTypes() RequestTypes {
	return RequestTypes{Header: deps.TypeOf[H]()}
}
func (h Header[H]) APIValidate(op *api.Operation, v *Validator) {
	schema := op.GetParametersSchema(api.ParameterInHeader)
	Validate(&schema, h.Value, v.Next("header"))
}
func (r *Header[H]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getHeader(&r.Value, request)
	if err != nil {
		return err
	}
	return ValidateInjectable(r, scope)
}

// Validates the injectable by pulling the validator and operation
// off of the scope and calling APIValidate. If there are any validation
// errors the validator (which implements error) is returned.
func ValidateInjectable(inj Injectable, scope *deps.Scope) error {
	v, _ := deps.GetScoped[Validator](scope)
	op, _ := deps.GetScoped[api.Operation](scope)

	inj.APIValidate(op, v)

	if len(*v.Validations) > 0 {
		return v
	}
	return nil
}

func getHeader(header any, r *http.Request) error {
	outNode := &queryNode{}

	for k, v := range r.Header {
		if len(v) > 0 {
			outNode.get(k).set(v[0])
		}
	}

	outNode.fixForType(nonAnyType(header))
	out := outNode.convert()
	enc := encodeMap(header, out)

	return enc
}

func getBody(body any, r *http.Request) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(body)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func getPath(path any, r *http.Request) error {
	outNode := &queryNode{}

	ctx := chi.RouteContext(r.Context())
	if ctx != nil {
		for i, key := range ctx.URLParams.Keys {
			value := ctx.URLParams.Values[i]
			outNode.get(key).set(value)
		}
	}

	outNode.fixForType(nonAnyType(path))
	out := outNode.convert()
	enc := encodeMap(path, out)

	return enc
}

func getQuery(query any, r *http.Request) error {
	outNode := &queryNode{}
	pathRegex := regexp.MustCompile(`[\]\[\.]+`)
	queryValues := r.URL.Query()

	for k, v := range queryValues {
		if len(v) == 0 {
			continue
		}
		path := pathRegex.Split(strings.TrimRight(k, "]"), -1)
		curr := outNode
		for _, node := range path {
			curr = curr.get(node)
		}
		curr.set(v[0])
	}

	outNode.fixForType(nonAnyType(query))
	out := outNode.convert()
	enc := encodeMap(query, out)

	return enc
}

func nonAnyType(val any) reflect.Type {
	rv := reflect.ValueOf(val)
	for rv.Kind() == reflect.Interface || rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return rv.Type()
}

func encodeMap(target any, m any) error {
	s := strings.Builder{}
	err := json.NewEncoder(&s).Encode(m)
	if err != nil {
		return err
	}
	reader := strings.NewReader(s.String())
	err = json.NewDecoder(reader).Decode(target)
	if err != nil {
		return err
	}
	return nil
}

type queryNodeKind int

const (
	queryNodeKindUnspecified queryNodeKind = iota
	queryNodeKindSlice
	queryNodeKindObject
	queryNodeKindValue
)

type queryNode struct {
	obj   map[string]*queryNode
	arr   []*queryNode
	value any
	kind  queryNodeKind
}

func (node *queryNode) get(x string) *queryNode {
	if i, err := strconv.Atoi(x); err == nil {
		node.kind = queryNodeKindSlice
		if len(node.arr) <= i {
			arr := make([]*queryNode, i+1)
			copy(arr, node.arr)
			node.arr = arr
		}
		n := node.arr[i]
		if n == nil {
			n = &queryNode{}
			node.arr[i] = n
		}
		return n
	} else {
		node.kind = queryNodeKindObject
		if node.obj == nil {
			node.obj = map[string]*queryNode{}
		}
		n := node.obj[x]
		if n == nil {
			n = &queryNode{}
			node.obj[x] = n
		}
		return n
	}
}

func (node *queryNode) set(value any) {
	node.value = value
	node.kind = queryNodeKindValue
}

func (node *queryNode) fixForType(typ reflect.Type) {
	typ = getConcrete(typ)

	switch node.kind {
	case queryNodeKindSlice:
		if typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array {
			for _, item := range node.arr {
				item.fixForType(typ.Elem())
			}
		}
	case queryNodeKindObject:
		jt := getType(typ)
		if jt != nil {
			for k, v := range node.obj {
				field := jt.fields[strings.ToLower(k)]
				if field != nil {
					v.fixForType(field.fieldType)
				}
			}
		}

	case queryNodeKindValue:
		if typ != reflect.TypeOf(node.value) && node.value != nil {
			str := toString(node.value)
			val, err := parseType(typ, str)
			if err == nil {
				node.value = val
			}
		}
	}
}

func (node *queryNode) convert() any {
	switch node.kind {
	case queryNodeKindSlice:
		c := make([]any, len(node.arr))
		for i, item := range node.arr {
			if item != nil {
				c[i] = item.convert()
			} else {
				c[i] = nil
			}
		}
		return c
	case queryNodeKindObject:
		c := make(map[string]any)
		for key, value := range node.obj {
			c[key] = value.convert()
		}
		return c
	}

	return node.value
}

func toString(value any) string {
	return fmt.Sprintf("%v", value)
}

var ErrUnsupportedType = errors.New("unsupported type")

func parseType(t reflect.Type, s string) (any, error) {
	switch t.Kind() {
	case reflect.Float32:
		return strconv.ParseFloat(s, 32) // float64, error
	case reflect.Float64:
		return strconv.ParseFloat(s, 64) // float64, error
	case reflect.Bool:
		return strconv.ParseBool(s) // bool, error
	case reflect.Complex64:
		return strconv.ParseComplex(s, 64) // complex128, error
	case reflect.Complex128:
		return strconv.ParseComplex(s, 128) // complex128, error
	case reflect.Int:
		return strconv.ParseInt(s, 10, 64) // int64, error
	case reflect.Int8:
		return strconv.ParseInt(s, 10, 8) // int64, error
	case reflect.Int16:
		return strconv.ParseInt(s, 10, 16) // int64, error
	case reflect.Int32:
		return strconv.ParseInt(s, 10, 32) // int64, error
	case reflect.Int64:
		return strconv.ParseInt(s, 10, 64) // int64, error
	case reflect.Uint:
		return strconv.ParseUint(s, 10, 64) // uint64, error
	case reflect.Uint8:
		return strconv.ParseUint(s, 10, 8) // uint64, error
	case reflect.Uint16:
		return strconv.ParseUint(s, 10, 16) // uint64, error
	case reflect.Uint32:
		return strconv.ParseUint(s, 10, 32) // uint64, error
	case reflect.Uint64:
		return strconv.ParseUint(s, 10, 64) // uint64, error
	case reflect.String:
		return s, nil
	case reflect.Pointer:
		if s == "" {
			return nil, nil
		} else {
			nonNil, err := parseType(t.Elem(), s)
			return &nonNil, err
		}
	case reflect.Array:
		parts := strings.Split(s, ",")
		array := reflect.New(t).Elem()
		length := len(parts)
		if length > t.Len() {
			length = t.Len()
		}
		for i := 0; i < length; i++ {
			item, err := parseType(t.Elem(), parts[i])
			if err != nil {
				return nil, err
			}
			array.Index(i).Set(reflect.ValueOf(item))
		}
		return array.Interface(), nil
	case reflect.Slice:
		parts := strings.Split(s, ",")
		slice := reflect.MakeSlice(reflect.SliceOf(t.Elem()), 0, len(parts))
		for i := 0; i < len(parts); i++ {
			item, err := parseType(t.Elem(), parts[i])
			if err != nil {
				return nil, err
			}
			slice = reflect.Append(slice, reflect.ValueOf(item))
		}
		return slice.Interface(), nil
	}

	return nil, ErrUnsupportedType
}

var jsonTypes map[reflect.Type]*jsonType = make(map[reflect.Type]*jsonType)

type jsonType struct {
	fields map[string]*jsonField
}

type jsonField struct {
	fieldType reflect.Type
	indices   []int
}

func getType(typ reflect.Type) *jsonType {
	jt := jsonTypes[typ]
	if jt != nil {
		return jt
	}
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	jt = &jsonType{
		fields: make(map[string]*jsonField),
	}
	jsonTypes[typ] = jt

	var iterateFields func(st reflect.Type, indices []int)

	iterateFields = func(st reflect.Type, indices []int) {
		for i := 0; i < st.NumField(); i++ {
			field := st.Field(i)
			fieldIndices := append(indices[:], i)
			if field.Anonymous {
				iterateFields(field.Type, fieldIndices)
			} else {
				key := field.Name
				if json := field.Tag.Get("json"); json != "" {
					options := strings.Split(json, ",")
					prop := options[0]
					if prop == "-" {
						continue
					}
					if prop != "" {
						key = prop
					}
				}
				key = strings.ToLower(key)

				jt.fields[key] = &jsonField{
					fieldType: field.Type,
					indices:   fieldIndices,
				}
			}
		}
	}
	iterateFields(typ, []int{})

	return jt
}

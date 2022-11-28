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
	"github.com/go-chi/chi/v5"
)

// An empty type to signal that a request has no parameters or body.
type None struct{}

// A value that holds types, so the dependency injection system can pick up on
// the types that represent the parts of a request.
type RequestTypes struct {
	Body   reflect.Type
	Param  reflect.Type
	Query  reflect.Type
	Header reflect.Type
}

// A type which has one or more injectable request types. This is how the dependency injected functions
// are inspected for types which are added to the Open API document.
type HasRequestTypes interface {
	GetRequestTypes() RequestTypes
}

// A function parameter that is injected with path parameters.
type Param[P any] struct {
	Value P
}

var _ deps.Dynamic = &Param[int]{}
var _ HasRequestTypes = &Param[int]{}

func (p Param[P]) GetRequestTypes() RequestTypes {
	return RequestTypes{Param: deps.TypeOf[P]()}
}
func (r *Param[P]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	return getParam(&r.Value, request)
}

// A function parameter that is injected with query parameters.
type Query[Q any] struct {
	Value Q
}

var _ deps.Dynamic = &Query[int]{}
var _ HasRequestTypes = &Query[int]{}

func (q Query[Q]) GetRequestTypes() RequestTypes {
	return RequestTypes{Query: deps.TypeOf[Q]()}
}
func (r *Query[Q]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	return getQuery(&r.Value, request)
}

// A function parameter that is injected with the request body.
type Body[B any] struct {
	Value B
}

var _ deps.Dynamic = &Body[int]{}
var _ HasRequestTypes = &Body[int]{}

func (b Body[B]) GetRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[B]()}
}
func (b *Body[B]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	return getBody(&b.Value, request)
}

// A function parameter that is injected with the body, path, and query parameters.
type Request[B any, P any, Q any] struct {
	Body  B
	Param P
	Query Q
}

var _ deps.Dynamic = &Request[int, int, int]{}
var _ HasRequestTypes = &Request[int, int, int]{}

func (r Request[B, P, Q]) GetRequestTypes() RequestTypes {
	return RequestTypes{Body: deps.TypeOf[B](), Param: deps.TypeOf[P](), Query: deps.TypeOf[Q]()}
}
func (r *Request[B, P, Q]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	err := getBody(&r.Body, request)
	if err != nil {
		return err
	}
	err = getParam(&r.Param, request)
	if err != nil {
		return err
	}
	err = getQuery(&r.Query, request)
	if err != nil {
		return err
	}
	return nil
}

// A function parameter that is injected with the request headers.
type Header[H any] struct {
	Value H
}

var _ deps.Dynamic = &Header[int]{}
var _ HasRequestTypes = &Header[int]{}

func (h Header[H]) GetRequestTypes() RequestTypes {
	return RequestTypes{Header: deps.TypeOf[H]()}
}
func (r *Header[H]) ProvideDynamic(scope *deps.Scope) error {
	request, _ := deps.GetScoped[http.Request](scope)
	return getHeader(&r.Value, request)
}

func getHeader(header any, r *http.Request) error {
	outNode := &queryNode{}

	for k, v := range r.Header {
		if len(v) > 0 {
			outNode.get(k).set(v[0])
		}
	}

	outNode.fixForType(reflect.TypeOf(header))
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

func getParam(param any, r *http.Request) error {
	outNode := &queryNode{}

	ctx := chi.RouteContext(r.Context())
	if ctx != nil {
		for i, key := range ctx.URLParams.Keys {
			value := ctx.URLParams.Values[i]
			outNode.get(key).set(value)
		}
	}

	outNode.fixForType(reflect.TypeOf(param))
	out := outNode.convert()
	enc := encodeMap(param, out)

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

	outNode.fixForType(reflect.TypeOf(query))
	out := outNode.convert()
	enc := encodeMap(query, out)

	return enc
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
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

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

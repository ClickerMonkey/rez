package rez

import (
	"net/http"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

// A function which handles the error and returns true or returns false for the error to be handled by default behavior.
type ErrorHandler = func(err error, response http.ResponseWriter, request *http.Request, scope *deps.Scope) (bool, error)

// A function which handles an error that couldn't be sent to the client.
type InternalErrorHandler = func(err error)

// An error type which has custom error handling.
type HandledError interface {
	error
	// Handle the error
	Handle(response http.ResponseWriter, request *http.Request, scope *deps.Scope) error
}

// A response type that has a known status but could return more than one status.
// HTTPStatus is used when returning a response and HTTPStatuses is used for documentation.
type HasStatus interface {
	// The status of this particular response
	HTTPStatus() int
	// All possible statuses for this response type
	HTTPStatuses() []int
}

// A response which has a custom content type.
type HasContentType interface {
	HTTPContentType() string
}

// A response which has custom sending logic.
type CanSend interface {
	HTTPSend(w http.ResponseWriter) error
}

// Router with OpenAPI integration and dependency injection
type Router interface {
	ValidationProvider

	// The internal chi.Router
	Chi() chi.Router

	// The url of the router at this point.
	URL() string

	// Sets the error handler at this router and all sub routers created after this is set.
	SetErrorHandler(handler ErrorHandler)

	// Sets the handler for errors we received outside of responding to the client.
	SetInternalErrorHandler(handle InternalErrorHandler)

	// Handles the given error if its a HandledError, is handled by the error handler, or is handled with default behavior.
	HandleError(err error, response http.ResponseWriter, request *http.Request, scope *deps.Scope) error

	// Enables or disables validation for all routes in this router or sub routers created after this is set.
	// By default validation is not enabled.
	EnableValidation(enabled bool)

	// Sets the validation options for the type or value's type.
	SetValidationOptions(valueOrType any, options ValidationOptions)

	// Sets the memory limit (in bytes) for multipart/form-data requests.
	// Any request larger than this will utilize temporary files.
	SetMemoryLimit(memoryLimit int64)

	// Gets the memory limit (in bytes) for multipart/form-data requests.
	// Any request larger than this will utilize temporary files.
	GetMemoryLimit() int64

	// Adds the types of the given values as injectable request bodies. This avoids
	// the necessity of rez.Body or rez.Request. If any of the values/types
	// have already been defined this will cause a panic.
	DefineBody(bodies ...any)

	// Adds the types of the given values as injectable parameters. This avoids
	// the necessity of rez.Param or rez.Request. If any of the values/types
	// have already been defined this will cause a panic.
	DefinePath(params ...any)

	// Adds the types of the given values as injectable query parameters. This avoids
	// the necessity of rez.Query or rez.Request. If any of the values/types
	// have already been defined this will cause a panic.
	DefineQuery(queries ...any)

	// Adds the types of the given values as injectable header values. This avoids
	// the necessity of rez.Header. If any of the values/types
	// have already been defined this will cause a panic.
	DefineHeader(headers ...any)

	// Gets the base operation which has all inherited tags and responses set at the current router.
	GetOperations() *api.Operation

	// Sets the base operation. All sub routes will inherit the properties on the base operation.
	SetOperations(op api.Operation)

	// Updates the base operation (merges in given properties into existing base operation).
	// All sub routes will inherit the properties on the base operation.
	UpdateOperations(op api.Operation)

	// Gets the path defined at the given pattern, if any
	GetPath(pattern string) *api.Path

	// Gets the path define at the given pattern, creating it if need be.
	CreatePath(pattern string) *api.Path

	// Merges the path definition into the path
	UpdatePath(pattern string, path api.Path)

	// Sets the tags for all child routes starting at this router.
	// This is similar to router.SetOperation(api.Operation{Tags: tags}).
	SetTags(tags []string)

	// Adds the tags for all child routes starting at this router
	AddTags(tags []string)

	// Sets the responses for all child routes starting at this router
	// This is similar to router.SetOperation(api.Operation{Responses: responses}).
	SetResponses(responses api.Responses)

	// Adds responses for all child routes starting at this router.
	AddResponses(responses api.Responses)

	// Adds a response for all child routes starting at this router.
	AddResponse(code string, response api.Response)

	// Gets or creates a scope for the given request if it doesn't exist yet.
	GetScope(response http.ResponseWriter, request *http.Request) (scope *deps.Scope, freeScope bool)

	// Use appends one or more middlewares onto the Router stack.
	// A middleware is a dependency injectable function.
	// The function has access to the request, response, scope, and any other
	// injectable request values. The values applied to the scope here will
	// be passed down to lower routes and can be injected in their functions.
	// If the function has problems injecting arguments or returns any errors
	// then the next handler in the stack will not be invoked and the error
	// will be handled like any other error.
	Use(middlewares ...any)

	// With adds inline middlewares for an endpoint handler and returns a
	// new router at the same URL.
	With(middlewares ...any) Router

	// Group adds a new inline-Router along the current routing
	// path, with a fresh middleware stack for the inline-Router.
	Group(fn func(r Router)) Router

	// Route mounts a sub-Router along a `pattern` string.
	Route(pattern string, fn func(r Router)) Router

	// Handle and HandleFunc adds routes for `pattern` that matches all HTTP methods.
	HandleFunc(pattern string, fn any, operations ...api.Operation) *api.Path

	// Method and MethodFunc adds routes for `pattern` that matches
	// the `method` HTTP method.
	MethodFunc(method, pattern string, fn any, operations ...api.Operation) RouterOperation

	// HTTP-method routing along `pattern`
	Connect(pattern string, fn any)
	Delete(pattern string, fn any, operations ...api.Operation) RouterOperation
	Get(pattern string, fn any, operations ...api.Operation) RouterOperation
	Head(pattern string, fn any, operations ...api.Operation) RouterOperation
	Options(pattern string, fn any, operations ...api.Operation) RouterOperation
	Patch(pattern string, fn any, operations ...api.Operation) RouterOperation
	Post(pattern string, fn any, operations ...api.Operation) RouterOperation
	Put(pattern string, fn any, operations ...api.Operation) RouterOperation
	Trace(pattern string, fn any, operations ...api.Operation) RouterOperation

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	NotFound(fn any)

	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	MethodNotAllowed(fn any)
}

// A router operation
type RouterOperation interface {
	// The operation documentation.
	Operation() *api.Operation

	// The router of the operation
	Router() Router

	// Adds the given type/instance as an input (body, param, query, header) to the operation.
	Input(input ...any)

	// Adds the given type/instance as a response type to the operation.
	Output(output ...any)
}

package rez

import (
	"net/http"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

// The middleware type alias
type Middleware = func(http.Handler) http.Handler

// A function which handles the error and returns true or returns false for the error to be handled by default behavior.
type ErrorHandler = func(err error, response http.ResponseWriter, request *http.Request, scope *deps.Scope) bool

// An error type which has custom error handling.
type HandledError interface {
	error
	// Handle the error
	Handle(response http.ResponseWriter, request *http.Request, scope *deps.Scope)
}

type HasStatus interface {
	// The status of this particular response
	HTTPStatus() int
	// All possible statuses for this response type
	HTTPStatuses() []int
}

// Router with OpenAPI integration and dependency injection
type Router interface {
	// The internal chi.Router
	Chi() chi.Router

	// The url of the router at this point.
	URL() string

	// Sets the error handler at this router and all sub routers created after this is set.
	SetErrorHandler(handler ErrorHandler)

	// Handles the given error if its a HandledError, is handled by the error handler, or is handled with default behavior.
	HandleError(err error, response http.ResponseWriter, request *http.Request, scope *deps.Scope)

	// Adds the types of the given values as injectable request bodies. This avoids
	// the necessity of rez.Body or rez.Request. If any of the values/types
	// have already been defined this will cause a panic.
	DefineBody(bodies ...any)

	// Adds the types of the given values as injectable parameters. This avoids
	// the necessity of rez.Param or rez.Request. If any of the values/types
	// have already been defined this will cause a panic.
	DefineParam(params ...any)

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

	// Sets the tags for all child routes starting at this router
	// This is similar to router.SetOperation(api.Operation{Responses: responses}).
	SetResponses(responses api.Responses)

	// Adds responses for all child routes starting at this router.
	AddResponses(responses api.Responses)

	// Adds a response for all child routes starting at this router.
	AddResponse(code string, response api.Response)

	// Gets or creates a scope for the given request if it doesn't exist yet.
	GetScope(response http.ResponseWriter, request *http.Request) *deps.Scope

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
	MethodFunc(method, pattern string, fn any, operations ...api.Operation) *api.Operation

	// HTTP-method routing along `pattern`
	Connect(pattern string, fn any)
	Delete(pattern string, fn any, operations ...api.Operation) *api.Operation
	Get(pattern string, fn any, operations ...api.Operation) *api.Operation
	Head(pattern string, fn any, operations ...api.Operation) *api.Operation
	Options(pattern string, fn any, operations ...api.Operation) *api.Operation
	Patch(pattern string, fn any, operations ...api.Operation) *api.Operation
	Post(pattern string, fn any, operations ...api.Operation) *api.Operation
	Put(pattern string, fn any, operations ...api.Operation) *api.Operation
	Trace(pattern string, fn any, operations ...api.Operation) *api.Operation

	// NotFound defines a handler to respond whenever a route could
	// not be found.
	NotFound(fn any)

	// MethodNotAllowed defines a handler to respond whenever a method is
	// not allowed.
	MethodNotAllowed(fn any)
}

# rez
REST (easy) framework in Go with out of the box OpenAPI generation, validation, dependency injection, and much more. 

- [Dependency Injection](#dependency-injection) How values are sent to middleware and endpoint functions.
- [Router](#router) Defining the routes, middlware, & documentation.
- [Middleware](#middleware) Code that runs before it reaches the final endpoint
- [Inspection](#inspection) How types are converted into documentation.
- [Validation](#validation) How to control validation.
- [Documentation](#documentation) All the ways to specify documentation.
- [Site](#methods) The main site type and its useful methods.

### Example
```go
type Echo struct {
	Message string `json:"message" api:"desc=A message to echo."`
}
site := rez.New(chi.NewRouter())
// /echo?message=HelloWorld!
site.Get("/echo", func(q rez.Query[Echo]) (*Echo, *rez.NotFound[string]) {
  if q.Value.Message == "" {
    return nil, rez.NewNotFound("message is required")
  }
  return &q.Value, nil
})
site.ServeSwaggerUI("/doc/swagger", nil)
site.ServeRedoc("/doc/redoc", nil)
site.Listen(":3000")
```
More can be found in [examples](examples).

## Dependency Injection

Dependency injection is used to pass arguments to a middleware or route functions. **rez** uses [deps](https://github.com/ClickerMonkey/deps) for dependency injection. There are several types that get injected out of the box:

- `context.Context`: The context of the request.
- `deps.Scope`: The scope which holds all the given values that can be injected for the current request. The root `Site` type has a parent scope which request scopes can inherit values from.
- `http.ResponseWriter`: The outgoing response.
- `http.Request`: The incoming request.
- `rez.Router`: The reference to the router where the middleware was used or the route was defined on.
- `rez.Path[P]`: A generic wrapper which holds the struct that is parsed from the path parameters. If the path is `/task/{taskID}` and the struct is `type TaskPath struct { TaskID int }` the `TaskID` property will be populated from the value in the URL.
- `rez.Query[Q]`: A generic wrapper which holds the struct that is parsed from the query string. If the url is `?message=Hi&times=4` and the struct is `type MyQuery struct { Message string, Times int }` the Message and Times fields will be populated from the query string.
- `rez.Header[H]`: A generic wrapper which holds the struct that is parsed from the headers. 
- `rez.Body[B]`: A generic wrapper which holds the type that is parsed from the request body.
- `rez.Request[B, P, Q]`: A generic wrapper which holds the body, params, and query structs that are to be parsed from the request.
- `rez.Validator`: A validator for the route or middleware.
- `api.Operation`: The operation (route only).
- `rez.MiddlewareNext`: Invoke the next handler (middleware only).

There are a few other methods to get other injectable values.
1. Use `rez.Site.Scope` to set global values and providers.
2. Implement `rez.Injectable`.
3. Use `rez.Router.DefineBody(bodies...)` to define types that will only be used as arguments that should come from the request body.
4. Use `rez.Router.DefinePath(paths...)` to define types that will only be used as arguments that should come from the request path parameters.
5. Use `rez.Router.DefineQuery(queries...)` to define types that will only be used as arguments that should come from the request query parameters.
6. Use `rez.Router.DefineHeader(headers...)` to define types that will only be used as arguments that should come from the request headers.
7. Use `*deps.Scope` as an argument in middleware and `Set` or `Provide` other values that the following handlers will be able to receive.

## Router
The [rez.Router](rez.go) is a wrapper of chi.Router where instead of `http.Hander`s and `http.HandlerFunc` you pass in a `func(args) results` which gets its arguments injected, and in the case of middleware is able to provide injected values for routes in the router. The function argument and result types are also inspected to build the OpenAPI documentation.

## Middleware
Middleware in **rez** is also a dependency injected function. The middleware can return nothing or can return an error which if non-nil will be sent as the response. The middleware has a special injected value `rez.MiddlewareNext` which is a function to call if we want to call the next handler. _Any arguments or return types that are identified as headers, queries, paths, request bodies, or responses are added as those objects in all routes that are in the router using the middleware._

Example:
```go
// site.Use(authMiddleware)
func authMiddleware(next rez.MiddlewareNext, r *http.Request) *rez.Unauthorized[string] {
  // Accessing headers this way doesn't add it as a parameter to all routes that use the middleware.
  auth := r.Header.Get("Authorization")
  if auth == "" {
    return rez.NewUnauthorized("No access")
  }
  next()
  return nil
}
```

You can also pass down injectable values to all middlewares and routes which are defined after middleware by setting the value on the scope.

```go
type User struct { ID int }
// site.Use(authMiddleware)
func authMiddleware(next rez.MiddlewareNext, s *deps.Scope) {
  // authenticate user and return error, if successful apply the user to the scope.
  s.Set(User{ID: 23})
  next()
}
type Task struct { ID int, Name string, Done bool }
// set.Get("/tasks", getTasks)
func getTasks(user User) Task[] {
  // get tasks the user can see, we only get here if authMiddleware called next
  return []Task{}
}
```

As mentioned what you reference with middleware could add to the operations that follow. This middleware is an example of authentication where the token is foolishly sent in the query string. All operations that follow will have the specified security scheme, accept `token` as a query parameter, and could respond with `rez.Unauthorized[string]`.

```go
type Auth struct { Token string }

type AuthMiddleware func(s *deps.Scope, next rez.MiddlewareNext, q rez.Query[Auth]) *rez.Unauthorized[string]
// All routes which use this middleware accept this type of security
func (auth AuthMiddleware) APIOperationUpdate(op *api.Operation) {
  op.Security = append(op.Security, map[string][]string{"queryAuth": {}})
}

var authMiddleware AuthMiddleware = func(s *deps.Scope, next rez.MiddlewareNext, q rez.Query[Auth]) *rez.Unauthorized[string] {
  if q.Value.Token == "" {
    return rez.NewUnauthorized("No access")
  }
  s.Set(q.Value)
  next()
  return nil
}
func echoToken(token Auth) Auth {
  return token
}

// Usage
site.Open.AddSecurity("queryAuth", &api.Security{
  Type: api.SecurityTypeApiKey,
  Name: "token",
  In:   api.ParameterInQuery,
})
site.Use(authMiddleware)
site.Get("/token", echoToken)
```

## Inspection

Function arguments are inspected to determine what path parameters, query parameters, headers, and body is used by a route. See [Dependency Injection](#dependency-injection) for more details on that. The types detected are converted into `api` objects and are added to the OpenAPI document and referenced in the path & operations in the path. The function return arguments are inspected for possible responses - most of the time these return types will be pointers for routes which can have multiple response types (or no specific response type). If the return type implements `rez.HasStatus` that is where the status code is pulled from. If the return type does not it's assumed to be a possible OK (200) result. The schemas built from the argument and return types are built once and can be controlled using various functions and interfaces. If the type is a struct then `json` and `api` tags can control the field visibility or schema options. See [Documentation](#documentation) for additional details on how to control the documentation & validation that is generated.

## Validation

Validation in rez is done if enabled and only for certain schema fields and after the data is marshalled into values. So any invalid type errors will not be triggered by the validation but when the JSON is parsed. General validation options can be applied per type, validation can be enabled or disabled for any router, and types can have custom validation code that takes over the validation process or runs after the validation process. If validation fails the error is returned to the user. How those validations are sent to the user can be controlled by calling `rez.Router.SetErrorHandler`.

- `rez.Router.EnabledValidation(bool)` enables or disables validation in this router and any sub-routers created after this call. By default validation is not enabled.
- `rez.Router.SetValidationOptions(any,ValidationOptions)` sets the validation options for the given type, which controls if validation is skipped, if format is enforced, or if specifying deprecated values triggers a validation error.
- `rez.CanValidateFull` if a type implements this it handles all validation logic.
- `rez.CanValidatePost` if a type implements this it will do additional validation logic after other validation logic has been done.
- `rez.Injectable` if a type implements this it must implement an `APIValidate` method.

The following schema fields are used during validation:
- `MultipleOf`, `Maximum`, `Minimum`, `ExclusiveMaximum`, `ExclusiveMinimum` are used for any int or float types.
- `MaxLength`, `MinLength` are used for string types.
- `Deprecated`, `Nullable`, `Pattern`, `Format`, `Enum`, `OneOf`, `AllOf`, `AnyOf`, `Not` are used for all types.
- `MinItems`, `MaxItems`, `Items`, `UniqueItems` are used for array and slice types.
- `MinProperties`, `MaxProperties`, `AdditionalProperties` are used for map types.
- `Properties`, `Required` are used for struct types.


## Documentation

Documentation is control by various ways on the types themselves or through router methods.

- `api.HasName`
A type's documented name is the name of the type in the GO code, but there might be collisions. If there are collisions the OpenAPI built will have schema names that include the types pkg path to be unique. To avoid those potentially lengthy names you can implement `api.HasName` like so:
```go
// tasks folder
type Search struct { Name string }
func (Search) APIName() string { return "TaskSearch" }
```
- `api.Description`
A request or response's description can be specified on tyhe type by implementing this interface. Using the code above.
```go
func (Search) APIDescription() string { return "This is used to determine what Tasks to return." }
```
- `api.HasBaseSchema`
A type's schema will be dynamically determined, but implementing this interface will provide the schema building logic with a starting point. You can define the preferred schema properties.
```go
func (Search) APIBaseSchema() *api.Schema {
  return &api.Schema{
    Title:       "Task Search",
    Description: "This is used to determine what Tasks to return.",
    Example      api.Any(Search{Name: "homework"}),
  }
}
```
- `api.HasFullSchema`
A type's schema will only be determined by what's returned, no further inspection is done. This is useful if you want to use some of the built-in struct types that support different formats.
```go
type Timestamp time.Time
func (Timestamp) APIFullSchema() *api.Schema {
  return &api.Schema{
    Type:    api.DataTypeString,
    Format:  "date-time",
    Pattern: `\\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d+\d\d:\d\d`,
    Example: api.Any("2018-11-13T20:20:39+00:00"),
  }
}
```
- `api.HasEnum`
A type can accept only a handful of values.
```go
type TodoAction string
const (
  TodoActionArchive  TodoAction = "archive"
  TodoActionDelete   TodoAction = "delete"
  TodoActionComplete TodoAction = "complete"
)
func (TodoAction) APIEnum() []any {
  return []any{TodoActionArchive, TodoActionDelete, TodoActionComplete}
}
```
- `api.HasExamples`
A type can provide several named examples for a given content type.
```go
func (Search) APIExamples(contentType api.ContentType) api.Examples {
  return api.Examples{
    "All tasks": api.Example{
      Summary: "This search will return all tasks the user can see.",
      Value:   api.Any(Search{}),
    },
    "Tasks with 'homework' in the name": api.Example{
      Summary: "This search will return all tasks with 'homework' in the name.",
      Value:   api.Any(Search{Name: "homework"}),
    },
  }
}
```
- `api.HasExample`
A type that can provide a single example for a type.
```go
func (Search) APIExample() *any {
  return api.Any(Search{Name: "homework"})
}
```
- `api.HasOperation`
A route function that has the operation fully defined here and no inspection needs to be done on the arguments or return types.
```go
type GetTask func(id string) *Task
func (GetTask) APIOperation() api.Operation {
  return api.Operation{
    Tags:        []string{"Task"},
    Summary:     "Get the task with the given ID",
    OperationID: "GET_TASK_BY_ID",
    Parameters:  []api.Parameter{{
      Name:     "id",
      In:       api.ParameterInPath,
      Required: true,
      Schema:   &api.Schema{Type: api.TypeString},
      Example:  api.Any("87y34"),
    }},
    Responses:  api.Responses{
      "200":    &api.Response{
        Description: "The task exists and has these values",
        Content:     api.Contents{
          api.ContentTypeJSON: &api.MediaType{
            Schema: site.Open.GetSchema(reflect.TypeOf(Task{})),
          },
        },
      },
    },
  }
}
```
- `api.HasOperationUpdate`
A route function that modifies the operation inspected after its done inspection.
```go
type GetTask func(id string) *Task
func (GetTask) APIOperationUpdate(op *api.Operation) {
  op.OperationID = "GET_TASK_BY_ID"
}
```
- `rez.Site.Open` is a reference to `api.Builder` which has a `Document` field which can be modified. This is the base document to use before building the final `api.Document`.
- `rez.Router` has a few methods to assist in documentation:
  - `GetOperations() *api.Operation` returns a reference to the operation template that has accumulated at this point in the router. Sub routers inherit this. Middlewares add to it.
  - `SetOperations(api.Operation)` sets the operation template in its entirety, overwriting what has been built so far.
  - `UpdateOperations(api.Operation)` merges in the fields set on the given operation into the operation template of this router.
  - `GetPath(pattern) *api.Path` returns a reference to the path with the given pattern. Defining methods will add operations to this path. If the path has not been defined yet nil is returned.
  - `CreatePath(pattern) *api.Path` returns a reference to the path with the given pattern, creating it if need be.
  - `UpdatePath(pattern, api.Path)` merges in the fields set on the given path with the path defined at the given pattern - creating it if need be.
  - `SetTags(tags []string)` sets the tags on the operation template to this value.
  - `AddTags(tags)` adds the tags to the operation template.
  - `SetResponses(api.Responses)` sets the responses for all operations defined after. This overwrites any responses specified previously by the user or middlewares.
  - `AddResponses(api.Responses)` adds the responses to the operation template.
  - `AddResponse(code, api.Response)` adds the response to the operation template.
  - `HandleFunc(pattern, fn, ...api.Operation) *api.Path` can accept zero or more operation definitions to merge into the operations defined at this path - and the reference to the path at the pattern is returned.
  - `"method"(pattern, fn, ...api.Operation) *api.Operation` is a method with the name of any of the HTTP methods which adds this method to the path with the pattern and merges in any given operations with the operation template and then returns the reference to the final built operation for this route.
- Struct tags. Fields on a struct can specify the `api` tag which is a comma-delimited list of key=value or flags. If you need to use a comma in a value you can escape it like `\,`.
  - `title` ex: `api:"title=A person's address"` (see `api.Schema.Title`)
  - `desc` or `description` ex: `api:"desc=The ten digit home phone number."` (see `api.Schema.Description`)
  - `format` ex: `api:"format=email"` (see `api.Schema.Format`)
  - `pattern` ex: `api:"pattern=\d+"` (see `api.Schema.Pattern`)
  - `deprecated` ex: `api:"deprecated"` (see `api.Schema.Deprecated`)
  - `required` ex: `api:"required"` (see `api.Schema.Nullable`)
  - `null` or `nullable` ex: `api:"null"` (see `api.Schema.Nullable`)
  - `readonly` ex: `api:"readonly"` (see `api.Schema.ReadOnly`)
  - `writeonly` ex: `api:"writeonly"` (see `api.Schema.WriteOnly`)
  - `enum` ex: `api:"enum=1|2|3"` (see `api.Schema.Enum`)
  - `minlength` ex: `api:"minlength=6"` (see `api.Schema.MinLength`)
  - `maxlength` ex: `api:"maxlength=6"` (see `api.Schema.MaxLength`)
  - `minitems` ex: `api:"minitems=6"` (see `api.Schema.MinItems`)
  - `maxitems` ex: `api:"maxitems=6"` (see `api.Schema.MaxItems`)
  - `multipleof` ex: `api:"multipleof=2"` (see `api.Schema.MultipleOf`)
  - `min` or `minimum` ex: `api:"min=1"` (see `api.Schema.Minimum`)
  - `max` or `maximum` ex: `api:"max=1"` (see `api.Schema.Maximum`)
  - `exclusivemaximum` or `exclusivemax` ex: `api:"exclusivemax=true"` (see `api.Schema.ExclusiveMaximum`)
  - `exclusiveminimum` or `exclusivemin` ex: `api:"exclusivemin"` (see `api.Schema.ExclusiveMinimum`)

## Site

`rez.Site` is the implementation of router that must be created with `rez.New(chi.Router)`. Site has a few additional methods:
- `BuildDocument() *api.Document` returns the built document based on the routes and middlewares defined thus far.
- `BuildJSON() []byte` calls `BuildDocument` and marshals it to JSON.
- `ServeOpenJSON(patten)` serves the `BuildJSON` to a GET route at the defined pattern. This gets called by the other `Serve` document related endpoints if it was not called yet with a default pattern of `openapi3.json`.
- `ServeSwaggerUI(pattern,options)` serves an HTML page at the given pattern which presents the SwaggerUI which points to the OpenAPI document JSON.
- `ServeRedoc(pattern)` serves an HTML page at the given pattern which presents the Redoc which points to the OpenAPI document JSON.
- `Listen(addr)` starts the site and blocks until it stops.
- `Run()` starts the site but looks at the CLI args for a `--host` argument to specify the port. It defaults to `:80`.
- `PrintPaths()` prints an ASCII grid to the console with the paths described in the site at this point in time. Includes the "Method", "URL", and "About" if any summary or descriptions are given. Example output:
```
┌───────┬───────────┬────────────────────┐
│Method │URL        │About               │
├───────┼───────────┼────────────────────┤
│GET    │/task/{id} │Get task by id      │
├───────┼───────────┼────────────────────┤
│DELETE │/task/{id} │Delete task by id   │
├───────┼───────────┼────────────────────┤
│GET    │/auth      │Get current session │
├───────┼───────────┼────────────────────┤
│POST   │/auth      │Login               │
├───────┼───────────┼────────────────────┤
│DELETE │/auth      │Logout              │
└───────┴───────────┴────────────────────┘
```

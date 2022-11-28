# rez
REST (easy) framework in Go with out of the box OpenAPI generation, validation, dependency injection, and much more. 

### Router
The rez.Router is a wrapper of chi.Router where instead of `http.Hander`s and `http.HandlerFunc` you pass in a `func(args) results` which gets its arguments injected, and in the case of middleware is able to provide injected values for endpoints in the router. The function argument and result types are also evaluated to build the OpenAPI documentation.

### Middleware

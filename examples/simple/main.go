package main

import (
	"fmt"
	"regexp"
	"time"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

type Headers struct {
	Authorization string
}
type TaskParams struct {
	ID int `json:"id"`
}
type Task struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Done   bool       `json:"done"`
	DoneAt *time.Time `json:"doneAt,omitempty"`
}
type AuthRequest struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}
type AuthResult struct {
	Token string `json:"token"`
}
type AuthError struct{}
type JWT string

func (err AuthError) Error() string       { return "AuthError" }
func (err AuthError) HTTPStatus() int     { return 403 }
func (err AuthError) HTTPStatuses() []int { return []int{403} }

func main() {
	// These functions can have the following types automatically injected:
	// - http.Request
	// - http.ResponseWriter
	// - context.Context
	// - deps.Scope
	// - rez.Router
	// - rez.Param
	// - rez.Body
	// - rez.Query
	// - rez.Header
	// - rez.Request (contains Body, Query, & Param)
	// Any types returned will be added to possible response types.
	// The response types can have HTTPStatus[es] methods, otherwise 200 is assumed.
	r := chi.NewRouter()

	site := rez.New(r)
	site.Open.AddTag(api.Tag{
		Name:        "Task",
		Description: "A collection of task related operations",
	})
	site.Open.AddTag(api.Tag{
		Name:        "Authentication",
		Description: "A collection of auth related operations",
	})
	api.SetFullSchema[time.Time](site.Open, &api.Schema{
		Type:        api.DataTypeString,
		Description: "Date & time",
		Format:      "date-time",
	})

	site.Route("/task", func(r rez.Router) {
		r.Use(authMiddleware)
		r.UpdateOperations(api.Operation{
			Tags: []string{"Task"},
		})
		r.UpdatePath("/{id}", api.Path{
			Summary: "Operations on a specific task",
		})

		r.Get("/{id}", getTask, api.Operation{
			Summary: "Get task by id",
		})
		r.Delete("/{id}", deleteTask, api.Operation{
			Summary: "Delete task by id",
		})
	})
	site.Group(func(r rez.Router) {
		r.UpdateOperations(api.Operation{
			Tags: []string{"Authentication"},
		})
		r.UpdatePath("/auth", api.Path{
			Summary: "Operations for authentication",
		})

		r.Post("/auth", authLogin, api.Operation{Summary: "Login"})
		r.With(authMiddleware).Get("/auth", authGet, api.Operation{Summary: "Get current session"})
		r.Delete("/auth", authLogout, api.Operation{Summary: "Logout"})
	})

	fmt.Println("Listening on http://localhost:3000")

	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc")

	site.Listen(":3000")
}

func getTask(params rez.Param[TaskParams]) (*Task, *rez.NotFound[string]) {
	if params.Value.ID == 0 {
		return nil, &rez.NotFound[string]{}
	}
	return &Task{ID: params.Value.ID, Name: "New Task"}, nil
}
func deleteTask(params rez.Param[TaskParams]) *rez.NotFound[string] {
	if params.Value.ID == 0 {
		return &rez.NotFound[string]{}
	}
	return nil
}
func authLogin(body rez.Body[AuthRequest]) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: body.Value.Password + body.Value.Password}, nil
}
func authGet(token JWT) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: string(token)}, nil
}
func authLogout() (bool, *rez.Unauthorized[string]) {
	return true, nil
}

// All return types will be added to all responses of routes in the router where this is used
func authMiddleware(headers rez.Header[Headers], scope *deps.Scope, router rez.Router, next rez.MiddlewareNext) *rez.Unauthorized[string] {
	bearer, err := regexp.Compile(`^[Bb]earer (.+)$`)
	if err != nil {
		return &rez.Unauthorized[string]{}
	}
	token := JWT(bearer.FindStringSubmatch(headers.Value.Authorization)[1])
	if token == "" {
		fmt.Printf("In middleware, error found because no token")
		return &rez.Unauthorized[string]{}
	}
	deps.SetScoped(scope, &token)
	next()
	return nil
}

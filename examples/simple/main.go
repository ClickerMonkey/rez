package main

import (
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
	Done   bool       `json:"done" api:"desc=If the task is complete\\, if true doneAt should be given."`
	DoneAt *time.Time `json:"doneAt,omitempty" api:"desc=When the task was marked done."`
}
type AuthRequest struct {
	ID       string `json:"id" api:"desc=The email/username of the user to authenticate."`
	Password string `json:"password"`
}
type AuthResult struct {
	Token string `json:"token" api:"desc=The JWT generated from a successful login."`
}

func (r AuthRequest) APIDescription() string {
	return "The result of a successful authentication."
}

type JWT string
type AuthError struct{}

func (err AuthError) Error() string       { return "AuthError" }
func (err AuthError) HTTPStatus() int     { return 403 }
func (err AuthError) HTTPStatuses() []int { return []int{403} }

func main() {
	site := rez.New(chi.NewRouter())
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

	site.DefineParam(TaskParams{})
	site.DefineBody(AuthRequest{})
	site.DefineHeader(Headers{})

	site.Route("/task", func(r rez.Router) {
		r.Use(authMiddleware)
		r.UpdatePath("/{id}", api.Path{Summary: "Operations on a specific task"})
		r.UpdateOperations(api.Operation{Tags: []string{"Task"}})

		r.Get("/{id}", getTask, api.Operation{Summary: "Get task by id"})
		r.Delete("/{id}", deleteTask, api.Operation{Summary: "Delete task by id"})
	})
	site.Group(func(r rez.Router) {
		r.UpdatePath("/auth", api.Path{Summary: "Operations for authentication"})
		r.UpdateOperations(api.Operation{Tags: []string{"Authentication"}})

		r.Post("/auth", authLogin, api.Operation{Summary: "Login"})
		r.With(authMiddleware).Get("/auth", authGet, api.Operation{Summary: "Get current session"})
		r.Delete("/auth", authLogout, api.Operation{Summary: "Logout"})
	})

	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc")
	site.Listen(":3000")
}

func getTask(params TaskParams) (*Task, *rez.NotFound[string]) {
	if params.ID == 0 {
		return nil, &rez.NotFound[string]{}
	}
	return &Task{ID: params.ID, Name: "New Task"}, nil
}
func deleteTask(params TaskParams) *rez.NotFound[string] {
	if params.ID == 0 {
		return &rez.NotFound[string]{}
	}
	return nil
}
func authLogin(body AuthRequest) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: body.ID + body.Password}, nil
}
func authGet(token JWT) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: string(token)}, nil
}
func authLogout() (*rez.OK[string], *rez.Unauthorized[string]) {
	return &rez.OK[string]{Result: "OK"}, nil
}
func authMiddleware(headers Headers, scope *deps.Scope, router rez.Router, next rez.MiddlewareNext) *rez.Unauthorized[string] {
	bearer, err := regexp.Compile(`^[Bb]earer (.+)$`)
	if err != nil {
		return &rez.Unauthorized[string]{}
	}
	matches := bearer.FindStringSubmatch(headers.Authorization)
	if len(matches) < 2 {
		return &rez.Unauthorized[string]{}
	}
	token := JWT(matches[1])
	if token == "" {
		return &rez.Unauthorized[string]{}
	}
	deps.SetScoped(scope, &token)
	next()
	return nil
}

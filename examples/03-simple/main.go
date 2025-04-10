package main

import (
	"net/http"
	"regexp"
	"time"

	"github.com/ClickerMonkey/deps"
	"github.com/ClickerMonkey/rez"
	"github.com/ClickerMonkey/rez/api"
	"github.com/go-chi/chi/v5"
)

type SearchQuery struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

var _ rez.CanValidateFull = SearchQuery{}

func (q SearchQuery) FullValidate(v *rez.Validator) {
	if q.Limit < 0 {
		v.Next("limit").Add(rez.Validation{Message: "limit must be positive"})
	} else if q.Limit > 1000 {
		v.Next("limit").Add(rez.Validation{Message: "limit must be less than 1000"})
	}
	if q.Offset < 0 {
		v.Next("offset").Add(rez.Validation{Message: "offset must be positive"})
	}
}

func (q SearchQuery) Resolve() (limit int, offset int) {
	limit = q.Limit
	offset = q.Offset
	if limit <= 0 {
		limit = 20
	}
	return
}

type TaskPath struct {
	ID int `json:"id" api:"min=1"`
}
type Task struct {
	ID     int        `json:"id"`
	Name   string     `json:"name"`
	Done   bool       `json:"done" api:"desc=If the task is complete\\, if true doneAt should be given."`
	DoneAt *time.Time `json:"doneAt,omitempty" api:"desc=When the task was marked done."`
}
type TaskSearchRequest struct {
	Name *string `json:"name"`
	Done *bool   `json:"done"`
}
type TaskSearchResponse struct {
	Offset  int    `json:"offset"`
	Total   int    `json:"total"`
	Results []Task `json:"results"`
}
type AuthRequest struct {
	ID       string `json:"id" api:"desc=The email/username of the user to authenticate.,format=email"`
	Password string `json:"password"`
}
type AuthResult struct {
	Token string `json:"token" api:"desc=The JWT generated from a successful login."`
}

func (r AuthRequest) APIDescription() string {
	return "The result of a successful authentication."
}

type JWT string

func main() {
	site := rez.New(chi.NewRouter())
	site.Open.Document.Info.Title = "REZ Simple Example"
	site.Open.Document.Info.Description = "REZ Simple Example API Documentation"
	site.Open.AddSecurity("bearer", &api.Security{
		Type:         api.SecurityTypeHTTP,
		Description:  "Authentication with a 'Bearer {token}' in the Authorization header.",
		Scheme:       "bearer",
		BearerFormat: "JWT",
	})
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

	site.EnableValidation(true)
	site.Route("/task", func(r rez.Router) {
		r.Use(authMiddleware)
		r.UpdatePath("/{id}", api.Path{Summary: "Operations on a specific task"})
		r.UpdateOperations(api.Operation{Tags: []string{"Task"}})
		r.DefinePath(TaskPath{})
		r.DefineQuery(SearchQuery{})
		r.DefineBody(TaskSearchRequest{})

		r.Get("/{id}", getTask, api.Operation{Summary: "Get task by id"})
		r.Delete("/{id}", deleteTask, api.Operation{Summary: "Delete task by id"})
		r.Post("/search", searchTask, api.Operation{Summary: "Search tasks"})
	})
	site.Group(func(r rez.Router) {
		r.UpdatePath("/auth", api.Path{Summary: "Operations for authentication"})
		r.UpdateOperations(api.Operation{Tags: []string{"Authentication"}})
		r.DefineBody(AuthRequest{})
		r.SetValidationOptions("", rez.ValidationOptions{EnforceFormat: true})

		r.Post("/auth", authLogin, api.Operation{Summary: "Login"})
		r.With(authMiddleware).Get("/auth", authGet, api.Operation{Summary: "Get current session"})
		r.Delete("/auth", authLogout, api.Operation{Summary: "Logout"})
	})

	site.PrintPaths()
	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc", nil)
	site.Listen(":3000")
}

func getTask(params TaskPath) (*Task, *rez.NotFound[string]) {
	if params.ID == 0 {
		return nil, rez.NewNotFound("Task not found")
	}
	return &Task{ID: params.ID, Name: "New Task"}, nil
}
func deleteTask(params TaskPath) *rez.NotFound[string] {
	if params.ID == 0 {
		return rez.NewNotFound("Task not found")
	}
	return nil
}
func searchTask(body TaskSearchRequest, query SearchQuery) (*TaskSearchResponse, *rez.Unauthorized[string]) {
	_, offset := query.Resolve()

	return &TaskSearchResponse{
		Total:   1,
		Results: []Task{{Name: "REZ Examples", Done: false}},
		Offset:  offset,
	}, nil
}
func authLogin(body AuthRequest) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: body.ID + body.Password}, nil
}
func authGet(token JWT) (*AuthResult, *rez.Unauthorized[string]) {
	return &AuthResult{Token: string(token)}, nil
}
func authLogout() (*rez.OK[string], *rez.Unauthorized[string]) {
	return rez.NewOK("OK"), nil
}

// Middleware which also updates the operations it's applied to.

type AuthMiddleware func(r *http.Request, scope *deps.Scope, router rez.Router, next rez.MiddlewareNext) *rez.Unauthorized[string]

var _ api.HasOperationUpdate = AuthMiddleware(nil)

func (am AuthMiddleware) APIOperationUpdate(op *api.Operation) {
	op.Security = append(op.Security, map[string][]string{"bearer": {}})
}

var bearerRegex *regexp.Regexp = regexp.MustCompile(`^[Bb]earer (.+)$`)

var authMiddleware AuthMiddleware = func(r *http.Request, scope *deps.Scope, router rez.Router, next rez.MiddlewareNext) *rez.Unauthorized[string] {
	matches := bearerRegex.FindStringSubmatch(r.Header.Get("Authorization"))
	if len(matches) < 2 {
		return rez.NewUnauthorized("")
	}
	token := JWT(matches[1])
	if token == "" {
		return rez.NewUnauthorized("")
	}
	deps.SetScoped(scope, &token)
	next()
	return nil
}

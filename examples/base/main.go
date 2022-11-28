package main

import (
	"github.com/ClickerMonkey/rez"
	"github.com/go-chi/chi/v5"
)

type MessageResult struct {
	Message string `json:"message"`
}

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

	site.Get("/message", func() MessageResult {
		return MessageResult{Message: "Hello World"}
	})

	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc")

	site.Listen(":3000")
}

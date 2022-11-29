package main

import (
	"github.com/ClickerMonkey/rez"
	"github.com/go-chi/chi/v5"
)

type Echo struct {
	Message string `json:"message" api:"desc=A message to echo."`
}

func main() {
	site := rez.New(chi.NewRouter())
	// /echo?message=HelloWorld!
	site.Get("/echo", func(q rez.Query[Echo]) (*Echo, *rez.NotFound[string]) {
		if q.Value.Message == "" {
			return nil, &rez.NotFound[string]{Result: "message is required"}
		}
		return &q.Value, nil
	})

	site.PrintPaths()
	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc")
	site.Listen(":3000")
}

package main

import (
	"github.com/ClickerMonkey/rez"
	"github.com/go-chi/chi/v5"
)

type MessageResult struct {
	Message string `json:"message" api:"desc=A message"`
}

func main() {
	site := rez.New(chi.NewRouter())

	site.Get("/message", func() MessageResult {
		return MessageResult{Message: "Hello World"}
	})

	site.ServeSwaggerUI("/doc/swagger", nil)
	site.ServeRedoc("/doc/redoc")
	site.Listen(":3000")
}

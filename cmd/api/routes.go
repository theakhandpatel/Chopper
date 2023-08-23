package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", app.healthCheckHandler)
	r.Get("/api/shorten", app.shortenURLHandler)
	r.Get("/{shortURL}", app.expandURLHandler)

	return r
}

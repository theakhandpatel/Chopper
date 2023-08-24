package main

import (
	"net/http"

	"github.com/go-chi/chi"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/", app.HealthCheckHandler)
	r.Get("/api/shorten", app.ShortenURLHandler)
	r.Get("/api/stats", app.AnalyticsHandler)
	r.Get("/{shortURL}", app.ExpandURLHandler)

	return r
}

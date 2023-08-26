package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// routes configures the application's routing using Chi router.
func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.HealthCheckHandler)
	r.Get("/api/shorten", app.rateLimit(app.ShortenURLHandler))
	r.Get("/api/stats", app.AnalyticsHandler)
	r.Get("/{shortURL}", app.ExpandURLHandler)
	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	// Apply middleware for logging and error recovery
	return app.metrics(app.recoverPanic(middleware.Logger(r)))
}

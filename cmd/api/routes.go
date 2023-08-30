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
	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	r.Get("/", app.HealthCheckHandler)

	r.Post("/api/shorten", app.rateLimitForShortner(app.ShortenURLHandler))
	r.Get("/api/stats", app.AnalyticsHandler)
	r.Post("/api/signup", app.registerUserHandler)
	r.Post("/api/signin", app.loginUserHandler)
	r.Post("/api/signout", app.requireAuthenticatedUser(app.logoutUserHandler))

	r.Get("/{shortURL}", app.rateLimit(app.ExpandURLHandler))
	// Apply middleware for logging and error recovery
	return app.metrics(middleware.Logger(app.recoverPanic(app.authenticate(r))))
}

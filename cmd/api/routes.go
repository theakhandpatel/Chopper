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

	r.Post("/api/short", app.rateLimitForShortner(app.CreateShortURLHandler))
	r.Get("/api/short/{shortCode}", app.requireAuthorizedUser(app.GetShortURLHandler))
	r.Put("/api/short/{shortCode}", app.requireAuthorizedUser(app.EditShortURLHandler))
	r.Delete("/api/short/{shortCode}", app.requireAuthorizedUser(app.DeleteShortURLHandler))
	r.Get("/api/stats/{shortCode}", app.requireAuthenticatedUser(app.AnalyticsHandler))

	r.Post("/api/signup", app.registerUserHandler)
	r.Post("/api/signin", app.loginUserHandler)
	r.Post("/api/signout", app.requireAuthenticatedUser(app.logoutUserHandler))

	r.Post("/api/premium", app.requireAuthenticatedUser(app.registerPremiumHandler))

	r.Get("/{shortURL}", app.rateLimit(app.ExpandURLHandler))
	// Apply middleware for logging and error recovery
	return app.metrics(middleware.Logger(app.recoverPanic(app.authenticate(r))))
}

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

	// Apply middleware for logging and error recovery
	r.Use(app.metrics)
	r.Use(middleware.Logger)
	r.Use(app.recoverPanic)
	r.Use(app.authenticate)

	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	r.Get("/", app.HealthCheckHandler)

	r.Get("/api/stats/{shortCode}", app.requirePremiumUser(app.AnalyticsHandler))

	r.Route("/api/short", func(sr chi.Router) {
		sr.Post("/", app.dailyLimiter(app.CreateShortURLHandler))
		sr.Get("/{shortCode}", app.requireAuthorizedUser(app.GetShortURLHandler))
		sr.Put("/{shortCode}", app.requirePremiumUser(app.EditShortURLHandler))
		sr.Delete("/{shortCode}", app.requirePremiumUser(app.DeleteShortURLHandler))
	})

	r.Post("/api/signup", app.registerUserHandler)
	r.Post("/api/signin", app.loginUserHandler)
	r.Post("/api/signout", app.requireAuthenticatedUser(app.logoutUserHandler))
	r.Post("/api/premium", app.requireAuthenticatedUser(app.registerPremiumHandler))

	r.Get("/qr/{shortCode}", app.rateLimit(app.requirePremiumUser(app.QRCodeHandler)))

	r.Get("/{shortCode}", app.rateLimit(app.ExpandURLHandler))

	return r
}

package main

import (
	"expvar"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

// routes configures the application's routing using Chi router.
func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	// Apply middleware for logging and error recovery
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(app.metrics)
	r.Use(middleware.Logger)
	r.Use(app.recoverPanic)
	r.Use(app.authenticate)

	r.Get("/debug/vars", expvar.Handler().ServeHTTP)

	r.Get("/", app.HealthCheckHandler)

	r.Route("/api/short", func(sr chi.Router) {
		sr.Get("/", app.requireAuthenticatedUser(app.GetAllShortsHandler))
		sr.Post("/", app.dailyLimiter(app.CreateShortURLHandler))
		sr.Get("/{shortCode}", app.requireAuthorizedUser(app.GetShortURLHandler))
		sr.Put("/{shortCode}", app.requirePremiumUser(app.EditShortURLHandler))
		sr.Delete("/{shortCode}", app.requirePremiumUser(app.DeleteShortURLHandler))
	})

	r.Get("/api/stats/{shortCode}", app.requirePremiumUser(app.AnalyticsHandler))

	r.Post("/api/signup", app.registerUserHandler)
	r.Post("/api/signin", app.loginUserHandler)
	r.Post("/api/resetpassword", app.resetPasswordHandler)
	r.Post("/api/changepassword", app.requireAuthenticatedUser(app.changePassswordHandler))
	r.Post("/api/changeemail", app.requireAuthenticatedUser(app.changeEmailHandler))
	r.Post("/api/signout", app.requireAuthenticatedUser(app.logoutUserHandler))
	r.Post("/api/premium", app.requireAuthenticatedUser(app.registerPremiumHandler))

	r.Get("/qr/{shortCode}", app.rateLimit(app.requirePremiumUser(app.QRCodeHandler)))

	r.Get("/{shortCode}", app.rateLimit(app.ExpandURLHandler))

	return r
}

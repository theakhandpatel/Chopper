package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
	"url_shortner/internal/data"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
)

// health check message.
func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"message": "OK"})
}

// URL shortening requests.
func (app *application) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("URL")

	// Validate and process the provided long URL.
	if longURL == "" || !govalidator.IsURL(longURL) {
		http.Error(w, "Provide a Valid URL", http.StatusBadRequest)
		return
	}

	// if the URL already exists in the database.
	existingURL, err := app.Model.URL.GetByLongURL(longURL)
	if err != nil && err != data.ErrRecordNotFound {
		app.serverErrorResponse(w, r, err)
		return
	}

	if existingURL != nil && existingURL.Long == longURL {
		app.writeJSON(w, http.StatusOK, envelope{"url": existingURL})
		return
	}

	// Generate a new URL and handle collisions.
	url := data.NewURL(longURL)
	urlInserted := false

	for retriesLeft := app.config.MaxCollisionRetries; retriesLeft > 0; retriesLeft-- {
		err := app.Model.URL.Insert(url)
		if err == nil {
			urlInserted = true
			break
		}

		if err != data.ErrDuplicateEntry {
			app.serverErrorResponse(w, r, err)
			return
		}

		if err == data.ErrDuplicateEntry {
			url.ReShorten() //  modify the short code
		}
	}

	if !urlInserted {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"url": url})
}

// expanding short URLs.
func (app *application) ExpandURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")

	url, err := app.Model.URL.Get(shortURL)
	if err != nil {
		switch {

		case errors.Is(err, data.ErrRecordNotFound):
			http.NotFound(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// Redirect to the long URL.
	longURL := url.Long
	if longURL == "" {
		http.NotFound(w, r)
		return
	}

	// Update access count and record analytics.
	err = app.Model.URL.UpdateCount(shortURL)
	if err != nil {
		fmt.Println(err)
	}

	analyticsEntry := data.AnalyticsEntry{
		ShortURL:  shortURL,
		IP:        r.RemoteAddr,
		UserAgent: r.UserAgent(),
		Referrer:  r.Referer(),
		Timestamp: time.Now(),
	}

	err = app.Model.Analytics.Insert(&analyticsEntry)
	if err != nil {
		fmt.Println("Analytics insertion error:", err)
	}

	http.Redirect(w, r, longURL, app.config.StatusRedirectType)
}

// analytics for a given short URL.
func (app *application) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("URL")
	if shortURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortCode, err := extractShortcode(shortURL)
	if err != nil {
		http.Error(w, "URL is not valid", http.StatusBadRequest)
		return
	}

	analytics, err := app.Model.Analytics.Get(shortCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"short_url": shortURL, "analytics": analytics})
}

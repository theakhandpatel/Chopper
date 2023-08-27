package main

import (
	"errors"
	"net/http"
	"time"
	"url_shortner/internal/data"
	"url_shortner/internal/validator"

	"github.com/go-chi/chi"
)

// health check message.
func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"message": "OK"})
}

// URL shortening requests.
func (app *application) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {

	var input inputURL
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	ValidateInput(v, &input)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	var url *data.URL

	//If no custom code is required
	if input.ShortURL == "" {
		// if the URL already exists in the database.
		existingURL, err := app.Model.URL.GetByLongURL(input.LongURL)
		if err != nil && err != data.ErrRecordNotFound {
			app.serverErrorResponse(w, r, err)
			return
		}

		if existingURL != nil && existingURL.Long == input.LongURL {
			app.writeJSON(w, http.StatusOK, envelope{"url": existingURL})
			return
		}
	}

	url = data.NewURL(input.LongURL, input.ShortURL)

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
			app.NotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Redirect to the long URL.
	longURL := url.Long
	if longURL == "" {
		app.NotFoundResponse(w, r)
		return
	}

	// Update access count and record analytics.
	err = app.Model.URL.UpdateCount(shortURL)
	if err != nil {
		app.logResponse(r, err)
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
		app.logResponse(r, err)
	}

	http.Redirect(w, r, longURL, app.config.StatusRedirectType)
}

// analytics for a given short URL.
func (app *application) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("URL")
	if shortURL == "" {
		app.badRequestResponse(w, r, errors.New("url parameter is missing"))
		return
	}

	shortCode, err := extractShortcode(shortURL)
	if err != nil {
		app.NotFoundResponse(w, r)
		return
	}

	analytics, err := app.Model.Analytics.Get(shortCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"short_url": shortURL, "analytics": analytics})
}

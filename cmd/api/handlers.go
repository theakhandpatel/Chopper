package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"url_shortner/internal/data"

	"github.com/go-chi/chi"
)

func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"message": "OK"})
}

func (app *application) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("URL")
	if longURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	url := data.NewURL(longURL)
	fmt.Println("Handler:", url.Short, url.Long)
	err := app.Model.URL.Insert(url)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	fmt.Println("Handler2:", url.Short, url.Long)
	app.writeJSON(w, http.StatusCreated, envelope{"url": url})
}

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

	longURL := url.Long
	if longURL == "" {
		http.NotFound(w, r)
		return
	}

	err = app.Model.URL.UpdateCount(shortURL)
	if err != nil {
		fmt.Println(err)
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
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

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

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

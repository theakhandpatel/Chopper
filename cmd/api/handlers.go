package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"url_shortner/internal/data"

	"github.com/go-chi/chi"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("alive and kicking")
}

func (app *application) shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hit short")
	longURL := r.URL.Query().Get("URL")
	if longURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	url := data.NewURL(longURL)
	app.URL.Insert(url)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(url)
}

func (app *application) expandURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")

	longURL := app.URL.Get(shortURL).Long
	if longURL == "" {
		http.NotFound(w, r)
		return
	}

	if !strings.HasPrefix(longURL, "http://") && !strings.HasPrefix(longURL, "https://") {
		longURL = "https://" + longURL
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

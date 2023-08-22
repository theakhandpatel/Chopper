package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"url_shortner/internal/data"

	"github.com/go-chi/chi"
)

func main() {
	r := chi.NewRouter()

	r.Get("/shorten", shortenURLHandler)
	r.Get("/{shortURL}", expandURLHandler)

	http.ListenAndServe(":8080", r)
}

var DB data.URLModel

func shortenURLHandler(w http.ResponseWriter, r *http.Request) {
	longURL := r.URL.Query().Get("URL")
	if longURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	url := data.NewURL(longURL)

	DB.Insert(url)
	fmt.Println(url)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(url)
}

func expandURLHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.Path)
	shortURL := chi.URLParam(r, "shortURL")
	fmt.Println("bypass", shortURL)
	url := DB.Get(shortURL)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(url)

}

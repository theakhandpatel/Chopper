package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type envelope map[string]any

func (app *application) writeJSON(w http.ResponseWriter, statusCode int, data envelope) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(js)

	return nil
}

func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println("Error:", r.Method, r.URL.String(), "  : ", err)
	message := "the server encountered a problem and couldnot process your request"
	http.Error(w, message, http.StatusInternalServerError)
}

func extractShortcode(url string) (shortcode string, err error) {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid URL format")
	}
	shortcode = parts[3]
	return shortcode, nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// A generic map structure to package response.
type envelope map[string]interface{}

// writeJSON writes the provided data as JSON to the response writer with the specified status code.
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

// serverErrorResponse sends a JSON-encoded error response with a generic error message and status code.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Println("Error:", r.Method, r.URL.String(), "  : ", err)
	message := "the server encountered a problem and could not process your request"
	http.Error(w, message, http.StatusInternalServerError)
}

// rateLimitExceededResponse sends a JSON-encoded response indicating that the rate limit has been exceeded.
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	http.Error(w, message, http.StatusTooManyRequests)
}

// extractShortcode extracts the shortcode from the given URL.
func extractShortcode(url string) (shortcode string, err error) {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid URL format")
	}
	shortcode = parts[3]
	return shortcode, nil
}

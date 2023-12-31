package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/skip2/go-qrcode"
)

// A generic map structure to package response.
type envelope map[string]interface{}

// writes the provided data as JSON to the response writer with the specified status code.
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

// reads response data into the specified variable
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {

		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d) ", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}

func (app *application) isAuthenticated(r *http.Request) bool {
	user := app.getUserFromContext(r)
	return !user.IsAnonymous()
}

func getRedirectCode(redirectType string) int {
	if redirectType == "temporary" {
		return 307
	} else if redirectType == "permanent" {
		return 308
	}

	return 0
}

func addHTTPPrefix(url string) string {
	// Check if the URL starts with "http://" or "https://"
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		// If not, add "http://" as the prefix
		url = "http://" + url
	}
	return url
}

func getDeployedURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	hostURL := fmt.Sprintf("%s://%s/", scheme, r.Host)
	return hostURL
}

func generateAndSaveQRCode(shortCode string, imagePath string) error {

	// Generate the QR code
	code, err := qrcode.New(shortCode, qrcode.Medium)
	if err != nil {
		return err
	}

	// Create and save the image file
	file, err := os.Create(imagePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the QR code image to the file
	err = code.Write(256, file)
	if err != nil {
		return err
	}

	return nil
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"url_shortner/internal/validator"

	"github.com/asaskevich/govalidator"
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

// extractShortcode extracts the shortcode from the given URL.
func extractShortcode(url string) (shortcode string, err error) {
	parts := strings.Split(url, "/")
	if len(parts) < 4 {
		return "", fmt.Errorf("invalid URL format")
	}
	shortcode = parts[3]
	return shortcode, nil
}

type inputURL struct {
	LongURL  string `json:"long"`
	ShortURL string `json:"short"`
	Redirect string `json:"redirect"`
}

func ValidateInput(v *validator.Validator, input *inputURL) {
	v.Check(input.LongURL != "", "long", "cannot be empty")
	v.Check(govalidator.IsURL(input.LongURL), "long", "must be valid url")

	if input.ShortURL != "" {
		v.Check(len(input.ShortURL) >= 4, "short", "must be greater than 3 chars")
		v.Check(v.Matches(input.ShortURL, validator.ShortCodeRX), "short", "should containe characters from a-z,A-Z, 0-9")
	}

	if input.Redirect != "" {
		v.Check(input.Redirect == "permanent" || input.Redirect == "temporary", "redirect", "must be either 'permanent' or  'temporary'")
	}
}

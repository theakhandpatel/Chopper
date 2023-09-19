package main

import (
	"fmt"
	"net/http"
)

func (app *application) logResponse(r *http.Request, err error) {
	fmt.Println("Error:", r.Method, r.URL.String(), "  : ", err)
}

func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{"error": message}
	err := app.writeJSON(w, status, env)
	if err != nil {
		app.logResponse(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// sends a JSON-encoded error response with a generic error message and status code.
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	message := "the server encountered a problem and could not process your request"
	http.Error(w, message, http.StatusInternalServerError)
	app.logResponse(r, err)
}

func (app *application) NotFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

func (app *application) expiredLinkResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested link has expired"
	app.errorResponse(w, r, http.StatusGone, message)
}

func (app *application) createConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to create the record due to a conflict, please try again with different value"
	app.errorResponse(w, r, http.StatusConflict, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) authorizationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "You are not authorized to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

func (app *application) premiumRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "You require premium to access this feature"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

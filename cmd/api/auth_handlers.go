package main

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"url_shortner/internal/data"
	"url_shortner/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Username: input.Username,
		Email:    strings.ToLower(input.Email),
		Type:     1,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "A user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	go func() {
		data := map[string]interface{}{
			"username": user.Username,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logResponse(r, err)
		}
	}()

	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user})
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) loginUserHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Email = strings.ToLower(input.Email)
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

	token, err := app.Models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) changeEmailHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Email = strings.ToLower(input.Email)
	v := validator.New()
	data.ValidateEmail(v, input.Email)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.getUserFromContext(r)

	err = app.Models.Users.SetEmail(user.ID, input.Email)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "Updated"})
}

func (app *application) changePassswordHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NewPassword string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidatePasswordPlaintext(v, input.NewPassword)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.getUserFromContext(r)
	err = user.Password.Set(input.NewPassword)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.Models.Users.SetPassword(user.ID, string(user.Password.Hash))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.Models.Tokens.DeleteAllForUser(data.ScopeAuthentication, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "Updated"})
}

func (app *application) resetPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	input.Email = strings.ToLower(input.Email)
	v := validator.New()
	data.ValidateEmail(v, input.Email)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user, err := app.Models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.writeJSON(w, http.StatusOK, envelope{"message": "An email has been sent to your account with instructions"})
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	token, err := app.Models.Tokens.New(user.ID, 10*time.Minute, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	go func() {
		data := map[string]interface{}{
			"username": user.Username,
			"token":    token,
		}

		err = app.mailer.Send(user.Email, "reset_password.tmpl", data)
		if err != nil {
			app.logResponse(r, err)
		}
	}()

	app.writeJSON(w, http.StatusOK, envelope{"message": "An email has been sent to your account with instructions"})
}

func (app *application) logoutUserHandler(w http.ResponseWriter, r *http.Request) {
	tokenPlaintext := app.getAuthTokenPlaintextFromContext(r)

	if tokenPlaintext == "" {
		app.authenticationRequiredResponse(w, r)
	}

	err := app.Models.Tokens.DeleteOneForUser(tokenPlaintext)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"message": "logged out"})
}

func (app *application) registerPremiumHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)
	if user.Type != 2 {
		err := app.Models.Users.SetUserType(user.ID, 2)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "You are premium user"})
}

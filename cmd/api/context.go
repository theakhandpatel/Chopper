package main

import (
	"context"
	"net/http"
	"url_shortner/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")
const authTokenPlaintextContextKey = contextKey("auth_token_plaintext")
const urlContextKey = contextKey("url")

func (app *application) setUserInContext(r *http.Request, user *data.User) *http.Request {

	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

func (app *application) getUserFromContext(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("misisng user value in request context")
	}
	return user
}

func (app *application) setURLInContext(r *http.Request, url *data.URL) *http.Request {

	ctx := context.WithValue(r.Context(), urlContextKey, url)
	return r.WithContext(ctx)
}

func (app *application) getURLFromContext(r *http.Request) *data.URL {
	url, ok := r.Context().Value(urlContextKey).(*data.URL)
	if !ok {
		panic("missing url value in request context")
	}
	return url
}

func (app *application) setAuthTokenPlaintextInContext(r *http.Request, token string) *http.Request {
	ctx := context.WithValue(r.Context(), authTokenPlaintextContextKey, token)
	return r.WithContext(ctx)
}

func (app *application) getAuthTokenPlaintextFromContext(r *http.Request) string {
	token, ok := r.Context().Value(authTokenPlaintextContextKey).(string)
	if !ok {
		panic("missing token value in request context")
	}
	return token
}

package main

import (
	"errors"
	"net/http"
	"time"
	"url_shortner/internal/data"
	"url_shortner/internal/validator"

	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
)

type inputURL struct {
	LongURL  string `json:"long"`
	ShortURL string `json:"short"`
	Redirect string `json:"redirect"`
	UserID   int64  `json:"-"`
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

// health check message.
func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"message": "OK"})
}

// URL shortening requests.
func (app *application) ShortenURLHandler(w http.ResponseWriter, r *http.Request) {

	var input inputURL
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	ValidateInput(v, &input)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	user := app.getUserFromContext(r)
	input.UserID = user.ID
	if user == data.AnonymousUser {
		app.AnonymousShortenURLHandler(w, r, &input)
	} else {
		app.AuthenticatedShortenURLHandler(w, r, &input)
	}
}

func (app *application) AuthenticatedShortenURLHandler(w http.ResponseWriter, r *http.Request, input *inputURL) {

	var url *data.URL

	redirectType := http.StatusPermanentRedirect
	if input.Redirect == "temporary" {
		redirectType = http.StatusTemporaryRedirect
	}
	//If no custom code is required
	if input.ShortURL == "" {

		existingURL, err := app.Model.URLS.GetByLongURL(input.LongURL, redirectType, input.UserID)

		if err != nil && err != data.ErrRecordNotFound {
			app.serverErrorResponse(w, r, err)
			return
		}

		if existingURL != nil {
			if redirectType == existingURL.Redirect || input.Redirect == "" {
				existingURL.Modified = time.Now()
				app.Model.URLS.Update(existingURL)
				app.writeJSON(w, http.StatusOK, envelope{"url": existingURL})
				return
			}
		}
	}

	maxTriesForInsertion := 3
	if input.ShortURL != "" {
		maxTriesForInsertion = 1
	}

	url = data.NewURL(input.LongURL, input.ShortURL, redirectType, input.UserID)

	urlInserted := false

	for retriesLeft := maxTriesForInsertion; retriesLeft > 0; retriesLeft-- {
		err := app.Model.URLS.Insert(url)
		if err == nil {
			urlInserted = true
			break
		}

		if err != data.ErrDuplicateEntry {
			app.serverErrorResponse(w, r, err)
			return
		}

		if err == data.ErrDuplicateEntry {
			url.ReShorten() //  modify the short code
		}
	}

	if !urlInserted {
		app.serverErrorResponse(w, r, data.ErrMaxCollision)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"url": url})
}

func (app *application) AnonymousShortenURLHandler(w http.ResponseWriter, r *http.Request, input *inputURL) {
	var url *data.URL

	// if the URL already exists in the database.
	existingURL, err := app.Model.URLS.GetByLongURL(input.LongURL, http.StatusPermanentRedirect, input.UserID)
	if err != nil && err != data.ErrRecordNotFound {
		app.serverErrorResponse(w, r, err)
		return
	}

	if existingURL != nil {
		existingURL.Modified = time.Now()
		app.Model.URLS.Update(existingURL)
		app.writeJSON(w, http.StatusOK, envelope{"url": existingURL})
		return
	}

	maxTriesForInsertion := 3
	url = data.NewURL(input.LongURL, "", http.StatusPermanentRedirect, input.UserID)

	urlInserted := false

	for retriesLeft := maxTriesForInsertion; retriesLeft > 0; retriesLeft-- {
		err := app.Model.URLS.Insert(url)
		if err == nil {
			urlInserted = true
			break
		}

		if err != data.ErrDuplicateEntry {
			app.serverErrorResponse(w, r, err)
			return
		}

		if err == data.ErrDuplicateEntry {
			url.ReShorten() //  modify the short code
		}
	}

	if !urlInserted {
		app.serverErrorResponse(w, r, data.ErrMaxCollision)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"url": url})
}

// expanding short URLs.
func (app *application) ExpandURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortURL")

	url, err := app.Model.URLS.Get(shortURL)
	if err != nil {
		switch {

		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	longURL := url.Long
	if longURL == "" {
		app.NotFoundResponse(w, r)
		return
	}

	currentTime := time.Now()
	expiryTime := url.Modified.Add(6 * time.Hour)
	if expiryTime.Before(currentTime) {
		app.expiredLinkResponse(w, r)
		return
	}

	if url.UserID != data.AnonymousUser.ID {
		analyticsEntry := data.AnalyticsEntry{
			ShortURL:  shortURL,
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Referrer:  r.Referer(),
			Timestamp: time.Now(),
			UserID:    url.UserID,
		}

		err = app.Model.Analytics.Insert(&analyticsEntry)
		if err != nil {
			app.logResponse(r, err)
		}
	}

	http.Redirect(w, r, longURL, url.Redirect)
}

// analytics for a given short URL.
func (app *application) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Query().Get("URL")
	if shortURL == "" {
		app.badRequestResponse(w, r, errors.New("url parameter is missing"))
		return
	}

	shortCode, err := extractShortcode(shortURL)
	if err != nil {
		app.NotFoundResponse(w, r)
		return
	}
	user := app.getUserFromContext(r)
	analytics, err := app.Model.Analytics.GetAll(shortCode, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"short_url": shortURL, "analytics": analytics})
}

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

// health check message.
func (app *application) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, envelope{"message": "OK"})
}

// URL shortening requests.
func (app *application) CreateShortURLHandler(w http.ResponseWriter, r *http.Request) {

	var input inputURL
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.getUserFromContext(r)

	v := validator.New()
	v.Check(input.LongURL != "", "long", "cannot be empty")
	v.Check(govalidator.IsURL(input.LongURL), "long", "must be valid url")
	if input.ShortURL != "" {
		if user.IsPremium() {
			v.Check(len(input.ShortURL) >= 4, "short", "must be greater than or equal to 4 chars")
		} else {
			v.Check(len(input.ShortURL) >= 6, "short", "must be greater than or equal to 6  chars")
		}
		v.Check(v.Matches(input.ShortURL, validator.ShortCodeRX), "short", "should containe characters from a-z,A-Z, 0-9")
	}
	if input.Redirect != "" {
		v.Check(input.Redirect == "permanent" || input.Redirect == "temporary", "redirect", "must be either 'permanent' or  'temporary'")
	}
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	input.LongURL = addHTTPPrefix(input.LongURL)
	input.UserID = user.ID
	if user.IsAnonymous() {
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
			url.Reshorten() //  modify the short code
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
			url.Reshorten() //  modify the short code
		}
	}

	if !urlInserted {
		app.serverErrorResponse(w, r, data.ErrMaxCollision)
		return
	}

	app.writeJSON(w, http.StatusCreated, envelope{"url": url})
}

func (app *application) EditShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url := app.getURLFromContext(r)
	var input inputURL
	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if input.LongURL != "" {
		v.Check(govalidator.IsURL(input.LongURL), "long", "must be valid url")
	}
	if input.ShortURL != "" {
		v.Check(len(input.ShortURL) >= 4, "short", "must be greater than 3 chars")
		v.Check(v.Matches(input.ShortURL, validator.ShortCodeRX), "short", "should containe characters from a-z,A-Z, 0-9")
	}
	if input.Redirect != "" {
		v.Check(input.Redirect == "permanent" || input.Redirect == "temporary", "redirect", "must be either 'permanent' or  'temporary'")
	}
	v.Check(input.LongURL != "" || input.ShortURL != "" || input.Redirect != "", "all", "Need Updated Data")

	updateNeeded := true
	if input.LongURL != "" && input.LongURL != url.LongForm {
		url.LongForm = input.LongURL
		updateNeeded = true
	}
	if input.ShortURL != "" && input.ShortURL != url.ShortCode {
		url.ShortCode = input.ShortURL
		updateNeeded = true
	}
	if input.Redirect != "" && getRedirectCode(input.Redirect) != url.Redirect {
		url.Redirect = getRedirectCode(input.Redirect)
		updateNeeded = true
	}
	v.Check(updateNeeded, "all", "Nothing to Update")

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	url.Modified = time.Now()

	err = app.Model.URLS.Update(url)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusAccepted, envelope{"url": url})
}

func (app *application) DeleteShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url := app.getURLFromContext(r)

	err := app.Model.URLS.DeleteByShort(url.ShortCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusNoContent, envelope{})
	err = app.Model.Analytics.DeleteByURLID(url.ID)
	if err != nil {
		app.logResponse(r, err)
	}
}

func (app *application) GetShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url := app.getURLFromContext(r)
	app.writeJSON(w, http.StatusOK, envelope{"url": url})
}

// expanding short URLs.
func (app *application) ExpandURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := chi.URLParam(r, "shortCode")

	url, err := app.Model.URLS.GetByShort(shortURL)
	if err != nil {
		switch {

		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	longURL := url.LongForm
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
			IP:        r.RemoteAddr,
			UserAgent: r.UserAgent(),
			Referrer:  r.Referer(),
			Timestamp: time.Now(),
			URLID:     url.ID,
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

	url := app.getURLFromContext(r)
	analytics, err := app.Model.Analytics.GetByURLID(url.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"short_url": url.ShortCode, "analytics": analytics})
}

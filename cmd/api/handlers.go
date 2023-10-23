package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	New      bool   `json:"new"`
	Once     bool   `json:"once"` // Once is used to specify that this short url will be deleted as soon as it is used once.
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
	user := app.getUserFromContext(r)

	if !user.IsPremium() {
		input.New = false
		input.Once = false
	}

	if input.Once {
		input.New = true
	}

	redirectType := http.StatusPermanentRedirect
	if input.Redirect == "temporary" {
		redirectType = http.StatusTemporaryRedirect
	}
	//If no custom code is required
	if input.ShortURL == "" && !input.New {
		existingURL, err := app.Models.URLS.GetByLongURL(input.LongURL, redirectType, input.UserID)
		if err != nil && err != data.ErrRecordNotFound {
			app.serverErrorResponse(w, r, err)
			return
		}
		if existingURL != nil {
			existingURL.Expired = time.Now().Add(24 * time.Hour)
			app.Models.URLS.Update(existingURL)
			app.writeJSON(w, http.StatusOK, envelope{"url": existingURL, "short_url": (getDeployedURL(r) + existingURL.ShortCode)})
			return
		}
	}

	url = data.NewURL(input.LongURL, input.ShortURL, redirectType, user, input.Once)

	err := app.Models.URLS.Insert(url)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEntry):
			app.createConflictResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	hostURL := getDeployedURL(r)
	app.writeJSON(w, http.StatusCreated, envelope{"url": url, "short_url": (hostURL + url.ShortCode)})
}

func (app *application) AnonymousShortenURLHandler(w http.ResponseWriter, r *http.Request, input *inputURL) {
	var url *data.URL
	user := app.getUserFromContext(r)

	// if the URL already exists in the database.
	existingURL, err := app.Models.URLS.GetByLongURL(input.LongURL, http.StatusPermanentRedirect, input.UserID)
	if err != nil && err != data.ErrRecordNotFound {
		app.serverErrorResponse(w, r, err)
		return
	}
	expiryTime := time.Now().Add(6 * time.Hour)
	if existingURL != nil {
		existingURL.Expired = expiryTime
		app.Models.URLS.Update(existingURL)
		app.writeJSON(w, http.StatusOK, envelope{"url": existingURL, "short_url": (getDeployedURL(r) + existingURL.ShortCode)})
		return
	}

	url = data.NewURL(input.LongURL, "", http.StatusPermanentRedirect, user, false)

	err = app.Models.URLS.Insert(url)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEntry):
			app.createConflictResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	hostURL := getDeployedURL(r)
	app.writeJSON(w, http.StatusCreated, envelope{"url": url, "short_url": (hostURL + url.ShortCode)})
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

	url.Expired = time.Now()

	err = app.Models.URLS.Update(url)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	hostURL := getDeployedURL(r)
	app.writeJSON(w, http.StatusAccepted, envelope{"url": url, "short_url": (hostURL + url.ShortCode)})
}

func (app *application) DeleteShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url := app.getURLFromContext(r)

	err := app.Models.URLS.DeleteByShort(url.ShortCode)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusNoContent, envelope{})
	err = app.Models.Analytics.DeleteByURLID(url.ID)
	if err != nil {
		app.logResponse(r, err)
	}
}

func (app *application) GetShortURLHandler(w http.ResponseWriter, r *http.Request) {
	url := app.getURLFromContext(r)
	hostURL := getDeployedURL(r)
	app.writeJSON(w, http.StatusOK, envelope{"url": url, "short_url": (hostURL + url.ShortCode)})
}

func (app *application) GetAllShortsHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)

	urls, err := app.Models.URLS.GetAllForUser(user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	app.writeJSON(w, http.StatusOK, envelope{"urls": urls})
}

// expanding short URLs.
func (app *application) ExpandURLHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")

	url, err := app.Models.URLS.GetByShort(shortCode)
	if err != nil {
		switch {

		case errors.Is(err, data.ErrRecordNotFound):
			app.NotFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if url.Expired.Before(time.Now()) {
		app.expiredLinkResponse(w, r)
		app.Models.URLS.DeleteByShort(shortCode)
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

		err = app.Models.Analytics.Insert(&analyticsEntry)
		if err != nil {
			app.logResponse(r, err)
		}
	}

	if url.Once {
		url.Expired = time.Now()
		err = app.Models.URLS.Update(url)
		if err != nil {
			app.logResponse(r, err)
		}
	}

	http.Redirect(w, r, url.LongForm, url.Redirect)

}

// analytics for a given short URL.
func (app *application) AnalyticsHandler(w http.ResponseWriter, r *http.Request) {

	url := app.getURLFromContext(r)
	analytics, err := app.Models.Analytics.GetByURLID(url.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	hostURL := getDeployedURL(r)
	app.writeJSON(w, http.StatusOK, envelope{"short_url": (hostURL + url.ShortCode), "analytics": analytics})
}

func (app *application) QRCodeHandler(w http.ResponseWriter, r *http.Request) {
	shortCode := chi.URLParam(r, "shortCode")
	imagePath := filepath.Join("./qrcodes", shortCode+".png")
	_, err := os.Stat(imagePath)

	if os.IsNotExist(err) {
		// Generate and save the QR code image
		err := generateAndSaveQRCode(getDeployedURL(r)+shortCode, imagePath)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	w.Header().Set("Content-Type", "image/png")

	file, err := os.Open(imagePath)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

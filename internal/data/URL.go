package data

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
	"url_shortner/internal/utils"

	"github.com/mattn/go-sqlite3"
)

// ErrRecordNotFound is returned when a record is not found in the database.
var ErrRecordNotFound = errors.New("record not found")

// ErrMaxCollision is returned when the collision limit is exceeded during URL shortening.
var ErrMaxCollision = errors.New("collision limit exceeded")

// ErrDuplicateEntry is returned when a duplicate entry already exists in the database.
var ErrDuplicateEntry = errors.New("entry already exists")

// URL represents a shortened URL record.
type URL struct {
	ID       int64
	Long     string
	Short    string
	Accessed int64 `json:"-"`
	Redirect int
	UserID   int64     `json:"-"`
	Created  time.Time `json:"-"`
	Modified time.Time
}

// NewURL creates a new URL instance with a shortened version of the provided long URL.
func NewURL(longURL string, shortURL string, redirect int, userID int64) *URL {
	if shortURL == "" {
		if userID == AnonymousUser.ID {
			shortURL = utils.GetShortCode(8)
		} else {
			shortURL = utils.GetShortCode(6)
		}
	}
	if redirect != http.StatusTemporaryRedirect {
		redirect = http.StatusPermanentRedirect
	}
	return &URL{
		Long:     longURL,
		Short:    shortURL,
		Accessed: 0,
		Redirect: redirect,
		Created:  time.Now(),
		Modified: time.Now(),
		UserID:   int64(userID),
	}
}

// Reshorten generates a new short URL by appending a timestamp to the long URL and shortening it again.
func (url *URL) Reshorten() {
	if url.UserID == AnonymousUser.ID {
		url.Short = utils.GetShortCode(8)
	} else {
		url.Short = utils.GetShortCode(6)
	}
}

// URLModel represents the database model for URL operations.
type URLModel struct {
	DB *sql.DB
}

// Insert inserts a new URL record into the database.
func (model *URLModel) Insert(url *URL) error {
	query := `
		INSERT INTO urls (long_url, short_url, accessed, redirect, user_id, created, modified) VALUES (?, ?, ?,?,?,?,?);
	`

	res, err := model.DB.Exec(query, url.Long, url.Short, url.Accessed, url.Redirect, url.UserID, url.Created, url.Modified)

	if err != nil {
		// Check for duplicate entry error and return a predefined error.
		sqliteErr, isSQLError := err.(sqlite3.Error)
		if isSQLError && sqliteErr.Code == sqlite3.ErrConstraint {
			return ErrDuplicateEntry
		}
		// Use errors.Is to check for specific error type.
		if errors.Is(err, sqlite3.ErrConstraint) {
			return ErrDuplicateEntry
		}
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	url.ID = id

	return nil
}

// GetByShort retrieves a URL record based on the short URL.
func (model *URLModel) GetByShort(shortURL string) (*URL, error) {
	query := `
		SELECT id, long_url, accessed, redirect, user_id, created, modified FROM urls WHERE short_url = ?;
	`
	row := model.DB.QueryRow(query, shortURL)

	url := &URL{
		Short: shortURL,
	}
	err := row.Scan(&url.ID, &url.Long, &url.Accessed, &url.Redirect, &url.UserID, &url.Created, &url.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return url, nil
}

func (model *URLModel) DeleteByShort(shortURL string) error {
	query := `
			DELETE FROM urls WHERE short_url = ?;
	`
	_, err := model.DB.Exec(query, shortURL)
	if err != nil {
		return err
	}

	return nil
}

// Update modifies an existing URL record in the database.
func (model *URLModel) Update(url *URL) error {
	query := `
		UPDATE urls
		SET long_url = ?, short_url = ?, accessed = ?, redirect = ?, user_id = ?, modified = ?
		WHERE id = ?;
	`

	_, err := model.DB.Exec(query, url.Long, url.Short, url.Accessed, url.Redirect, url.UserID, url.Modified, url.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetByLongURL retrieves a URL record based on the long URL.
func (model *URLModel) GetByLongURL(longURL string, redirectType int, userID int64) (*URL, error) {
	query := `
		SELECT id, long_url, short_url, accessed, redirect, created, modified FROM urls WHERE long_url = ? AND redirect=? AND user_id = ?;
	`
	row := model.DB.QueryRow(query, longURL, redirectType, userID)

	url := &URL{
		UserID: userID,
	}
	err := row.Scan(&url.ID, &url.Long, &url.Short, &url.Accessed, &url.Redirect, &url.Created, &url.Modified)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return url, nil
}

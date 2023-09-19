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
	ID        int64 `json:"-"`
	LongForm  string
	ShortCode string
	Redirect  int
	UserID    int64     `json:"-"`
	Created   time.Time `json:"-"`
	Expired   time.Time
}

// NewURL creates a new URL instance with a shortened version of the provided long URL.
func NewURL(longURL string, shortCode string, redirect int, user *User) *URL {
	if shortCode == "" {
		if user.IsAnonymous() {
			shortCode = utils.GetShortCode(8)
		} else {
			shortCode = utils.GetShortCode(6)
		}
	}
	if redirect != http.StatusTemporaryRedirect {
		redirect = http.StatusPermanentRedirect
	}
	var expiry time.Time
	if user.IsPremium() {
		expiry = time.Now().Add(24 * 30 * time.Hour)
	} else if !user.IsAnonymous() {
		expiry = time.Now().Add(24 * 7 * time.Hour)
	} else {
		expiry = time.Now().Add(12 * time.Hour)
	}

	return &URL{
		LongForm:  longURL,
		ShortCode: shortCode,
		Redirect:  redirect,
		Created:   time.Now(),
		Expired:   expiry,
		UserID:    user.ID,
	}
}

// Reshorten generates a new short URL by appending a timestamp to the long URL and shortening it again.
func (url *URL) Reshorten() {
	if url.UserID == AnonymousUser.ID {
		url.ShortCode = utils.GetShortCode(8)
	} else {
		url.ShortCode = utils.GetShortCode(6)
	}
}

// URLModel represents the database model for URL operations.
type URLModel struct {
	DB *sql.DB
}

// Insert inserts a new URL record into the database.
func (model *URLModel) Insert(url *URL) error {
	query := `
		INSERT INTO urls (long_url, short_url, redirect, user_id, created, expired) VALUES (?,?,?,?,?,?);
	`

	res, err := model.DB.Exec(query, url.LongForm, url.ShortCode, url.Redirect, url.UserID, url.Created, url.Expired)

	if err != nil {
		// Check for duplicate entry error
		sqliteErr, isSQLError := err.(sqlite3.Error)
		if isSQLError && sqliteErr.Code == sqlite3.ErrConstraint {
			return ErrDuplicateEntry
		}
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
func (model *URLModel) GetByShort(shortCode string) (*URL, error) {
	query := `
		SELECT id, long_url,  redirect, user_id, created, expired FROM urls WHERE short_url = ?;
	`
	row := model.DB.QueryRow(query, shortCode)

	url := &URL{
		ShortCode: shortCode,
	}
	err := row.Scan(&url.ID, &url.LongForm, &url.Redirect, &url.UserID, &url.Created, &url.Expired)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return url, nil
}

func (model *URLModel) DeleteByShort(shortCode string) error {
	tx, err := model.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // Rollback the transaction

	query := `
		DELETE FROM urls WHERE short_url = ?;
	`
	_, err = tx.Exec(query, shortCode)
	if err != nil {
		return err
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// Update modifies an existing URL record in the database.
func (model *URLModel) Update(url *URL) error {
	query := `
		UPDATE urls
		SET long_url = ?, short_url = ?, redirect = ?, user_id = ?, expired = ?
		WHERE id = ?;
	`

	_, err := model.DB.Exec(query, url.LongForm, url.ShortCode, url.Redirect, url.UserID, url.Expired, url.ID)
	if err != nil {
		return err
	}

	return nil
}

// GetByLongURL retrieves a URL record based on the long URL.
func (model *URLModel) GetByLongURL(longURL string, redirectType int, userID int64) (*URL, error) {
	query := `
		SELECT id, long_url, short_url, redirect, created, expired FROM urls WHERE long_url = ? AND redirect=? AND user_id = ?;
	`
	row := model.DB.QueryRow(query, longURL, redirectType, userID)

	url := &URL{
		UserID: userID,
	}
	err := row.Scan(&url.ID, &url.LongForm, &url.ShortCode, &url.Redirect, &url.Created, &url.Expired)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return url, nil
}

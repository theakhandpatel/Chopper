package data

import (
	"database/sql"
	"errors"
	"strconv"
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
	Long     string
	Short    string
	Accessed int64
}

// NewURL creates a new URL instance with a shortened version of the provided long URL.
func NewURL(longURL string) *URL {
	shortURL := utils.Shorten(longURL)
	return &URL{
		Long:     longURL,
		Short:    shortURL,
		Accessed: 0,
	}
}

// ReShorten generates a new short URL by appending a timestamp to the long URL and shortening it again.
func (url *URL) ReShorten() {
	timestamp := time.Now().UnixNano()
	url.Short = utils.Shorten(url.Long + strconv.FormatInt(timestamp, 10))
}

// URLModel represents the database model for URL operations.
type URLModel struct {
	DB                  *sql.DB
	MaxCollisionRetries int64
}

// Insert inserts a new URL record into the database.
func (model *URLModel) Insert(url *URL) error {
	query := `
		INSERT INTO urls (long_url, short_url, accessed) VALUES (?, ?, ?);
	`

	_, err := model.DB.Exec(query, url.Long, url.Short, url.Accessed)

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

	return nil
}

// Get retrieves a URL record based on the short URL.
func (model *URLModel) Get(shortURL string) (*URL, error) {
	query := `
		SELECT long_url, short_url, accessed FROM urls WHERE short_url = ?;
	`
	row := model.DB.QueryRow(query, shortURL)

	url := &URL{}
	err := row.Scan(&url.Long, &url.Short, &url.Accessed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return url, nil
}

// UpdateCount updates the access count of a URL record.
func (model *URLModel) UpdateCount(shortURL string) error {
	query := `
		UPDATE urls SET accessed = accessed + 1 WHERE short_url = ?;
	`
	_, err := model.DB.Exec(query, shortURL)

	return err
}

// GetByLongURL retrieves a URL record based on the long URL.
func (model *URLModel) GetByLongURL(longURL string) (*URL, error) {
	query := `
		SELECT long_url, short_url, accessed FROM urls WHERE long_url = ?;
	`
	row := model.DB.QueryRow(query, longURL)

	url := &URL{}
	err := row.Scan(&url.Long, &url.Short, &url.Accessed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return url, nil
}

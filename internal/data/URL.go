package data

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"
	"url_shortner/internal/utils"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrMaxCollision   = errors.New("collision limit excedded")
)

type URL struct {
	Long     string
	Short    string
	Accessed int64
}

func NewURL(longURL string) *URL {
	shortURL := utils.Shorten(longURL)
	return &URL{
		Long:     longURL,
		Short:    shortURL,
		Accessed: 0,
	}
}

type URLModel struct {
	DB         *sql.DB
	MaxRetries int
}

func (model *URLModel) Insert(url *URL) error {
	query := `
					INSERT INTO urls (long_url, short_url, accessed) VALUES (?, ?, ?);
			`

	for retriesLeft := model.MaxRetries; retriesLeft > 0; retriesLeft-- {

		_, err := model.DB.Exec(query, url.Long, url.Short, url.Accessed)

		// Successful insertion
		if err == nil {
			retriesLeft = 0
			fmt.Println(url.Short, url.Long)
			return nil
		}

		fmt.Println("Insert:", retriesLeft, url.Short, url.Long)

		sqliteErr, isSQLError := err.(sqlite3.Error)
		if isSQLError && sqliteErr.Code == sqlite3.ErrConstraint {

			existingURL, _ := model.Get(url.Short)

			if existingURL.Long == url.Long {
				return nil
			} else {
				// Handle duplicate entry by modifying the short code
				timestamp := time.Now().UnixNano() // Using timestamp as a unique identifier
				url.Short = utils.Shorten(url.Long + strconv.FormatInt(timestamp, 10))
			}

		} else {
			return err
		}

	}

	return ErrMaxCollision
}

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

func (model *URLModel) UpdateCount(shortURL string) error {
	query := `
		UPDATE urls SET accessed = accessed + 1 WHERE short_url = ?;
	`
	_, err := model.DB.Exec(query, shortURL)

	return err
}

package data

import (
	"database/sql"
	"errors"
	"url_shortner/internal/utils"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrDuplicateEntry = errors.New("Duplicate Entry")
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
	DB *sql.DB
}

func (model *URLModel) Insert(url *URL) error {
	query := `
		INSERT INTO urls (long_url, short_url, accessed) VALUES (?, ?, ?);
	`
	_, err := model.DB.Exec(query, url.Long, url.Short, url.Accessed)

	if err != nil {
		sqliteErr, isSQLError := err.(sqlite3.Error)
		if isSQLError && sqliteErr.Code == sqlite3.ErrConstraint {

			return ErrDuplicateEntry
		}
		return err
	}

	return nil
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
			return nil, nil
		}
		return nil, err
	}

	_, err = model.DB.Exec(`
		UPDATE urls SET accessed = accessed + 1 WHERE short_url = ?;
	`, shortURL)

	if err != nil {
		return nil, err
	}

	return url, nil
}

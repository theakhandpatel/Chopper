package main

import (
	"database/sql"
	"net/http"
	"url_shortner/internal/data"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	Model data.Model
}

func main() {

	db, err := openDB("./database.db")
	if err != nil {
		panic(err)
	}

	app := application{
		Model: data.NewModels(db),
	}
	http.ListenAndServe(":8080", app.routes())
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		long_url TEXT NOT NULL,
		short_url TEXT NOT NULL,
		accessed INTEGER DEFAULT 0,
		UNIQUE(short_url)
	);
	`)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_long_url ON urls(long_url);`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS analytics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			short_url TEXT NOT NULL,
			ip TEXT NOT NULL,
			user_agent TEXT NOT NULL,
			referrer TEXT,
			timestamp DATETIME NOT NULL
	);
	`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

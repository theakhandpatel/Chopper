package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"url_shortner/internal/data"

	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	Model              data.Model
	StatusRedirectType int
	Port               int64
}

func main() {

	port := *flag.Int64("port", 8080, "Port number")
	maxCollisionRetries := *flag.Int64("maxRetry", 5, "Maximum Collision Retries")
	enableTemporaryRedirect := *flag.Bool("temp-redirect", true, "temporary redirect enabled")
	dsn := *flag.String("dsn", "./database.db", "path to the SQLite file")
	flag.Parse()

	db, err := openDB(dsn)
	if err != nil {
		panic(err)
	}
	statusRedirectType := http.StatusPermanentRedirect
	if enableTemporaryRedirect {
		statusRedirectType = http.StatusTemporaryRedirect
	}

	app := application{
		Model:              data.NewModels(db, maxCollisionRetries),
		StatusRedirectType: statusRedirectType,
		Port:               port,
	}

	fmt.Println("Intializing server at :", app.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", app.Port), app.routes())
	panic(err)
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

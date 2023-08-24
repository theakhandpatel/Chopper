package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"url_shortner/internal/data"

	_ "github.com/mattn/go-sqlite3"
)

type config struct {
	Port    int64
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	StatusRedirectType  int
	MaxCollisionRetries int64
	dsn                 string
}

type application struct {
	Model  data.Model
	config config
}

// TODO: Implemnet Server Metrics
func main() {
	var cfg config

	flag.Int64Var(&cfg.Port, "port", 8080, "Port number")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.Int64Var(&cfg.MaxCollisionRetries, "maxRetry", 5, "")
	flag.StringVar(&cfg.dsn, "dsn", "./database.db", "path to the SQLite file")
	enableTemporaryRedirect := *flag.Bool("temp-redirect", false, "temporary redirect enabled")

	flag.Parse()

	cfg.StatusRedirectType = http.StatusPermanentRedirect
	if enableTemporaryRedirect {
		cfg.StatusRedirectType = http.StatusTemporaryRedirect
	}

	db, err := openDB(cfg.dsn)
	if err != nil {
		panic(err)
	}

	app := application{
		Model:  data.NewModels(db, cfg.MaxCollisionRetries),
		config: cfg,
	}

	fmt.Println("Intializing server at :", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), app.routes())
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

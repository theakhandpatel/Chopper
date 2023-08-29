package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"url_shortner/internal/data"

	_ "github.com/mattn/go-sqlite3"
)

// config represents the configuration parameters for the application.
type config struct {
	Port    int64 // Port number
	limiter struct {
		rps     float64 // Rate limiter maximum requests per second
		burst   int     // Rate limiter maximum burst
		enabled bool    // Enable rate limiter
	}
	dsn string // Path to the SQLite database file
}

// application represents the main application structure.
type application struct {
	Model  data.Model // Data model for the application
	config config     // Application configuration
}

func main() {
	var cfg config

	// Parsing command line flags
	flag.Int64Var(&cfg.Port, "port", 8080, "Port number")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", false, "Enable rate limiter")
	flag.StringVar(&cfg.dsn, "dsn", "./database.db", "Path to the SQLite file")

	flag.Parse()

	// Initializing the database
	db, err := openDB(cfg.dsn)
	if err != nil {
		panic(err)
	}

	app := application{
		Model:  data.NewModel(db),
		config: cfg,
	}

	// Starting the server
	fmt.Println("Initializing server at port:", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%v", cfg.Port), app.routes())
	panic(err)
}

// openDB  initializes the SQLite database.
func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Creating the 'urls' table for storing URL records
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

	// Creating indexes for efficient queries
	_, err = db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);`)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_long_url ON urls(long_url);`)
	if err != nil {
		return nil, err
	}

	// Creating the 'analytics' table for storing URL analytics data
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

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash BLOB NOT NULL,
    activated INTEGER NOT NULL,
    version INTEGER NOT NULL DEFAULT 1
	);
	`)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tokens (
			hash BLOB PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
			expiry TIMESTAMP NOT NULL,
			scope TEXT NOT NULL
	);
	`)

	if err != nil {
		return nil, err
	}

	return db, nil
}

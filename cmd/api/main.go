package main

import (
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"url_shortner/internal/data"

	"github.com/golang-migrate/migrate/v4"
	sqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// config represents the configuration parameters for the application.
type config struct {
	Port        int64 // Port number
	rateLimiter struct {
		rps     float64 // Rate limiter maximum requests per second
		burst   int     // Rate limiter maximum burst
		enabled bool    // Enable rate limiter
	}
	dailyLimiter struct {
		anonymous     float64 // anonymous
		authenticated float64 // loggedin users
		enabled       bool
	}
	database struct {
		dsn            string // Path to the SQLite database file
		migrationsPath string
	}
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
	flag.Float64Var(&cfg.rateLimiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.rateLimiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.rateLimiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.StringVar(&cfg.database.dsn, "dsn", "./database.db", "Path to the SQLite file")
	flag.BoolVar(&cfg.dailyLimiter.enabled, "dailylimiter-enabled", true, "Enable daily limiter")
	flag.Float64Var(&cfg.dailyLimiter.anonymous, "dailyLimiter-ip", 3.0, "Daily limit for Anonymous Users(By IP)")
	flag.Float64Var(&cfg.dailyLimiter.authenticated, "dailyLimiter-id", 10.0, "Daily limit for Authenticated Users")
	flag.StringVar(&cfg.database.migrationsPath, "migrations", "./migrations", "Relative Path to the migrations folder")

	flag.Parse()

	// Initializing the database
	db, err := openDB(cfg.database.dsn, cfg.database.migrationsPath)
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
func openDB(dsn string, migrationsPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath, // Path to your migration files
		"sqlite3",                // Database driver name
		driver,
	)
	if err != nil {
		panic(err)
	}

	// Run all "up" migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(err)
	}
	fmt.Println("Migrations are up....")
	return db, nil
}

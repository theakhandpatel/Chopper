package data

import "database/sql"

type Model struct {
	URL       URLModel
	Analytics AnalyticsModel
}

func NewModels(db *sql.DB, maxCollisionRetries int64) Model {
	return Model{
		URL:       URLModel{DB: db, MaxRetries: maxCollisionRetries},
		Analytics: AnalyticsModel{DB: db},
	}
}

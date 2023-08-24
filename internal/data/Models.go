package data

import "database/sql"

type Model struct {
	URL       URLModel
	Analytics AnalyticsModel
}

func NewModels(db *sql.DB) Model {
	return Model{
		URL:       URLModel{DB: db, MaxRetries: 5},
		Analytics: AnalyticsModel{DB: db},
	}
}

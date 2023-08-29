package data

import "database/sql"

type Model struct {
	URLS      URLModel
	Analytics AnalyticsModel
	Tokens    TokenModel
	Users     UserModel
}

func NewModel(db *sql.DB) Model {
	return Model{
		URLS:      URLModel{DB: db},
		Analytics: AnalyticsModel{DB: db},
		Tokens:    TokenModel{DB: db},
		Users:     UserModel{DB: db},
	}
}

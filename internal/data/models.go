package data

import "database/sql"

type Model struct {
	URLS      URLModel
	Analytics AnalyticsModel
	Tokens    TokenModel
	Users     UserModel
}

func NewModel(db *sql.DB, maxCollisionRetries int64) Model {
	return Model{
		URLS:      URLModel{DB: db, MaxCollisionRetries: maxCollisionRetries},
		Analytics: AnalyticsModel{DB: db},
		Tokens:    TokenModel{DB: db},
		Users:     UserModel{DB: db},
	}
}

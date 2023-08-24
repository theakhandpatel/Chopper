package data

import (
	"database/sql"
	"time"
)

type AnalyticsEntry struct {
	ID        int64
	ShortURL  string
	IP        string
	UserAgent string
	Referrer  string
	Timestamp time.Time
}

type AnalyticsModel struct {
	DB *sql.DB
}

func (model *AnalyticsModel) Insert(entry *AnalyticsEntry) error {
	query := `
			INSERT INTO analytics (short_url, ip, user_agent, referrer, timestamp)
			VALUES (?, ?, ?, ?, ?);
	`
	_, err := model.DB.Exec(query, entry.ShortURL, entry.IP, entry.UserAgent, entry.Referrer, entry.Timestamp)
	return err
}

func (model *AnalyticsModel) Get(shortURL string) ([]*AnalyticsEntry, error) {
	query := `
			SELECT id, short_url, ip, user_agent, referrer, timestamp
			FROM analytics
			WHERE short_url = ?;
	`
	rows, err := model.DB.Query(query, shortURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []*AnalyticsEntry

	for rows.Next() {
		var entry AnalyticsEntry
		err := rows.Scan(&entry.ID, &entry.ShortURL, &entry.IP, &entry.UserAgent, &entry.Referrer, &entry.Timestamp)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, &entry)
	}

	return analytics, nil
}

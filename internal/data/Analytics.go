package data

import (
	"database/sql"
	"time"
)

// AnalyticsEntry represents a single entry of analytics data.
type AnalyticsEntry struct {
	ID        int64     `json:"omit"`
	ShortURL  string    `json:"omit"`
	IP        string    `json:"ip_address"`
	UserAgent string    `json:"user-agent"`
	Referrer  string    `json:"referrer"`
	Timestamp time.Time `json:"accessed_at"`
}

// AnalyticsModel provides methods to interact with the analytics data in the database.
type AnalyticsModel struct {
	DB *sql.DB
}

// Insert adds a new analytics entry into the database.
func (model *AnalyticsModel) Insert(entry *AnalyticsEntry) error {
	query := `
			INSERT INTO analytics (short_url, ip, user_agent, referrer, timestamp)
			VALUES (?, ?, ?, ?, ?);
	`
	_, err := model.DB.Exec(query, entry.ShortURL, entry.IP, entry.UserAgent, entry.Referrer, entry.Timestamp)
	return err
}

// Get retrieves analytics entries for a specific short URL from the database.
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

package data

import (
	"database/sql"
	"time"
)

// AnalyticsEntry represents a single entry of analytics data.
type AnalyticsEntry struct {
	ID        int64     `json:"-"`
	URLID     int64     `json:"url-id"`
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
			INSERT INTO analytics (url_id, ip, user_agent, referrer, timestamp)
			VALUES (?, ?, ?, ?, ?);
	`
	_, err := model.DB.Exec(query, entry.URLID, entry.IP, entry.UserAgent, entry.Referrer, entry.Timestamp)
	return err
}

// Get retrieves analytics entries for a specific short URL from the database.
func (model *AnalyticsModel) GetByURLID(urlID int64) ([]*AnalyticsEntry, error) {
	query := `
			SELECT id, url_id,ip, user_agent, referrer, timestamp
			FROM analytics
			WHERE url_id = ?;
	`
	rows, err := model.DB.Query(query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []*AnalyticsEntry

	for rows.Next() {
		var entry AnalyticsEntry
		err := rows.Scan(&entry.ID, &entry.URLID, &entry.IP, &entry.UserAgent, &entry.Referrer, &entry.Timestamp)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, &entry)
	}

	return analytics, nil
}

-- 1. Create a new table with the desired schema
CREATE TABLE analytics_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    url_id INTEGER NOT NULL,
    ip TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    referrer TEXT,
    timestamp DATETIME NOT NULL,
    FOREIGN KEY(url_id) REFERENCES urls(id) ON DELETE CASCADE
);

-- 2. Copy data from the old table to the new table
INSERT INTO analytics_new (id, url_id, ip, user_agent, referrer, timestamp)
SELECT id, short_url, ip, user_agent, referrer, timestamp FROM analytics;

-- 3. Drop the old table
DROP TABLE analytics;

-- 4. Rename the new table to 'analytics'
ALTER TABLE analytics_new RENAME TO analytics;

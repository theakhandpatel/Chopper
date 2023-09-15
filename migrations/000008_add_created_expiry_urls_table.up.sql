CREATE TABLE new_urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url TEXT NOT NULL,
    short_url TEXT NOT NULL,
    accessed INTEGER DEFAULT 0,
    redirect INTEGER DEFAULT 301,
    user_id INTEGER NOT NULL DEFAULT 0,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    UNIQUE(short_url)
);


INSERT INTO new_urls (
    id, long_url, short_url, accessed, redirect, user_id, created, modified
)
SELECT
    id, long_url, short_url, accessed, redirect, user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM
    urls;

DROP TABLE urls;

ALTER TABLE new_urls RENAME TO urls;

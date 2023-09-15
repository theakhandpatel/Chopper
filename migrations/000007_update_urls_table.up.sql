CREATE TABLE urls_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url TEXT NOT NULL,
    short_url TEXT NOT NULL,
    accessed INTEGER DEFAULT 0,
    redirect INTEGER DEFAULT 308,
    user_id INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY(user_id) REFERENCES users(id)
    UNIQUE(short_url)
);

INSERT INTO urls_new (id, long_url, short_url, accessed, redirect, user_id)
SELECT id, long_url, short_url, accessed, redirect, 0 FROM urls;

DROP TABLE urls;

ALTER TABLE urls_new RENAME TO urls;

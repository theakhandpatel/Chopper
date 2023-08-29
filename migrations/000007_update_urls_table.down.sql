CREATE TABLE urls_temp (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url TEXT NOT NULL,
    short_url TEXT NOT NULL,
    accessed INTEGER DEFAULT 0,
    redirect INTEGER DEFAULT 301,
    UNIQUE(short_url)
);

INSERT INTO urls_temp (id, long_url, short_url, accessed, redirect)
SELECT id, long_url, short_url, accessed, redirect FROM urls;

DROP TABLE urls;

ALTER TABLE urls_temp RENAME TO urls;

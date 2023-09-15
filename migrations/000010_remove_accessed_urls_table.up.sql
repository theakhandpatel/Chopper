-- 1. Create a new table with the desired schema
CREATE TABLE urls_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    long_url TEXT NOT NULL,
    short_url TEXT NOT NULL,
    redirect INTEGER DEFAULT 308,
    user_id INTEGER NOT NULL DEFAULT 0,
    created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id),
    UNIQUE(short_url)
);

-- 2. Copy data from the old table to the new table
INSERT INTO urls_new (id, long_url, short_url, redirect, user_id, created, modified)
SELECT id, long_url, short_url, redirect, user_id, created, modified FROM urls;

-- 3. Drop the old table
DROP TABLE urls;

-- 4. Rename the new table to 'urls'
ALTER TABLE urls_new RENAME TO urls;

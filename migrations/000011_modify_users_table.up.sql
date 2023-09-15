-- 1. Create a new table with the desired schema
CREATE TABLE users_new (
    id INTEGER PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash BLOB NOT NULL,
    type INTEGER NOT NULL CHECK (type IN (0, 1, 2)) DEFAULT 1
);

-- 2. Copy data from the old table to the new table
INSERT INTO users_new (id, created_at, name, email, password_hash, type)
SELECT id, created_at, name, email, password_hash, version FROM users;

-- 3. Drop the old table
DROP TABLE users;

-- 4. Rename the new table to 'users'
ALTER TABLE users_new RENAME TO users;

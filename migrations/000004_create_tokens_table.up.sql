CREATE TABLE IF NOT EXISTS tokens (
			hash BLOB PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
			expiry TIMESTAMP NOT NULL,
			scope TEXT NOT NULL
	);
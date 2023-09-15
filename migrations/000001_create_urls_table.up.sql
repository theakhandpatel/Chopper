	CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		long_url TEXT NOT NULL,
		short_url TEXT NOT NULL UNIQUE,
		accessed INTEGER DEFAULT 0,
		redirect INTEGER DEFAULT 301,
		UNIQUE(short_url)
	);


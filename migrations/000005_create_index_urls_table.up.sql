CREATE UNIQUE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);
CREATE INDEX IF NOT EXISTS idx_long_url ON urls(long_url);
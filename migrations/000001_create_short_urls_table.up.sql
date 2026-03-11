CREATE TABLE IF NOT EXISTS short_urls (
    id SERIAL PRIMARY KEY,
    short_id VARCHAR(8) UNIQUE NOT NULL,
    original_url VARCHAR(2048) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_original_url ON short_urls (original_url);
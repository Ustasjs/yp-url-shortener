ALTER TABLE short_urls ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_short_urls_is_deleted ON short_urls (is_deleted);

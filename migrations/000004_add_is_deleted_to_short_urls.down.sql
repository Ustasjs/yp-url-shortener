DROP INDEX IF EXISTS idx_short_urls_is_deleted;

ALTER TABLE short_urls DROP COLUMN IF EXISTS is_deleted;
